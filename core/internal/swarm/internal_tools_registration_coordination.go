package swarm

func (r *InternalToolRegistry) registerCoordinationTools() {
	r.tools["consult_council"] = &InternalTool{
		Name:        "consult_council",
		Description: "Send a question to a specific council member (architect, coder, creative, sentry) and get their response. Use for specialist expertise.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"member":   map[string]any{"type": "string", "description": "Council member ID: council-architect, council-coder, council-creative, or council-sentry"},
				"question": map[string]any{"type": "string", "description": "The question or task for the council member"},
			},
			"required": []string{"member", "question"},
		},
		Handler: r.handleConsultCouncil,
	}
	r.tools["delegate_task"] = &InternalTool{
		Name:        "delegate_task",
		Description: "Publish a structured ask to a specific team's command topic for processing.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"team_id": map[string]any{"type": "string", "description": "The target team ID"},
				"task":    map[string]any{"type": "string", "description": "Legacy plain-text task description. Prefer the structured ask object when possible."},
				"ask": map[string]any{
					"type":        "object",
					"description": "Structured team ask carrying goal, lane posture, and proof expectations.",
					"properties": map[string]any{
						"schema_version":        map[string]any{"type": "string"},
						"ask_kind":              map[string]any{"type": "string", "enum": []string{"coordination", "research", "implementation", "validation", "review"}},
						"lane_role":             map[string]any{"type": "string", "enum": []string{"coordinator", "researcher", "implementer", "validator", "reviewer"}},
						"goal":                  map[string]any{"type": "string"},
						"owned_scope":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"constraints":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"required_capabilities": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"approval_posture":      map[string]any{"type": "string", "enum": []string{"auto_allowed", "optional", "required"}},
						"exit_criteria":         map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"evidence_required":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"context":               map[string]any{"type": "object"},
					},
				},
				"hint": map[string]any{
					"type":        "object",
					"description": "Optional scoring hints for delegation priority",
					"properties": map[string]any{
						"confidence": map[string]any{"type": "number", "description": "0.0-1.0 confidence this is the right team"},
						"urgency":    map[string]any{"type": "string", "enum": []string{"low", "medium", "high", "critical"}},
						"complexity": map[string]any{"type": "integer", "description": "1-5 complexity rating"},
						"risk":       map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}},
					},
				},
			},
			"required": []string{"team_id"},
		},
		Handler: r.handleDelegateTask,
	}
	r.tools["create_team"] = &InternalTool{
		Name:        "create_team",
		Description: "Create and start a new team at runtime with a minimal manifest.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"team_id":       map[string]any{"type": "string", "description": "Unique team ID"},
				"name":          map[string]any{"type": "string", "description": "Display name (optional)"},
				"type":          map[string]any{"type": "string", "enum": []string{"action", "expression"}, "description": "Team type (default action)"},
				"role":          map[string]any{"type": "string", "description": "Primary agent role (default worker)"},
				"agent_id":      map[string]any{"type": "string", "description": "Optional first agent ID"},
				"system_prompt": map[string]any{"type": "string", "description": "Optional first agent system prompt"},
				"tools":         map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional first agent tools"},
				"inputs":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional team input subjects"},
				"deliveries":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional team delivery subjects"},
				"ask_routing": map[string]any{
					"type":                 "object",
					"description":          "Optional per-team ask-kind to lane-role routing hints.",
					"additionalProperties": map[string]any{"type": "string"},
				},
			},
			"required": []string{"team_id"},
		},
		Handler: r.handleCreateTeam,
	}
	r.tools["list_teams"] = &InternalTool{
		Name:        "list_teams",
		Description: "Returns the active team roster with member counts.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Handler:     r.handleListTeams,
	}
	r.tools["list_missions"] = &InternalTool{
		Name:        "list_missions",
		Description: "Returns active missions with team/agent counts.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Handler:     r.handleListMissions,
	}
	r.tools["get_system_status"] = &InternalTool{
		Name:        "get_system_status",
		Description: "Returns system telemetry: goroutines, heap memory, LLM tokens/sec.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Handler:     r.handleGetSystemStatus,
	}
	r.tools["list_available_tools"] = &InternalTool{
		Name:        "list_available_tools",
		Description: "Returns all available tools (internal + MCP) with descriptions.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Handler:     r.handleListAvailableTools,
	}
}
