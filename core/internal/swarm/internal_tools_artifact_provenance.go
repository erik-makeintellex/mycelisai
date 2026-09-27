package swarm

import (
	"context"
	"encoding/json"
	"strings"
)

// artifactProvenanceKeys are owned by Core. Model-supplied metadata with the
// same keys is discarded so an agent cannot claim another team's artifact.
var artifactProvenanceKeys = []string{"provenance", "provenance_agent_id", "provenance_team_id", "provenance_run_id", "provenance_source", "provenance_source_channel", "provenance_sensitivity"}

// attributeArtifact returns the artifacts.agent_id and metadata for a new
// artifact. With an agent invocation the real agent, team and run are
// recorded as "attributed"; without one the row keeps agent_id 'internal'
// and is marked "unattributed", which can never be handed off.
func attributeArtifact(ctx context.Context, metaJSON string) (string, string) {
	meta := map[string]any{}
	if strings.TrimSpace(metaJSON) != "" {
		_ = json.Unmarshal([]byte(metaJSON), &meta)
	}
	for _, key := range artifactProvenanceKeys {
		delete(meta, key)
	}
	agentID := "internal"
	inv, ok := ToolInvocationContextFromContext(ctx)
	if ok && strings.TrimSpace(inv.AgentID) != "" {
		agentID = strings.TrimSpace(inv.AgentID)
		meta["provenance"] = "attributed"
		meta["provenance_agent_id"] = agentID
		meta["provenance_team_id"] = strings.TrimSpace(inv.TeamID)
		meta["provenance_run_id"] = strings.TrimSpace(inv.RunID)
		meta["provenance_source"] = "tool_invocation"
		meta["provenance_sensitivity"] = coreArtifactSensitivity(stringValue(meta["sensitivity_class"]))
		if channel := strings.TrimSpace(inv.SourceChannel); channel != "" {
			meta["provenance_source_channel"] = channel
		}
	} else {
		meta["provenance"] = "unattributed"
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return agentID, `{"provenance":"unattributed"}`
	}
	return agentID, string(raw)
}

// artifactExchangeAgentID names the producing agent on the Exchange mirror.
func artifactExchangeAgentID(ctx context.Context) string {
	if inv, ok := ToolInvocationContextFromContext(ctx); ok && strings.TrimSpace(inv.AgentID) != "" {
		return strings.TrimSpace(inv.AgentID)
	}
	return "internal"
}

// artifactSensitivityRank orders the classifications Core understands.
var artifactSensitivityRank = map[string]int{"public": 0, "org_visible": 1, "team_scoped": 2, "role_scoped": 3, "restricted": 4, "admin_only": 5}

// coreArtifactSensitivity is Core's stored classification for a new team
// artifact: team_scoped by default. A model label may raise it, never lower it.
func coreArtifactSensitivity(modelLabel string) string {
	label := strings.ToLower(strings.TrimSpace(modelLabel))
	if rank, ok := artifactSensitivityRank[label]; ok && rank > artifactSensitivityRank["team_scoped"] {
		return label
	}
	return "team_scoped"
}

// ArtifactSensitivityAllowsHandoff reports whether Core's stored class permits
// an in-run handoff without approval. Unknown or missing counts as restricted.
func ArtifactSensitivityAllowsHandoff(coreClass string) bool {
	rank, ok := artifactSensitivityRank[strings.ToLower(strings.TrimSpace(coreClass))]
	return ok && rank < artifactSensitivityRank["restricted"]
}
