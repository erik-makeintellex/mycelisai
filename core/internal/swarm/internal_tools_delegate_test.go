package swarm

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestNormalizeDelegateTaskArgs_ObjectTask(t *testing.T) {
	teamID, ask, err := normalizeDelegateTaskArgs(map[string]any{
		"team_id": "admin-core",
		"task": map[string]any{
			"ask_kind":  "research",
			"lane_role": "researcher",
			"goal":      "Create a web research team",
		},
	})
	if err != nil {
		t.Fatalf("normalize error: %v", err)
	}
	if teamID != "admin-core" {
		t.Fatalf("teamID = %q, want admin-core", teamID)
	}
	if ask.AskKind != protocol.TeamAskKindResearch {
		t.Fatalf("ask_kind = %q", ask.AskKind)
	}
	if ask.LaneRole != protocol.TeamLaneRoleResearcher {
		t.Fatalf("lane_role = %q", ask.LaneRole)
	}
	if ask.Goal != "Create a web research team" {
		t.Fatalf("goal = %q", ask.Goal)
	}
}

func TestNormalizeDelegateTaskArgs_AliasFields(t *testing.T) {
	teamID, ask, err := normalizeDelegateTaskArgs(map[string]any{
		"teamId":   "council-core",
		"ask_kind": "research",
		"intent":   "Find best path",
		"context": map[string]any{
			"topic": "openclaw",
		},
	})
	if err != nil {
		t.Fatalf("normalize error: %v", err)
	}
	if teamID != "council-core" {
		t.Fatalf("teamID = %q, want council-core", teamID)
	}
	if ask.Goal != "Find best path" {
		t.Fatalf("goal = %q", ask.Goal)
	}
	if ask.AskKind != protocol.TeamAskKindResearch {
		t.Fatalf("ask_kind = %q", ask.AskKind)
	}
	if ask.LaneRole != protocol.TeamLaneRoleResearcher {
		t.Fatalf("lane_role = %q", ask.LaneRole)
	}
	if ask.Context["topic"] != "openclaw" {
		t.Fatalf("context.topic = %v", ask.Context["topic"])
	}
}

func TestHandleDelegateTask_NormalizedInputStillExecutes(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	_, err := r.handleDelegateTask(context.Background(), map[string]any{
		"team": map[string]any{"id": "admin-core"},
		"task": map[string]any{
			"operation": "generate_blueprint",
			"intent":    "Create team",
		},
	})
	if err == nil {
		t.Fatal("expected error because NATS is unavailable in this test")
	}
	if !strings.Contains(err.Error(), "NATS not available") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleDelegateTask_MissingRequired(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	_, err := r.handleDelegateTask(context.Background(), map[string]any{
		"operation": "generate_blueprint",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "requires 'team_id' unless Soma can resolve a routed team") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalizeDelegateTaskArgs_LegacyStringUpgradesToStructuredAsk(t *testing.T) {
	teamID, ask, err := normalizeDelegateTaskArgs(map[string]any{
		"team_id": "admin-core",
		"task":    "inspect gate state",
	})
	if err != nil {
		t.Fatalf("normalize error: %v", err)
	}
	if teamID != "admin-core" {
		t.Fatalf("teamID = %q", teamID)
	}
	if ask.Goal != "inspect gate state" {
		t.Fatalf("goal = %q", ask.Goal)
	}
	if ask.AskKind != protocol.TeamAskKindImplementation {
		t.Fatalf("ask_kind = %q", ask.AskKind)
	}
	if ask.LaneRole != protocol.TeamLaneRoleImplementer {
		t.Fatalf("lane_role = %q", ask.LaneRole)
	}
}

func TestNormalizeDelegateTaskArgs_StructuredAskInput(t *testing.T) {
	teamID, ask, err := normalizeDelegateTaskArgs(map[string]any{
		"team_id": "prime-development",
		"ask": map[string]any{
			"ask_kind":          "validation",
			"lane_role":         "validator",
			"goal":              "Prove the latest branch state with focused runtime checks.",
			"exit_criteria":     []any{"return pass/fail with failing gate list"},
			"evidence_required": []any{"test_output", "summary"},
		},
	})
	if err != nil {
		t.Fatalf("normalize error: %v", err)
	}
	if teamID != "prime-development" {
		t.Fatalf("teamID = %q", teamID)
	}
	raw, marshalErr := json.Marshal(ask)
	if marshalErr != nil {
		t.Fatalf("marshal ask: %v", marshalErr)
	}
	if !strings.Contains(string(raw), `"ask_kind":"validation"`) {
		t.Fatalf("structured ask = %s", string(raw))
	}
	if ask.LaneRole != protocol.TeamLaneRoleValidator {
		t.Fatalf("lane_role = %q", ask.LaneRole)
	}
}

func TestHandleDelegateTask_ResolvesTargetTeamFromAskRoutingHints(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	r.SetSoma(NewTestSoma([]*TeamManifest{
		{
			ID:         "research-team",
			Name:       "Research Team",
			Type:       TeamTypeAction,
			AskRouting: map[string]string{"research": "researcher"},
		},
	}))

	_, err := r.handleDelegateTask(context.Background(), map[string]any{
		"ask": map[string]any{
			"ask_kind":  "research",
			"lane_role": "researcher",
			"goal":      "Find the best supported path.",
		},
	})
	if err == nil {
		t.Fatal("expected NATS unavailable after route resolution")
	}
	if !strings.Contains(err.Error(), "NATS not available") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleDelegateTask_RoutingHintsRejectAmbiguousMatches(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	r.SetSoma(NewTestSoma([]*TeamManifest{
		{
			ID:         "research-a",
			Name:       "Research A",
			Type:       TeamTypeAction,
			AskRouting: map[string]string{"research": "researcher"},
		},
		{
			ID:         "research-b",
			Name:       "Research B",
			Type:       TeamTypeAction,
			AskRouting: map[string]string{"research": "researcher"},
		},
	}))

	_, err := r.handleDelegateTask(context.Background(), map[string]any{
		"ask": map[string]any{
			"ask_kind":  "research",
			"lane_role": "researcher",
			"goal":      "Find the best supported path.",
		},
	})
	if err == nil {
		t.Fatal("expected routing ambiguity error")
	}
	if !strings.Contains(err.Error(), "routing is ambiguous") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleDelegateTask_RoutingHintsRejectNoMatch(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	r.SetSoma(NewTestSoma([]*TeamManifest{
		{
			ID:         "review-team",
			Name:       "Review Team",
			Type:       TeamTypeAction,
			AskRouting: map[string]string{"review": "reviewer"},
		},
	}))

	_, err := r.handleDelegateTask(context.Background(), map[string]any{
		"ask": map[string]any{
			"ask_kind":  "validation",
			"lane_role": "validator",
			"goal":      "Run focused proof.",
		},
	})
	if err == nil {
		t.Fatal("expected no-match routing error")
	}
	if !strings.Contains(err.Error(), "could not resolve a team") {
		t.Fatalf("unexpected error: %v", err)
	}
}
