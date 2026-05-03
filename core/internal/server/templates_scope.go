package server

import "github.com/mycelis/core/pkg/protocol"

// buildScopeFromBlueprint extracts scope validation metadata from a blueprint.
func buildScopeFromBlueprint(bp *protocol.MissionBlueprint) *protocol.ScopeValidation {
	toolSet := make(map[string]bool)
	for _, team := range bp.Teams {
		for _, agent := range team.Agents {
			for _, tool := range agent.Tools {
				toolSet[tool] = true
			}
		}
	}

	tools := make([]string, 0, len(toolSet))
	for t := range toolSet {
		tools = append(tools, t)
	}

	totalAgents := 0
	for _, team := range bp.Teams {
		totalAgents += len(team.Agents)
	}

	risk := "low"
	if totalAgents > 5 || len(bp.Teams) > 2 {
		risk = "medium"
	}
	if totalAgents > 10 || len(bp.Teams) > 4 {
		risk = "high"
	}

	return &protocol.ScopeValidation{
		Tools:             tools,
		AffectedResources: []string{"missions", "teams", "service_manifests"},
		RiskLevel:         risk,
	}
}
