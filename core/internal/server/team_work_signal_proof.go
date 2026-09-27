package server

import (
	"context"
	"fmt"
	"sort"
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
	readback *teamOutputReadback,
) (string, error) {
	return p.recordCompletionProof(ctx, exec, item, payloadKind, outputRefs, finalResult, nil, readback)
}

// teamOutputReadback is Core's own readback of the refs a team result claims.
type teamOutputReadback struct {
	ByRef  map[string]outputReadback
	Failed *outputReadback
}

func (r *teamOutputReadback) failureCode() string {
	if r == nil || r.Failed == nil {
		return ""
	}
	if r.Failed.Status == outputReadbackMismatch {
		return "output_digest_mismatch"
	}
	return "output_" + r.Failed.Status
}

func readbackTeamOutputRefs(item protocol.TeamWorkItem, refs []protocol.TeamOutputRef, claimed map[string]string) *teamOutputReadback {
	readback := &teamOutputReadback{ByRef: map[string]outputReadback{}}
	echoes := []string{item.Objective}
	if item.WorkIntent != nil {
		echoes = append(echoes, item.WorkIntent.Objective)
	}
	echoes = normalizeStringSlice(echoes)
	for _, ref := range refs {
		key := teamOutputRefKey(ref)
		storage, entrypoint := strings.TrimSpace(ref.StorageRef), strings.TrimSpace(ref.Entrypoint)
		var result outputReadback
		switch {
		case storage != "" && teamOutputRefIsWorkspacePath(storage):
			result = readbackWorkspaceOutput(storage, claimed[key], echoes)
			if result.verified() && result.Folder && entrypoint != "" {
				if entry := readbackWorkspaceOutput(teamWorkEntrypointPath(ref), "", echoes); !entry.verified() {
					result = entry
				}
			}
		case storage == "" && entrypoint != "" && teamOutputRefIsWorkspacePath(entrypoint):
			result = readbackWorkspaceOutput(entrypoint, claimed[key], echoes)
		default:
			result = outputReadback{Path: firstNonEmptyString(storage, entrypoint, ref.OutputID), Status: outputReadbackUnresolvable,
				Detail: "the output ref names no workspace path Core can read back"}
		}
		readback.ByRef[key] = result
		if readback.Failed == nil && !result.verified() && result.Status != outputReadbackUnresolvable {
			failed := result
			readback.Failed = &failed
		}
	}
	return readback
}

func teamReadbackStatuses(readback *teamOutputReadback) []map[string]any {
	statuses := make([]map[string]any, 0, len(readback.ByRef))
	for _, result := range readback.ByRef {
		statuses = append(statuses, map[string]any{"path": result.Path, "status": result.Status, "sha256": result.Checksum, "bytes": result.Bytes})
	}
	sort.Slice(statuses, func(i, j int) bool { return fmt.Sprint(statuses[i]["path"]) < fmt.Sprint(statuses[j]["path"]) })
	return statuses
}

func teamOutputRefIsWorkspacePath(value string) bool {
	return !strings.Contains(value, "://") && !strings.HasPrefix(value, "/api/")
}

func applyTeamOutputReadback(outputs []protocol.ExecutionOutput, refs []protocol.TeamOutputRef, readback *teamOutputReadback) {
	if readback == nil {
		return
	}
	byOutput := map[string]outputReadback{}
	for _, ref := range refs {
		if result, ok := readback.ByRef[teamOutputRefKey(ref)]; ok {
			byOutput[firstNonEmptyString(ref.OutputID, ref.Label)] = result
		}
	}
	for i := range outputs {
		result, ok := byOutput[outputs[i].ID]
		if !ok || outputs[i].Proof == nil {
			continue
		}
		proof := outputs[i].Proof
		proof.PathBoundaryStatus = firstNonEmptyString(result.PathBoundary, outputReadbackNotApplicable)
		proof.ReadbackStatus = result.Status
		if result.Checksum != "" {
			proof.Checksum, proof.ChecksumAlgorithm, proof.Bytes = result.Checksum, "sha256", result.Bytes
		}
		if !result.verified() {
			proof.RecoveryHint = result.Detail
		}
	}
}

type completionValidationEvidence struct {
	ValidationRef string
	ContentDigest string
	EvidenceRefs  []string
}

