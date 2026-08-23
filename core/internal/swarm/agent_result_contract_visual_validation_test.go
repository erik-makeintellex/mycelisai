package swarm

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestVisualChangeValidationMustMarkRenderedSurface(t *testing.T) {
	entrypoint := "groups/team/generated/app/index.html"
	plan := &protocol.OutputValidationPlan{
		Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
		Probe: &protocol.OutputValidationProbe{
			Action:  protocol.OutputValidationAction{Kind: protocol.OutputValidationActionKeyHold, Key: "ArrowRight"},
			Observe: protocol.OutputValidationObservation{Kind: protocol.OutputValidationObserveVisualChange, Target: "[data-mycelis-validation-surface]"},
		},
	}
	requirement := &teamResultRequirement{
		Kind: "project_package", ExpectedOutputs: []string{"playable browser game"},
		EntrypointRequired: true, ReadbackRequired: true, OutputValidation: plan,
	}
	staticLabel := `<!doctype html><div data-mycelis-validation-surface>Control: Hold ArrowRight to move</div><canvas id="game"></canvas><script>
const ctx = document.getElementById('game').getContext('2d');
let x = 0;
addEventListener('keydown', e => { if (e.key === 'ArrowRight') x += 5; });
function draw(){ ctx.clearRect(0,0,100,100); ctx.fillRect(x,0,10,10); requestAnimationFrame(draw); }
draw();
</script>`
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: staticLabel},
		{ToolName: "read_file", Path: entrypoint, Content: staticLabel},
	}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create a playable browser game.")
	issues := strings.Join(resultContractIssues(requirement, artifacts, evidence), ";")
	if !strings.Contains(issues, "visual-change observation target") {
		t.Fatalf("issues = %q, want observed visual-surface mutation failure", issues)
	}

	canvasMarked := strings.Replace(staticLabel, `<canvas id="game"`, `<canvas data-mycelis-validation-surface id="game"`, 1)
	canvasMarked = strings.Replace(canvasMarked, `<div data-mycelis-validation-surface>Control: Hold ArrowRight to move</div>`, `<div>Control: Hold ArrowRight to move</div>`, 1)
	evidence = []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: canvasMarked},
		{ToolName: "read_file", Path: entrypoint, Content: canvasMarked},
	}
	artifacts = reconcileToolBackedArtifacts(nil, evidence, "Create a playable browser game.")
	if issues := resultContractIssues(requirement, artifacts, evidence); len(issues) != 0 {
		t.Fatalf("marked rendered surface was rejected: %v", issues)
	}
}
