package swarm

import (
	"context"
	"strings"
	"testing"
)

func TestNormalizeDelegateTaskArgs_ObjectTask(t *testing.T) {
	teamID, task := normalizeDelegateTaskArgs(map[string]any{
		"team_id": "admin-core",
		"task": map[string]any{
			"operation": "generate_blueprint",
			"intent":    "Create a web research team",
		},
	})
	if teamID != "admin-core" {
		t.Fatalf("teamID = %q, want admin-core", teamID)
	}
	if !strings.Contains(task, `"operation":"generate_blueprint"`) {
		t.Fatalf("task payload missing operation: %s", task)
	}
}

func TestNormalizeDelegateTaskArgs_AliasFields(t *testing.T) {
	teamID, task := normalizeDelegateTaskArgs(map[string]any{
		"teamId":    "council-core",
		"operation": "research",
		"intent":    "Find best path",
		"context": map[string]any{
			"topic": "openclaw",
		},
	})
	if teamID != "council-core" {
		t.Fatalf("teamID = %q, want council-core", teamID)
	}
	if !strings.Contains(task, `"intent":"Find best path"`) {
		t.Fatalf("task payload missing intent: %s", task)
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
	if !strings.Contains(err.Error(), "requires 'team_id' and 'task'") {
		t.Fatalf("unexpected error: %v", err)
	}
}
