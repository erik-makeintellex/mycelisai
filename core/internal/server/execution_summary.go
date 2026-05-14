package server

import (
	"strings"

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
		Kind:           "answer",
		Title:          "Soma answer",
		Summary:        summaryText,
		Retained:       boolPtr(false),
		RetentionClass: protocol.ExecutionRetentionClassNonRetained,
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
			RunClass:     protocol.ExecutionRunClassNoRun,
			NoRunReason:  "Direct Soma answers are audit-only and do not create execution runs.",
			ProofClass:   protocol.ExecutionProofClassAuditOnly,
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
			ID:             proofID,
			Kind:           "proposal",
			Title:          "Guided proposal",
			Summary:        display.OperatorSummary,
			Retained:       boolPtr(true),
			RetentionClass: protocol.ExecutionRetentionClassRetained,
		}},
		Proof: protocol.ExecutionProof{
			RunClass:      protocol.ExecutionRunClassNoRun,
			NoRunReason:   "Guided proposals create intent proof first; no execution run exists until confirmation.",
			ProofClass:    protocol.ExecutionProofClassIntentOnly,
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

func buildConfirmActionExecutionSummary(proofID, runID, auditID string, scope *protocol.ScopeValidation, results []plannedToolExecutionResult) *protocol.ExecutionSummary {
	capabilities := []protocol.CapabilityUse{}
	if scope != nil {
		capabilities = capabilityUseFromPlannedCalls(scope.PlannedToolCalls, scope.Tools, scope.RiskLevel)
	}
	outputs := executionOutputsFromToolResults(results)

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
		Outputs:       outputs,
		Proof: protocol.ExecutionProof{
			RunID:         runID,
			RunClass:      protocol.ExecutionRunClassLinked,
			ProofClass:    protocol.ExecutionProofClassRunAudit,
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

func buildConfirmActionFailureExecutionSummary(proofID, runID, auditID string, err error) *protocol.ExecutionSummary {
	blocker := "The approved execution failed."
	if err != nil {
		blocker = firstNonEmptyString(strings.TrimSpace(err.Error()), blocker)
	}
	runClass := protocol.ExecutionRunClassNoRun
	noRunReason := "The approved execution did not retain a run id."
	if strings.TrimSpace(runID) != "" {
		runClass = protocol.ExecutionRunClassLinked
		noRunReason = ""
	}
	return &protocol.ExecutionSummary{
		Intent: protocol.ExecutionIntent{
			Resolved: "chat-action",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: "Soma accepted the approval, then execution degraded before completion.",
		},
		Execution: protocol.ExecutionState{
			Shape:   protocol.ExecutionShapeGuidedProposal,
			Status:  protocol.ExecutionStatusFailed,
			Summary: "Soma could not complete the approved proposal.",
		},
		Proof: protocol.ExecutionProof{
			RunID:         runID,
			RunClass:      runClass,
			NoRunReason:   noRunReason,
			ProofClass:    protocol.ExecutionProofClassRunAudit,
			AuditEventID:  auditID,
			IntentProofID: proofID,
			Verified:      boolPtr(false),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "confirmed",
			RecoveryState:  "failed",
			Blocker:        blocker,
			Retryable:      boolPtr(true),
			Degradation: &protocol.ExecutionDegradation{
				Code:              "approved_execution_failed",
				WhatFailed:        blocker,
				TrustedState:      "The approval, intent proof, failed run record, and audit event remain trusted.",
				InvalidatedProof:  "No completed execution proof or retained output should be trusted for this attempt.",
				SafeContinuation:  "Review the failed run, adjust the request or runtime dependency, then retry the proposal.",
				RequiresAttention: true,
			},
		},
		NextStep: &protocol.ExecutionNextStep{
			Label:  "Review failed run",
			Action: "view_run",
			Href:   "/api/v1/runs/" + runID,
		},
	}
}

func buildMCPToolCallExecutionSummary(serverName, toolName, summary, exchangeItemID string) *protocol.ExecutionSummary {
	retained := strings.TrimSpace(exchangeItemID) != ""
	capabilityLabel := strings.TrimSpace(toolName)
	if server := strings.TrimSpace(serverName); server != "" && capabilityLabel != "" {
		capabilityLabel = server + ":" + capabilityLabel
	}
	return &protocol.ExecutionSummary{
		Intent: protocol.ExecutionIntent{
			Resolved: "mcp_tool_call",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: firstNonEmptyString(summary, "MCP tool call completed."),
			Assumptions: []string{
				"Direct MCP tool calls are operator-triggered tool invocations.",
				"No execution run is created unless a run id is supplied by an enclosing runtime path.",
			},
		},
		Execution: protocol.ExecutionState{
			Shape:   protocol.ExecutionShapeToolAssistedWork,
			Status:  protocol.ExecutionStatusCompleted,
			Summary: firstNonEmptyString(summary, "MCP tool call completed."),
		},
		CapabilityUse: []protocol.CapabilityUse{{
			ID:     strings.TrimSpace(toolName),
			Label:  capabilityLabel,
			Kind:   protocol.CapabilityUseMCP,
			Reason: "Direct MCP tool call requested through the admin runtime.",
		}},
		Outputs: []protocol.ExecutionOutput{{
			ID:             exchangeItemID,
			Kind:           "mcp_tool_result",
			Title:          firstNonEmptyString(strings.TrimSpace(toolName), "MCP tool result"),
			Summary:        firstNonEmptyString(summary, "MCP tool call completed."),
			Retained:       boolPtr(retained),
			RetentionClass: retentionClassForBool(retained),
		}},
		Proof: protocol.ExecutionProof{
			RunClass:       protocol.ExecutionRunClassNoRun,
			NoRunReason:    "Direct MCP tool calls are audit/exchange proof only unless invoked inside a run.",
			ProofClass:     protocol.ExecutionProofClassAuditOnly,
			ExchangeItemID: exchangeItemID,
			Verified:       boolPtr(retained),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "allow",
			RecoveryState:  "exchange_recorded",
			Retryable:      boolPtr(true),
		},
		NextStep: &protocol.ExecutionNextStep{
			Label: "Review MCP result",
		},
	}
}
