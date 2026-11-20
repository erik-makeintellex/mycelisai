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

func TestBuildRuntimeTeamManifest_GuaranteesIdealReviewerMember(t *testing.T) {
	withoutReviewer := buildRuntimeTeamManifest(map[string]any{
		"team_id": "arcade-builders",
		"members": map[string]any{
			"architect": "council-architect",
			"developer": "prime-development-agent",
		},
	})
	if withoutReviewer == nil {
		t.Fatal("expected runtime manifest")
	}
	if len(withoutReviewer.Members) != 3 {
		t.Fatalf("members = %+v, want 3 (including ideal reviewer)", withoutReviewer.Members)
	}
	last := withoutReviewer.Members[len(withoutReviewer.Members)-1]
	if last.Role != "delivery_reviewer" || last.ID != "arcade-builders-delivery-reviewer" {
		t.Fatalf("expected auto-injected delivery_reviewer, got %+v", last)
	}

	withReviewer := buildRuntimeTeamManifest(map[string]any{
		"team_id": "governed-pipeline",
		"members": map[string]any{
			"lead-integrator": "prime-development-agent",
			"output-verifier": "council-sentry",
		},
	})
	if withReviewer == nil {
		t.Fatal("expected runtime manifest")
	}
	if len(withReviewer.Members) != 2 {
		t.Fatalf("members = %+v, want 2 without duplicate reviewer", withReviewer.Members)
	}
}
