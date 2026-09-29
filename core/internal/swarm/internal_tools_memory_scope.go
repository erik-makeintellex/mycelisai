package swarm

import (
	"context"
	"strings"
)

type memoryScope struct {
	TenantID   string
	TeamID     string
	AgentID    string
	RunID      string
	Visibility string
}

func normalizedVisibility(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "private", "team", "global":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func resolveMemoryScope(ctx context.Context, args map[string]any) memoryScope {
	scope := memoryScope{
		TenantID: "default",
	}
	userTurn := false
	if inv, ok := ToolInvocationContextFromContext(ctx); ok {
		scope.TeamID = strings.TrimSpace(inv.TeamID)
		scope.AgentID = strings.TrimSpace(inv.AgentID)
		scope.RunID = strings.TrimSpace(inv.RunID)
		userTurn = inv.Recall.User
	}

	// MEM-LANES: a signed-in user's rows keep Core's tenant; the model never
	// picks it.
	if value := strings.TrimSpace(stringValue(args["tenant_id"])); value != "" && !userTurn {
		scope.TenantID = value
	}
	// Model-supplied ids only fill gaps: they never replace the runtime's
	// invocation identity, so forged args cannot widen recall or write scope.
	if value := strings.TrimSpace(stringValue(args["team_id"])); value != "" && scope.TeamID == "" {
		scope.TeamID = value
	}
	if value := strings.TrimSpace(stringValue(args["agent_id"])); value != "" && scope.AgentID == "" {
		scope.AgentID = value
	}
	if value := strings.TrimSpace(stringValue(args["run_id"])); value != "" {
		scope.RunID = value
	}

	scope.Visibility = normalizedVisibility(stringValue(args["visibility"]))
	if scope.Visibility == "" {
		switch {
		case userTurn:
			// A user's own words stay theirs unless they explicitly share them
			// with the team (proven membership) or org-wide (MEM-LANES).
			scope.Visibility = "private"
		case scope.TeamID != "":
			scope.Visibility = "team"
		case scope.AgentID != "":
			scope.Visibility = "private"
		default:
			scope.Visibility = "global"
		}
	}

	return scope
}

func dedupeStringValues(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
