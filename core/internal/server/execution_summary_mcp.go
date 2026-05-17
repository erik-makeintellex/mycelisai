package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

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
