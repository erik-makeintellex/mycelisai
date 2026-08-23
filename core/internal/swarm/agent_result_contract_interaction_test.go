package swarm

import (
	"strings"
	"testing"
)

func TestInteractivePackageReadbackRejectsInertHandler(t *testing.T) {
	requirement := &teamResultRequirement{
		Kind: "project_package", ExpectedOutputs: []string{"interactive browser app"},
		AcceptanceCriteria: []string{"primary interaction changes the application"}, ReadbackRequired: true,
	}
	entrypoint := "groups/team/generated/app/index.html"
	content := `<p>Click Run to update the status.</p><button data-mycelis-primary-action onclick="console.log('run')">Run</button><main data-mycelis-validation-surface>Ready</main>`
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: content},
		{ToolName: "read_file", Path: entrypoint, Content: content},
	}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create an interactive browser app.")

	issues := strings.Join(resultContractIssues(requirement, artifacts, evidence), ";")
	if !strings.Contains(issues, "inspectable primary interaction") {
		t.Fatalf("issues = %q, want inert primary interaction repair", issues)
	}
}

func TestInteractivePackageReadbackWaitsForStateChangingInteraction(t *testing.T) {
	entrypoint := "groups/team/generated/app/index.html"
	requirement := &teamResultRequirement{
		Kind: "project_package", ExpectedOutputs: []string{"interactive browser app"},
		EntrypointRequired: true, ReadbackRequired: true,
	}
	inert := `<p>Click Run to update the status.</p><button data-mycelis-primary-action onclick="console.log('run')">Run</button><main data-mycelis-validation-surface>Ready</main>`
	evidence := []successfulToolEvidence{{ToolName: "write_file", Path: entrypoint, Content: inert}}
	artifacts := reconcileToolBackedArtifacts(nil, evidence, "Create an interactive browser app.")
	if resultContractEvidenceToolAllowed(requirement, "read_file", artifacts, evidence) {
		t.Fatal("readback was allowed before the interactive write exposed a state-changing effect")
	}

	repaired := `<p>Click Run to update the status.</p><button data-mycelis-primary-action onclick="document.getElementById('status').textContent='Changed'">Run</button><main data-mycelis-validation-surface id="status">Ready</main>`
	evidence = append(evidence, successfulToolEvidence{ToolName: "write_file", Path: entrypoint, Content: repaired})
	artifacts = reconcileToolBackedArtifacts(artifacts, evidence, "Create an interactive browser app.")
	if !resultContractEvidenceToolAllowed(requirement, "read_file", artifacts, evidence) {
		t.Fatal("readback remained blocked after the interactive write gained a state-changing effect")
	}
}
