package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/trust"
	"github.com/mycelis/core/pkg/protocol"
)

func (p *teamWorkSignalProjection) recordAsyncCompletionProof(
	ctx context.Context,
	exec trust.SQLExecutor,
	item protocol.TeamWorkItem,
	payloadKind protocol.SignalPayloadKind,
	outputRefs []protocol.TeamOutputRef,
	finalResult bool,
) (string, error) {
	if payloadKind != protocol.PayloadKindResult || item.State != protocol.TeamWorkStateOutputReady || len(outputRefs) == 0 {
		return "", nil
	}
	if strings.TrimSpace(item.ContractID) == "" && strings.TrimSpace(item.IntentProofID) == "" {
		return "", nil
	}
	outputs := executionOutputsFromTeamOutputRefs(item, outputRefs)
	if len(outputs) == 0 {
		return "", nil
	}
	proofID := uuid.NewString()
	for i := range outputs {
		outputs[i].ProofArtifactID = proofID
		if outputs[i].Proof == nil {
			outputs[i].Proof = &protocol.OutputProofEnvelope{}
		}
		outputs[i].Proof.ProofID = proofID
	}
	recordedID, err := trust.RecordProofArtifact(ctx, exec, trust.ProofArtifactInput{
		ID:               proofID,
		ArtifactKind:     "team_signal_result",
		ContractID:       item.ContractID,
		IntentProofID:    item.IntentProofID,
		RunID:            item.RunID,
		Status:           protocol.ProofArtifactStatusSuccess,
		ProofClass:       protocol.ExecutionProofClassRunAudit,
		ValidationSource: protocol.TrustValidationSourceRetainedOutput,
		EvidenceStrength: protocol.TrustEvidenceStrengthRetainedOutput,
		ProofQuality:     protocol.TrustProofQualityVerified,
		OutputRefs:       outputs,
		AuditRefs:        auditRefsForAsyncCompletion(item, outputRefs),
		ReviewLineage: []map[string]string{{
			"event":        "team_signal_result",
			"source":       string(protocol.TrustValidationSourceRetainedOutput),
			"team_id":      item.TeamID,
			"work_item_id": item.WorkItemID,
		}},
		Payload: map[string]any{
			"team_id":            item.TeamID,
			"work_item_id":       item.WorkItemID,
			"run_id":             item.RunID,
			"contract_id":        item.ContractID,
			"intent_proof_id":    item.IntentProofID,
			"expected_outputs":   item.ExpectedOutputs,
			"expected_proof":     item.ExpectedProof,
			"output_refs":        outputRefs,
			"validation_scope":   "retained files, package containment, referenced local assets, and static interaction contract",
			"runtime_validation": "not_performed_by_completion_projection",
		},
		Intermediate: !finalResult,
	})
	if err != nil {
		return "", fmt.Errorf("record async team completion proof: %w", err)
	}
	return recordedID, nil
}

func executionOutputsFromTeamOutputRefs(item protocol.TeamWorkItem, refs []protocol.TeamOutputRef) []protocol.ExecutionOutput {
	outputs := make([]protocol.ExecutionOutput, 0, len(refs))
	retained := true
	for _, ref := range refs {
		storageRef := strings.TrimSpace(ref.StorageRef)
		entrypoint := strings.TrimSpace(ref.Entrypoint)
		if storageRef == "" && entrypoint == "" && strings.TrimSpace(ref.OutputID) == "" {
			continue
		}
		output := protocol.ExecutionOutput{
			ID:             firstNonEmptyString(ref.OutputID, ref.Label),
			Kind:           firstNonEmptyString(ref.Kind, "output"),
			OutputClass:    outputClassForTeamRef(ref),
			Title:          firstNonEmptyString(ref.Label, ref.OutputID, "Team output"),
			Href:           workspaceFileOutputHref(firstNonEmptyString(storageRef, entrypoint)),
			OpenURL:        workspaceFileOutputHref(firstNonEmptyString(storageRef, entrypoint)),
			Entrypoint:     entrypoint,
			Folder:         storageRef,
			Validation:     ref.ValidationRef,
			Retained:       &retained,
			RetentionClass: protocol.ExecutionRetentionClassRetained,
			Proof: &protocol.OutputProofEnvelope{
				OutputRefID:      ref.OutputID,
				StorageRef:       storageRef,
				SourceRunID:      firstNonEmptyString(ref.RunID, item.RunID),
				SourceContractID: firstNonEmptyString(ref.ContractID, item.ContractID),
				ExecutionStatus:  string(protocol.ExecutionStatusCompleted),
				ReadbackStatus:   "retained_ref",
			},
		}
		outputs = append(outputs, output)
	}
	return outputs
}

func auditRefsForAsyncCompletion(item protocol.TeamWorkItem, refs []protocol.TeamOutputRef) []map[string]string {
	values := []map[string]string{}
	for _, ref := range refs {
		for _, auditRef := range ref.AuditRefs {
			if strings.TrimSpace(auditRef) != "" {
				values = append(values, map[string]string{"audit_ref": auditRef, "source": "team_output_ref"})
			}
		}
	}
	if len(values) == 0 && strings.TrimSpace(item.WorkItemID) != "" {
		values = append(values, map[string]string{"work_item_id": item.WorkItemID, "source": "team_work_item"})
	}
	return values
}
