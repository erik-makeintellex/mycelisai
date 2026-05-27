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
		systemPrompt = fmt.Sprintf("You are %s in team %s. Execute assigned tasks and report outcomes. Start as the only team member; request a temporary specialist only when you can name the missing capability, owned task, proof expected, and removal point.", role, teamID)
	}
	members := runtimeTeamMembersFromArgs(merged, teamID, agentID, role, systemPrompt)

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
	askRouting := parseTeamAskRouting(merged["ask_routing"])
	if len(askRouting) == 0 {
		askRouting = defaultTeamAskRouting()
	}

	return &TeamManifest{
		ID:          teamID,
		Name:        name,
		Type:        teamType,
		Description: runtimeTeamDescription(members),
		AskRouting:  askRouting,
		Members:     members,
		Inputs:      inputs,
		Deliveries:  deliveries,
	}
}

func runtimeTeamDescription(members []protocol.AgentManifest) string {
	if len(members) <= 1 {
		return "Runtime-created lead-only team; expand only with operator action or justified temporary specialist request."
	}
	return "Runtime-created specialist delivery team with bounded roles, retained outputs, and Soma-owned governance."
}

func runtimeTeamMembersFromArgs(merged map[string]any, teamID, fallbackAgentID, fallbackRole, fallbackSystemPrompt string) []protocol.AgentManifest {
	tools := stringSlice(merged["tools"])
	if agents := runtimeAgentsFromRaw(merged["agents"], teamID, tools); len(agents) > 0 {
		return agents
	}
	return []protocol.AgentManifest{{
		ID:            fallbackAgentID,
		Role:          fallbackRole,
		SystemPrompt:  fallbackSystemPrompt,
		Tools:         tools,
		MaxIterations: 6,
	}}
}

func runtimeAgentsFromRaw(raw any, teamID string, fallbackTools []string) []protocol.AgentManifest {
	var sources []map[string]any
	switch typed := raw.(type) {
	case []map[string]any:
		sources = typed
	case []any:
		for _, item := range typed {
			if source, ok := item.(map[string]any); ok {
				sources = append(sources, source)
			}
		}
	}
	members := make([]protocol.AgentManifest, 0, len(sources))
	seen := map[string]struct{}{}
	for idx, source := range sources {
		member := runtimeAgentFromMap(source, teamID, idx, fallbackTools)
		if member.ID == "" {
			continue
		}
		if _, exists := seen[member.ID]; exists {
			continue
		}
		seen[member.ID] = struct{}{}
		members = append(members, member)
	}
	return members
}

func runtimeAgentFromMap(source map[string]any, teamID string, idx int, fallbackTools []string) protocol.AgentManifest {
	role := pickFirstString(source, "role", "name")
	if role == "" {
		role = "specialist"
	}
	id := normalizeRuntimeID(pickFirstString(source, "id", "agent_id"))
	if id == "" {
		id = normalizeRuntimeID(fmt.Sprintf("%s-%s-%d", teamID, role, idx+1))
	}
	tools := stringSlice(source["tools"])
	if len(tools) == 0 {
		tools = fallbackTools
	}
	if len(tools) == 0 {
		tools = []string{"store_artifact"}
	}
	maxIterations := intValue(source["max_iterations"])
	if maxIterations <= 0 {
		maxIterations = 6
	}
	return protocol.AgentManifest{
		ID:            id,
		Role:          role,
		SystemPrompt:  firstNonEmptyString(stringValue(source["system_prompt"]), fmt.Sprintf("You are the %s for team %s. Own your bounded specialist contribution and report concise output/proof to Soma.", role, teamID)),
		Model:         stringValue(source["model"]),
		Provider:      stringValue(source["provider"]),
		Inputs:        stringSlice(source["inputs"]),
		Outputs:       stringSlice(source["outputs"]),
		Tools:         tools,
		MaxIterations: maxIterations,
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
