package swarm

import (
	"context"
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/memory"
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
		default:
			// No user and no team, including a call with no invocation
			// context: never org-wide by default (MEM-LANES-2).
			scope.Visibility = "private"
		}
	}

	return scope
}

// ownerLaneVisibility is the visibility a remembered fact or conversation
// summary is saved with (MEM-LANES-2). The model's visibility is a request,
// never authority:
//
//   - global needs the M2 authority Core set on the turn (OrgWideWrite: a
//     verified root admin with memory:write); a turn with no user, and a call
//     with no invocation context, never has it;
//   - team needs a team, and in a user's turn proven membership of it (the
//     reader's MEM-LIST team keys);
//   - everything else is private.
//
// A narrowed request returns a note for the tool result.
func ownerLaneVisibility(access RecallAccess, scope memoryScope) (string, string) {
	switch scope.Visibility {
	case "global":
		if access.User && !access.Unavailable && access.OrgWideWrite {
			return "global", ""
		}
		return "private", "Kept private: sharing saved memory with the whole organization needs a root admin with memory:write."
	case "team":
		if scope.TeamID == "" {
			return "private", "Kept private: this turn has no team to share it with."
		}
		if !access.User {
			return "team", ""
		}
		key := memory.GovernedTeamKey(scope.TenantID, scope.TeamID)
		for _, member := range access.Reader.TeamKeys {
			if member == key {
				return "team", ""
			}
		}
		return "private", fmt.Sprintf("Kept private: you are not a verified member of team %s, so it was not shared with the team.", scope.TeamID)
	}
	return "private", ""
}

// withVisibilityNote appends a narrowed-visibility note to a tool result.
func withVisibilityNote(result, note string) string {
	if note == "" {
		return result
	}
	return result + " " + note
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
