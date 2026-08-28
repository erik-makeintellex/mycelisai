package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func projectedHeadlineForItem(item protocol.TeamWorkItem, kind protocol.SignalPayloadKind, payload map[string]any) string {
	if item.State == protocol.TeamWorkStateReviewing {
		return "Checking deliverable"
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "missing_retained_output" {
		return "Team result missing retained output"
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "invalid_deliverable_shape" {
		return "Team result is not a usable deliverable"
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "incomplete_deliverable_files" {
		return "Team deliverable is incomplete"
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "unverified_primary_interaction" {
		return "Team deliverable needs interaction proof"
	}
	if item.State == protocol.TeamWorkStateDegraded && strings.HasPrefix(item.DegradationState, "runtime_validation_") {
		return normalizedGeneratedPackageFailure(item.DegradationState).Headline
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "result_contract_unsatisfied" {
		return normalizedGeneratedPackageFailure(item.DegradationState).Headline
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "validation_plan_incomplete" {
		return "Team deliverable cannot be checked safely"
	}
	return projectedHeadline(kind, payload)
}

func projectedDetailsForItem(item protocol.TeamWorkItem, payload map[string]any) string {
	if item.State == protocol.TeamWorkStateReviewing {
		return "The team returned retained files. Soma is checking the approved primary workflow before calling the work complete."
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "missing_retained_output" {
		return "The team reported completion, but did not include retained output references for the expected deliverable."
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "invalid_deliverable_shape" {
		return "The team returned output, but the expected package entrypoint was not retained inside the team's workspace."
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "incomplete_deliverable_files" {
		return "The package entrypoint or one of its required local files is missing from the team's retained workspace."
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "unverified_primary_interaction" {
		return "The package does not expose enough retained evidence that its documented primary control can be used."
	}
	if item.State == protocol.TeamWorkStateDegraded && strings.HasPrefix(item.DegradationState, "runtime_validation_") {
		return normalizedGeneratedPackageFailure(item.DegradationState).Summary
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "result_contract_unsatisfied" {
		return normalizedGeneratedPackageFailure(item.DegradationState).Summary
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "validation_plan_incomplete" {
		return "The approved validation contract or retained launch target is incomplete."
	}
	return projectedDetails(payload)
}

func projectedNextActionForItem(item protocol.TeamWorkItem, payload map[string]any) string {
	if item.State == protocol.TeamWorkStateReviewing {
		return "Keep working with Soma while the deliverable check runs in the background."
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "missing_retained_output" {
		return "Ask Soma to have the team attach or regenerate the retained deliverable."
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "invalid_deliverable_shape" {
		return "Ask Soma to have the same team package, validate, and return the deliverable with a direct entrypoint."
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "incomplete_deliverable_files" {
		return "Ask Soma to have the same team restore the missing package files, validate the primary interaction, and return the repaired deliverable."
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "unverified_primary_interaction" {
		return "Ask Soma to have the same team expose a clear primary control, verify that it changes the application, and return the repaired package."
	}
	if item.State == protocol.TeamWorkStateDegraded && (strings.HasPrefix(item.DegradationState, "runtime_validation_") || item.DegradationState == "validation_plan_incomplete") {
		return normalizedGeneratedPackageFailure(item.DegradationState).Recovery
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "result_contract_unsatisfied" {
		return normalizedGeneratedPackageFailure(item.DegradationState).Recovery
	}
	return stringField(payload, "next_action")
}

func projectedRecoveryOptionsForItem(item protocol.TeamWorkItem, payload map[string]any) []string {
	if item.State != protocol.TeamWorkStateDegraded {
		return nil
	}
	switch item.DegradationState {
	case "missing_retained_output", "invalid_deliverable_shape", "incomplete_deliverable_files", "unverified_primary_interaction", "validation_plan_incomplete", "runtime_validation_failed", "runtime_validation_unavailable", "runtime_validation_deadline", "runtime_validation_stale", "result_contract_unsatisfied":
		nextAction := strings.TrimSpace(projectedNextActionForItem(item, payload))
		if nextAction != "" {
			return []string{nextAction}
		}
	}
	if options := projectedRecoveryOptionsFromPayload(payload); len(options) > 0 {
		return options
	}
	return normalizeStringSlice(item.RecoveryOptions)
}

func projectedRecoveryOptionsFromPayload(payload map[string]any) []string {
	if payload == nil {
		return nil
	}
	options := []string{}
	switch raw := payload["recovery_options"].(type) {
	case []any:
		for _, value := range raw {
			if text, ok := value.(string); ok {
				options = append(options, text)
			}
		}
	case []string:
		options = append(options, raw...)
	case string:
		options = append(options, raw)
	}
	return normalizeStringSlice(options)
}
