package swarm

import (
	"errors"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestApprovedResultContractRuntimeFallbackCompletesPackageAfterProviderDrop(t *testing.T) {
	incomplete := `<p>Draft package.</p><button id="play">Play</button><p id="score">Score 0</p>`
	provider := &resultContractProvider{
		responses: []string{writeToolCall("index.html", incomplete), "unreachable"},
		errors:    map[int]error{1: errors.New("provider dropped after partial write")},
	}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := runtimeFallbackPackageRequirement()

	result := agent.processMessageStructuredWithRequirement("Build a playable browser app package.", nil, false, requirement)

	if result.Availability != nil {
		t.Fatalf("runtime fallback left package degraded: %+v", result.Availability)
	}
	if provider.calls != 2 {
		t.Fatalf("provider calls = %d, want initial write plus failed continuation", provider.calls)
	}
	assertRuntimeFallbackPackageResult(t, result, executor, "write_file,write_file,read_file")
	for _, want := range []string{"index.html", "README.md", "PROOF.md", "project-package.json"} {
		if !testStringSliceContains(result.Artifacts[0].Files, want) {
			t.Fatalf("artifact files = %#v, missing %s", result.Artifacts[0].Files, want)
		}
	}
}

func TestApprovedResultContractRuntimeFallbackCompletesPackageAfterInitialProviderFailure(t *testing.T) {
	provider := &resultContractProvider{
		responses: []string{"unreachable"},
		errors:    map[int]error{0: errors.New("provider unavailable before tools")},
	}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := runtimeFallbackPackageRequirement()

	result := agent.processMessageStructuredWithRequirement("Build a playable browser app package.", nil, false, requirement)

	if result.Availability == nil || !result.Availability.Available {
		t.Fatalf("runtime fallback did not return available proof: %+v", result.Availability)
	}
	if result.Availability.Code != "runtime_owned_package_recovery" {
		t.Fatalf("availability code = %q", result.Availability.Code)
	}
	if provider.calls != 1 {
		t.Fatalf("provider calls = %d, want initial failed inference only", provider.calls)
	}
	assertRuntimeFallbackPackageResult(t, result, executor, "write_file,read_file")
}

func TestApprovedResultContractRuntimeFallbackCompletesAfterNoToolEvidence(t *testing.T) {
	provider := &resultContractProvider{
		responses: []string{
			"The package is ready.",
			"The package is ready.",
			"The package is ready.",
			"The package is ready.",
			"The package is ready.",
		},
	}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := runtimeFallbackPackageRequirement()
	requirement.PackageTitle = "Requested First Playable"

	result := agent.processMessageStructuredWithRequirement(
		"Build a playable browser app package. Use the package title Requested First Playable.",
		nil,
		false,
		requirement,
	)

	if result.Availability != nil {
		t.Fatalf("no-evidence runtime fallback left package degraded: %+v", result.Availability)
	}
	assertRuntimeFallbackPackageResult(t, result, executor, "write_file,read_file")
	if result.Artifacts[0].Title != "Requested First Playable" {
		t.Fatalf("title = %q, want requested title", result.Artifacts[0].Title)
	}
	if !strings.Contains(executor.files[result.Artifacts[0].Entrypoint], `id="game"`) {
		t.Fatalf("fallback entrypoint did not include playable canvas: %s", executor.files[result.Artifacts[0].Entrypoint])
	}
}

func TestRuntimeFallbackDoesNotClaimRichGameAcceptance(t *testing.T) {
	requirement := runtimeFallbackPackageRequirement()
	requirement.AcceptanceCriteria = []string{
		"Player can move, jump, attack, and collide with the level",
		"Enemies, health, key pickup, locked door, win, fail, and restart states work",
		"Generated music and action sounds respond to play",
	}

	if initialProjectPackageRuntimeFallbackAllowed(requirement) {
		t.Fatal("generic runtime fallback must not certify a rich game contract")
	}
}

func TestRuntimeFallbackAllowsGenericPrimaryInteractionAcceptance(t *testing.T) {
	requirement := runtimeFallbackPackageRequirement()
	requirement.AcceptanceCriteria = []string{"Primary interaction changes the application state"}

	if !initialProjectPackageRuntimeFallbackAllowed(requirement) {
		t.Fatal("generic primary interaction remains eligible for bounded recovery")
	}
}

func TestApprovedResultContractRuntimeFallbackRepairsCompleteWritesWithoutReadback(t *testing.T) {
	requirement := runtimeFallbackPackageRequirement()
	entrypoint := requirement.PackageEntrypoint
	partial := `<canvas data-mycelis-validation-surface></canvas>`
	evidence := []successfulToolEvidence{
		{ToolName: "write_file", Path: entrypoint, Content: partial},
		{ToolName: "write_file", Path: requirement.PackageFolder + "/README.md", Content: "# Usage"},
		{ToolName: "write_file", Path: requirement.PackageFolder + "/PROOF.md", Content: "# Proof"},
		{ToolName: "write_file", Path: requirement.PackageFolder + "/project-package.json", Content: `{}`},
	}
	artifact := protocol.ChatArtifactRef{
		Type: "project_package", Title: "Partial package", SavedPath: requirement.PackageFolder,
		Folder: requirement.PackageFolder, Entrypoint: entrypoint,
		Files: []string{"index.html", "README.md", "PROOF.md", "project-package.json"},
	}
	result := &agentToolLoopResult{artifacts: []protocol.ChatArtifactRef{artifact}, toolEvidence: evidence}
	executor := &resultContractToolExecutor{files: map[string]string{entrypoint: partial}}
	agent := resultContractTestAgent(&resultContractProvider{responses: []string{"unused"}}, executor)

	if !agent.completeProjectPackageRuntimeFallback("Build the approved package.", requirement, result, false) {
		t.Fatal("runtime fallback did not repair complete writes that lacked trusted readback")
	}
	if got := strings.Join(executor.calls, ","); got != "write_file,read_file" {
		t.Fatalf("tool calls = %s, want bounded replacement and readback", got)
	}
	if issues := resultContractIssues(requirement, result.artifacts, result.toolEvidence); len(issues) != 0 {
		t.Fatalf("runtime fallback left contract issues: %v", issues)
	}
}

func runtimeFallbackPackageRequirement() *teamResultRequirement {
	return &teamResultRequirement{
		Kind: "project_package", TeamID: "delivery-team",
		PackageFolder:      "groups/delivery-team/generated/package",
		PackageEntrypoint:  "groups/delivery-team/generated/package/index.html",
		FilesRequired:      []string{"index.html", "README.md", "PROOF.md", "project-package.json"},
		ExpectedOutputs:    []string{"interactive browser app"},
		EntrypointRequired: true, FolderRequired: true, ReadbackRequired: true,
		OutputValidation: &protocol.OutputValidationPlan{
			Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
			Probe: &protocol.OutputValidationProbe{
				Action:  protocol.OutputValidationAction{Kind: protocol.OutputValidationActionClick, Target: "[data-mycelis-primary-action]"},
				Observe: protocol.OutputValidationObservation{Kind: protocol.OutputValidationObserveTextChange, Target: "[data-mycelis-validation-surface]"},
			},
		},
	}
}

func assertRuntimeFallbackPackageResult(t *testing.T, result ProcessResult, executor *resultContractToolExecutor, calls string) {
	t.Helper()
	if got := strings.Join(executor.calls, ","); got != calls {
		t.Fatalf("tool calls = %s", got)
	}
	if len(result.Artifacts) != 1 {
		t.Fatalf("artifacts = %#v", result.Artifacts)
	}
	artifact := result.Artifacts[0]
	if artifact.Entrypoint != "groups/delivery-team/generated/package/index.html" {
		t.Fatalf("entrypoint = %q", artifact.Entrypoint)
	}
	if !strings.Contains(artifact.Validation, "Structural readback") {
		t.Fatalf("validation = %q", artifact.Validation)
	}
}
