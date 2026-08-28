package swarm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestApprovedResultContractRecoversAfterBlankIntermediateResponse(t *testing.T) {
	provider := &resultContractProvider{rejectBlankTurns: true, responses: []string{
		`{"tool_call":{"name":"write_file","arguments":{"path":"groups/delivery-team/generated/app/index.html","content":"<!doctype html><title>Ready</title>"}}}`,
		"",
		`{"tool_call":{"name":"read_file","arguments":{"path":"groups/delivery-team/generated/app/index.html"}}}`,
		"Completed with structural readback evidence.",
	}}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := &teamResultRequirement{
		Kind: "project_package", FilesRequired: []string{"index.html"}, EntrypointRequired: true,
		FolderRequired: true, ReadbackRequired: true, ProofRequirements: []string{"Readback"},
	}

	result := agent.processMessageStructuredWithRequirement("Build an approved project package.", nil, false, requirement)

	if result.Availability != nil {
		t.Fatalf("blank intermediate response prevented bounded recovery: %+v", result.Availability)
	}
	if strings.Join(executor.calls, ",") != "write_file,read_file" {
		t.Fatalf("tool calls = %v", executor.calls)
	}
}

func TestProjectPackageContractRedirectsArtifactStoreToPhysicalFileEvidence(t *testing.T) {
	provider := &resultContractProvider{responses: []string{
		`{"tool_call":{"name":"store_artifact","arguments":{"type":"project_package","title":"Claim","content":"done"}}}`,
		`{"tool_call":{"name":"write_file","arguments":{"path":"groups/delivery-team/generated/app/index.html","content":"<!doctype html><title>Ready</title>"}}}`,
		`{"tool_call":{"name":"read_file","arguments":{"path":"groups/delivery-team/generated/app/index.html"}}}`,
		"Completed with structural readback evidence.",
	}}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := &teamResultRequirement{Kind: "project_package", FilesRequired: []string{"index.html"}, EntrypointRequired: true, ReadbackRequired: true}

	result := agent.processMessageStructuredWithRequirement("Build an approved project package.", nil, false, requirement)

	if result.Availability != nil {
		t.Fatalf("artifact-store redirect prevented completion: %+v", result.Availability)
	}
	if strings.Join(executor.calls, ",") != "write_file,read_file" {
		t.Fatalf("tool calls = %v, store_artifact must not substitute for package files", executor.calls)
	}
}

