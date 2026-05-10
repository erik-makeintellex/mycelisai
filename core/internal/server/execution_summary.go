package server

import (
	"strings"

	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

func boolPtr(value bool) *bool {
	return &value
}

func buildDirectChatExecutionSummary(originalIntent, replyText, auditEventID string, agentResult chatAgentResult) *protocol.ExecutionSummary {
	summaryText := firstNonEmptyString(replyText, "Soma answered directly.")
	shape := protocol.ExecutionShapeDirectSoma
	executionSummary := "Soma completed a direct response."
	if len(agentResult.ToolsUsed) > 0 || len(agentResult.Artifacts) > 0 {
		shape = protocol.ExecutionShapeToolAssistedWork
		executionSummary = "Soma completed tool-assisted work."
	}
	outputs := []protocol.ExecutionOutput{{
		Kind:    "answer",
		Title:   "Soma answer",
		Summary: summaryText,
	}}
	outputs = append(outputs, executionOutputsFromArtifacts(agentResult.Artifacts)...)

	return &protocol.ExecutionSummary{
		Intent: protocol.ExecutionIntent{
			Original: strings.TrimSpace(originalIntent),
			Resolved: "answer",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: summaryText,
		},
		Execution: protocol.ExecutionState{
			Shape:   shape,
			Status:  protocol.ExecutionStatusCompleted,
			Summary: executionSummary,
		},
		CapabilityUse: capabilityUseFromChat(agentResult.ToolsUsed, agentResult.Consultations, ""),
		Outputs:       outputs,
		Proof: protocol.ExecutionProof{
			AuditEventID: auditEventID,
			Verified:     boolPtr(strings.TrimSpace(auditEventID) != ""),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "allow",
			RecoveryState:  "completed",
			Retryable:      boolPtr(true),
		},
		NextStep: &protocol.ExecutionNextStep{
			Label:  "Ask a follow-up",
			Action: "chat",
		},
	}
}

func buildProposalExecutionSummary(originalIntent string, planned []protocol.PlannedToolCall, mutTools []string, display proposalDisplayContract, proofID, auditEventID string, approval *protocol.ApprovalPolicy) *protocol.ExecutionSummary {
	approvalStatus := approvalStatusValue(approval)
	recoveryState := "awaiting_confirmation"
	if approvalStatus == "approval_required" {
		recoveryState = "approval_required"
	}
	blocker := approvalReasonValue(approval)

	return &protocol.ExecutionSummary{
		Intent: protocol.ExecutionIntent{
			Original: strings.TrimSpace(originalIntent),
			Resolved: "chat-action",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: display.OperatorSummary,
		},
		Execution: protocol.ExecutionState{
			Shape:   protocol.ExecutionShapeGuidedProposal,
			Status:  protocol.ExecutionStatusProposed,
			Summary: display.ExpectedResult,
		},
		CapabilityUse: capabilityUseFromPlannedCalls(planned, mutTools, chatToolRisk(mutTools)),
		Outputs: []protocol.ExecutionOutput{{
			ID:      proofID,
			Kind:    "proposal",
			Title:   "Guided proposal",
			Summary: display.OperatorSummary,
		}},
		Proof: protocol.ExecutionProof{
			AuditEventID:  auditEventID,
			IntentProofID: proofID,
			Verified:      boolPtr(strings.TrimSpace(proofID) != "" && strings.TrimSpace(auditEventID) != ""),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: approvalStatus,
			RecoveryState:  recoveryState,
			Blocker:        blocker,
			Retryable:      boolPtr(true),
		},
		NextStep: &protocol.ExecutionNextStep{
			Label:  "Confirm proposal",
			Action: "confirm_action",
			Href:   "/api/v1/intent/confirm-action",
		},
	}
}

func buildConfirmActionExecutionSummary(proofID, runID, auditID string, scope *protocol.ScopeValidation) *protocol.ExecutionSummary {
	capabilities := []protocol.CapabilityUse{}
	if scope != nil {
		capabilities = capabilityUseFromPlannedCalls(scope.PlannedToolCalls, scope.Tools, scope.RiskLevel)
	}

	return &protocol.ExecutionSummary{
		Intent: protocol.ExecutionIntent{
			Resolved: "chat-action",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: "Confirmed proposal execution completed.",
		},
		Execution: protocol.ExecutionState{
			Shape:   protocol.ExecutionShapeGuidedProposal,
			Status:  protocol.ExecutionStatusCompleted,
			Summary: "Soma executed the confirmed proposal and recorded durable proof.",
		},
		CapabilityUse: capabilities,
		Proof: protocol.ExecutionProof{
			RunID:         runID,
			AuditEventID:  auditID,
			IntentProofID: proofID,
			Verified:      boolPtr(true),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "confirmed",
			RecoveryState:  "verified",
			Retryable:      boolPtr(false),
		},
		NextStep: &protocol.ExecutionNextStep{
			Label:  "Review run",
			Action: "view_run",
			Href:   "/api/v1/runs/" + runID,
		},
	}
}

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
		if tool == "delegate" || tool == "delegate_task" {
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
			ID:       artifact.ID,
			Kind:     firstNonEmptyString(artifact.Type, "artifact"),
			Title:    title,
			Href:     firstNonEmptyString(artifact.URL, artifact.SavedPath),
			Retained: boolPtr(retained),
		})
	}
	return outputs
}

func confirmActionResponseData(proofID, runID, auditID string, scope *protocol.ScopeValidation) map[string]any {
	return map[string]any{
		"confirmed":         true,
		"verified":          true,
		"execution_state":   "verified",
		"proof_id":          proofID,
		"audit_event_id":    auditID,
		"run_id":            runID,
		"run_status":        runs.StatusCompleted,
		"execution_summary": buildConfirmActionExecutionSummary(proofID, runID, auditID, scope),
	}
}
