package swarm

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

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
	writes := []successfulToolEvidence{{ToolName: "write_file", Path: entrypoint}}
	withoutInstructions := append(writes, successfulToolEvidence{
		ToolName: "read_file", Path: entrypoint,
		Content: `<canvas></canvas><script>document.addEventListener('keydown', move)</script>`,
	})
	artifacts := reconcileToolBackedArtifacts(nil, withoutInstructions, "Create a playable project package.")
	issues := strings.Join(resultContractIssues(requirement, artifacts, withoutInstructions), ";")
	if !strings.Contains(issues, "visible control instructions") {
		t.Fatalf("issues = %q, want visible-control repair", issues)
	}

	withInstructions := append(writes, successfulToolEvidence{
		ToolName: "read_file", Path: entrypoint,
		Content: `<p>Use ArrowRight to move.</p><canvas></canvas><script>document.addEventListener('keydown', move)</script>`,
	})
	artifacts = reconcileToolBackedArtifacts(nil, withInstructions, "Create a playable project package.")
	if issues := resultContractIssues(requirement, artifacts, withInstructions); len(issues) != 0 {
		t.Fatalf("valid interactive evidence issues = %v", issues)
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
	evidence = append(evidence, successfulToolEvidence{ToolName: "write_file", Path: entrypoint, Content: repaired})
	artifacts = reconcileToolBackedArtifacts(nil, evidence, "Create an interactive browser app.")
	if issues := resultContractIssues(requirement, artifacts, evidence); len(issues) != 0 {
		t.Fatalf("repaired validation targets still fail: %v", issues)
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
	for _, required := range []string{"ArrowRight", "render loop actually consumes", "before/after state must differ", "unused intermediate"} {
		if !strings.Contains(instruction, required) {
			t.Fatalf("instruction = %q, want %q", instruction, required)
		}
	}
}

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
	dormant := `<script>function gameLoop(){render();requestAnimationFrame(gameLoop)}</script>`
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

func TestInteractivePackageUsesLaterSuccessfulRepairWriteAfterReadback(t *testing.T) {
	requirement := &teamResultRequirement{
		Kind: "project_package", ExpectedOutputs: []string{"playable browser game package"}, ReadbackRequired: true,
	}
	entrypoint := "groups/team/generated/app/index.html"
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: `<canvas></canvas><script>document.addEventListener('keydown', move)</script>`},
		{ToolName: "read_file", Path: entrypoint, Content: `<canvas></canvas><script>document.addEventListener('keydown', move)</script>`},
		{ToolName: "write_file", Path: entrypoint, Content: `<p>Controls: Hold ArrowRight to move.</p><canvas></canvas><script>document.addEventListener('keydown', move)</script>`},
	}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create a playable project package.")

	if issues := resultContractIssues(requirement, artifacts, evidence); len(issues) != 0 {
		t.Fatalf("successful repair write was not treated as latest evidence: %v", issues)
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

	repaired := `<p>Use Run to start.</p><button data-mycelis-primary-action onclick="status.textContent='changed'">Run</button><p data-mycelis-validation-surface id="status">Ready</p>`
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

func TestAgentTriggerRequestReplyReturnsDegradedTruthInsteadOfModelProse(t *testing.T) {
	server, nc := startTestNATS(t)
	defer server.Shutdown()
	defer nc.Close()

	provider := &resultContractProvider{responses: []string{"The requested package is complete."}}
	agent := resultContractTestAgent(provider, nil)
	agent.nc = nc
	subject := "test.team.agent.result-contract"
	if _, err := nc.Subscribe(subject, agent.handleTrigger); err != nil {
		t.Fatal(err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(protocol.TeamAsk{
		Goal: "Create retained work.",
		Context: map[string]any{
			"run_id": "run-1", "contract_id": "contract-1", "intent_proof_id": "proof-1",
			"result_contract": map[string]any{"kind": "project_package", "entrypoint_required": true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	reply, err := nc.Request(subject, raw, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	text := string(reply.Data)
	for _, want := range []string{"Work unavailable", "result_contract_unsatisfied", "Recovery:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("reply = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, "The requested package is complete") {
		t.Fatalf("degraded reply leaked model completion prose: %q", text)
	}
}
