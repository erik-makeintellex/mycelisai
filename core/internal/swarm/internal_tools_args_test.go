package swarm

import (
	"context"
	"strings"
	"testing"
)

func TestHandleReadSignals_AcceptsTopicPatternAlias(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	_, err := r.handleReadSignals(context.Background(), map[string]any{
		"topic_pattern": "swarm.team.admin-core.signal.status",
	})
	if err == nil {
		t.Fatal("expected error because NATS is unavailable in test")
	}
	if strings.Contains(err.Error(), "requires 'subject'") {
		t.Fatalf("expected alias normalization to provide subject, got: %v", err)
	}
}

func TestHandleConsultCouncil_AcceptsAliases(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	_, err := r.handleConsultCouncil(context.Background(), map[string]any{
		"agent":  "Architect",
		"query":  "Assess API approach",
		"unused": "x",
	})
	if err == nil {
		t.Fatal("expected error because NATS is unavailable in test")
	}
	if strings.Contains(err.Error(), "requires 'member' and 'question'") {
		t.Fatalf("expected alias normalization for member/question, got: %v", err)
	}
}

func TestHandleCreateTeam_RequiresSoma(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	_, err := r.handleCreateTeam(context.Background(), map[string]any{"team_id": "x"})
	if err == nil {
		t.Fatal("expected soma unavailable error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "soma") {
		t.Fatalf("unexpected error: %v", err)
	}
}
