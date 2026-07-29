package server

import "github.com/mycelis/core/pkg/protocol"

func projectedHeadlineForItem(item protocol.TeamWorkItem, kind protocol.SignalPayloadKind, payload map[string]any) string {
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "missing_retained_output" {
		return "Team result missing retained output"
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "invalid_deliverable_shape" {
		return "Team result is not a usable deliverable"
	}
	return projectedHeadline(kind, payload)
}

func projectedDetailsForItem(item protocol.TeamWorkItem, payload map[string]any) string {
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "missing_retained_output" {
		return "The team reported completion, but did not include retained output references for the expected deliverable."
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "invalid_deliverable_shape" {
		return "The team returned output, but the expected package entrypoint was not retained inside the team's workspace."
	}
	return projectedDetails(payload)
}

func projectedNextActionForItem(item protocol.TeamWorkItem, payload map[string]any) string {
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "missing_retained_output" {
		return "Ask Soma to have the team attach or regenerate the retained deliverable."
	}
	if item.State == protocol.TeamWorkStateDegraded && item.DegradationState == "invalid_deliverable_shape" {
		return "Ask Soma to have the same team package, validate, and return the deliverable with a direct entrypoint."
	}
	return stringField(payload, "next_action")
}
