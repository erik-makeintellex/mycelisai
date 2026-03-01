package swarm

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestApplyProviderOverrides(t *testing.T) {
	s := &Soma{}
	s.SetProviderOverrides(
		map[string]string{"council-core": "team-provider"},
		map[string]string{"council-architect": "agent-provider"},
	)

	manifest := &TeamManifest{
		ID:   "council-core",
		Name: "Council Core",
		Members: []protocol.AgentManifest{
			{ID: "council-architect", Role: "architect"},
			{ID: "council-sentry", Role: "sentry"},
		},
	}

	out := s.applyProviderOverrides(manifest)
	if out.Provider != "team-provider" {
		t.Fatalf("team provider = %q", out.Provider)
	}
	if out.Members[0].Provider != "agent-provider" {
		t.Fatalf("expected agent override for architect, got %q", out.Members[0].Provider)
	}
	if out.Members[1].Provider != "team-provider" {
		t.Fatalf("expected team fallback for sentry, got %q", out.Members[1].Provider)
	}
}
