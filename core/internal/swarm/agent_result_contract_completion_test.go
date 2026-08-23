package swarm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestApprovedResultContractRuntimeReadsCompletedEntrypointWithoutAnotherInference(t *testing.T) {
	provider := &resultContractProvider{responses: []string{
		`{"tool_call":{"name":"write_file","arguments":{"path":"groups/delivery-team/generated/app/index.html","content":"<!doctype html><title>Ready</title>"}}}`,
		"unexpected extra inference",
	}}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := &teamResultRequirement{
		Kind: "project_package", FilesRequired: []string{"index.html"},
		EntrypointRequired: true, FolderRequired: true, ReadbackRequired: true,
	}

	result := agent.processMessageStructuredWithRequirement("Build an approved project package.", nil, false, requirement)

	if result.Availability != nil {
		t.Fatalf("runtime readback degraded valid work: %+v", result.Availability)
	}
	if provider.calls != 1 {
		t.Fatalf("provider calls = %d, want no inference after mechanical readback", provider.calls)
	}
	if got := strings.Join(executor.calls, ","); got != "write_file,read_file" {
		t.Fatalf("tool calls = %s", got)
	}
	if len(result.Artifacts) != 1 || !strings.Contains(result.Artifacts[0].Validation, "Structural readback") {
		t.Fatalf("artifacts = %#v", result.Artifacts)
	}
}

func TestApprovedResultContractResetsCorrectionBudgetAfterEvidenceProgress(t *testing.T) {
	provider := &resultContractProvider{responses: []string{
		"The package is complete.",
		writeToolCall("index.html", "<!doctype html><title>Ready</title>"),
		"The package is complete.",
		writeToolCall("README.md", "# Ready"),
		"The package is complete.",
		writeToolCall("PROOF.md", "# Proof"),
		"The package is complete.",
		writeToolCall("project-package.json", `{"entrypoint":"index.html"}`),
		"The package is complete.",
		writeToolCall("LICENSE.txt", "Test license"),
	}}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := &teamResultRequirement{
		Kind:               "project_package",
		FilesRequired:      []string{"index.html", "README.md", "PROOF.md", "project-package.json", "LICENSE.txt"},
		EntrypointRequired: true, FolderRequired: true, ReadbackRequired: true,
	}

	result := agent.processMessageStructuredWithRequirement("Build an approved project package.", nil, false, requirement)

	if result.Availability != nil {
		t.Fatalf("progressive corrections exhausted despite successful evidence: %+v", result.Availability)
	}
	if got := strings.Join(executor.calls, ","); got != "write_file,write_file,write_file,write_file,write_file,read_file" {
		t.Fatalf("tool calls = %s", got)
	}
}

func TestApprovedResultContractRuntimeReadFailureRemainsDegraded(t *testing.T) {
	provider := &resultContractProvider{responses: []string{
		writeToolCall("index.html", "<!doctype html><title>Ready</title>"),
		"The package is complete.",
		"The package is complete.",
		"The package is complete.",
		"The package is complete.",
		"The package is complete.",
	}}
	executor := &resultContractToolExecutor{failRead: true}
	agent := resultContractTestAgent(provider, executor)
	requirement := &teamResultRequirement{
		Kind: "project_package", FilesRequired: []string{"index.html"},
		EntrypointRequired: true, ReadbackRequired: true,
	}

	result := agent.processMessageStructuredWithRequirement("Build an approved project package.", nil, false, requirement)

	if result.Availability == nil || result.Availability.Code != "result_contract_unsatisfied" {
		t.Fatalf("availability = %#v", result.Availability)
	}
	if !strings.Contains(result.Availability.Summary, "missing successful structural readback") {
		t.Fatalf("summary = %q", result.Availability.Summary)
	}
}

func TestApprovedResultContractRuntimeReadsLatestRepairedEntrypoint(t *testing.T) {
	invalid := `<button onclick="status.textContent='changed'">Run</button><p id="status">Ready</p>`
	repaired := `<p>Use Run to start.</p><button data-mycelis-primary-action onclick="status.textContent='changed'">Run</button><p data-mycelis-validation-surface id="status">Ready</p>`
	provider := &resultContractProvider{responses: []string{
		writeToolCall("index.html", invalid),
		writeToolCall("index.html", repaired),
		"unexpected extra inference",
	}}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := &teamResultRequirement{
		Kind: "project_package", ExpectedOutputs: []string{"interactive browser app"},
		EntrypointRequired: true, ReadbackRequired: true,
		OutputValidation: &protocol.OutputValidationPlan{
			Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
			Probe: &protocol.OutputValidationProbe{
				Action:  protocol.OutputValidationAction{Kind: protocol.OutputValidationActionClick, Target: "[data-mycelis-primary-action]"},
				Observe: protocol.OutputValidationObservation{Kind: protocol.OutputValidationObserveTextChange, Target: "[data-mycelis-validation-surface]"},
			},
		},
	}

	result := agent.processMessageStructuredWithRequirement("Build an interactive browser app.", nil, false, requirement)

	if result.Availability != nil {
		t.Fatalf("repaired package degraded: %+v", result.Availability)
	}
	if provider.calls != 2 {
		t.Fatalf("provider calls = %d, want repair write followed by runtime readback", provider.calls)
	}
	if got := strings.Join(executor.calls, ","); got != "write_file,write_file,read_file" {
		t.Fatalf("tool calls = %s", got)
	}
}

