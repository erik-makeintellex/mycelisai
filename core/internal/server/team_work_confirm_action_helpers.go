package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func baseConfirmedActionWorkItem(link confirmedActionTeamWorkLink, teamID, objective string) protocol.TeamWorkItem {
	return protocol.NormalizeTeamWorkItem(protocol.TeamWorkItem{
		WorkItemID:        uuid.NewString(),
		TeamID:            teamID,
		RunID:             link.RunID,
		IntentProofID:     link.ProofID,
		ContractID:        link.ContractID,
		ProofID:           link.ProofArtifactID,
		Objective:         firstNonEmptyString(objective, "Confirmed team work"),
		Owner:             "Soma",
		GovernancePosture: approvalPostureFromScope(link.Scope),
		ProofRefs:         compactProofRefs(link.ProofID, link.ProofArtifactID),
		AuditRefs:         compactProofRefs(link.AuditID),
	})
}

func confirmedActionStatusEvent(link confirmedActionTeamWorkLink, item protocol.TeamWorkItem, state protocol.TeamWorkState, headline, details, confidence, next string) *protocol.TeamStatusEvent {
	return &protocol.TeamStatusEvent{
		EventID:           uuid.NewString(),
		TeamID:            item.TeamID,
		WorkItemID:        item.WorkItemID,
		RunID:             link.RunID,
		IntentProofID:     link.ProofID,
		ContractID:        link.ContractID,
		ProofID:           link.ProofArtifactID,
		State:             state,
		Headline:          headline,
		Details:           details,
		ConfidencePosture: confidence,
		NextAction:        next,
		SourceKind:        string(protocol.SourceKindWebAPI),
		SourceChannel:     "api.intent.confirm-action",
		PayloadKind:       string(protocol.PayloadKindStatus),
		AuditRefs:         compactProofRefs(link.AuditID),
		Version:           "v1",
	}
}

func confirmedActionInteraction(link confirmedActionTeamWorkLink, item protocol.TeamWorkItem, verb, summary, toolName string, args map[string]any) protocol.TeamInteraction {
	return protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(),
		TeamID:        item.TeamID,
		WorkItemID:    item.WorkItemID,
		RunID:         link.RunID,
		IntentProofID: link.ProofID,
		ContractID:    link.ContractID,
		ProofID:       link.ProofArtifactID,
		SourceKind:    string(protocol.SourceKindWebAPI),
		SourceChannel: "api.intent.confirm-action",
		ActorRef:      firstNonEmptyString(link.AuditUser, "Soma"),
		Verb:          verb,
		Summary:       summary,
		PayloadKind:   string(protocol.PayloadKindCommand),
		Payload: map[string]any{
			"tool":      strings.TrimSpace(toolName),
			"arguments": args,
		},
		ApprovalRef: link.ProofID,
		AuditRefs:   compactProofRefs(link.AuditID),
	})
}

func executionOutputsForResult(result plannedToolExecutionResult) []protocol.ExecutionOutput {
	return executionOutputsFromToolResults([]plannedToolExecutionResult{result})
}

func outputRefsForTeamWork(link confirmedActionTeamWorkLink, workItemID, teamID string, outputs []protocol.ExecutionOutput) []protocol.TeamOutputRef {
	refs := make([]protocol.TeamOutputRef, 0, len(outputs))
	for _, output := range outputs {
		storageRef := teamOutputStorageRefFromExecutionOutput(output)
		refs = append(refs, protocol.TeamOutputRef{
			OutputID:      firstNonEmptyString(output.ID, output.Title),
			TeamID:        teamID,
			WorkItemID:    workItemID,
			RunID:         link.RunID,
			Kind:          firstNonEmptyString(output.Kind, "output"),
			Label:         firstNonEmptyString(output.Title, output.ID, "Team output"),
			StorageRef:    storageRef,
			Entrypoint:    relativeTeamOutputEntrypoint(storageRef, output.Entrypoint),
			ValidationRef: output.Validation,
			ProofRef:      link.ProofArtifactID,
			ContractID:    link.ContractID,
			ProofID:       link.ProofArtifactID,
			AuditRefs:     compactProofRefs(link.AuditID),
			CreatedAt:     time.Now().UTC(),
		})
	}
	return refs
}

func isTeamWorkTool(toolName string) bool {
	return strings.TrimSpace(toolName) == "create_team" || isDelegateTool(toolName) || isDeliverableTool(toolName)
}

func isDelegateTool(toolName string) bool {
	toolName = strings.TrimSpace(toolName)
	return toolName == "delegate" || toolName == "delegate_task"
}

