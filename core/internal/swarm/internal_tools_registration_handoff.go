package swarm

func (r *InternalToolRegistry) registerHandoffTools() {
	r.tools["hand_off"] = &InternalTool{
		Name:        "hand_off",
		Description: "Team leads only: hand stored artifacts from your team to another team in the same approved run. Core records a Team handoff, queues work for the receiving team, and notifies it. The result is queued, never delivered, until that team accepts. A team outside the run, or restricted content, returns handoff_needs_approval: ask Soma to propose it.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"target_team_id":  map[string]any{"type": "string", "description": "The receiving team's id"},
			"target_role":     map[string]any{"type": "string", "description": "Optional role on the receiving team that should pick this up"},
			"artifact_ids":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "1 to 10 artifact ids your team stored in this run with store_artifact"},
			"note":            map[string]any{"type": "string", "description": "Concise handoff note (at most 1,000 characters): what you are handing over and what the receiving team should do"},
			"expected_action": map[string]any{"type": "string", "enum": []string{"review", "use", "continue"}, "description": "What the receiving team should do with the inputs"},
		}, "required": []string{"target_team_id", "artifact_ids", "note"}},
		Handler: r.handleHandOff,
	}
	r.tools["read_handoff_input"] = &InternalTool{
		Name:        "read_handoff_input",
		Description: "Team leads only: read one artifact that another team handed to your team. Works only for handoffs addressed to your team and artifacts listed in them; content is limited to 32 KB. Core records the read on your team's handoff work item.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"handoff_id":  map[string]any{"type": "string", "description": "The handoff id from your team's handoff work"},
			"artifact_id": map[string]any{"type": "string", "description": "An artifact id listed in that handoff"},
		}, "required": []string{"handoff_id", "artifact_id"}},
		Handler: r.handleReadHandoffInput,
	}
}
