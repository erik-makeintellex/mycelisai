package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestDeliverableResultOutputIssue_RejectsGenericFileForPackageOutcome(t *testing.T) {
	item := protocol.TeamWorkItem{
		TeamID:          "game-team",
		ExecutionShape:  protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{"openable browser game package"},
	}
	refs := []protocol.TeamOutputRef{{
		Kind:       "file",
		StorageRef: "generated/how-do-you-think-you-could-improve-it.py",
	}}
	if got := deliverableResultOutputIssue(item, protocol.PayloadKindResult, refs); got != "invalid_deliverable_shape" {
		t.Fatalf("issue = %q, want invalid_deliverable_shape", got)
	}
}

func TestDeliverableResultOutputIssue_AcceptsIsolatedOpenablePackage(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	folder := filepath.Join(workspace, "groups", "game-team", "generated", "first-game")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "index.html"), []byte(`<p>Click Play to score.</p><button id="play">Play</button><script src="game.js"></script>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "game.js"), []byte(`document.querySelector("#play").addEventListener("click", () => { document.body.dataset.ready = "true"; });`), 0o644); err != nil {
		t.Fatal(err)
	}
	item := protocol.TeamWorkItem{
		TeamID:          "game-team",
		ExecutionShape:  protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{"openable browser game package"},
	}
	refs := []protocol.TeamOutputRef{{
		Kind:       "project_package",
		StorageRef: "groups/game-team/generated/first-game",
		Entrypoint: "index.html",
	}}
	if got := deliverableResultOutputIssue(item, protocol.PayloadKindResult, refs); got != "" {
		t.Fatalf("issue = %q, want accepted package", got)
	}
}

func TestDeliverableResultOutputIssueRejectsPackageWithoutUsablePrimaryInteraction(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	folder := filepath.Join(workspace, "groups", "game-team", "generated", "first-game")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "index.html"), []byte(`<div class="target"></div><p>Score: 0</p><button>Restart</button>`), 0o644); err != nil {
		t.Fatal(err)
	}
	item := protocol.TeamWorkItem{
		TeamID:          "game-team",
		Objective:       "Develop a playable browser application with controls.",
		ExecutionShape:  protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{"openable application package"},
	}
	refs := []protocol.TeamOutputRef{{
		Kind:       "project_package",
		StorageRef: "groups/game-team/generated/first-game",
		Entrypoint: "index.html",
	}}
	if got := deliverableResultOutputIssue(item, protocol.PayloadKindResult, refs); got != "unverified_primary_interaction" {
		t.Fatalf("issue = %q, want unverified_primary_interaction", got)
	}
}

func TestDeliverableResultOutputIssueRejectsMissingLocalPackageAsset(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	folder := filepath.Join(workspace, "groups", "game-team", "generated", "first-game")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "index.html"), []byte(`<script src="missing.js"></script>`), 0o644); err != nil {
		t.Fatal(err)
	}
	item := protocol.TeamWorkItem{
		TeamID:          "game-team",
		ExecutionShape:  protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{"openable browser game package"},
	}
	refs := []protocol.TeamOutputRef{{
		Kind:       "project_package",
		StorageRef: "groups/game-team/generated/first-game",
		Entrypoint: "index.html",
	}}
	if got := deliverableResultOutputIssue(item, protocol.PayloadKindResult, refs); got != "incomplete_deliverable_files" {
		t.Fatalf("issue = %q, want incomplete_deliverable_files", got)
	}
}
