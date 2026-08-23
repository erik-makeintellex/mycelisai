package swarm

import (
	"fmt"
	"slices"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

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
	wantInput := fmt.Sprintf(protocol.TopicTeamInternalCommand, "research-team")
	if len(manifest.Inputs) != 1 || manifest.Inputs[0] != wantInput {
		t.Fatalf("runtime team inputs = %#v, want %q", manifest.Inputs, wantInput)
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

func TestBuildRuntimeTeamManifest_UsesApprovedCapabilitiesAsWorkerTools(t *testing.T) {
	manifest := buildRuntimeTeamManifest(map[string]any{
		"team_id": "application-delivery-team",
		"required_capabilities": []any{
			"team_orchestration",
			"write_file",
			"store_artifact",
			"read_file",
			"local_command",
		},
	})
	if manifest == nil || len(manifest.Members) != 1 {
		t.Fatalf("manifest = %#v, want one runtime worker", manifest)
	}
	for _, want := range []string{"write_file", "store_artifact", "read_file", "local_command"} {
		if !slices.Contains(manifest.Members[0].Tools, want) {
			t.Fatalf("worker tools = %#v, want %q from approved capabilities", manifest.Members[0].Tools, want)
		}
	}
	if slices.Contains(manifest.Members[0].Tools, "team_orchestration") {
		t.Fatalf("worker tools = %#v, generic capability is not an executable tool", manifest.Members[0].Tools)
	}
}

func TestBuildRuntimeTeamManifest_TreatsWorkerInputsAsMetadataNotBusSubjects(t *testing.T) {
	manifest := buildRuntimeTeamManifest(map[string]any{
		"team_id": "voxel-game-team",
		"inputs":  []any{"game_brief", "style_reference", "target_platform"},
		"outputs": []any{"project_package", "playtest_notes"},
	})
	if manifest == nil || len(manifest.Members) != 1 {
		t.Fatalf("manifest = %#v, want one runtime worker", manifest)
	}
	wantInput := fmt.Sprintf(protocol.TopicTeamInternalCommand, "voxel-game-team")
	if len(manifest.Inputs) != 1 || manifest.Inputs[0] != wantInput {
		t.Fatalf("team bus inputs = %#v, want only %q", manifest.Inputs, wantInput)
	}
	if !slices.Equal(manifest.Members[0].Inputs, []string{"game_brief", "style_reference", "target_platform"}) {
		t.Fatalf("worker input metadata = %#v", manifest.Members[0].Inputs)
	}
	if !slices.Equal(manifest.Members[0].Outputs, []string{"project_package", "playtest_notes"}) {
		t.Fatalf("worker output metadata = %#v", manifest.Members[0].Outputs)
	}
}

func TestRuntimeTeamInputSubjectsKeepsExplicitNATSSubjects(t *testing.T) {
	got := runtimeTeamInputSubjects("research-team", []string{
		"research_question",
		"swarm.global.input.research",
		"swarm.team.research-team.internal.command",
	})
	want := []string{
		fmt.Sprintf(protocol.TopicTeamInternalCommand, "research-team"),
		"swarm.global.input.research",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("runtime team input subjects = %#v, want %#v", got, want)
	}
}
