package cognitive

import (
	"context"
	"testing"
)

type captureAdapter struct {
	lastOpts InferOptions
}

func (a *captureAdapter) Infer(_ context.Context, _ string, opts InferOptions) (*InferResponse, error) {
	a.lastOpts = opts
	return &InferResponse{Text: "ok", Provider: "capture", ModelUsed: "capture"}, nil
}

func (a *captureAdapter) Probe(_ context.Context) (bool, error) {
	return true, nil
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
