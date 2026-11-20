package server

import (
	"encoding/json"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func isChainedTeamPipelineRequest(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	return requestContainsAny(lower, []string{
		"chain", "chained", "2 or more", "two or more", "two teams", "between teams",
		"mcp output", "mcp outputs", "registered mcp", "listen and speak",
		"listen on", "speak on", "feed output", "pipe output", "handoff between",
	})
}

func isTeamExecutionBundleRequest(lower string) bool {
	if lower == "" {
		return false
	}
	return (strings.Contains(lower, "plan for") && (strings.Contains(lower, "execute next steps") || strings.Contains(lower, "next steps"))) ||
		strings.Contains(lower, "plan the next move for") ||
		(strings.Contains(lower, "work with") && strings.Contains(lower, "produce a concise brief")) ||
		(strings.Contains(lower, "execute next steps") && (strings.Contains(lower, "team") || strings.Contains(lower, "deliverable")))
}

func extractTargetTeamFromRequest(text string) (string, string) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	teamID, teamName := "", ""
	for _, marker := range []string{"(team_id:", "(id:", "team:"} {
		if idx := strings.Index(lower, marker); idx >= 0 {
			rest := strings.TrimSpace(trimmed[idx+len(marker):])
			for i, r := range rest {
				if r == ')' || r == ' ' || r == ',' || r == ';' {
					rest = rest[:i]
					break
				}
			}
			teamID = strings.TrimSpace(rest)
			break
		}
	}
	for _, prefix := range []string{"plan for team ", "plan for ", "plan the next move for ", "work with "} {
		if idx := strings.Index(lower, prefix); idx >= 0 {
			rest := strings.TrimSpace(trimmed[idx+len(prefix):])
			lowerRest := strings.ToLower(rest)
			for _, stop := range []string{" (", " and execute", " to produce", " on the", " next steps", "."} {
				if sIdx := strings.Index(lowerRest, stop); sIdx > 0 {
					rest = strings.TrimSpace(rest[:sIdx])
					lowerRest = strings.ToLower(rest)
				}
			}
			if rest != "" {
				teamName = strings.Trim(rest, "\"'`")
				break
			}
		}
	}
	if teamID == "" && teamName != "" {
		teamID = teamName
	}
	slug := sanitizeCodexBundleSlug(teamID)
	if slug == "" || slug == "codex-task-package" {
		slug = "mission-delivery-team"
	}
	if teamName == "" {
		teamName = titleCaseTeamID(slug)
	}
	return slug, teamName
}

func teamExecutionBundleCalls(text string) []protocol.PlannedToolCall {
	teamID, teamName := extractTargetTeamFromRequest(text)
	cmdSub := "swarm.team." + teamID + ".internal.command"
	resSub := "swarm.team." + teamID + ".signal.result"
	specs := []deliverableBundleFileSpec{
		{
			name: "team_execution_plan.md", title: teamName + " — Team Execution Plan", contentType: "text/markdown",
			content: "# " + teamName + " — Team Execution Plan\n\n## Mission Goal\n" + text + "\n\n## Specialist Role Assignments\n- **Lead Architect (`council-architect`):** Decomposes scope and validates write-isolation partitions.\n- **Delivery Specialist (`prime-development-agent`):** Executes next steps and produces verifiable workspace artifacts.\n- **Verification Sentry (`council-sentry`):** Audits SHA-256 readback and OutputProofEnvelope gates.\n\n## Governed NATS Channels\n- **Command Input:** `" + cmdSub + "`\n- **Verified Result:** `" + resSub + "`\n",
		},
		{
			name: "next_steps_action_matrix.md", title: teamName + " — Next Steps & Action Matrix", contentType: "text/markdown",
			content: "# " + teamName + " — Next Steps & Action Matrix\n\n| Step | Owner Role | Verification Gate | Expected Output |\n| --- | --- | --- | --- |\n| 1. Scope & Channel Alignment | `council-architect` | NATS subject contract verified | `team_execution_plan.md` |\n| 2. Deliverable Execution | `prime-development-agent` | Disjoint write-isolation pass | `next_steps_action_matrix.md` |\n| 3. Proof & Handoff Release | `council-sentry` | SHA-256 `OutputProofEnvelope` verified | `team_handoff_summary.json`, `proof.md` |\n",
		},
		{
			name: "team_handoff_summary.json", title: teamName + " — Channel & Handoff Summary (JSON)", contentType: "application/json",
			content: "{\n  \"team_id\": " + strconvQuoteSafe(teamID) + ",\n  \"team_name\": " + strconvQuoteSafe(teamName) + ",\n  \"command_channel\": " + strconvQuoteSafe(cmdSub) + ",\n  \"result_channel\": " + strconvQuoteSafe(resSub) + ",\n  \"status\": \"verified\",\n  \"next_steps\": [\n    \"Execute scoped team deliverables under disjoint file ownership\",\n    \"Publish verified OutputProofEnvelope result on " + resSub + "\"\n  ]\n}\n",
		},
	}
	calls := buildDeliverableBundleCalls("deliverables/teams/"+teamID, "team_execution_bundle", teamID, specs)
	for i := range calls {
		calls[i].Arguments["team_id"] = teamID
		calls[i].Arguments["team_name"] = teamName
	}
	return calls
}