func TestTeamResultRequirementFromTriggerParsesApprovedContract(t *testing.T) {
	raw, err := json.Marshal(protocol.TeamAsk{
		Goal: "Create retained work.",
		Context: map[string]any{
			"run_id": "run-1", "contract_id": "contract-1", "intent_proof_id": "proof-1",
			"result_contract": map[string]any{
				"kind": "project_package", "files_required": []string{"index.html"},
				"expected_outputs": []string{"Openable package"}, "acceptance_criteria": []string{"Opens cleanly"},
				"proof_required": []string{"Readback"}, "entrypoint_required": true, "validation_required": true,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	requirement := teamResultRequirementFromTrigger(raw, false)
	if requirement == nil || requirement.Kind != "project_package" || len(requirement.FilesRequired) != 1 {
		t.Fatalf("requirement = %#v", requirement)
	}
	if requirement.ExpectedOutputs[0] != "Openable package" || requirement.AcceptanceCriteria[0] != "Opens cleanly" || requirement.ProofRequirements[0] != "Readback" {
		t.Fatalf("requirement detail = %#v", requirement)
	}
	if teamResultRequirementFromTrigger(raw, true) != nil {
		t.Fatal("planning-only trigger must not enforce an execution result contract")
	}
}

func TestResultContractLoopLimitReservesBoundedMultiFileRepairBudget(t *testing.T) {
	requirement := &teamResultRequirement{Kind: "project_package", FilesRequired: []string{"README.md", "PROOF.md", "project-package.json"}}
	if got := resultContractLoopLimit(6, requirement); got != maxResultContractToolIterations {
		t.Fatalf("result contract loop limit = %d, want %d", got, maxResultContractToolIterations)
	}
	if got := resultContractLoopLimit(6, nil); got != 6 {
		t.Fatalf("ordinary loop limit = %d, want 6", got)
	}
}

func TestResultContractEvidenceToolAllowedRestrictsIncompleteProjectPackage(t *testing.T) {
	requirement := &teamResultRequirement{Kind: "project_package", FilesRequired: []string{"index.html"}}
	if !resultContractEvidenceToolAllowed(requirement, "write_file", nil, nil) {
		t.Fatal("write_file should advance package evidence")
	}
	for _, toolName := range []string{"read_file", "read_text_file"} {
		if resultContractEvidenceToolAllowed(requirement, toolName, nil, nil) {
			t.Fatalf("%s should wait until required package writes finish", toolName)
		}
	}
	for _, toolName := range []string{"local_command", "store_artifact", "consult_council"} {
		if resultContractEvidenceToolAllowed(requirement, toolName, nil, nil) {
			t.Fatalf("%s should be deferred while package evidence is incomplete", toolName)
		}
	}
}

func TestResultContractExecutionPromptReferencesSingleAuthoritativePackageBrief(t *testing.T) {
	requirement := &teamResultRequirement{
		Kind: "project_package", FilesRequired: []string{"README.md"},
		AcceptanceCriteria: []string{"primary control changes the application"},
		OutputValidation: &protocol.OutputValidationPlan{Probe: &protocol.OutputValidationProbe{
			Action:  protocol.OutputValidationAction{Kind: protocol.OutputValidationActionClick, Target: "[data-mycelis-primary-action]"},
			Observe: protocol.OutputValidationObservation{Kind: protocol.OutputValidationObserveTextChange, Target: "[data-mycelis-validation-surface]"},
		}},
	}
	prompt := resultContractExecutionPrompt(requirement)
	for _, want := range []string{
		"PACKAGE CONTRACT v1",
		"exactly one tool_call JSON object",
		"visibly explain the primary control",
		"keydown",
		"never a positional selector",
		"addEventListener",
		"text_change validation",
		"textContent",
		"different user-visible text",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("execution prompt = %q, missing %q", prompt, want)
		}
	}
	for _, duplicated := range []string{"README.md", "primary control changes the application"} {
		if strings.Contains(prompt, duplicated) {
			t.Fatalf("execution policy repeated package brief value %q: %s", duplicated, prompt)
		}
	}
}

func TestApprovedResultContractCorrectsProseUntilWriteAndReadbackExist(t *testing.T) {
	provider := &resultContractProvider{responses: []string{
		"The approved package is complete.",
		`{"tool_call":{"name":"write_file","arguments":{"path":"groups/delivery-team/generated/app/index.html","content":"<!doctype html><title>Ready</title>"}}}`,
		"The approved package is complete.",
		`{"tool_call":{"name":"read_file","arguments":{"path":"groups/delivery-team/generated/app/index.html"}}}`,
		"Completed with readback evidence.",
	}}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := &teamResultRequirement{
		Kind: "project_package", FilesRequired: []string{"index.html"}, ExpectedOutputs: []string{"Openable package"},
		EntrypointRequired: true, FolderRequired: true, ReadbackRequired: true, ProofRequirements: []string{"Readback"},
	}

	result := agent.processMessageStructuredWithRequirement("Build an approved project package.", nil, false, requirement)

	if result.Availability != nil {
		t.Fatalf("unexpected degraded result: %+v", result.Availability)
	}
	if strings.Join(executor.calls, ",") != "write_file,read_file" {
		t.Fatalf("tool calls = %v", executor.calls)
	}
	if len(result.Artifacts) != 1 ||
		!testStringSliceContains(result.Artifacts[0].Files, "index.html") ||
		!testStringSliceContains(result.Artifacts[0].Files, "project-package.json") {
		t.Fatalf("artifacts = %#v", result.Artifacts)
	}
	if !strings.Contains(result.Artifacts[0].Validation, "Structural readback") || !strings.Contains(result.Artifacts[0].Validation, "server/live validation") {
		t.Fatalf("validation = %q", result.Artifacts[0].Validation)
	}
}

func TestApprovedResultContractDegradesWithoutRequiredWritesAndReadback(t *testing.T) {
	provider := &resultContractProvider{responses: []string{
		`{"tool_call":{"name":"write_file","arguments":{"path":"groups/delivery-team/generated/app/index.html","content":"<!doctype html>"}}}`,
		"Everything requested exists and is proven.",
		"Everything requested exists and is proven.",
		"Everything requested exists and is proven.",
	}}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)
	requirement := &teamResultRequirement{
		Kind: "project_package", FilesRequired: []string{"README.md", "PROOF.md", "project-package.json"},
		EntrypointRequired: true, FolderRequired: true, ReadbackRequired: true,
	}

	result := agent.processMessageStructuredWithRequirement("Build an approved project package.", nil, false, requirement)

	if result.Availability == nil || result.Availability.Code != "result_contract_unsatisfied" {
		t.Fatalf("availability = %#v", result.Availability)
	}
	for _, missing := range []string{"README.md", "PROOF.md", "readback"} {
		if !strings.Contains(result.Availability.Summary, missing) {
			t.Fatalf("summary = %q, missing %q", result.Availability.Summary, missing)
		}
	}
	if result.Availability.RecommendedAction == "" {
		t.Fatal("expected concrete recovery action")
	}
	if strings.Contains(result.Text, "Everything requested exists") || strings.Contains(result.Text, "Created retained output") {
		t.Fatalf("degraded result returned completion-style text: %q", result.Text)
	}
	if !strings.Contains(result.Text, "needs repair") {
		t.Fatalf("degraded result text = %q, want repair-oriented wording", result.Text)
	}
	if len(result.Artifacts) != 1 ||
		!testStringSliceContains(result.Artifacts[0].Files, "index.html") ||
		!testStringSliceContains(result.Artifacts[0].Files, "project-package.json") ||
		testStringSliceContains(result.Artifacts[0].Files, "README.md") ||
		testStringSliceContains(result.Artifacts[0].Files, "PROOF.md") {
		t.Fatalf("artifact file evidence crossed the support-file boundary: %#v", result.Artifacts)
	}
	if result.Artifacts[0].Validation != "" {
		t.Fatalf("artifact invented validation: %#v", result.Artifacts[0])
	}
}

func TestApprovedResultContractDoesNotCountFailedToolCallsAsEvidence(t *testing.T) {
	provider := &resultContractProvider{responses: []string{
		`{"tool_call":{"name":"write_file","arguments":{"path":"groups/delivery-team/generated/app/index.html","content":"<!doctype html>"}}}`,
		"The write could not be completed.",
		"The package is complete.",
		"The package is complete.",
	}}
	executor := &resultContractToolExecutor{fail: true}
	agent := resultContractTestAgent(provider, executor)
	requirement := &teamResultRequirement{Kind: "project_package", FilesRequired: []string{"index.html"}, EntrypointRequired: true}

	result := agent.processMessageStructuredWithRequirement("Build an approved project package.", nil, false, requirement)

	if result.Availability == nil || result.Availability.Code != "result_contract_unsatisfied" {
		t.Fatalf("availability = %#v", result.Availability)
	}
	if !strings.Contains(result.Availability.Summary, "no successful retained-output write") {
		t.Fatalf("failed tool was counted as evidence: %q", result.Availability.Summary)
	}
	if len(result.Artifacts) != 0 {
		t.Fatalf("failed write produced artifacts: %#v", result.Artifacts)
	}
}