func isDeliverableTool(toolName string) bool {
	switch strings.TrimSpace(toolName) {
	case "generate_image", "save_cached_image", "write_file", "store_artifact":
		return true
	default:
		return false
	}
}

func executionShapeForTeamWorkTool(toolName string) protocol.TeamExecutionShape {
	if strings.TrimSpace(toolName) == "create_team" {
		return protocol.TeamExecutionShapeCreateTeam
	}
	if isDeliverableTool(toolName) {
		return protocol.TeamExecutionShapeDeliverable
	}
	return protocol.TeamExecutionShapeDelegatedWork
}

func objectiveForPlannedTool(toolName string, args map[string]any) string {
	if isDelegateTool(toolName) {
		if ask, ok := args["ask"].(map[string]any); ok {
			return firstNonEmptyString(ask["goal"], ask["task"], ask["intent"])
		}
		if task, ok := args["task"].(map[string]any); ok {
			return firstNonEmptyString(task["goal"], task["operation"], task["intent"])
		}
		return firstNonEmptyString(args["goal"], args["task"], args["intent"], "Delegated team work")
	}
	return firstNonEmptyString(args["goal"], args["objective"], args["description"], fmt.Sprintf("Confirmed %s action", strings.TrimSpace(toolName)))
}

func objectiveForDeliverableResult(result plannedToolExecutionResult) string {
	packageOutput := projectPackageOutputFromArgs(result.Arguments)
	if packageOutput != nil {
		return "Produce retained deliverable package " + packageOutput.Title
	}
	path := firstNonEmptyString(result.Arguments["path"], result.Arguments["file_path"], result.Arguments["target_path"])
	if path != "" {
		return "Produce retained output " + path
	}
	if strings.TrimSpace(result.Name) == "generate_image" {
		return firstNonEmptyString(result.Arguments["goal"], result.Arguments["prompt"], "Generate retained media output")
	}
	if strings.TrimSpace(result.Name) == "save_cached_image" {
		return firstNonEmptyString(result.Arguments["goal"], result.Arguments["filename"], "Save generated media output")
	}
	return objectiveForPlannedTool(result.Name, result.Arguments)
}

func expectedOutputsFromDeliverableResult(result plannedToolExecutionResult, outputs []protocol.ExecutionOutput) []string {
	items := make([]string, 0, len(outputs))
	for _, output := range outputs {
		items = append(items, firstNonEmptyString(output.Title, output.ID, output.Kind))
	}
	if len(items) == 0 {
		items = append(items, firstNonEmptyString(result.Name, "Deliverable output"))
	}
	return normalizeStringSlice(items)
}

func expectedOutputsFromDelegateArgs(args map[string]any) []string {
	if ask, ok := args["ask"].(map[string]any); ok {
		return confirmedActionStringSlice(ask["exit_criteria"])
	}
	return firstStringSliceArgument(args["expected_outputs"], args["exit_criteria"])
}

func expectedProofFromDelegateArgs(args map[string]any) []string {
	if ask, ok := args["ask"].(map[string]any); ok {
		return confirmedActionStringSlice(ask["evidence_required"])
	}
	return firstStringSliceArgument(args["expected_proof"], args["evidence_required"])
}

func requiredCapabilitiesFromDelegateArgs(args map[string]any) []string {
	if ask, ok := args["ask"].(map[string]any); ok {
		return confirmedActionStringSlice(ask["required_capabilities"])
	}
	return firstStringSliceArgument(args["required_capabilities"], args["capabilities"], args["tools"])
}

func approvalPostureFromScope(scope *protocol.ScopeValidation) protocol.ApprovalPosture {
	if scope != nil && scope.Approval != nil {
		switch strings.TrimSpace(scope.Approval.ApprovalMode) {
		case string(protocol.ApprovalPostureAutoAllowed):
			return protocol.ApprovalPostureAutoAllowed
		case string(protocol.ApprovalPostureOptional):
			return protocol.ApprovalPostureOptional
		}
	}
	return protocol.ApprovalPostureRequired
}

func confirmedActionCreatedTeamIDFromScope(scope *protocol.ScopeValidation) string {
	if scope == nil {
		return ""
	}
	for _, planned := range scope.PlannedToolCalls {
		if strings.TrimSpace(planned.Name) == "create_team" {
			return confirmedActionTeamID(planned.Arguments)
		}
	}
	return ""
}

func teamDisplayName(args map[string]any, teamID string) string {
	return firstNonEmptyString(mergedTeamArgs(args)["name"], teamID)
}

func failureSummary(err error) string {
	if err == nil {
		return "Confirmed action failed before team work completed."
	}
	return "Confirmed action failed before team work completed: " + err.Error()
}

func compactProofRefs(values ...string) []string {
	return normalizeStringSlice(values)
}
