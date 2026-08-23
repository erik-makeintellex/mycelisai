package swarm

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestInteractivePackageRejectsAnimationLoopThatNeverStarts(t *testing.T) {
	plan := &protocol.OutputValidationPlan{Kind: protocol.OutputValidationInteractiveBrowser, Required: true}
	content := `<script>
function gameLoop() { render(); requestAnimationFrame(gameLoop); }
</script>`
	issues := strings.Join(outputValidationAnimationLoopIssues(plan, content), ";")
	if !strings.Contains(issues, "animation loop gameLoop is defined but never started") {
		t.Fatalf("issues = %q, want dormant animation-loop issue", issues)
	}
	if !strings.Contains(outputValidationCorrectionInstruction(plan, []string{issues}), "invoking the loop once") {
		t.Fatal("correction does not tell the worker how to start the retained loop")
	}
}

func TestInteractivePackageAcceptsAnimationLoopBootstrapForms(t *testing.T) {
	plan := &protocol.OutputValidationPlan{Kind: protocol.OutputValidationInteractiveBrowser, Required: true}
	declaration := `function gameLoop() { render(); requestAnimationFrame(gameLoop); }`
	for name, bootstrap := range map[string]string{
		"direct call":        `gameLoop();`,
		"animation callback": `requestAnimationFrame(gameLoop);`,
		"event callback":     `addEventListener('load', gameLoop);`,
		"timer callback":     `setTimeout(gameLoop, 0);`,
	} {
		t.Run(name, func(t *testing.T) {
			if issues := outputValidationAnimationLoopIssues(plan, declaration+bootstrap); len(issues) != 0 {
				t.Fatalf("bootstrap was rejected: %v", issues)
			}
		})
	}
}

func TestAnimationLoopPreflightBlocksReadbackUntilRepair(t *testing.T) {
	entrypoint := "groups/team/generated/app/index.html"
	plan := &protocol.OutputValidationPlan{Kind: protocol.OutputValidationInteractiveBrowser, Required: true}
	requirement := &teamResultRequirement{Kind: "project_package", EntrypointRequired: true, ReadbackRequired: true, OutputValidation: plan}
	dormant := `<p>Click Run to start.</p><button onclick="status.textContent='Started'">Run</button><main id="status">Ready</main><script>function gameLoop(){render();requestAnimationFrame(gameLoop)}</script>`
	evidence := []successfulToolEvidence{{ToolName: "write_file", Path: entrypoint, Content: dormant}}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create an interactive browser app.")
	if resultContractEvidenceToolAllowed(requirement, "read_file", artifacts, evidence) {
		t.Fatal("readback was allowed before the self-scheduling loop was started")
	}

	started := dormant + `<script>gameLoop()</script>`
	evidence = append(evidence, successfulToolEvidence{ToolName: "write_file", Path: entrypoint, Content: started})
	artifacts = reconcileToolBackedArtifacts(artifacts, evidence, "Create an interactive browser app.")
	if !resultContractEvidenceToolAllowed(requirement, "read_file", artifacts, evidence) {
		t.Fatal("readback remained blocked after the loop gained a bootstrap")
	}
	evidence = append(evidence, successfulToolEvidence{ToolName: "read_file", Path: entrypoint, Content: started})
	if issues := resultContractIssues(requirement, artifacts, evidence); len(issues) != 0 {
		t.Fatalf("repaired entrypoint issues = %v", issues)
	}
}

func TestAnimationLoopPreflightIgnoresUnownedPatterns(t *testing.T) {
	interactive := &protocol.OutputValidationPlan{Kind: protocol.OutputValidationInteractiveBrowser, Required: true}
	for _, content := range []string{
		`const loop = () => requestAnimationFrame(loop);`,
		`function renderOnce() { render(); }`,
		`// function fake(){requestAnimationFrame(fake)}`,
		`<p>It's a function demo.</p><script>function loop(){requestAnimationFrame(loop)}; loop()</script>`,
	} {
		if issues := outputValidationAnimationLoopIssues(interactive, content); len(issues) != 0 {
			t.Fatalf("unowned pattern %q produced issues %v", content, issues)
		}
	}
	if issues := outputValidationAnimationLoopIssues(&protocol.OutputValidationPlan{Kind: "text", Required: true}, `function loop(){requestAnimationFrame(loop)}`); len(issues) != 0 {
		t.Fatalf("noninteractive plan produced issues %v", issues)
	}
}
