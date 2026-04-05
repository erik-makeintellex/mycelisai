package swarm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func pickFirstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if v := stringValue(m[key]); v != "" {
			return v
		}
	}
	return ""
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func normalizeRuntimeID(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

func buildRuntimeTeamManifest(args map[string]any) *TeamManifest {
	manifestMap, _ := args["manifest"].(map[string]any)
	merged := map[string]any{}
	for k, v := range args {
		merged[k] = v
	}
	for k, v := range manifestMap {
		if _, exists := merged[k]; !exists {
			merged[k] = v
		}
	}
	if agentsRaw, ok := merged["agents"].([]any); ok && len(agentsRaw) > 0 {
		if first, ok := agentsRaw[0].(map[string]any); ok {
			mergeAgentCompatFields(merged, first)
		}
	}

	teamID := normalizeRuntimeID(pickFirstString(merged, "team_id", "id", "team_name"))
	if teamID == "" {
		return nil
	}
	name := pickFirstString(merged, "name")
	if name == "" {
		name = teamID
	}
	teamType := TeamType(pickFirstString(merged, "type"))
	if teamType != TeamTypeAction && teamType != TeamTypeExpression {
		teamType = TeamTypeAction
	}
	role := pickFirstString(merged, "role")
	if role == "" {
		role = "worker"
	}
	agentID := normalizeRuntimeID(pickFirstString(merged, "agent_id"))
	if agentID == "" {
		agentID = teamID + "-agent"
	}
	systemPrompt := pickFirstString(merged, "system_prompt")
	if systemPrompt == "" {
		systemPrompt = fmt.Sprintf("You are %s in team %s. Execute assigned tasks and report outcomes.", role, teamID)
	}

	inputs := stringSlice(merged["inputs"])
	if len(inputs) == 0 {
		inputs = []string{protocol.TopicGlobalBroadcast}
	}
	deliveries := stringSlice(merged["deliveries"])
	if len(deliveries) == 0 {
		if teamType == TeamTypeExpression {
			deliveries = []string{fmt.Sprintf(protocol.TopicTeamSignalStatus, teamID)}
		} else {
			deliveries = []string{fmt.Sprintf(protocol.TopicTeamSignalResult, teamID)}
		}
	}
	tools := stringSlice(merged["tools"])
	askRouting := parseTeamAskRouting(merged["ask_routing"])
	if len(askRouting) == 0 {
		askRouting = defaultTeamAskRouting()
	}

	return &TeamManifest{
		ID:          teamID,
		Name:        name,
		Type:        teamType,
		Description: "Runtime-created team",
		AskRouting:  askRouting,
		Members: []protocol.AgentManifest{{
			ID:            agentID,
			Role:          role,
			SystemPrompt:  systemPrompt,
			Tools:         tools,
			MaxIterations: 6,
		}},
		Inputs:     inputs,
		Deliveries: deliveries,
	}
}

func parseTeamAskRouting(raw any) map[string]string {
	source, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	resolved := map[string]string{}
	for key, value := range source {
		askKind := strings.TrimSpace(key)
		laneRole := strings.TrimSpace(stringValue(value))
		if askKind == "" || laneRole == "" {
			continue
		}
		resolved[askKind] = laneRole
	}
	if len(resolved) == 0 {
		return nil
	}
	return resolved
}

func defaultTeamAskRouting() map[string]string {
	return map[string]string{
		string(protocol.TeamAskKindCoordination):   string(protocol.TeamLaneRoleCoordinator),
		string(protocol.TeamAskKindResearch):       string(protocol.TeamLaneRoleResearcher),
		string(protocol.TeamAskKindImplementation): string(protocol.TeamLaneRoleImplementer),
		string(protocol.TeamAskKindValidation):     string(protocol.TeamLaneRoleValidator),
		string(protocol.TeamAskKindReview):         string(protocol.TeamLaneRoleReviewer),
	}
}

func mergeAgentCompatFields(merged map[string]any, first map[string]any) {
	if _, exists := merged["agent_id"]; !exists {
		if v := pickFirstString(first, "agent_id", "id"); v != "" {
			merged["agent_id"] = v
		}
	}
	if _, exists := merged["role"]; !exists {
		if v := pickFirstString(first, "role"); v != "" {
			merged["role"] = v
		}
	}
	if _, exists := merged["tools"]; !exists {
		if v, ok := first["tools"]; ok {
			merged["tools"] = v
		}
	}
	if _, exists := merged["system_prompt"]; !exists {
		if v := pickFirstString(first, "system_prompt"); v != "" {
			merged["system_prompt"] = v
		}
	}
}
