package swarm

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestProviderPolicyResolveManifestOrganizationDefaultApplies(t *testing.T) {
	manifest := &TeamManifest{
		ID:   "prime-development",
		Name: "Prime Development",
		Members: []protocol.AgentManifest{
			{ID: "prime-development-agent", Role: "coder"},
		},
	}

	resolved, blocked := (ProviderPolicy{Provider: "org-provider"}).ResolveManifest(manifest)
	if len(blocked) != 0 {
		t.Fatalf("expected no blocked overrides, got %+v", blocked)
	}
	if resolved.Provider != "org-provider" {
		t.Fatalf("team provider = %q", resolved.Provider)
	}
	if resolved.Members[0].Provider != "org-provider" {
		t.Fatalf("member provider = %q", resolved.Members[0].Provider)
	}
}

func TestProviderPolicyResolveManifestKernelAndCouncilRoleDefaultsApply(t *testing.T) {
	policy := ProviderPolicy{
		Provider: "org-provider",
		Kernel: ProviderScope{
			AllowedProviders: []string{"org-provider", "kernel-provider"},
			RoleProviders:    map[string]string{"admin": "kernel-provider"},
		},
		Council: ProviderScope{
			AllowedProviders: []string{"org-provider", "council-provider"},
			RoleProviders:    map[string]string{"architect": "council-provider"},
		},
	}

	adminManifest := &TeamManifest{
		ID:   "admin-core",
		Name: "Admin",
		Members: []protocol.AgentManifest{
			{ID: "admin", Role: "admin"},
		},
	}
	adminResolved, blocked := policy.ResolveManifest(adminManifest)
	if len(blocked) != 0 {
		t.Fatalf("expected no blocked admin overrides, got %+v", blocked)
	}
	if adminResolved.Members[0].Provider != "kernel-provider" {
		t.Fatalf("admin provider = %q", adminResolved.Members[0].Provider)
	}

	councilManifest := &TeamManifest{
		ID:   "council-core",
		Name: "Council",
		Members: []protocol.AgentManifest{
			{ID: "council-architect", Role: "architect"},
			{ID: "council-coder", Role: "coder"},
		},
	}
	councilResolved, blocked := policy.ResolveManifest(councilManifest)
	if len(blocked) != 0 {
		t.Fatalf("expected no blocked council overrides, got %+v", blocked)
	}
	if councilResolved.Members[0].Provider != "council-provider" {
		t.Fatalf("architect provider = %q", councilResolved.Members[0].Provider)
	}
	if councilResolved.Members[1].Provider != "org-provider" {
		t.Fatalf("coder provider = %q", councilResolved.Members[1].Provider)
	}
}

func TestProviderPolicyResolveManifestTeamOverrideApplies(t *testing.T) {
	policy := ProviderPolicy{
		Provider:         "org-provider",
		AllowedProviders: []string{"org-provider", "team-provider"},
		Teams: map[string]ProviderScope{
			"prime-development": {Provider: "team-provider"},
		},
	}
	manifest := &TeamManifest{
		ID:   "prime-development",
		Name: "Prime Development",
		Members: []protocol.AgentManifest{
			{ID: "prime-development-agent", Role: "coder"},
		},
	}

	resolved, blocked := policy.ResolveManifest(manifest)
	if len(blocked) != 0 {
		t.Fatalf("expected no blocked overrides, got %+v", blocked)
	}
	if resolved.Provider != "team-provider" {
		t.Fatalf("team provider = %q", resolved.Provider)
	}
	if resolved.Members[0].Provider != "team-provider" {
		t.Fatalf("member provider = %q", resolved.Members[0].Provider)
	}
}

func TestProviderPolicyResolveManifestAgentOverrideApplies(t *testing.T) {
	policy := ProviderPolicy{
		Provider:         "org-provider",
		AllowedProviders: []string{"org-provider", "agent-provider"},
		Agents: map[string]ProviderScope{
			"council-architect": {Provider: "agent-provider"},
		},
	}
	manifest := &TeamManifest{
		ID:   "council-core",
		Name: "Council",
		Members: []protocol.AgentManifest{
			{ID: "council-architect", Role: "architect"},
			{ID: "council-sentry", Role: "sentry"},
		},
	}

	resolved, blocked := policy.ResolveManifest(manifest)
	if len(blocked) != 0 {
		t.Fatalf("expected no blocked overrides, got %+v", blocked)
	}
	if resolved.Members[0].Provider != "agent-provider" {
		t.Fatalf("architect provider = %q", resolved.Members[0].Provider)
	}
	if resolved.Members[1].Provider != "org-provider" {
		t.Fatalf("sentry provider = %q", resolved.Members[1].Provider)
	}
}

func TestProviderPolicyResolveManifestBlockedOverridesAreIgnored(t *testing.T) {
	policy := ProviderPolicy{
		Provider:         "org-provider",
		AllowedProviders: []string{"org-provider"},
		Teams: map[string]ProviderScope{
			"council-core": {Provider: "team-provider"},
		},
		Agents: map[string]ProviderScope{
			"council-architect": {Provider: "agent-provider"},
		},
	}
	manifest := &TeamManifest{
		ID:       "council-core",
		Name:     "Council",
		Provider: "manifest-team-provider",
		Members: []protocol.AgentManifest{
			{ID: "council-architect", Role: "architect", Provider: "manifest-agent-provider"},
		},
	}

	resolved, blocked := policy.ResolveManifest(manifest)
	if resolved.Provider != "org-provider" {
		t.Fatalf("team provider = %q", resolved.Provider)
	}
	if resolved.Members[0].Provider != "org-provider" {
		t.Fatalf("agent provider = %q", resolved.Members[0].Provider)
	}
	if len(blocked) != 6 {
		t.Fatalf("expected 6 blocked overrides, got %+v", blocked)
	}
}
