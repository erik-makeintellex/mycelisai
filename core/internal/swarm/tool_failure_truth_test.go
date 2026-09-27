package swarm

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/pkg/protocol"
)

func (r *recordingEmitter) waitFor(t *testing.T, eventType protocol.EventType) map[string]interface{} {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		r.mu.Lock()
		for i, got := range r.events {
			if got == eventType {
				payload := r.loads[i]
				r.mu.Unlock()
				return payload
			}
		}
		r.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no %s event emitted", eventType)
	return nil
}

func (r *recordingEmitter) has(eventType protocol.EventType) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, got := range r.events {
		if got == eventType {
			return true
		}
	}
	return false
}

// An MCP result flagged isError reaches the agent as a tool failure: a
// tool.failed event, failure feedback to the model, and no success evidence.
func TestMCPIsErrorResultIsReportedAsToolFailure(t *testing.T) {
	provider := &scriptedDenialProvider{replies: []string{
		`{"tool_call":{"name":"fs_write","arguments":{"path":"/etc/passwd","content":"x"}}}`,
		"The write was refused by the filesystem server.",
	}}
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles:  map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{"mock": {Type: "mock", Enabled: true, ModelID: "test-model"}},
		},
		Adapters: map[string]cognitive.LLMProvider{"mock": provider},
	}
	exec := &mcpExchangeExecutor{serverID: uuid.New(), err: &mcp.ToolResultError{Tool: "fs_write", Message: "access denied: outside allowed directories"}}
	agent := NewAgent(context.Background(), agentManifestWithTools("worker", []string{"fs_write"}), "team-a", nil, router, exec)
	agent.SetToolDescriptions(map[string]string{"fs_write": "Write through the filesystem MCP server."})
	emitter := &recordingEmitter{}
	agent.SetEventEmitter(emitter, "run-1")

	result := agent.processMessageStructuredWithPosture("Write the file.", nil, false)

	payload := emitter.waitFor(t, protocol.EventToolFailed)
	if payload["tool"] != "fs_write" || payload["phase"] != "execute" || !strings.Contains(stringValue(payload["error"]), "access denied") {
		t.Fatalf("tool.failed payload = %#v", payload)
	}
	if emitter.has(protocol.EventToolCompleted) {
		t.Fatal("isError MCP result emitted tool.completed")
	}
	provider.mu.Lock()
	lastPrompt := provider.prompts[len(provider.prompts)-1]
	provider.mu.Unlock()
	if !strings.Contains(lastPrompt, "Tool fs_write failed") || !strings.Contains(lastPrompt, "access denied") {
		t.Fatalf("model did not receive failure feedback: %q", lastPrompt)
	}
	if result.Text != "The write was refused by the filesystem server." {
		t.Fatalf("result text = %q", result.Text)
	}
}

// The bus can still carry runtime_fallback_eligible, but nothing reads it:
// inference failure is always the provider_inference_failed blocker.
func TestInferenceFailureIsBlockerEvenWhenAskRequestsRuntimeFallback(t *testing.T) {
	ask, _ := json.Marshal(protocol.TeamAsk{Goal: "Build the package.", Context: map[string]any{"result_contract": map[string]any{
		"kind": "project_package", "team_id": "delivery-team",
		"package_folder":            "groups/delivery-team/generated/package",
		"package_entrypoint":        "groups/delivery-team/generated/package/index.html",
		"files_required":            []any{"index.html", "README.md", "PROOF.md", "project-package.json"},
		"acceptance_criteria":       []any{"Primary interaction changes the application state"},
		"runtime_fallback_eligible": true, "entrypoint_required": true, "folder_required": true, "validation_required": true,
	}}})
	requirement := teamResultRequirementFromTrigger(ask, false)
	if requirement == nil {
		t.Fatal("contract did not parse")
	}
	provider := &resultContractProvider{responses: []string{"unreachable"}, errors: map[int]error{0: errors.New("provider unavailable")}}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)

	result := agent.processMessageStructuredWithRequirement("Build the package.", nil, false, requirement)

	if result.Availability == nil || result.Availability.Available || result.Availability.Code != "provider_inference_failed" {
		t.Fatalf("availability = %+v, want provider_inference_failed blocker", result.Availability)
	}
	if len(executor.calls) != 0 || len(result.Artifacts) != 0 || len(result.ToolsUsed) != 0 {
		t.Fatalf("runtime produced output after inference failure: calls=%v artifacts=%#v tools=%v", executor.calls, result.Artifacts, result.ToolsUsed)
	}
}

// When inference stops mid-delivery, the runtime writes nothing further; the
// contract gate reports what is still missing.
func TestInferenceStopAfterPartialWriteLeavesContractUnsatisfied(t *testing.T) {
	provider := &resultContractProvider{
		responses: []string{supportFilePackageWrite("index.html", supportFileEntrypoint), "unreachable"},
		errors:    map[int]error{1: errors.New("provider dropped after partial write")},
	}
	executor := &resultContractToolExecutor{}
	agent := resultContractTestAgent(provider, executor)

	result := agent.processMessageStructuredWithRequirement("Build the package.", nil, false, supportFilePackageRequirement())

	if result.Availability == nil || result.Availability.Code != "result_contract_unsatisfied" {
		t.Fatalf("availability = %+v, want result_contract_unsatisfied", result.Availability)
	}
	if got := strings.Join(executor.calls, ","); got != "write_file" {
		t.Fatalf("tool calls = %s, want only the model's write", got)
	}
	if _, synthesized := executor.files["groups/delivery-team/generated/package/README.md"]; synthesized {
		t.Fatal("runtime synthesized README.md after inference stopped")
	}
}
