package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func buildProposalWorkIntent(planned []protocol.PlannedToolCall, latestRequest string, mutTools []string, display proposalDisplayContract) *protocol.WorkIntent {
	text := strings.ToLower(strings.Join(strings.Fields(latestRequest+" "+strings.Join(mutTools, " ")), " "))
	paths := plannedWriteFilePaths(planned)
	kind := "one_shot"
	cadence := "run_once"
	runtimePosture := "Run once after approval, then return the result to Soma."
	switch {
	case requestContainsAny(text, []string{"schedule", "every ", "daily", "weekly", "monthly"}):
		kind = "scheduled"
		cadence = "scheduled"
		runtimePosture = "Create or update a scheduled handoff after approval."
	case requestContainsAny(text, []string{"watch", "monitor", "keep running", "service", "daemon"}):
		kind = "service"
		cadence = "continuous"
		runtimePosture = "Keep the approved work active until the operator or policy stops it."
	case requestContainsAny(text, []string{"extend soma", "soma itself", "plugin", "capability extension"}):
		kind = "self_extension"
	case requestContainsAny(text, []string{"app", "application", "game", "playable", "package", "project"}):
		kind = "project"
	}
	output := inferProposalOutputContract(latestRequest, paths, display)
	if output.Shape == "app_package" || output.Shape == "mixed_output" {
		kind = "project"
	}
	return &protocol.WorkIntent{
		Kind:            kind,
		Objective:       display.OperatorSummary,
		Cadence:         cadence,
		RuntimePosture:  runtimePosture,
		BusScope:        display.BusScope,
		NATSSubjects:    display.NATSSubjects,
		OutputContract:  output,
		ScheduleSummary: scheduleSummaryForWorkIntent(cadence),
	}
}

func inferProposalOutputContract(latestRequest string, paths []string, display proposalDisplayContract) *protocol.WorkOutputContract {
	text := strings.ToLower(strings.Join(append([]string{latestRequest, display.ExpectedResult}, paths...), " "))
	shape := "document"
	launchHint := ""
	validation := []string{"Retained output is reviewable from Soma, Groups, or Resources."}
	switch {
	case requestContainsAny(text, []string{"game", "playable", "browser app", "index.html", "project-package", "application"}):
		shape = "app_package"
		launchHint = "Return an openable entrypoint and folder access."
		validation = []string{"Open the entrypoint.", "Confirm the primary interaction works.", "Retain proof and repair notes."}
	case requestContainsAny(text, []string{"image", "audio", "video", "media", "music"}):
		shape = "media"
		validation = []string{"Preview the generated media.", "Retain source prompt/context and proof."}
	case requestContainsAny(text, []string{"table", "spreadsheet", "csv", "matrix"}):
		shape = "table"
		validation = []string{"Render as a readable table.", "Keep source/context available for review."}
	case requestContainsAny(text, []string{"code", "script", ".go", ".ts", ".tsx", ".py", ".js"}):
		shape = "code_script"
		validation = []string{"Return file path and verification output when available."}
	case requestContainsAny(text, []string{"dataset", "json", "data export"}):
		shape = "dataset"
		validation = []string{"Retain schema or format notes with the output."}
	}
	return &protocol.WorkOutputContract{
		Shape:              shape,
		PrimaryDeliverable: firstNonEmptyString(firstString(paths), display.ExpectedResult),
		Retention:          "user_deliverable",
		LaunchHint:         launchHint,
		Validation:         validation,
	}
}

func scheduleSummaryForWorkIntent(cadence string) string {
	switch cadence {
	case "scheduled":
		return "Run on the requested cadence after approval."
	case "continuous":
		return "Continue running until stopped or recovered."
	default:
		return ""
	}
}

func proposalExecutionMode(intent *protocol.WorkIntent) string {
	if intent == nil {
		return "confirm_then_execute"
	}
	switch intent.Cadence {
	case "scheduled":
		return "schedule_handoff"
	case "continuous", "event_driven":
		return "team_async"
	default:
		return "confirm_then_execute"
	}
}
