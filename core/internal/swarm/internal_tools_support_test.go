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
	if len(manifest.Members) != 1 {
		t.Fatalf("runtime team members = %d, want lead-only start", len(manifest.Members))
	}
	if manifest.Description == "" {
		t.Fatal("expected lead-only team description")
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

func TestBuildRuntimeTeamManifest_PreservesSpecialistRoster(t *testing.T) {
	manifest := buildRuntimeTeamManifest(map[string]any{
		"team_id": "comic-page-cell",
		"agents": []any{
			map[string]any{"id": "comic-page-cell-lead", "role": "creative lead", "tools": []any{"store_artifact"}},
			map[string]any{"id": "comic-page-cell-artist", "role": "panel layout artist", "tools": []any{"generate_image", "save_cached_image"}},
			map[string]any{"id": "comic-page-cell-proof", "role": "proof editor"},
		},
	})
	if manifest == nil {
		t.Fatal("expected runtime manifest")
	}
	if len(manifest.Members) != 3 {
		t.Fatalf("runtime team members = %d, want specialist roster", len(manifest.Members))
	}
	if manifest.Description == "" || manifest.Description == "Runtime-created lead-only team; expand only with operator action or justified temporary specialist request." {
		t.Fatalf("description = %q, want specialist delivery description", manifest.Description)
	}
	if manifest.Members[1].Role != "panel layout artist" {
		t.Fatalf("member role = %q, want panel layout artist", manifest.Members[1].Role)
	}
	if len(manifest.Members[1].Tools) != 2 || manifest.Members[1].Tools[0] != "generate_image" {
		t.Fatalf("artist tools = %#v, want generated media tools", manifest.Members[1].Tools)
	}
	if len(manifest.Members[2].Tools) == 0 || manifest.Members[2].Tools[0] != "store_artifact" {
		t.Fatalf("fallback tools = %#v, want store_artifact fallback", manifest.Members[2].Tools)
	}
}
