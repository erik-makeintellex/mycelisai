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

// TestInferWithContract_ProviderOverrideDisabledFailsClosedWithoutExplicitList
// proves the negative case: an explicit provider override that is disabled
// must fail closed with no substitution when no ProfileFallbacks entry names
// an alternate — there is no more ambient/silent fallback.
func TestInferWithContract_ProviderOverrideDisabledFailsClosedWithoutExplicitList(t *testing.T) {
	other := &MockProvider{OutputSequence: []string{"from-b"}}
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"provider-a": {Enabled: false, ModelID: "disabled-model", Location: "local"},
				"provider-b": {Enabled: true, ModelID: "fallback-model", Location: "local"},
			},
			Profiles: map[string]string{
				"chat": "provider-a",
			},
		},
		Adapters: map[string]LLMProvider{
			"provider-b": other,
		},
	}

	_, err := r.InferWithContract(context.Background(), InferRequest{
		Profile:  "chat",
		Provider: "provider-a",
		Prompt:   "hello",
	})
	if err == nil {
		t.Fatal("expected error: disabled provider override must fail closed without an explicit fallback list")
	}
}

// TestInferWithContract_ProviderOverrideFallsBackWithExplicitSameBoundaryList
// proves the positive case: the same scenario succeeds once the operator
// configures an explicit, same-data-boundary ProfileFallbacks entry.
func TestInferWithContract_ProviderOverrideFallsBackWithExplicitSameBoundaryList(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"provider-a": {Enabled: false, ModelID: "disabled-model", Location: "local", DataBoundary: "local_only"},
				"provider-b": {Enabled: true, ModelID: "fallback-model", Location: "local", DataBoundary: "local_only"},
			},
			Profiles: map[string]string{
				"chat": "provider-a",
			},
			ProfileFallbacks: map[string][]string{
				"chat": {"provider-b"},
			},
		},
		Adapters: map[string]LLMProvider{
			"provider-b": &MockProvider{OutputSequence: []string{"from-b"}},
		},
	}

	resp, err := r.InferWithContract(context.Background(), InferRequest{
		Profile:  "chat",
		Provider: "provider-a",
		Prompt:   "hello",
	})
	if err != nil {
		t.Fatalf("InferWithContract: %v", err)
	}
	if resp.Text != "from-b" {
		t.Fatalf("expected fallback response from provider-b, got %q", resp.Text)
	}
}
