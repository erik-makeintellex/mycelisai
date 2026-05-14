package server

import (
	"strings"

	"github.com/mycelis/core/internal/runs"
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
		name := strings.TrimSpace(call.Name)
		if name == "" {
			continue
		}
		tools = append(tools, name)
		risks[name] = capabilityRiskForTool(name, call.Arguments)
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
		outputs = append(outputs, protocol.ExecutionOutput{
			ID:             artifact.ID,
			Kind:           firstNonEmptyString(artifact.Type, "artifact"),
			Title:          title,
			Href:           firstNonEmptyString(artifact.URL, artifact.SavedPath),
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

func confirmActionResponseData(proofID, runID, auditID string, scope *protocol.ScopeValidation, results []plannedToolExecutionResult) map[string]any {
	return map[string]any{
		"confirmed":         true,
		"verified":          true,
		"execution_state":   "verified",
		"proof_id":          proofID,
		"audit_event_id":    auditID,
		"run_id":            runID,
		"run_status":        runs.StatusCompleted,
		"execution_summary": buildConfirmActionExecutionSummary(proofID, runID, auditID, scope, results),
	}
}

func confirmActionFailureResponseData(proofID, runID, auditID string, err error) map[string]any {
	return map[string]any{
		"confirmed":         false,
		"verified":          false,
		"execution_state":   "failed",
		"proof_id":          proofID,
		"audit_event_id":    auditID,
		"run_id":            runID,
		"run_status":        runs.StatusFailed,
		"execution_summary": buildConfirmActionFailureExecutionSummary(proofID, runID, auditID, err),
	}
}
