package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func buildTeamLeadGuidanceExecutionSummary(response TeamLeadGuidanceResponse, requestContext string) *protocol.ExecutionSummary {
	contract := response.ExecutionContract
	if contract == nil {
		return nil
	}

	outputs := make([]protocol.ExecutionOutput, 0, len(contract.TargetOutputs))
	for _, target := range contract.TargetOutputs {
		title := strings.TrimSpace(target)
		if title == "" {
			continue
		}
		retained := contract.ExecutionMode == TeamLeadExecutionModeContinuityResume
		outputs = append(outputs, protocol.ExecutionOutput{
			Kind:     "retained_output",
			Title:    title,
			Retained: boolPtr(retained),
		})
	}

	capabilityUse := []protocol.CapabilityUse{{
		ID:     "team_lead_guidance",
		Label:  contract.OwnerLabel,
		Kind:   protocol.CapabilityUseTeam,
		Reason: "Used to shape the Team Lead execution path.",
	}}
	if contract.WorkflowGroup != nil {
		capabilityUse = append(capabilityUse, protocol.CapabilityUse{
			ID:     firstNonEmptyString(contract.WorkflowGroup.GroupID, contract.WorkflowGroup.Name),
			Label:  contract.WorkflowGroup.Name,
			Kind:   protocol.CapabilityUseTeam,
			Reason: "Retained output or group workflow context attached to the guidance.",
		})
	}

	return &protocol.ExecutionSummary{
		Intent: protocol.ExecutionIntent{
			Original: strings.TrimSpace(requestContext),
			Resolved: string(response.Action),
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: response.Summary,
			Assumptions: []string{
				"Guidance is retained as an operator-facing execution contract, not a completed run.",
				"No run id is attached until a concrete execution run exists.",
			},
		},
		Execution: protocol.ExecutionState{
			Shape:   protocol.ExecutionShapeTeamExecution,
			Status:  protocol.ExecutionStatusProposed,
			Summary: contract.Summary,
		},
		CapabilityUse: capabilityUse,
		Outputs:       outputs,
		Proof: protocol.ExecutionProof{
			Verified: boolPtr(false),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "guidance_only",
			RecoveryState:  "ready_for_operator_review",
			Retryable:      boolPtr(true),
		},
		NextStep: &protocol.ExecutionNextStep{
			Label: firstNonEmptyString(firstString(response.PrioritySteps), "Review the Team Lead execution contract."),
		},
	}
}

func buildGroupBroadcastExecutionSummary(group *CollaborationGroup, message, auditEventID string, subjects []string) *protocol.ExecutionSummary {
	if group == nil {
		return nil
	}

	outputs := []protocol.ExecutionOutput{{
		Kind:    "group_broadcast",
		Title:   "Group broadcast accepted",
		Summary: "Broadcast queued for the group collaboration channel and active team command lanes.",
	}}
	for _, teamID := range group.TeamIDs {
		teamID = strings.TrimSpace(teamID)
		if teamID == "" {
			continue
		}
		outputs = append(outputs, protocol.ExecutionOutput{
			ID:    teamID,
			Kind:  "team_signal",
			Title: teamID,
		})
	}

	return &protocol.ExecutionSummary{
		Intent: protocol.ExecutionIntent{
			Original: strings.TrimSpace(message),
			Resolved: "group_broadcast",
		},
		Understanding: protocol.ExecutionUnderstanding{
			Summary: "Accepted a group coordination message for fanout to retained group/team channels.",
		},
		Execution: protocol.ExecutionState{
			Shape:   protocol.ExecutionShapeTeamExecution,
			Status:  protocol.ExecutionStatusRunning,
			Summary: "Broadcast publish completed locally and durable audit proof was recorded.",
		},
		CapabilityUse: []protocol.CapabilityUse{{
			ID:     group.ID,
			Label:  group.Name,
			Kind:   protocol.CapabilityUseTeam,
			Reason: "Group broadcast fanout across " + itoa(len(subjects)) + " subject(s).",
		}},
		Outputs: outputs,
		Proof: protocol.ExecutionProof{
			AuditEventID: auditEventID,
			Verified:     boolPtr(strings.TrimSpace(auditEventID) != ""),
		},
		AuditRecovery: protocol.AuditRecovery{
			ApprovalStatus: "accepted",
			RecoveryState:  "audit_recorded",
			Retryable:      boolPtr(true),
		},
		NextStep: &protocol.ExecutionNextStep{
			Label: "Review group retained outputs or team responses.",
		},
	}
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
