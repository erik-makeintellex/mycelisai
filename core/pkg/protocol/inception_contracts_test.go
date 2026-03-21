package protocol

import "testing"

func TestDefaultInceptionContractBundle(t *testing.T) {
	bundle := DefaultInceptionContractBundle()

	if len(bundle.AllowedPaths) != 4 {
		t.Fatalf("allowed_paths len = %d, want 4", len(bundle.AllowedPaths))
	}
	if bundle.AllowedPaths[0] != DecisionPathDirect {
		t.Fatalf("allowed_paths[0] = %q, want %q", bundle.AllowedPaths[0], DecisionPathDirect)
	}
	if bundle.AllowedPaths[1] != DecisionPathManifestTeam {
		t.Fatalf("allowed_paths[1] = %q, want %q", bundle.AllowedPaths[1], DecisionPathManifestTeam)
	}
	if bundle.AllowedPaths[2] != DecisionPathPropose {
		t.Fatalf("allowed_paths[2] = %q, want %q", bundle.AllowedPaths[2], DecisionPathPropose)
	}
	if bundle.AllowedPaths[3] != DecisionPathScheduledRepeat {
		t.Fatalf("allowed_paths[3] = %q, want %q", bundle.AllowedPaths[3], DecisionPathScheduledRepeat)
	}

	if len(bundle.AllowedLifetimes) != 3 {
		t.Fatalf("allowed_lifetimes len = %d, want 3", len(bundle.AllowedLifetimes))
	}
	if bundle.AllowedLifetimes[0] != TeamLifetimeEphemeral {
		t.Fatalf("allowed_lifetimes[0] = %q, want %q", bundle.AllowedLifetimes[0], TeamLifetimeEphemeral)
	}
	if bundle.AllowedLifetimes[1] != TeamLifetimePersistent {
		t.Fatalf("allowed_lifetimes[1] = %q, want %q", bundle.AllowedLifetimes[1], TeamLifetimePersistent)
	}
	if bundle.AllowedLifetimes[2] != TeamLifetimeAuto {
		t.Fatalf("allowed_lifetimes[2] = %q, want %q", bundle.AllowedLifetimes[2], TeamLifetimeAuto)
	}

	if bundle.DecisionFrame.PathSelected != DecisionPathDirect {
		t.Fatalf("path_selected = %q, want %q", bundle.DecisionFrame.PathSelected, DecisionPathDirect)
	}
	if bundle.DecisionFrame.TeamLifetime != TeamLifetimeEphemeral {
		t.Fatalf("team_lifetime = %q, want %q", bundle.DecisionFrame.TeamLifetime, TeamLifetimeEphemeral)
	}
	if bundle.UniversalInvoke.ProviderType == "" {
		t.Fatal("universal_invoke.provider_type should be non-empty")
	}
	if bundle.UniversalInvoke.Status == "" {
		t.Fatal("universal_invoke.status should be non-empty")
	}
	if bundle.HeartbeatBudget.MaxDurationMinutes <= 0 {
		t.Fatal("heartbeat_budget.max_duration_minutes should be > 0")
	}
	if bundle.HeartbeatBudget.MaxActions <= 0 {
		t.Fatal("heartbeat_budget.max_actions should be > 0")
	}
}
