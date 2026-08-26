package server

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestBuildPlannedToolCalls_ComicTeamIncludesSpecialistsAndMediaDeliverable(t *testing.T) {
	request := "Create a team to write a comic book page. We need an artist, someone who comes up with characters, and someone who writes the lines. Generate a comic image using local ComfyUI."
	result, ok := deterministicGovernedMutationResult(request, []string{"create_team", "generate_image", "save_cached_image"})
	if !ok || !strings.Contains(result.Text, "/media") {
		t.Fatalf("deterministic proposal = %#v, ok=%v, want group media target", result, ok)
	}
	calls := buildPlannedToolCalls(chatAgentResult{}, request, []string{"create_team", "generate_image", "save_cached_image"})
	requirePlannedCallNames(t, calls, "create_team", "generate_image", "save_cached_image")
	agents, ok := calls[0].Arguments["agents"].([]map[string]any)
	if !ok || len(agents) < 5 || calls[0].Arguments["staffing_mode"] != "specialist_delivery" {
		t.Fatalf("team args = %#v, agents = %#v", calls[0].Arguments, agents)
	}
	name, _ := calls[0].Arguments["name"].(string)
	if name != "Media Generation Team" {
		t.Fatalf("team name = %q, want inferred media team name", name)
	}
	teamID, _ := calls[0].Arguments["team_id"].(string)
	if !strings.HasPrefix(teamID, "media-generation-team-") || len(teamID) != len("media-generation-team-")+5 {
		t.Fatalf("team_id = %q, want readable media team id", teamID)
	}
	if calls[1].Arguments["size"] != "768x1024" || calls[2].Arguments["folder"] != "groups/"+teamID+"/media" {
		t.Fatalf("media calls = %#v %#v", calls[1], calls[2])
	}
	if validation, _ := calls[1].Arguments["validation"].(string); !strings.Contains(validation, "panel composition") {
		t.Fatalf("media validation = %q", validation)
	}
	tools := toolsForPlannedCalls(calls, nil)
	if !containsString(tools, "generate_image") || !containsString(tools, "save_cached_image") {
		t.Fatalf("tools = %#v, want media tools", tools)
	}
}

func TestRequestAsksForMedia_AllowsCodeGeneratedGameVisualsWithoutImageProvider(t *testing.T) {
	requests := []string{
		"Create a runnable browser game with code-generated voxel-style media and no external assets.",
		"Build a browser-native game using code generated graphics and sound cues.",
		"Use generated in code visuals; do not generate an image.",
		"Use code-generated voxel visuals; if image generation is unavailable, keep everything browser-native.",
		"Create a web page and using SVG code imagine and create an image of the Mycelis infrastructure and your place in it.",
		"Draw the architecture as a scalable vector graphic embedded in HTML.",
	}
	for _, request := range requests {
		if requestAsksForMedia(request) {
			t.Fatalf("requestAsksForMedia(%q) = true, want false for self-contained browser output", request)
		}
	}
}

func TestRequestAsksForMedia_StillRoutesExplicitImageRequests(t *testing.T) {
	requests := []string{
		"Generate an image of a voxel game hero.",
		"Create artwork for the marketing card.",
		"Make a comic book page using local ComfyUI.",
	}
	for _, request := range requests {
		if !requestAsksForMedia(request) {
			t.Fatalf("requestAsksForMedia(%q) = false, want true for explicit media generation", request)
		}
	}
}

func TestFilterOptionalMediaGenerationPlan_DropsProviderCallsForCodeGeneratedOutput(t *testing.T) {
	planned := []protocol.PlannedToolCall{
		{Name: "create_team"},
		{Name: "generate_image"},
		{Name: "save_cached_image"},
		{Name: "write_file"},
	}
	filtered := filterOptionalMediaGenerationPlan(
		"Create a runnable browser game with code-generated voxel visuals, no external assets, and browser-native audio.",
		planned,
	)
	requirePlannedCallNames(t, filtered, "create_team", "write_file")
}

func TestFilterOptionalMediaGenerationPlan_KeepsProviderCallsForExplicitImages(t *testing.T) {
	planned := []protocol.PlannedToolCall{
		{Name: "create_team"},
		{Name: "generate_image"},
		{Name: "save_cached_image"},
	}
	filtered := filterOptionalMediaGenerationPlan("Generate an image of a voxel game hero.", planned)
	requirePlannedCallNames(t, filtered, "create_team", "generate_image", "save_cached_image")
}
