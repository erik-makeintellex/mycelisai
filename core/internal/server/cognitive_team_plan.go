package server

import (
	"encoding/json"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func hasTeamCreationIntent(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	if requestContainsAny(lower, []string{
		"put a team together",
		"put together a team",
		"create a team",
		"build a team",
		"launch a team",
		"assemble a team",
		"spawn a team",
		"start a team",
	}) {
		return true
	}
	return requestContainsAny(lower, []string{"create", "build", "launch", "put together", "assemble", "spawn"}) &&
		requestContainsAny(lower, []string{"team", "teams", "specialists"})
}

func containsRawTeamJSON(text string) (map[string]any, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, false
	}

	candidates := []string{trimmed}
	if start := strings.Index(trimmed, "```json"); start >= 0 {
		rest := trimmed[start+len("```json"):]
		if end := strings.Index(rest, "```"); end >= 0 {
			candidates = append(candidates, strings.TrimSpace(rest[:end]))
		}
	}
	if start := strings.Index(trimmed, "{"); start >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > start {
			candidates = append(candidates, strings.TrimSpace(trimmed[start:end+1]))
		}
	}

	for _, candidate := range candidates {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(candidate), &parsed); err != nil {
			continue
		}
		if _, isToolCall := parsed["tool_call"]; isToolCall {
			continue
		}
		if firstNonEmptyString(parsed["team_id"]) != "" {
			return parsed, true
		}
	}
	return nil, false
}

func containsRawJSONBlock(text string) bool {
	if _, ok := containsRawTeamJSON(text); ok {
		return true
	}
	return strings.Contains(text, "\"team_id\"") || strings.Contains(text, "```json")
}

func titleCaseTeamID(teamID string) string {
	parts := strings.Fields(strings.ReplaceAll(strings.TrimSpace(teamID), "-", " "))
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, " ")
}

func inferCreateTeamPlanFromRequest(latestRequest, agentReply string) (protocol.PlannedToolCall, bool) {
	cleanRequest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(latestRequest), "Confirmed action from prior user request:\n"))
	if rawMap, ok := containsRawTeamJSON(agentReply); ok {
		teamID := firstNonEmptyString(rawMap["team_id"], rawMap["id"])
		name := firstNonEmptyString(rawMap["name"])
		if name == "" {
			name = titleCaseTeamID(teamID)
		}
		return protocol.PlannedToolCall{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id":     teamID,
				"name":        name,
				"description": firstNonEmptyString(rawMap["instructions"], rawMap["description"], cleanRequest),
				"members":     rawMap["members"],
				"tools":       rawMap["tools"],
			},
		}, true
	}

	if !hasTeamCreationIntent(cleanRequest) {
		return protocol.PlannedToolCall{}, false
	}

	lower := strings.ToLower(cleanRequest)
	if isTetrisGameBundleRequest(lower) || strings.Contains(lower, "tetris") || strings.Contains(lower, "street fighter") {
		return protocol.PlannedToolCall{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id":     "tetris-street-fighter",
				"name":        "Street Fighter II Tetris Team",
				"description": "Arcade game design, architecture, development, and character theme team",
				"members": map[string]any{
					"designer":  "agui-design-architect-agent",
					"architect": "council-architect",
					"developer": "prime-development-agent",
					"artist":    "council-creative",
				},
				"tools": []string{"generate_image", "local_command", "write_file"},
			},
		}, true
	}

	return protocol.PlannedToolCall{
		Name: "create_team",
		Arguments: map[string]any{
			"team_id":     "mission-delivery-team",
			"name":        "Mission Delivery Team",
			"description": cleanRequest,
			"members": map[string]any{
				"architect": "council-architect",
				"developer": "prime-development-agent",
				"designer":  "agui-design-architect-agent",
				"reviewer":  "council-sentry",
			},
			"tools": []string{"write_file", "delegate_task"},
		},
	}, true
}
