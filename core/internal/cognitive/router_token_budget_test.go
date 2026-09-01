package cognitive

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type captureAdapter struct {
	lastOpts InferOptions
	response *InferResponse
	calls    int
}

type failingGatewayAdapter struct {
	probeCalls int
}

func (a *failingGatewayAdapter) Infer(context.Context, string, InferOptions) (*InferResponse, error) {
	return nil, errors.New("gateway inference failed")
}

func (a *failingGatewayAdapter) Probe(context.Context) (bool, error) {
	a.probeCalls++
	return false, errors.New("gateway unavailable")
}

func (a *captureAdapter) Infer(_ context.Context, _ string, opts InferOptions) (*InferResponse, error) {
	a.calls++
	a.lastOpts = opts
	if a.response != nil {
		return a.response, nil
	}
	return &InferResponse{Text: "ok", Provider: "capture", ModelUsed: "capture"}, nil
}

func (a *captureAdapter) Probe(_ context.Context) (bool, error) {
	return true, nil
}

func TestInferWithContract_PreservesRouteModelAndReportedUsage(t *testing.T) {
	adapter := &captureAdapter{response: &InferResponse{
		Text:               "gateway response",
		Provider:           "openai_compatible",
		ModelUsed:          "openai/gpt-4.1-mini-2025-04-14",
		UpstreamResponseID: "chatcmpl-router-proof",
		PromptTokens:       14,
		CompletionTokens:   9,
		TokensUsed:         23,
	}}
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"litellm": {Type: "openai_compatible", ModelID: "team-coder", Enabled: true},
			},
			Profiles: map[string]string{"chat": "litellm"},
		},
		Adapters: map[string]LLMProvider{"litellm": adapter},
	}

	resp, err := r.InferWithContract(context.Background(), InferRequest{Profile: "chat", Prompt: "hello"})
	if err != nil {
		t.Fatalf("InferWithContract: %v", err)
	}
	if resp.Provider != "litellm" {
		t.Fatalf("Provider = %q, want configured route identity litellm", resp.Provider)
	}
	if resp.ModelUsed != "openai/gpt-4.1-mini-2025-04-14" {
		t.Fatalf("ModelUsed = %q, want gateway response model", resp.ModelUsed)
	}
	if resp.UpstreamResponseID != "chatcmpl-router-proof" || resp.PromptTokens != 14 || resp.CompletionTokens != 9 {
		t.Fatalf("upstream accounting changed during route normalization: %#v", resp)
	}
	if resp.TokensUsed != 23 || r.totalTokens.Load() != 23 {
		t.Fatalf("reported tokens = %d, recorded tokens = %d, want 23", resp.TokensUsed, r.totalTokens.Load())
	}
}

func TestInferWithContract_DoesNotEstimateMissingUsage(t *testing.T) {
	adapter := &captureAdapter{response: &InferResponse{
		Text:      strings.Repeat("long response ", 100),
		Provider:  "openai_compatible",
		ModelUsed: "gateway-model",
	}}
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"litellm": {Type: "openai_compatible", ModelID: "team-coder", Enabled: true},
			},
			Profiles: map[string]string{"chat": "litellm"},
		},
		Adapters: map[string]LLMProvider{"litellm": adapter},
	}

	resp, err := r.InferWithContract(context.Background(), InferRequest{Profile: "chat", Prompt: "hello"})
	if err != nil {
		t.Fatalf("InferWithContract: %v", err)
	}
	if resp.TokensUsed != 0 || r.totalTokens.Load() != 0 {
		t.Fatalf("missing usage was estimated: response %d, recorded %d", resp.TokensUsed, r.totalTokens.Load())
	}
}

func TestInferWithContract_ModelGatewayFailureDoesNotCrossBoundary(t *testing.T) {
	gateway := &failingGatewayAdapter{}
	local := &captureAdapter{}
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"gateway": {
					Type: "openai_compatible", ModelID: "mycelis-default", Enabled: true,
					ModelGateway: true, DataBoundary: "leaves_org", Location: "remote",
				},
				"ollama": {
					Type: "openai_compatible", ModelID: "qwen3:8b", Enabled: true,
					DataBoundary: "local_only", Location: "local",
				},
			},
			Profiles: map[string]string{"chat": "gateway"},
		},
		Adapters: map[string]LLMProvider{"gateway": gateway, "ollama": local},
	}

	_, err := r.InferWithContract(context.Background(), InferRequest{Profile: "chat", Prompt: "hello"})
	if err == nil || err.Error() != "gateway inference failed" {
		t.Fatalf("InferWithContract error = %v, want original gateway failure", err)
	}
	if gateway.probeCalls != 0 || local.calls != 0 {
		t.Fatalf("gateway probe calls = %d, local inference calls = %d; gateway failure must fail closed", gateway.probeCalls, local.calls)
	}
	if r.Config.Profiles["chat"] != "gateway" {
		t.Fatalf("chat profile rerouted to %q across gateway boundary", r.Config.Profiles["chat"])
	}
}

func TestInferWithContract_UsesProviderTokenBudgetDefaults(t *testing.T) {
	adapter := &captureAdapter{}
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"local-ollama-dev": {
					Type:               "ollama",
					ModelID:            "qwen3:8b",
					TokenBudgetProfile: TokenBudgetExtended,
					MaxOutputTokens:    2048,
					Enabled:            true,
				},
			},
			Profiles: map[string]string{
				"chat": "local-ollama-dev",
			},
		},
		Adapters: map[string]LLMProvider{
			"local-ollama-dev": adapter,
		},
	}

	_, err := r.InferWithContract(context.Background(), InferRequest{
		Profile: "chat",
		Prompt:  "hello",
	})
	if err != nil {
		t.Fatalf("InferWithContract: %v", err)
	}

	if adapter.lastOpts.MaxTokens != 2048 {
		t.Fatalf("expected MaxTokens=2048, got %d", adapter.lastOpts.MaxTokens)
	}
}
