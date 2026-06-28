package server

import (
	"strings"
	"testing"
)

func TestBuildPlannedToolCalls_TeamWatchRegisterUsesRequestedMarkdownTarget(t *testing.T) {
	request := strings.Join([]string{
		"Create a consistent standing team with team_id standing-app-steward-qa named App Steward Watch Team.",
		"Have that team watch the generated app content at groups/temp-app-builder-qa/generated/first-game.",
		"Create a retained watch register markdown file at groups/standing-app-steward-qa/watch/CONTENT_WATCH.md.",
		"The register should name the watched folder, the build team, expected transaction signals, and how the steward should react.",
		"Return retained output and proof.",
	}, " ")
	calls := plannedCallsFromWrongBlueprint(request, []string{"write_file", "generate_blueprint", "delegate"})

	requirePlannedCallNames(t, calls, "create_team", "write_file")
	if calls[1].Arguments["path"] != "groups/standing-app-steward-qa/watch/CONTENT_WATCH.md" {
		t.Fatalf("path = %#v", calls[1].Arguments["path"])
	}
	if _, hasPackageKind := calls[1].Arguments["package_kind"]; hasPackageKind {
		t.Fatalf("watch register should not become a project package: %#v", calls[1].Arguments)
	}
	content, _ := calls[1].Arguments["content"].(string)
	if !strings.Contains(content, "groups/temp-app-builder-qa/generated/first-game") {
		t.Fatalf("content = %q, want watched folder retained", content)
	}
}

func TestBuildPlannedToolCalls_ReactionFilePrefersLastWriteTarget(t *testing.T) {
	request := strings.Join([]string{
		"Use the existing standing team standing-app-steward-qa to react to the transaction output at groups/temp-content-transaction-qa/generated/transaction/TRANSACTION_REPORT.md.",
		"Create a retained steward reaction file at groups/standing-app-steward-qa/watch/STEWARD_REACTION.md.",
		"The reaction should say what remains trusted and what needs review.",
	}, " ")
	result, ok := deterministicGovernedMutationResult(request, []string{"write_file"})
	if !ok {
		t.Fatal("expected deterministic governed mutation result")
	}
	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)

	requirePlannedCallNames(t, calls, "write_file")
	if calls[0].Arguments["path"] != "groups/standing-app-steward-qa/watch/STEWARD_REACTION.md" {
		t.Fatalf("path = %#v", calls[0].Arguments["path"])
	}
	content, _ := calls[0].Arguments["content"].(string)
	if !strings.Contains(content, "TRANSACTION_REPORT.md") {
		t.Fatalf("content = %q, want source transaction retained", content)
	}
}

func TestBuildPlannedToolCalls_ReactionFileKeepsTargetWhenSourceReportComesLater(t *testing.T) {
	request := strings.Join([]string{
		"Create a retained markdown file at groups/standing-game-steward-qa/watch/STEWARD_REACTION.md.",
		"Review generated game folder groups/temp-game-builder-qa/generated/first-game and playtest report workspace/logs/game-proof/PLAYTEST_REPORT.json.",
		"Write a steward reaction with one next requested change.",
	}, " ")
	result, ok := deterministicGovernedMutationResult(request, []string{"write_file"})
	if !ok {
		t.Fatal("expected deterministic governed mutation result")
	}
	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)

	requirePlannedCallNames(t, calls, "write_file")
	if calls[0].Arguments["path"] != "groups/standing-game-steward-qa/watch/STEWARD_REACTION.md" {
		t.Fatalf("path = %#v", calls[0].Arguments["path"])
	}
	content, _ := calls[0].Arguments["content"].(string)
	if !strings.Contains(content, "PLAYTEST_REPORT.json") {
		t.Fatalf("content = %q, want source playtest report retained", content)
	}
}
