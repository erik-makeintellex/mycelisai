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
