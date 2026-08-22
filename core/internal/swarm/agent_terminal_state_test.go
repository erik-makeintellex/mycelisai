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

type recoveringTerminalStateProvider struct {
	calls int
}

func (p *recoveringTerminalStateProvider) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	p.calls++
	text := ""
	if p.calls > 1 {
		text = "Recovered response."
	}
	return &cognitive.InferResponse{Text: text, ModelUsed: "test-model", Provider: "mock"}, nil
}

func (p *recoveringTerminalStateProvider) Probe(context.Context) (bool, error) {
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

func TestProcessMessageStructured_InferenceErrorOverridesConfiguredAvailability(t *testing.T) {
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{
				"mock": {Type: "mock", Enabled: true, ModelID: "test-model"},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"mock": &terminalStateProvider{err: context.DeadlineExceeded},
		},
	}

	agent := NewAgent(context.Background(), protocol.AgentManifest{ID: "worker", Role: "worker"}, "delivery-team", nil, router, nil)
	result := agent.processMessageStructured("produce the retained package", nil)

	if result.Availability == nil {
		t.Fatal("expected structured availability blocker")
	}
	if result.Availability.Available {
		t.Fatal("inference failure must not report the configured provider as available")
	}
	if result.Availability.Code != "provider_inference_failed" {
		t.Fatalf("availability.code = %q, want provider_inference_failed", result.Availability.Code)
	}
	if result.Availability.RecommendedAction == "" {
		t.Fatal("expected a recovery action")
	}
}

func TestProcessMessageStructured_RetriesEmptyInitialProviderOutputOnce(t *testing.T) {
	provider := &recoveringTerminalStateProvider{}
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{
				"mock": {Type: "mock", Enabled: true, ModelID: "test-model"},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{"mock": provider},
	}

	agent := NewAgent(context.Background(), protocol.AgentManifest{ID: "worker", Role: "worker"}, "delivery-team", nil, router, nil)
	result := agent.processMessageStructured("produce the retained package", nil)

	if provider.calls != 2 {
		t.Fatalf("provider calls = %d, want one bounded retry", provider.calls)
	}
	if result.Availability != nil {
		t.Fatalf("unexpected availability blocker: %+v", result.Availability)
	}
	if result.Text != "Recovered response." {
		t.Fatalf("text = %q, want recovered response", result.Text)
	}
}
