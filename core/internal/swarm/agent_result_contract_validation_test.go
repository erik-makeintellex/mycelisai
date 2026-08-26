package swarm

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestProjectPackageValidationRequiresEntrypointReadback(t *testing.T) {
	artifacts := reconcileToolBackedArtifacts(nil, []successfulToolEvidence{
		{ToolName: "write_file", Path: "groups/team/generated/app/index.html"},
		{ToolName: "write_file", Path: "groups/team/generated/app/README.md"},
		{ToolName: "read_file", Path: "groups/team/generated/app/README.md"},
	}, "Create a project package.")
	requirement := &teamResultRequirement{Kind: "project_package", EntrypointRequired: true, ReadbackRequired: true}

	issues := resultContractIssues(requirement, artifacts, []successfulToolEvidence{
		{ToolName: "write_file", Path: "groups/team/generated/app/index.html"},
		{ToolName: "write_file", Path: "groups/team/generated/app/README.md"},
		{ToolName: "read_file", Path: "groups/team/generated/app/README.md"},
	})
	if !strings.Contains(strings.Join(issues, ";"), "missing successful structural readback") {
		t.Fatalf("issues = %v, want missing entrypoint readback", issues)
	}
}

func TestInteractivePackageReadbackRequiresHandlerAndVisibleInstructions(t *testing.T) {
	requirement := &teamResultRequirement{
		Kind: "project_package", ExpectedOutputs: []string{"playable browser game package"},
		AcceptanceCriteria: []string{"primary control changes the application"}, ReadbackRequired: true,
	}
	entrypoint := "groups/team/generated/app/index.html"
	withoutInstructionsContent := `<canvas></canvas><script>document.addEventListener('keydown', move)</script>`
	withoutInstructions := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: withoutInstructionsContent},
		{ToolName: "read_file", Path: entrypoint, Content: withoutInstructionsContent},
	}
	artifacts := reconcileToolBackedArtifacts(nil, withoutInstructions, "Create a playable project package.")
	issues := strings.Join(resultContractIssues(requirement, artifacts, withoutInstructions), ";")
	if !strings.Contains(issues, "visible control instructions") {
		t.Fatalf("issues = %q, want visible-control repair", issues)
	}

	withInstructionsContent := `<p>Use ArrowRight to move.</p><canvas></canvas><script>document.addEventListener('keydown', move); function move(){ document.body.dataset.moved = 'true'; }</script>`
	withInstructions := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: withInstructionsContent},
		{ToolName: "read_file", Path: entrypoint, Content: withInstructionsContent},
	}
	artifacts = reconcileToolBackedArtifacts(nil, withInstructions, "Create a playable project package.")
	if issues := resultContractIssues(requirement, artifacts, withInstructions); len(issues) != 0 {
		t.Fatalf("valid interactive evidence issues = %v", issues)
	}
}

func TestInteractivePackageAcceptsFamiliarVisibleControlLabel(t *testing.T) {
	requirement := &teamResultRequirement{
		Kind: "project_package", ExpectedOutputs: []string{"interactive browser app"},
		AcceptanceCriteria: []string{"primary control changes the application"}, ReadbackRequired: true,
	}
	entrypoint := "groups/team/generated/app/index.html"
	content := `<button data-mycelis-primary-action onclick="document.getElementById('status').textContent='restarted'">Restart</button><p data-mycelis-validation-surface id="status">Ready</p>`
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: content},
		{ToolName: "read_file", Path: entrypoint, Content: content},
	}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create an interactive browser app.")
	if issues := resultContractIssues(requirement, artifacts, evidence); len(issues) != 0 {
		t.Fatalf("familiar visible control label was rejected: %v", issues)
	}
}