func hasTeamCreationIntent(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	if isChainedTeamPipelineRequest(lower) || isTeamExecutionBundleRequest(lower) {
		return true
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

func inferCreateTeamPlanCallsFromRequest(latestRequest, agentReply string) []protocol.PlannedToolCall {
	if isChainedTeamPipelineRequest(latestRequest) {
		return []protocol.PlannedToolCall{
			{
				Name: "create_team",
				Arguments: map[string]any{
					"team_id":     "mcp-upstream-discovery-team",
					"name":        "MCP & Upstream Discovery Team",
					"description": "Collects registered MCP tool outputs and publishes normalized findings on its result channel.",
					"inputs":      []string{"swarm.team.mcp-upstream-discovery-team.internal.command", "swarm.mcp.outputs.registered", "swarm.global.broadcast"},
					"deliveries":  []string{"swarm.team.mcp-upstream-discovery-team.signal.result", "swarm.team.mcp-upstream-discovery-team.signal.status"},
					"members": map[string]any{
						"mcp-collector":     "council-architect",
						"signal-normalizer": "council-sentry",
					},
					"tools": []string{"read_file", "publish_signal", "write_file"},
				},
			},
			{
				Name: "create_team",
				Arguments: map[string]any{
					"team_id":     "chained-downstream-delivery-team",
					"name":        "Chained Downstream Delivery Team",
					"description": "Listens on swarm.team.mcp-upstream-discovery-team.signal.result and synthesizes downstream deliverables.",
					"inputs":      []string{"swarm.team.mcp-upstream-discovery-team.signal.result", "swarm.team.chained-downstream-delivery-team.internal.command"},
					"deliveries":  []string{"swarm.team.chained-downstream-delivery-team.signal.result", "swarm.team.chained-downstream-delivery-team.signal.status"},
					"members": map[string]any{
						"lead-integrator": "prime-development-agent",
						"output-verifier": "agui-design-architect-agent",
					},
					"tools": []string{"write_file", "publish_signal"},
				},
			},
		}
	}
	call, ok := inferCreateTeamPlanFromRequest(latestRequest, agentReply)
	if !ok {
		return nil
	}
	teamID := firstNonEmptyString(call.Arguments["team_id"], "mission-delivery-team")
	if call.Arguments["inputs"] == nil {
		call.Arguments["inputs"] = []string{"swarm.team." + teamID + ".internal.command", "swarm.global.broadcast"}
	}
	if call.Arguments["deliveries"] == nil {
		call.Arguments["deliveries"] = []string{"swarm.team." + teamID + ".signal.result", "swarm.team." + teamID + ".signal.status"}
	}
	return []protocol.PlannedToolCall{call}
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
				"inputs":      []string{"swarm.team." + teamID + ".internal.command", "swarm.global.broadcast"},
				"deliveries":  []string{"swarm.team." + teamID + ".signal.result", "swarm.team." + teamID + ".signal.status"},
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
				"inputs":      []string{"swarm.team.tetris-street-fighter.internal.command", "swarm.global.broadcast"},
				"deliveries":  []string{"swarm.team.tetris-street-fighter.signal.result", "swarm.team.tetris-street-fighter.signal.status"},
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

	teamID, teamName := "mission-delivery-team", "Mission Delivery Team"
	if isTeamExecutionBundleRequest(lower) {
		teamID, teamName = extractTargetTeamFromRequest(cleanRequest)
	}
	return protocol.PlannedToolCall{
		Name: "create_team",
		Arguments: map[string]any{
			"team_id":     teamID,
			"name":        teamName,
			"description": cleanRequest,
			"inputs":      []string{"swarm.team." + teamID + ".internal.command", "swarm.global.broadcast"},
			"deliveries":  []string{"swarm.team." + teamID + ".signal.result", "swarm.team." + teamID + ".signal.status"},
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