func TestApprovedResultContractWritesReferencedLocalDependencyBeforeReadback(t *testing.T) {
	entrypoint := `<!doctype html><title>Ready</title><button data-mycelis-primary-action>Play</button><p data-mycelis-validation-surface>Score 0</p><script src="script.js"></script>`
	provider := &resultContractProvider{responses: []string{
		writeToolCall("index.html", entrypoint),
		writeToolCall("README.md", "# Ready"),
		writeToolCall("PROOF.md", "# Proof"),
		writeToolCall("project-package.json", `{"entrypoint":"index.html","files":["index.html","script.js"]}`),
		"The package is complete.",
		writeToolCall("script.js", `document.querySelector('button').onclick=()=>document.querySelector('p').textContent='Score 1'`),
	}}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := &teamResultRequirement{
		Kind: "project_package", FilesRequired: []string{"README.md", "PROOF.md", "project-package.json"},
		ExpectedOutputs: []string{"interactive browser app"}, EntrypointRequired: true,
		FolderRequired: true, ReadbackRequired: true,
	}

	result := agent.processMessageStructuredWithRequirement("Build an approved project package.", nil, false, requirement)

	if result.Availability != nil {
		t.Fatalf("referenced dependency correction degraded: %+v", result.Availability)
	}
	if got := strings.Join(executor.calls, ","); got != "write_file,write_file,write_file,write_file,write_file,read_file" {
		t.Fatalf("tool calls = %s", got)
	}
	if !strings.Contains(strings.Join(result.Artifacts[0].Files, ","), "script.js") {
		t.Fatalf("artifact files = %#v", result.Artifacts[0].Files)
	}
}

func TestApprovedResultContractCountsAutoScaffoldedProjectPackageSupportFiles(t *testing.T) {
	entrypoint := `<p>Controls: Click Play.</p><button data-mycelis-primary-action onclick="score.textContent='Score 1'">Play</button><p data-mycelis-validation-surface id="score">Score 0</p>`
	provider := &resultContractProvider{responses: []string{
		`{"tool_call":{"name":"write_file","arguments":{"path":"groups/delivery-team/generated/package/index.html","content":` + quotedJSON(entrypoint) + `,"package_kind":"project_package","package_title":"Playable Package","package_folder":"groups/delivery-team/generated/package","package_entrypoint":"groups/delivery-team/generated/package/index.html","package_files":["index.html","README.md","PROOF.md","project-package.json"]}}}`,
		"unexpected extra inference",
	}}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := &teamResultRequirement{
		Kind: "project_package", TeamID: "delivery-team",
		FilesRequired:      []string{"index.html", "README.md", "PROOF.md", "project-package.json"},
		EntrypointRequired: true, FolderRequired: true, ReadbackRequired: true,
		OutputValidation: &protocol.OutputValidationPlan{
			Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
			Probe: &protocol.OutputValidationProbe{
				Action:  protocol.OutputValidationAction{Kind: protocol.OutputValidationActionClick, Target: "[data-mycelis-primary-action]"},
				Observe: protocol.OutputValidationObservation{Kind: protocol.OutputValidationObserveTextChange, Target: "[data-mycelis-validation-surface]"},
			},
		},
	}

	result := agent.processMessageStructuredWithRequirement("Build an approved project package.", nil, false, requirement)

	if result.Availability != nil {
		t.Fatalf("auto-scaffolded support files were not trusted as package evidence: %+v", result.Availability)
	}
	if provider.calls != 1 {
		t.Fatalf("provider calls = %d, want one package write followed by runtime readback", provider.calls)
	}
	if got := strings.Join(executor.calls, ","); got != "write_file,read_file" {
		t.Fatalf("tool calls = %s", got)
	}
	if len(result.Artifacts) != 1 || len(result.Artifacts[0].Files) != 4 {
		t.Fatalf("artifact files = %#v", result.Artifacts)
	}
}

func TestProjectPackageRejectsDivergentLatestReadback(t *testing.T) {
	entrypoint := "groups/team/generated/app/index.html"
	written := "<!doctype html><title>Expected</title>"
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: written},
		{ToolName: "read_file", Path: entrypoint, Content: "<!doctype html><title>Different</title>"},
	}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create a project package.")
	requirement := &teamResultRequirement{Kind: "project_package", EntrypointRequired: true, ReadbackRequired: true}

	issues := strings.Join(resultContractIssues(requirement, artifacts, evidence), ";")
	if !strings.Contains(issues, "missing successful structural readback") {
		t.Fatalf("divergent readback was trusted: %s", issues)
	}
	if artifacts[0].Validation != "" {
		t.Fatalf("divergent readback produced validation: %q", artifacts[0].Validation)
	}
}

func TestOutputValidationMarkersMustBeRenderedHTMLAttributes(t *testing.T) {
	plan := &protocol.OutputValidationPlan{Probe: &protocol.OutputValidationProbe{
		Action:  protocol.OutputValidationAction{Target: "[data-mycelis-primary-action]"},
		Observe: protocol.OutputValidationObservation{Target: "[data-mycelis-validation-surface]"},
	}}
	cssClaims := `<style>#play{data-mycelis-primary-action:run}#score{data-mycelis-validation-surface:score}</style><button id="play">Play</button><p id="score">0</p>`
	if issues := outputValidationTargetIssues(plan, cssClaims); len(issues) != 2 {
		t.Fatalf("CSS claims satisfied rendered marker contract: %v", issues)
	}
	rendered := `<button data-mycelis-primary-action="run">Play</button><p data-mycelis-validation-surface>0</p>`
	if issues := outputValidationTargetIssues(plan, rendered); len(issues) != 0 {
		t.Fatalf("rendered HTML attributes were rejected: %v", issues)
	}
}

func writeToolCall(name, content string) string {
	path := "groups/delivery-team/generated/app/" + name
	return `{"tool_call":{"name":"write_file","arguments":{"path":"` + path + `","content":` + quotedJSON(content) + `}}}`
}

func quotedJSON(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}
