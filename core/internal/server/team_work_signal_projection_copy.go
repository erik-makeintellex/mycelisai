package server

import "github.com/mycelis/core/pkg/protocol"

func projectedHeadlineForItem(item protocol.TeamWorkItem, kind protocol.SignalPayloadKind, payload map[string]any) string {
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
	return projectedHeadline(kind, payload)
}

func projectedDetailsForItem(item protocol.TeamWorkItem, payload map[string]any) string {
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
	return projectedDetails(payload)
}

func projectedNextActionForItem(item protocol.TeamWorkItem, payload map[string]any) string {
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
	return stringField(payload, "next_action")
}
