package server

import (
	"regexp"
	"strings"
	"testing"
)

func TestBuildPlannedToolCalls_GameImprovementNotifiesMarketingWithEvidence(t *testing.T) {
	request := strings.Join([]string{
		"We have a game delivery team and need a marketing team.",
		"Improve the game, let the marketing team know what changed,",
		"generate marketing about the game, and show gameplay examples of the changes.",
	}, " ")
	result, ok := deterministicGovernedMutationResult(request, []string{"create_team", "write_file", "delegate_task"})
	if !ok {
		t.Fatal("expected deterministic governed mutation result")
	}
	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)

	requirePlannedCallNames(t, calls, "create_team", "delegate_task", "write_file", "write_file", "delegate_task")
	marketingTeamID, _ := calls[0].Arguments["team_id"].(string)
	if !regexp.MustCompile(`^marketing-delivery-team-[0-9a-f]{5}$`).MatchString(marketingTeamID) {
		t.Fatalf("marketing team_id = %q, want readable stable marketing id", marketingTeamID)
	}
	if calls[1].Arguments["team_id"] != "game-delivery-team" {
		t.Fatalf("game delegate team_id = %#v", calls[1].Arguments["team_id"])
	}
	if calls[2].Arguments["path"] != "groups/game-delivery-team/proof/GAMEPLAY_CHANGE_EXAMPLES.md" {
		t.Fatalf("evidence path = %#v", calls[2].Arguments["path"])
	}
	if calls[3].Arguments["path"] != "groups/"+marketingTeamID+"/marketing/MARKETING_HANDOFF.md" {
		t.Fatalf("marketing handoff path = %#v", calls[3].Arguments["path"])
	}
	content, _ := calls[2].Arguments["content"].(string)
	for _, want := range []string{"Gameplay Change Examples", "Marketing-safe claim boundary", "Direct launch path"} {
		if !strings.Contains(content, want) {
			t.Fatalf("game evidence content missing %q: %.300q", want, content)
		}
	}
	handoff, _ := calls[3].Arguments["content"].(string)
	for _, want := range []string{"Marketing Handoff", "game-delivery-team", marketingTeamID, "Evidence examples"} {
		if !strings.Contains(handoff, want) {
			t.Fatalf("marketing handoff missing %q: %.300q", want, handoff)
		}
	}
}

func TestBuildPlannedToolCalls_AppMarketingUsesGenericEvidencePattern(t *testing.T) {
	request := strings.Join([]string{
		"Update the customer portal app output at groups/app-builder-abc/generated/portal.",
		"Notify marketing team_id launch-marketing-team and generate marketing from the changed app.",
		"Show usage examples proving what changed.",
	}, " ")
	result, ok := deterministicGovernedMutationResult(request, []string{"create_team", "write_file", "delegate_task"})
	if !ok {
		t.Fatal("expected deterministic governed mutation result")
	}
	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)

	requirePlannedCallNames(t, calls, "create_team", "delegate_task", "write_file", "write_file", "delegate_task")
	if calls[0].Arguments["team_id"] != "launch-marketing-team" {
		t.Fatalf("marketing team_id = %#v", calls[0].Arguments["team_id"])
	}
	if calls[1].Arguments["team_id"] != "app-builder-abc" {
		t.Fatalf("source team_id = %#v", calls[1].Arguments["team_id"])
	}
	if calls[2].Arguments["path"] != "groups/app-builder-abc/proof/USAGE_CHANGE_EXAMPLES.md" {
		t.Fatalf("generic evidence path = %#v", calls[2].Arguments["path"])
	}
	content, _ := calls[2].Arguments["content"].(string)
	if !strings.Contains(content, "Open/view path showing the changed deliverable") {
		t.Fatalf("generic evidence content = %.300q", content)
	}
}
