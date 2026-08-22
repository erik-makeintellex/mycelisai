package server

import "github.com/mycelis/core/pkg/protocol"

func (s *AdminServer) auditExecutedPlannedTool(planned protocol.PlannedToolCall, resolvedToolName, auditUser string) {
	resource := firstNonEmptyString(planned.Arguments["path"], planned.Arguments["subject"], planned.Arguments["channel"])
	capabilityID := capabilityForPlannedTool(firstNonEmptyString(planned.ToolRef, resolvedToolName))
	details := buildExecutionAuditDetailsForTool(planned, resolvedToolName)
	_, _ = s.createAuditEvent(
		protocol.TemplateChatToProposal, "confirm-action",
		"Approved capability usage executed",
		map[string]any{
			"actor":           "Soma",
			"user":            auditUser,
			"action":          "capability_usage",
			"result_status":   "completed",
			"capability_used": capabilityID,
			"resource":        resource,
			"details":         details,
		},
	)
	if resolvedToolName == "publish_signal" || resolvedToolName == "broadcast" {
		_, _ = s.createAuditEvent(
			protocol.TemplateChatToProposal, "confirm-action",
			"Governed channel write executed",
			map[string]any{
				"actor":           "Soma",
				"user":            auditUser,
				"action":          "channel_written",
				"result_status":   "completed",
				"capability_used": capabilityID,
				"resource":        resource,
			},
		)
	}
	if resolvedToolName == "write_file" || resolvedToolName == "promote_deployment_context" {
		_, _ = s.createAuditEvent(
			protocol.TemplateChatToProposal, "confirm-action",
			"Governed artifact created",
			map[string]any{
				"actor":           "Soma",
				"user":            auditUser,
				"action":          "artifact_created",
				"result_status":   "completed",
				"capability_used": capabilityID,
				"resource":        resource,
			},
		)
	}
}