func (p *teamWorkSignalProjection) recordRuntimeCompletionProof(
	ctx context.Context,
	exec trust.SQLExecutor,
	item protocol.TeamWorkItem,
	outputRefs []protocol.TeamOutputRef,
	finalResult bool,
	validation completionValidationEvidence,
) (string, error) {
	return p.recordCompletionProof(ctx, exec, item, protocol.PayloadKindResult, outputRefs, finalResult, &validation, nil)
}

func (p *teamWorkSignalProjection) recordCompletionProof(
	ctx context.Context,
	exec trust.SQLExecutor,
	item protocol.TeamWorkItem,
	payloadKind protocol.SignalPayloadKind,
	outputRefs []protocol.TeamOutputRef,
	finalResult bool,
	validation *completionValidationEvidence,
	readback *teamOutputReadback,
) (string, error) {
	readbackFailed := readback.failureCode() != ""
	readyState := item.State == protocol.TeamWorkStateOutputReady || (readbackFailed && item.State == protocol.TeamWorkStateDegraded)
	if payloadKind != protocol.PayloadKindResult || !readyState || len(outputRefs) == 0 {
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
	applyTeamOutputReadback(outputs, outputRefs, readback)
	status := protocol.ProofArtifactStatusSuccess
	evidenceStrength := protocol.TrustEvidenceStrengthRetainedOutput
	// Without a runtime validation plan Core has only read the refs back, so
	// the proof stays unverified; only runtime validation earns verified.
	proofQuality := protocol.TrustProofQualityUnverified
	var degradation any
	validationSource := protocol.TrustValidationSourceRetainedOutput
	validationScope := "Core readback of each claimed output ref: exists inside the workspace, readable, non-empty, not a request echo, and matching any declared digest; no runtime validation plan"
	runtimeValidation := "not_planned"
	reviewEvent := "team_signal_result"
	if validation != nil {
		validationSource = protocol.TrustValidationSourceRuntimeOutput
		validationScope = "digest-bound browser load, page errors, local assets, and approved primary interaction"
		runtimeValidation = "passed"
		reviewEvent = "runtime_output_validation"
		proofQuality = protocol.TrustProofQualityVerified
	}
	if readbackFailed {
		status, evidenceStrength, proofQuality = protocol.ProofArtifactStatusDegraded, protocol.TrustEvidenceStrengthDegraded, protocol.TrustProofQualityFailed
		degradation = map[string]any{
			"code": readback.failureCode(), "what_failed": readback.Failed.Detail, "path": readback.Failed.Path,
			"readback_status": readback.Failed.Status, "requires_attention": true,
		}
	}
	payload := map[string]any{
		"team_id": item.TeamID, "work_item_id": item.WorkItemID, "run_id": item.RunID,
		"contract_id": item.ContractID, "intent_proof_id": item.IntentProofID,
		"expected_outputs": item.ExpectedOutputs, "expected_proof": item.ExpectedProof,
		"output_refs": outputRefs, "validation_scope": validationScope,
		"runtime_validation": runtimeValidation,
	}
	if readback != nil {
		payload["readback_statuses"] = teamReadbackStatuses(readback)
	}
	if validation != nil {
		payload["validation_ref"] = validation.ValidationRef
		payload["content_digest"] = validation.ContentDigest
		payload["evidence_refs"] = validation.EvidenceRefs
	}
	recordedID, err := trust.RecordProofArtifact(ctx, exec, trust.ProofArtifactInput{
		ID:               proofID,
		ArtifactKind:     "team_signal_result",
		ContractID:       item.ContractID,
		IntentProofID:    item.IntentProofID,
		RunID:            item.RunID,
		Status:           status,
		ProofClass:       protocol.ExecutionProofClassRunAudit,
		ValidationSource: validationSource,
		EvidenceStrength: evidenceStrength,
		ProofQuality:     proofQuality,
		OutputRefs:       outputs,
		Degradation:      degradation,
		AuditRefs:        auditRefsForAsyncCompletion(item, outputRefs),
		ReviewLineage: []map[string]string{{
			"event":        reviewEvent,
			"source":       string(validationSource),
			"team_id":      item.TeamID,
			"work_item_id": item.WorkItemID,
		}},
		Payload:      payload,
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
