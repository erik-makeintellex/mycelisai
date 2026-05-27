package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

type recoveryAction struct {
	ID              string
	Label           string
	ApprovalPosture protocol.ApprovalPosture
	CapabilityID    string
	RetryTarget     string
	ExpectedProof   []string
	TrustedState    string
	TargetState     protocol.TeamWorkState
}

func recoveryActionForTeamWorkItem(item protocol.TeamWorkItem) recoveryAction {
	capabilityID := firstRecoveryString(item.CapabilityRequirements)
	retryTarget := firstNonEmptyString(item.RunID, item.WorkItemID, item.TeamID)
	expectedProof := item.ExpectedProof
	if len(expectedProof) == 0 {
		expectedProof = []string{"Recovered team status event", "Retained output or degraded retry proof"}
	}

	action := recoveryAction{
		ID:              "recover_team_work",
		Label:           "Recover team work",
		ApprovalPosture: item.GovernancePosture,
		CapabilityID:    capabilityID,
		RetryTarget:     retryTarget,
		ExpectedProof:   compactRecoveryStrings(expectedProof),
		TrustedState:    recoveryTrustedState(item),
		TargetState:     protocol.TeamWorkStateQueued,
	}
	if action.ApprovalPosture == "" {
		action.ApprovalPosture = protocol.ApprovalPostureAutoAllowed
	}
	if strings.Contains(strings.ToLower(item.DegradationState), "media") || capabilityID == "media_generation" || capabilityID == "media_output" {
		action.ID = "retry_media_capability"
		action.Label = "Retry media capability"
		action.CapabilityID = firstNonEmptyString(capabilityID, "media_generation")
		action.ExpectedProof = compactRecoveryStrings(append(action.ExpectedProof, "Media provider availability proof"))
	}
	return action
}

func recoveryOptionFromAction(action recoveryAction) string {
	parts := []string{"Retry"}
	if action.CapabilityID != "" {
		parts = append(parts, action.CapabilityID)
	} else {
		parts = append(parts, "the capability")
	}
	if action.RetryTarget != "" {
		parts = append(parts, "for", action.RetryTarget)
	}
	if action.ApprovalPosture == protocol.ApprovalPostureRequired {
		parts = append(parts, "after operator approval")
	}
	if len(action.ExpectedProof) > 0 {
		parts = append(parts, "and retain proof:", strings.Join(action.ExpectedProof, ", "))
	}
	if action.TargetState != "" {
		parts = append(parts, "then return to", string(action.TargetState))
	}
	return strings.Join(parts, " ") + "."
}

func recoveryTrustedState(item protocol.TeamWorkItem) string {
	if item.DegradationState == "" {
		return "Current work item state and retained proof refs remain trusted."
	}
	return "Current work item state remains trusted; " + item.DegradationState + " invalidates only the failed execution attempt."
}

func compactRecoveryStrings(values []string) []string {
	seen := map[string]struct{}{}
	compacted := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		compacted = append(compacted, trimmed)
	}
	return compacted
}

func firstRecoveryString(values []string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
