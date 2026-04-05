package swarm

import "testing"

func TestBuildRuntimeTeamManifest_DefaultsAskRoutingHints(t *testing.T) {
	manifest := buildRuntimeTeamManifest(map[string]any{
		"team_id": "research-team",
		"role":    "researcher",
	})
	if manifest == nil {
		t.Fatal("expected runtime manifest")
	}
	if manifest.AskRouting["research"] != "researcher" {
		t.Fatalf("research ask routing = %q", manifest.AskRouting["research"])
	}
	if manifest.AskRouting["implementation"] != "implementer" {
		t.Fatalf("implementation ask routing = %q", manifest.AskRouting["implementation"])
	}
}

func TestBuildRuntimeTeamManifest_PreservesExplicitAskRoutingHints(t *testing.T) {
	manifest := buildRuntimeTeamManifest(map[string]any{
		"team_id": "review-team",
		"ask_routing": map[string]any{
			"review":       "reviewer",
			"coordination": "coordinator",
		},
	})
	if manifest == nil {
		t.Fatal("expected runtime manifest")
	}
	if len(manifest.AskRouting) != 2 {
		t.Fatalf("ask routing = %#v", manifest.AskRouting)
	}
	if manifest.AskRouting["review"] != "reviewer" {
		t.Fatalf("review ask routing = %q", manifest.AskRouting["review"])
	}
	if manifest.AskRouting["coordination"] != "coordinator" {
		t.Fatalf("coordination ask routing = %q", manifest.AskRouting["coordination"])
	}
}
