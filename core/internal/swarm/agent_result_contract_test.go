package swarm

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

type resultContractProvider struct {
	responses        []string
	calls            int
	rejectBlankTurns bool
}

func (provider *resultContractProvider) Infer(_ context.Context, _ string, options cognitive.InferOptions) (*cognitive.InferResponse, error) {
	if provider.rejectBlankTurns {
		for _, message := range options.Messages {
			if strings.TrimSpace(message.Content) == "" {
				return nil, errors.New("provider rejected blank conversation turn")
			}
		}
	}
	index := provider.calls
	provider.calls++
	if index >= len(provider.responses) {
		index = len(provider.responses) - 1
	}
	return &cognitive.InferResponse{Text: provider.responses[index], Provider: "mock", ModelUsed: "contract-test"}, nil
}

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

func (provider *resultContractProvider) Probe(context.Context) (bool, error) { return true, nil }

type resultContractToolExecutor struct {
	calls []string
	fail  bool
}

func (executor *resultContractToolExecutor) FindToolByName(_ context.Context, name string) (uuid.UUID, string, error) {
	return InternalServerID, name, nil
}

func (executor *resultContractToolExecutor) CallTool(_ context.Context, _ uuid.UUID, name string, _ map[string]any) (string, error) {
	executor.calls = append(executor.calls, name)
	if executor.fail {
		return "", errors.New("write unavailable")
	}
	return "completed:" + name, nil
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

func TestResultContractExecutionPromptCarriesAcceptanceIntoInteractivePackage(t *testing.T) {
	requirement := &teamResultRequirement{
		Kind: "project_package", FilesRequired: []string{"README.md"},
		AcceptanceCriteria: []string{"primary control changes the application"},
		OutputValidation: &protocol.OutputValidationPlan{Probe: &protocol.OutputValidationProbe{
			Action:  protocol.OutputValidationAction{Kind: protocol.OutputValidationActionClick, Target: "[data-mycelis-primary-action]"},
			Observe: protocol.OutputValidationObservation{Kind: protocol.OutputValidationObserveTextChange, Target: "[data-mycelis-validation-surface]"},
		}},
	}
	prompt := resultContractExecutionPrompt(requirement)
	for _, want := range []string{"README.md", "primary control changes the application", "visibly explain the primary control", "keydown", "never a positional selector", "addEventListener"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("execution prompt = %q, missing %q", prompt, want)
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
	if len(result.Artifacts) != 1 || len(result.Artifacts[0].Files) != 1 || result.Artifacts[0].Files[0] != "index.html" {
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
	for _, missing := range []string{"README.md", "PROOF.md", "project-package.json", "readback"} {
		if !strings.Contains(result.Availability.Summary, missing) {
			t.Fatalf("summary = %q, missing %q", result.Availability.Summary, missing)
		}
	}
	if result.Availability.RecommendedAction == "" {
		t.Fatal("expected concrete recovery action")
	}
	if len(result.Artifacts) != 1 || len(result.Artifacts[0].Files) != 1 || result.Artifacts[0].Files[0] != "index.html" {
		t.Fatalf("artifact invented support files: %#v", result.Artifacts)
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

func resultContractTestAgent(provider cognitive.LLMProvider, executor MCPToolExecutor) *Agent {
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles:  map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{"mock": {Type: "mock", Enabled: true, ModelID: "contract-test"}},
		},
		Adapters: map[string]cognitive.LLMProvider{"mock": provider},
	}
	agent := NewAgent(context.Background(), protocol.AgentManifest{
		ID: "worker", Role: "implementer", Provider: "mock", Tools: []string{"write_file", "read_file"}, MaxIterations: 6,
	}, "delivery-team", nil, router, executor)
	agent.SetToolDescriptions(map[string]string{"write_file": "Write retained output.", "read_file": "Read retained output."})
	return agent
}
