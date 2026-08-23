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

func TestDeliverableResultOutputIssue_AcceptsGenericInteractiveApplicationPackage(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	folder := filepath.Join(workspace, "groups", "app-team", "generated", "browser-app")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `<p>Use Run to update the status.</p><button data-mycelis-primary-action onclick="status.textContent='Changed'">Run</button><main data-mycelis-validation-surface id="status">Ready</main>`
	if err := os.WriteFile(filepath.Join(folder, "index.html"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	item := protocol.TeamWorkItem{
		TeamID:         "app-team",
		ExecutionShape: protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{
			"openable application package",
		},
		WorkIntent: &protocol.WorkIntent{OutputContract: &protocol.WorkOutputContract{
			Shape: "app_package",
			OutputValidation: &protocol.OutputValidationPlan{
				Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
			},
		}},
	}
	refs := []protocol.TeamOutputRef{{
		Kind:       "project_package",
		StorageRef: "groups/app-team/generated/browser-app",
		Entrypoint: "index.html",
	}}
	if got := deliverableResultOutputIssue(item, protocol.PayloadKindResult, refs); got != "" {
		t.Fatalf("issue = %q, want accepted interactive package", got)
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

func TestDeliverableResultOutputIssueRejectsDormantInteractivePackage(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	folder := filepath.Join(workspace, "groups", "game-team", "generated", "first-game")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `<!doctype html>
<canvas width="320" height="180"></canvas>
<p>Use arrow keys to move.</p>
<button data-mycelis-primary-action>Start</button>
<main data-mycelis-validation-surface>Ready</main>
<script>
const surface = document.querySelector("[data-mycelis-validation-surface]");
document.querySelector("button").addEventListener("click", () => { surface.textContent = "Started"; });
function update() {
  requestAnimationFrame(update);
}
</script>`
	if err := os.WriteFile(filepath.Join(folder, "index.html"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	item := protocol.TeamWorkItem{
		TeamID:          "game-team",
		Objective:       "Develop a playable browser application with controls.",
		ExecutionShape:  protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{"openable browser game package"},
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

func TestDeliverableResultOutputIssueRejectsInertInteractiveHandler(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	folder := filepath.Join(workspace, "groups", "app-team", "generated", "browser-app")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `<p>Click Run to update the status.</p><button data-mycelis-primary-action onclick="console.log('run')">Run</button><main data-mycelis-validation-surface>Ready</main>`
	if err := os.WriteFile(filepath.Join(folder, "index.html"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	item := protocol.TeamWorkItem{
		TeamID:          "app-team",
		Objective:       "Build an interactive browser application.",
		ExecutionShape:  protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{"openable application package"},
	}
	refs := []protocol.TeamOutputRef{{
		Kind:       "project_package",
		StorageRef: "groups/app-team/generated/browser-app",
		Entrypoint: "index.html",
	}}
	if got := deliverableResultOutputIssue(item, protocol.PayloadKindResult, refs); got != "unverified_primary_interaction" {
		t.Fatalf("issue = %q, want unverified_primary_interaction", got)
	}
}

func TestDeliverableResultOutputIssueRejectsDormantInteractiveRenderLoop(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	folder := filepath.Join(workspace, "groups", "app-team", "generated", "browser-app")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `<p>Use ArrowRight to move.</p><canvas data-mycelis-validation-surface></canvas><script>
document.addEventListener('keydown', () => { document.body.dataset.moved = 'true'; });
function renderLoop() { requestAnimationFrame(renderLoop); }
</script>`
	if err := os.WriteFile(filepath.Join(folder, "index.html"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	item := protocol.TeamWorkItem{
		TeamID:          "app-team",
		Objective:       "Build an interactive browser application.",
		ExecutionShape:  protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{"openable application package"},
	}
	refs := []protocol.TeamOutputRef{{
		Kind:       "project_package",
		StorageRef: "groups/app-team/generated/browser-app",
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
