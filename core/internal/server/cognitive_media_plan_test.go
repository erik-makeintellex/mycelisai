package server

import (
	"strings"
	"testing"
)

func TestBuildPlannedToolCalls_ComicTeamIncludesSpecialistsAndMediaDeliverable(t *testing.T) {
	request := "Create a team to write a comic book page. We need an artist, someone who comes up with characters, and someone who writes the lines. Generate a comic image using local ComfyUI."
	result, ok := deterministicGovernedMutationResult(request, []string{"create_team", "generate_image", "save_cached_image"})
	if !ok || !strings.Contains(result.Text, "groups/soma-requested-team/media") {
		t.Fatalf("deterministic proposal = %#v, ok=%v, want group media target", result, ok)
	}
	calls := buildPlannedToolCalls(chatAgentResult{}, request, []string{"create_team", "generate_image", "save_cached_image"})
	requirePlannedCallNames(t, calls, "create_team", "generate_image", "save_cached_image")
	agents, ok := calls[0].Arguments["agents"].([]map[string]any)
	if !ok || len(agents) < 5 || calls[0].Arguments["staffing_mode"] != "specialist_delivery" {
		t.Fatalf("team args = %#v, agents = %#v", calls[0].Arguments, agents)
	}
	if calls[1].Arguments["size"] != "768x1024" || calls[2].Arguments["folder"] != "groups/soma-requested-team/media" {
		t.Fatalf("media calls = %#v %#v", calls[1], calls[2])
	}
	tools := toolsForPlannedCalls(calls, nil)
	if !containsString(tools, "generate_image") || !containsString(tools, "save_cached_image") {
		t.Fatalf("tools = %#v, want media tools", tools)
	}
}
