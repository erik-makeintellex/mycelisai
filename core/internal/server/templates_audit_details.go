package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// Delegation is the first governed tool where operators benefit from seeing the
// underlying structured ask without exposing raw command payloads by default.
func buildExecutionAuditDetailsForTool(planned protocol.PlannedToolCall, resolvedToolName string) map[string]any {
	details := map[string]any{"tool": resolvedToolName}
	if resolvedToolName != "delegate_task" {
		return details
	}

	teamID, ask := extractDelegationAuditFields(planned.Arguments)
	if teamID != "" {
		details["team_id"] = teamID
	}
	if ask.IsZero() {
		return details
	}

	details["ask_kind"] = string(ask.AskKind)
	details["lane_role"] = string(ask.LaneRole)
	if goal := strings.TrimSpace(ask.Goal); goal != "" {
		details["goal"] = goal
	}
	if summary := protocol.SummarizeTeamAsk(ask); summary != "" {
		details["operator_summary"] = summary
	}
	return details
}

func extractDelegationAuditFields(args map[string]any) (string, protocol.TeamAsk) {
	if len(args) == 0 {
		return "", protocol.TeamAsk{}
	}

	teamID := firstNonEmptyString(args["team_id"], args["teamId"], args["target_team"])
	if teamID == "" {
		switch team := args["team"].(type) {
		case map[string]any:
			teamID = firstNonEmptyString(team["id"], team["team_id"], team["name"])
		case string:
			teamID = strings.TrimSpace(team)
		}
	}

	if rawAsk, ok := args["ask"].(map[string]any); ok {
		return teamID, protocol.TeamAskFromMap(rawAsk)
	}
	switch task := args["task"].(type) {
	case map[string]any:
		return teamID, protocol.TeamAskFromMap(task)
	case string:
		return teamID, protocol.NormalizeTeamAsk(protocol.TeamAsk{Goal: strings.TrimSpace(task)})
	}

	return teamID, protocol.NormalizeTeamAsk(protocol.TeamAsk{
		AskKind:  protocol.TeamAskKind(firstNonEmptyString(args["ask_kind"])),
		LaneRole: protocol.TeamLaneRole(firstNonEmptyString(args["lane_role"])),
		Goal: firstNonEmptyString(
			args["goal"],
			args["intent"],
			args["message"],
			args["operation"],
		),
	})
}
