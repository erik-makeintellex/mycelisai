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
