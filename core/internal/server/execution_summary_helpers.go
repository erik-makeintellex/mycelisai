package server

import (
	"context"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/trust"
	"github.com/mycelis/core/pkg/protocol"
)

func capabilityUseFromChat(tools []string, consultations []protocol.ConsultationEntry, risk string) []protocol.CapabilityUse {
	uses := capabilityUseFromTools(tools, risk)
	seen := map[string]struct{}{}
	for _, item := range uses {
		seen[item.ID] = struct{}{}
	}
	for _, consultation := range consultations {
		id := strings.TrimSpace(consultation.Member)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uses = append(uses, protocol.CapabilityUse{
			ID:     id,
			Label:  id,
			Kind:   protocol.CapabilityUseTeam,
			Reason: firstNonEmptyString(consultation.Summary, "Consulted during response preparation."),
		})
	}
	return uses
}

func capabilityUseFromPlannedCalls(planned []protocol.PlannedToolCall, fallbackTools []string, fallbackRisk string) []protocol.CapabilityUse {
	tools := make([]string, 0, len(planned)+len(fallbackTools))
	risks := map[string]string{}
	for _, call := range planned {
		name := strings.TrimSpace(firstNonEmptyString(call.ToolRef, call.Name))
		if name == "" {
			continue
		}
		tools = append(tools, name)
		risks[name] = capabilityRiskForTool(call.Name, call.Arguments)
	}
	tools = append(tools, fallbackTools...)

	uses := capabilityUseFromTools(tools, fallbackRisk)
	for i := range uses {
		if risk := strings.TrimSpace(risks[uses[i].ID]); risk != "" && risk != "low" {
			uses[i].Risk = risk
		}
	}
	return uses
}

func capabilityUseFromTools(tools []string, fallbackRisk string) []protocol.CapabilityUse {
	deduped := uniqueOrderedTools(tools)
	uses := make([]protocol.CapabilityUse, 0, len(deduped))
	for _, tool := range deduped {
		uses = append(uses, protocol.CapabilityUse{
			ID:     tool,
			Label:  tool,
			Kind:   capabilityUseKindForTool(tool),
			Reason: "Used to satisfy the requested execution path.",
			Risk:   fallbackRisk,
		})
	}
	return uses
}

func capabilityUseKindForTool(tool string) protocol.CapabilityUseKind {
	switch inferAdapterKindFromTool(tool) {
	case "mcp":
		return protocol.CapabilityUseMCP
	case "host", "openapi":
		return protocol.CapabilityUseTool
	default:
		if tool == "create_team" || tool == "delegate" || tool == "delegate_task" {
			return protocol.CapabilityUseTeam
		}
		return protocol.CapabilityUseTool
	}
}

func executionOutputsFromArtifacts(artifacts []protocol.ChatArtifactRef) []protocol.ExecutionOutput {
	outputs := make([]protocol.ExecutionOutput, 0, len(artifacts))
	for _, artifact := range artifacts {
		title := strings.TrimSpace(artifact.Title)
		if title == "" {
			title = "Artifact"
		}
		retained := artifact.ID != "" || artifact.Cached || artifact.SavedPath != ""
		href := firstNonEmptyString(artifact.URL, artifact.SavedPath)
		if artifact.Type == "project_package" {
			retained = true
			href = firstNonEmptyString(href, workspaceFileOutputHref(artifact.Entrypoint))
			if artifact.Folder == "" {
				artifact.Folder = parentWorkspacePath(artifact.Entrypoint)
			}
		}
		outputs = append(outputs, protocol.ExecutionOutput{
			ID:             artifact.ID,
			Kind:           firstNonEmptyString(artifact.Type, "artifact"),
			Title:          title,
			Href:           href,
			Entrypoint:     artifact.Entrypoint,
			Folder:         artifact.Folder,
			Files:          artifact.Files,
			Validation:     artifact.Validation,
			Retained:       boolPtr(retained),
			RetentionClass: retentionClassForBool(retained),
		})
	}
	return outputs
}

func retentionClassForBool(retained bool) protocol.ExecutionRetentionClass {
	if retained {
		return protocol.ExecutionRetentionClassRetained
	}
	return protocol.ExecutionRetentionClassNonRetained
}