func TestInteractivePackageReadbackRequiresApprovedValidationTargets(t *testing.T) {
	entrypoint := "groups/team/generated/app/index.html"
	plan := &protocol.OutputValidationPlan{
		Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
		Checks: []protocol.OutputValidationCheck{protocol.OutputValidationCheckLoad},
		Probe: &protocol.OutputValidationProbe{
			Action:  protocol.OutputValidationAction{Kind: protocol.OutputValidationActionClick, Target: "[data-mycelis-primary-action]"},
			Observe: protocol.OutputValidationObservation{Kind: protocol.OutputValidationObserveTextChange, Target: "[data-mycelis-validation-surface]"},
		},
	}
	requirement := &teamResultRequirement{
		Kind: "project_package", ExpectedOutputs: []string{"interactive browser app"},
		ReadbackRequired: true, OutputValidation: plan,
	}
	content := `<p>Use the primary control.</p><button onclick="status.textContent='changed'">Run</button><p id="status">Ready</p>`
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: content},
		{ToolName: "read_file", Path: entrypoint, Content: content},
	}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create an interactive browser app.")
	issues := strings.Join(resultContractIssues(requirement, artifacts, evidence), ";")
	for _, selector := range []string{"[data-mycelis-primary-action]", "[data-mycelis-validation-surface]"} {
		if !strings.Contains(issues, selector) {
			t.Fatalf("issues = %q, want missing selector %s", issues, selector)
		}
	}

	repaired := strings.Replace(content, "<button ", `<button data-mycelis-primary-action `, 1)
	repaired = strings.Replace(repaired, `<p id="status"`, `<p data-mycelis-validation-surface id="status"`, 1)
	repaired = strings.Replace(repaired, "status.textContent", "document.getElementById('status').textContent", 1)
	evidence = append(evidence,
		successfulToolEvidence{ToolName: "write_file", Path: entrypoint, Content: repaired},
		successfulToolEvidence{ToolName: "read_file", Path: entrypoint, Content: repaired},
	)
	artifacts = reconcileToolBackedArtifacts(nil, evidence, "Create an interactive browser app.")
	if issues := resultContractIssues(requirement, artifacts, evidence); len(issues) != 0 {
		t.Fatalf("repaired validation targets still fail: %v", issues)
	}
}

func TestTextChangeValidationMustMutateObservedSurface(t *testing.T) {
	entrypoint := "groups/team/generated/app/index.html"
	plan := &protocol.OutputValidationPlan{
		Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
		Probe: &protocol.OutputValidationProbe{
			Action:  protocol.OutputValidationAction{Kind: protocol.OutputValidationActionClick, Target: "[data-mycelis-primary-action]"},
			Observe: protocol.OutputValidationObservation{Kind: protocol.OutputValidationObserveTextChange, Target: "[data-mycelis-validation-surface]"},
		},
	}
	requirement := &teamResultRequirement{
		Kind: "project_package", ExpectedOutputs: []string{"interactive browser app"},
		EntrypointRequired: true, ReadbackRequired: true, OutputValidation: plan,
	}
	wrongTarget := `<p>Use Run to start.</p><button data-mycelis-primary-action onclick="document.getElementById('other').textContent='changed'">Run</button><p id="status" data-mycelis-validation-surface>Ready</p><p id="other">Other</p>`
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: wrongTarget},
		{ToolName: "read_file", Path: entrypoint, Content: wrongTarget},
	}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create an interactive browser app.")
	issues := strings.Join(resultContractIssues(requirement, artifacts, evidence), ";")
	if !strings.Contains(issues, "text-change observation target") {
		t.Fatalf("issues = %q, want observed-surface mutation failure", issues)
	}

	repaired := strings.Replace(wrongTarget, "getElementById('other')", "getElementById('status')", 1)
	evidence = []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: repaired},
		{ToolName: "read_file", Path: entrypoint, Content: repaired},
	}
	artifacts = reconcileToolBackedArtifacts(nil, evidence, "Create an interactive browser app.")
	if issues := resultContractIssues(requirement, artifacts, evidence); len(issues) != 0 {
		t.Fatalf("observed-surface mutation was rejected: %v", issues)
	}
}

func TestOutputValidationExecutionInstructionRequiresRenderedStateTransition(t *testing.T) {
	plan := &protocol.OutputValidationPlan{
		Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
		Probe: &protocol.OutputValidationProbe{
			Action:  protocol.OutputValidationAction{Kind: protocol.OutputValidationActionKeyHold, Key: "ArrowRight"},
			Observe: protocol.OutputValidationObservation{Kind: protocol.OutputValidationObserveVisualChange, Target: "[data-mycelis-validation-surface]"},
		},
	}

	instruction := outputValidationExecutionInstruction(plan)
	for _, required := range []string{"ArrowRight", "one unambiguous state-changing effect", "must not also match that control", "render loop actually consumes", "before/after state must differ", "unused intermediate", "game canvas itself"} {
		if !strings.Contains(instruction, required) {
			t.Fatalf("instruction = %q, want %q", instruction, required)
		}
	}
}

