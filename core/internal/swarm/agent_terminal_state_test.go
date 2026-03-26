package swarm

import (
	"context"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

type terminalStateProvider struct {
	resp *cognitive.InferResponse
	err  error
}

func (p *terminalStateProvider) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return p.resp, p.err
}

func (p *terminalStateProvider) Probe(context.Context) (bool, error) {
	return true, nil
}

func TestProcessMessageStructured_EmptyProviderOutputBecomesBlocker(t *testing.T) {
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{
				"mock": {Type: "mock", Enabled: true, ModelID: "test-model"},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"mock": &terminalStateProvider{
				resp: &cognitive.InferResponse{
					Text:      "",
					ModelUsed: "test-model",
					Provider:  "mock",
				},
			},
		},
	}

	agent := NewAgent(context.Background(), protocol.AgentManifest{ID: "admin", Role: "admin"}, "admin-core", nil, router, nil)
	result := agent.processMessageStructured("summarize the workspace", nil)

	if result.Availability == nil {
		t.Fatal("expected structured availability blocker")
	}
	if result.Availability.Code != "empty_provider_output" {
		t.Fatalf("availability.code = %q, want empty_provider_output", result.Availability.Code)
	}
	if result.Text != "" {
		t.Fatalf("text = %q, want empty", result.Text)
	}
	if result.ProviderID != "mock" {
		t.Fatalf("provider_id = %q, want mock", result.ProviderID)
	}
	if result.ModelUsed != "test-model" {
		t.Fatalf("model_used = %q, want test-model", result.ModelUsed)
	}
}
