package swarm

import (
	"context"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

type directAnswerCorrectionProvider struct {
	calls   int
	prompts []string
}

func (p *directAnswerCorrectionProvider) Infer(_ context.Context, prompt string, _ cognitive.InferOptions) (*cognitive.InferResponse, error) {
	p.calls++
	p.prompts = append(p.prompts, prompt)
	if p.calls == 1 {
		return &cognitive.InferResponse{Text: `{"tool_call":{"name":"delegate","arguments":{"goal":"preview template"}}}`, Provider: "mock", ModelUsed: "test-model"}, nil
	}
	if p.calls == 2 {
		return &cognitive.InferResponse{Text: `{"tool_call":{"name":"preview_config_document","arguments":{"format":"yaml","content":"kind: outcome_template"}}}`, Provider: "mock", ModelUsed: "test-model"}, nil
	}
	return &cognitive.InferResponse{Text: "Preview ready.", Provider: "mock", ModelUsed: "test-model"}, nil
}

func (p *directAnswerCorrectionProvider) Probe(context.Context) (bool, error) { return true, nil }

func TestProcessMessageStructured_DirectAnswerRejectsMutationAndUsesReadOnlyTool(t *testing.T) {
	provider := &directAnswerCorrectionProvider{}
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{
				"mock": {Type: "mock", Enabled: true, ModelID: "test-model"},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{"mock": provider},
	}

	exec := &countingToolExecutor{serverID: InternalServerID}
	agent := NewAgent(context.Background(), protocol.AgentManifest{
		ID: "admin", Role: "admin", Provider: "mock", Tools: []string{"delegate", "preview_config_document"},
	}, "admin-core", nil, router, exec)
	agent.SetToolDescriptions(map[string]string{
		"delegate": "Delegate work to another team.", "preview_config_document": "Validate an Outcome Template.",
	})

	result := agent.processMessageStructured(directAnswerRouteMarker+"\nPreview an Outcome Template.", nil)
	if exec.findCalls != 1 || exec.callCalls != 1 {
		t.Fatalf("tool execution calls = find:%d call:%d, want one read-only execution", exec.findCalls, exec.callCalls)
	}
	if len(result.ToolsUsed) != 1 || result.ToolsUsed[0] != "preview_config_document" {
		t.Fatalf("tools_used = %#v, want preview_config_document", result.ToolsUsed)
	}
	if provider.calls != 3 || len(provider.prompts) == 0 || strings.Contains(provider.prompts[0], "**delegate**") {
		t.Fatalf("direct-answer provider state = calls:%d prompts:%#v", provider.calls, provider.prompts)
	}
}

func TestProcessMessageStructured_CapturesConfigMutationArgumentsDuringProposalPlanning(t *testing.T) {
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{
				"mock": {Type: "mock", Enabled: true, ModelID: "test-model"},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{"mock": &proposalPlanningProvider{resp: &cognitive.InferResponse{
			Text: `{"tool_call":{"name":"store_config_document","arguments":{"format":"yaml","content":"model-authored-config"}}}`, Provider: "mock", ModelUsed: "test-model",
		}}},
	}

	exec := &countingToolExecutor{serverID: InternalServerID}
	agent := NewAgent(context.Background(), protocol.AgentManifest{
		ID: "admin", Role: "admin", Provider: "mock", Tools: []string{"store_config_document"},
	}, "admin-core", nil, router, exec)
	agent.SetToolDescriptions(map[string]string{"store_config_document": "Save a validated ConfigDocument revision."})

	result := agent.processMessageStructured("Save this outcome template.", nil)
	if exec.findCalls != 0 || exec.callCalls != 0 || len(result.PlannedToolCalls) != 1 {
		t.Fatalf("planning result = %#v, find=%d call=%d", result, exec.findCalls, exec.callCalls)
	}
	call := result.PlannedToolCalls[0]
	if call.Name != "store_config_document" || call.Arguments["content"] != "model-authored-config" {
		t.Fatalf("planned call = %#v", call)
	}
}

func TestCompositeToolExecutor_AllowsProposalPlanningConfigPreview(t *testing.T) {
	reg := NewInternalToolRegistry(InternalToolDeps{})
	composite := NewCompositeToolExecutor(reg, nil)
	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{PlanningOnly: true})
	output, err := composite.CallTool(ctx, InternalServerID, "preview_config_document", map[string]any{
		"format": "yaml", "content": `apiVersion: mycelis.ai/v1
kind: OutcomeTemplate
metadata:
  id: quote-preview
  name: Quote preview
  version: "1"
  owner_id: soma
  scope: {kind: workspace, ref: primary}
  enabled: true
  source: {kind: soma, ref: conversation:test}
  governance: {risk_level: low, approval_posture: optional}
spec:
  defaults: {target_outcome: Preview the supplied request}
`,
	})
	if err != nil || !strings.Contains(output, `"valid":true`) {
		t.Fatalf("preview output = %s, err=%v", output, err)
	}
}
