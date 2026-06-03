package server

import "github.com/mycelis/core/pkg/protocol"

func visibilityConfirmActionSummary(scope *protocol.ScopeValidation, results ...plannedToolExecutionResult) *protocol.ExecutionSummary {
	return buildConfirmActionExecutionSummary(
		"proof-123",
		"contract-123",
		"artifact-123",
		"run-123",
		"audit-123",
		scope,
		results,
	)
}

func visibilityScopeValidation(tools ...string) *protocol.ScopeValidation {
	calls := make([]protocol.PlannedToolCall, 0, len(tools))
	for _, tool := range tools {
		calls = append(calls, protocol.PlannedToolCall{Name: tool})
	}
	return &protocol.ScopeValidation{
		Tools:            tools,
		PlannedToolCalls: calls,
	}
}

func visibilityCreateTeamResult(teamID, name string) plannedToolExecutionResult {
	return plannedToolExecutionResult{
		Name:      "create_team",
		Arguments: map[string]any{"team_id": teamID, "name": name},
	}
}

func visibilityWriteFileResult(path string) plannedToolExecutionResult {
	return plannedToolExecutionResult{
		Name:      "write_file",
		Arguments: map[string]any{"path": path},
	}
}

func visibilityToolResult(
	name string,
	arguments map[string]any,
	artifacts ...protocol.ChatArtifactRef,
) plannedToolExecutionResult {
	return plannedToolExecutionResult{
		Name:      name,
		Arguments: arguments,
		Artifacts: artifacts,
	}
}

func visibilitySavedImageArtifact(id, title string) protocol.ChatArtifactRef {
	return protocol.ChatArtifactRef{
		ID:        id,
		Type:      "image",
		Title:     title,
		SavedPath: "saved-media/comic-page.png",
	}
}