func TestHeldKeyVisualValidationRequiresCanvasMarker(t *testing.T) {
	plan := &protocol.OutputValidationPlan{
		Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
		Probe: &protocol.OutputValidationProbe{
			Action:  protocol.OutputValidationAction{Kind: protocol.OutputValidationActionKeyHold, Key: "ArrowRight"},
			Observe: protocol.OutputValidationObservation{Kind: protocol.OutputValidationObserveVisualChange, Target: "[data-mycelis-validation-surface]"},
		},
	}
	statusOnly := `<canvas id="game"></canvas><p data-mycelis-validation-surface>Ready</p><script>const ctx=game.getContext('2d');addEventListener('keydown',()=>ctx.fillRect(1,1,2,2));</script>`
	if issues := outputValidationVisualChangeIssues(plan, statusOnly); len(issues) != 1 || !strings.Contains(issues[0], "rendered canvas") {
		t.Fatalf("status marker issues = %v, want rendered canvas requirement", issues)
	}
	canvasMarked := strings.Replace(statusOnly, `<canvas id="game"`, `<canvas data-mycelis-validation-surface id="game"`, 1)
	canvasMarked = strings.Replace(canvasMarked, ` data-mycelis-validation-surface>Ready`, `>Ready`, 1)
	if issues := outputValidationVisualChangeIssues(plan, canvasMarked); len(issues) != 0 {
		t.Fatalf("canvas marker rejected: %v", issues)
	}
}

func TestInteractivePackageRequiresFreshReadbackAfterRepairWrite(t *testing.T) {
	requirement := &teamResultRequirement{
		Kind: "project_package", ExpectedOutputs: []string{"playable browser game package"}, ReadbackRequired: true,
	}
	entrypoint := "groups/team/generated/app/index.html"
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: `<canvas></canvas><script>document.addEventListener('keydown', move)</script>`},
		{ToolName: "read_file", Path: entrypoint, Content: `<canvas></canvas><script>document.addEventListener('keydown', move)</script>`},
		{ToolName: "write_file", Path: entrypoint, Content: `<p>Controls: Hold ArrowRight to move.</p><canvas></canvas><script>document.addEventListener('keydown', move); function move(){ document.body.dataset.moved = 'true'; }</script>`},
	}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create a playable project package.")

	issues := strings.Join(resultContractIssues(requirement, artifacts, evidence), ";")
	if !strings.Contains(issues, "missing successful structural readback") {
		t.Fatalf("later write did not invalidate stale readback: %s", issues)
	}
	if artifacts[0].Validation != "" {
		t.Fatalf("artifact retained stale validation: %q", artifacts[0].Validation)
	}

	evidence = append(evidence, successfulToolEvidence{ToolName: "read_file", Path: entrypoint, Content: evidence[len(evidence)-1].Content})
	artifacts = reconcileToolBackedArtifacts(artifacts, evidence, "Create a playable project package.")
	if issues := resultContractIssues(requirement, artifacts, evidence); len(issues) != 0 {
		t.Fatalf("fresh readback did not restore proof: %v", issues)
	}
}

func TestDormantGameOverwriteInvalidatesEarlierReadback(t *testing.T) {
	entrypoint := "groups/team/generated/first-game/index.html"
	plan := &protocol.OutputValidationPlan{Kind: protocol.OutputValidationInteractiveBrowser, Required: true}
	requirement := &teamResultRequirement{
		Kind: "project_package", EntrypointRequired: true, ReadbackRequired: true, OutputValidation: plan,
	}
	started := `<script>function gameLoop(){requestAnimationFrame(gameLoop)} gameLoop()</script>`
	dormant := `<p>Use ArrowRight to move.</p><canvas data-mycelis-validation-surface></canvas><script>
document.addEventListener('keydown', e => right = e.key === 'ArrowRight');
function gameLoop() { requestAnimationFrame(gameLoop); }
</script>`
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: started},
		{ToolName: "read_file", Path: entrypoint, Content: started},
		{ToolName: "write_file", Path: entrypoint, Content: dormant},
	}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create a playable browser game.")
	if resultContractEvidenceToolAllowed(requirement, "read_file", artifacts, evidence) {
		t.Fatal("readback was allowed for a dormant latest entrypoint write")
	}
	issues := strings.Join(resultContractIssues(requirement, artifacts, evidence), ";")
	for _, expected := range []string{"missing successful structural readback", "animation loop gameLoop is defined but never started"} {
		if !strings.Contains(issues, expected) {
			t.Fatalf("issues = %q, want %q", issues, expected)
		}
	}
}

