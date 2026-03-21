package cognitive

import (
	"context"
	"testing"
)

func TestInferWithContract_ProviderOverride(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Profiles: map[string]string{
				"chat": "provider-a",
			},
		},
		Adapters: map[string]LLMProvider{
			"provider-a": &MockProvider{OutputSequence: []string{"from-a"}},
			"provider-b": &MockProvider{OutputSequence: []string{"from-b"}},
		},
	}

	resp, err := r.InferWithContract(context.Background(), InferRequest{
		Profile:  "chat",
		Provider: "provider-b",
		Prompt:   "hello",
	})
	if err != nil {
		t.Fatalf("InferWithContract: %v", err)
	}
	if resp.Text != "from-b" {
		t.Fatalf("expected override response from provider-b, got %q", resp.Text)
	}
}

func TestInferWithContract_ProviderOverrideMissing(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Profiles: map[string]string{
				"chat": "provider-a",
			},
		},
		Adapters: map[string]LLMProvider{
			"provider-a": &MockProvider{OutputSequence: []string{"from-a"}},
		},
	}

	_, err := r.InferWithContract(context.Background(), InferRequest{
		Profile:  "chat",
		Provider: "provider-x",
		Prompt:   "hello",
	})
	if err == nil {
		t.Fatal("expected error for missing explicit provider override")
	}
}