func confirmActionResponseData(proofID, contractID, proofArtifactID, runID, auditID string, scope *protocol.ScopeValidation, results []plannedToolExecutionResult) map[string]any {
	return map[string]any{
		"confirmed":         true,
		"verified":          true,
		"execution_state":   "verified",
		"proof_id":          proofID,
		"intent_proof_id":   proofID,
		"contract_id":       contractID,
		"proof_artifact_id": proofArtifactID,
		"audit_event_id":    auditID,
		"run_id":            runID,
		"run_status":        runs.StatusCompleted,
		"execution_summary": buildConfirmActionExecutionSummary(proofID, contractID, proofArtifactID, runID, auditID, scope, results),
	}
}

func confirmActionFailureResponseData(proofID, contractID, proofArtifactID, runID, auditID string, err error) map[string]any {
	return map[string]any{
		"confirmed":         false,
		"verified":          false,
		"execution_state":   "failed",
		"proof_id":          proofID,
		"intent_proof_id":   proofID,
		"contract_id":       contractID,
		"proof_artifact_id": proofArtifactID,
		"audit_event_id":    auditID,
		"run_id":            runID,
		"run_status":        runs.StatusFailed,
		"execution_summary": buildConfirmActionFailureExecutionSummary(proofID, contractID, proofArtifactID, runID, auditID, err),
	}
}

func (s *AdminServer) persistConfirmActionSuccessProof(ctx context.Context, proofID, contractID, runID, auditID string, scope *protocol.ScopeValidation, results []plannedToolExecutionResult) string {
	artifactID := uuid.NewString()
	summary := buildConfirmActionExecutionSummary(proofID, contractID, artifactID, runID, auditID, scope, results)
	return s.recordConfirmActionProofArtifact(ctx, artifactID, proofID, contractID, runID, auditID, protocol.ProofArtifactStatusSuccess, summary, summary.Outputs, nil)
}

func (s *AdminServer) persistConfirmActionFailureProof(ctx context.Context, proofID, contractID, runID, auditID string, err error) string {
	artifactID := uuid.NewString()
	summary := buildConfirmActionFailureExecutionSummary(proofID, contractID, artifactID, runID, auditID, err)
	var degradation any
	if summary.AuditRecovery.Degradation != nil {
		degradation = summary.AuditRecovery.Degradation
	}
	return s.recordConfirmActionProofArtifact(ctx, artifactID, proofID, contractID, runID, auditID, protocol.ProofArtifactStatusFailure, summary, summary.Outputs, degradation)
}

func (s *AdminServer) recordConfirmActionProofArtifact(ctx context.Context, artifactID, proofID, contractID, runID, auditID string, status protocol.ProofArtifactStatus, summary *protocol.ExecutionSummary, outputs []protocol.ExecutionOutput, degradation any) string {
	db := s.getDB()
	if db == nil {
		return ""
	}
	evidenceStrength := protocol.TrustEvidenceStrengthRunAudit
	proofQuality := protocol.TrustProofQualityVerified
	if status == protocol.ProofArtifactStatusFailure || status == protocol.ProofArtifactStatusDegraded {
		evidenceStrength = protocol.TrustEvidenceStrengthDegraded
		proofQuality = protocol.TrustProofQualityFailed
	}
	recordedID, err := trust.NewStore(db).RecordProofArtifact(ctx, trust.ProofArtifactInput{
		ID:               artifactID,
		ContractID:       contractID,
		IntentProofID:    proofID,
		RunID:            runID,
		AuditEventID:     auditID,
		Status:           status,
		ProofClass:       protocol.ExecutionProofClassRunAudit,
		ValidationSource: protocol.TrustValidationSourceConfirmAction,
		EvidenceStrength: evidenceStrength,
		ProofQuality:     proofQuality,
		OutputRefs:       outputs,
		AuditRefs:        []map[string]string{{"audit_event_id": auditID, "source": "log_entries"}},
		ReviewLineage: []map[string]string{{
			"event":  "confirm_action_" + string(status),
			"source": string(protocol.TrustValidationSourceConfirmAction),
		}},
		Degradation: degradation,
		Recovery:    summary.AuditRecovery,
		Payload:     summary,
	})
	if err != nil {
		log.Printf("CE-1: confirm-action proof artifact persistence failed: %v", err)
		return ""
	}
	return recordedID
}