func TestProjectPackageRequiredFilesMustShareRetainedPackageFolder(t *testing.T) {
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: "groups/team/generated/app/index.html"},
		{ToolName: "write_file", Path: "groups/team/scratch/README.md"},
	}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create a project package.")
	requirement := &teamResultRequirement{Kind: "project_package", FilesRequired: []string{"README.md"}, EntrypointRequired: true}

	issues := strings.Join(resultContractIssues(requirement, artifacts, evidence), ";")
	if !strings.Contains(issues, "missing successful write for README.md") {
		t.Fatalf("wrong-folder file satisfied package contract: %s", issues)
	}
}

func TestProjectPackageReadbackWaitsForEveryRequiredWrite(t *testing.T) {
	requirement := &teamResultRequirement{
		Kind: "project_package", FilesRequired: []string{"index.html", "README.md", "project-package.json"},
		EntrypointRequired: true, ReadbackRequired: true,
	}
	evidence := []successfulToolEvidence{{ToolName: "write_file", Path: "groups/team/generated/app/index.html"}}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create a project package.")

	if resultContractEvidenceToolAllowed(requirement, "read_file", artifacts, evidence) {
		t.Fatal("readback was allowed before every required package file was written")
	}
	evidence = append(evidence,
		successfulToolEvidence{ToolName: "write_file", Path: "groups/team/generated/app/README.md"},
		successfulToolEvidence{ToolName: "write_file", Path: "groups/team/generated/app/project-package.json"},
	)
	artifacts = reconcileToolBackedArtifacts(artifacts, evidence, "Create a project package.")
	if !resultContractEvidenceToolAllowed(requirement, "read_file", artifacts, evidence) {
		t.Fatal("readback remained blocked after every required package file was written")
	}
}

func TestProjectPackageReadbackWaitsForInspectableInteractiveWrite(t *testing.T) {
	entrypoint := "groups/team/generated/app/index.html"
	plan := &protocol.OutputValidationPlan{
		Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
		Probe: &protocol.OutputValidationProbe{
			Action:  protocol.OutputValidationAction{Kind: protocol.OutputValidationActionClick, Target: "[data-mycelis-primary-action]"},
			Observe: protocol.OutputValidationObservation{Kind: protocol.OutputValidationObserveTextChange, Target: "[data-mycelis-validation-surface]"},
		},
	}
	requirement := &teamResultRequirement{
		Kind: "project_package", ExpectedOutputs: []string{"interactive browser app"},
		EntrypointRequired: true, ReadbackRequired: true, OutputValidation: plan,
	}
	evidence := []successfulToolEvidence{{
		ToolName: "write_file", Path: entrypoint,
		Content: `<button onclick="status.textContent='changed'">Run</button><p id="status">Ready</p>`,
	}}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create an interactive browser app.")
	if resultContractEvidenceToolAllowed(requirement, "read_file", artifacts, evidence) {
		t.Fatal("readback was allowed before the interactive write exposed approved validation targets")
	}

	repaired := `<p>Use Run to start.</p><button data-mycelis-primary-action onclick="document.getElementById('status').textContent='changed'">Run</button><p data-mycelis-validation-surface id="status">Ready</p>`
	evidence = append(evidence, successfulToolEvidence{ToolName: "write_file", Path: entrypoint, Content: repaired})
	artifacts = reconcileToolBackedArtifacts(artifacts, evidence, "Create an interactive browser app.")
	if !resultContractEvidenceToolAllowed(requirement, "read_file", artifacts, evidence) {
		t.Fatal("readback remained blocked after the interactive write satisfied static contract checks")
	}
}

func TestToolEvidencePreservesPathCaseWhileMatchingCaseInsensitively(t *testing.T) {
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: "Groups/Delivery-Team/Generated/App/Index.HTML"},
		{ToolName: "read_file", Path: "groups/delivery-team/generated/app/index.html"},
	}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create a project package.")
	if len(artifacts) != 1 {
		t.Fatalf("artifacts = %#v", artifacts)
	}
	artifact := artifacts[0]
	if artifact.Entrypoint != "Groups/Delivery-Team/Generated/App/Index.HTML" {
		t.Fatalf("entrypoint lost original case: %q", artifact.Entrypoint)
	}
	if artifact.Folder != "Groups/Delivery-Team/Generated/App" || len(artifact.Files) != 1 || artifact.Files[0] != "Index.HTML" {
		t.Fatalf("artifact paths lost original case: %#v", artifact)
	}
	requirement := &teamResultRequirement{
		Kind: "project_package", FilesRequired: []string{"index.html"}, EntrypointRequired: true,
		FolderRequired: true, ReadbackRequired: true,
	}
	if issues := resultContractIssues(requirement, artifacts, evidence); len(issues) != 0 {
		t.Fatalf("case-insensitive evidence comparison failed: %v", issues)
	}
}
