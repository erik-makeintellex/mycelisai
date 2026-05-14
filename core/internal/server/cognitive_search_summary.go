package server

import (
	"strings"

	"github.com/mycelis/core/internal/searchcap"
	"github.com/mycelis/core/pkg/protocol"
)

func buildSearchExecutionSummary(originalIntent, replyText, auditEventID string, tools []string, status protocol.ExecutionStatus, blocker string, degradation *protocol.ExecutionDegradation) *protocol.ExecutionSummary {
	shape := protocol.ExecutionShapeDirectSoma
	executionSummary := "Soma completed a direct response."
	if len(tools) > 0 {
		shape = protocol.ExecutionShapeToolAssistedWork
		executionSummary = "Soma completed tool-assisted work."
	}
	recoveryState := "completed"
	if status == protocol.ExecutionStatusBlocked {
		recoveryState = "blocked"
		executionSummary = firstNonEmptyString(blocker, "Soma could not complete tool-assisted work.")
	}
	return &protocol.ExecutionSummary{
		Intent: protocol.ExecutionIntent{
			Original: strings.TrimSpace(originalIntent),
			Resolved: "answer",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: firstNonEmptyString(replyText, executionSummary),
		},
		Execution: protocol.ExecutionState{
			Shape:   shape,
			Status:  status,
			Summary: executionSummary,
		},
		CapabilityUse: capabilityUseFromTools(tools, ""),
		Outputs: []protocol.ExecutionOutput{{
			Kind:           "answer",
			Title:          "Soma answer",
			Summary:        firstNonEmptyString(replyText, executionSummary),
			Retained:       boolPtr(false),
			RetentionClass: protocol.ExecutionRetentionClassNonRetained,
		}},
		Proof: protocol.ExecutionProof{
			RunClass:     protocol.ExecutionRunClassNoRun,
			NoRunReason:  "Direct search answers are audit-only and do not create execution runs.",
			ProofClass:   protocol.ExecutionProofClassAuditOnly,
			AuditEventID: auditEventID,
			Verified:     boolPtr(strings.TrimSpace(auditEventID) != ""),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "allow",
			RecoveryState:  recoveryState,
			Blocker:        strings.TrimSpace(blocker),
			Retryable:      boolPtr(true),
			Degradation:    degradation,
		},
		NextStep: &protocol.ExecutionNextStep{
			Label:  "Ask a follow-up",
			Action: "chat",
		},
	}
}

func searchBlockerDegradation(blocker *searchcap.Blocker) *protocol.ExecutionDegradation {
	if blocker == nil {
		return nil
	}
	return searchDegradation(
		firstNonEmptyString(strings.TrimSpace(blocker.Code), "search_blocked"),
		blocker.Message,
		firstNonEmptyString(strings.TrimSpace(blocker.NextAction), "Retry after search capability configuration is corrected."),
	)
}

func searchDegradation(code, whatFailed, safeContinuation string) *protocol.ExecutionDegradation {
	return &protocol.ExecutionDegradation{
		Code:              strings.TrimSpace(code),
		WhatFailed:        firstNonEmptyString(strings.TrimSpace(whatFailed), "Search did not produce trusted results."),
		TrustedState:      "The chat response and audit record remain trusted; no result set was accepted as execution proof.",
		InvalidatedProof:  "No public-web result proof is available for this search attempt.",
		SafeContinuation:  firstNonEmptyString(strings.TrimSpace(safeContinuation), "Retry after the search provider is configured and reachable."),
		RequiresAttention: true,
	}
}
