package cognitive

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type deadAdapter struct {
	probeCalls int
	inferCalls int
}

func (a *deadAdapter) Infer(context.Context, string, InferOptions) (*InferResponse, error) {
	a.inferCalls++
	return nil, errors.New("dial tcp 127.0.0.1:8000: connect: connection refused")
}

func (a *deadAdapter) Probe(context.Context) (bool, error) {
	a.probeCalls++
	return false, errors.New("dial tcp 127.0.0.1:8000: connect: connection refused")
}

// TestInferWithContract_UnavailableProviderFailsClosedWithNormalizedError
// proves the negative case from Packet L: vLLM (or any bound provider) down,
// with another provider configured but NOT in an explicit ProfileFallbacks
// list, must fail with a normalized blocker and must never call the other
// provider.
func TestInferWithContract_UnavailableProviderFailsClosedWithNormalizedError(t *testing.T) {
	dead := &deadAdapter{}
	other := &captureAdapter{}
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"vllm": {
					Type: "openai_compatible", ModelID: "qwen2.5-coder-14b", Enabled: true,
					DataBoundary: "local_only", Location: "local",
				},
				"ollama": {
					Type: "openai_compatible", ModelID: "qwen2.5-coder:7b", Enabled: true,
					DataBoundary: "local_only", Location: "local",
				},
			},
			Profiles: map[string]string{"chat": "vllm"},
			// No ProfileFallbacks configured: ollama must never be called.
		},
		Adapters: map[string]LLMProvider{"vllm": dead, "ollama": other},
	}

	resp, err := r.InferWithContract(context.Background(), InferRequest{Profile: "chat", Prompt: "hello"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if resp != nil {
		t.Fatalf("expected nil response on failure, got %+v", resp)
	}
	if !errors.Is(err, ErrAIEngineUnavailable) {
		t.Fatalf("error = %q, want ErrAIEngineUnavailable", err.Error())
	}
	if !strings.Contains(err.Error(), "AI engine unavailable") {
		t.Fatalf("error = %q, want normalized \"AI engine unavailable\" blocker", err.Error())
	}
	if strings.Contains(err.Error(), "127.0.0.1") || strings.Contains(err.Error(), "dial") || strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("error leaked adapter/transport detail: %q", err.Error())
	}
	if dead.probeCalls != 1 {
		t.Fatalf("expected exactly one health probe on the dead provider, got %d", dead.probeCalls)
	}
	if other.calls != 0 {
		t.Fatalf("ollama must not be called without an explicit fallback list, got %d calls", other.calls)
	}
	if r.Config.Profiles["chat"] != "vllm" {
		t.Fatalf("chat profile rerouted to %q without operator configuration", r.Config.Profiles["chat"])
	}
}

// TestInferWithContract_ExplicitSameBoundaryFallbackIsUsed proves the
// positive case: an operator-configured, same-data-boundary fallback list
// lets Core substitute a different provider, and the substitution is
// recorded via the resolved provider on the response.
func TestInferWithContract_ExplicitSameBoundaryFallbackIsUsed(t *testing.T) {
	other := &captureAdapter{}
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"vllm": {
					Type: "openai_compatible", ModelID: "qwen2.5-coder-14b", Enabled: false, // not executable
					DataBoundary: "local_only", Location: "local",
				},
				"ollama": {
					Type: "openai_compatible", ModelID: "qwen2.5-coder:7b", Enabled: true,
					DataBoundary: "local_only", Location: "local",
				},
			},
			Profiles: map[string]string{"chat": "vllm"},
			ProfileFallbacks: map[string][]string{
				"chat": {"ollama"},
			},
		},
		Adapters: map[string]LLMProvider{"ollama": other},
	}

	resp, err := r.InferWithContract(context.Background(), InferRequest{Profile: "chat", Prompt: "hello"})
	if err != nil {
		t.Fatalf("InferWithContract: %v", err)
	}
	if resp.Provider != "ollama" {
		t.Fatalf("resp.Provider = %q, want ollama (explicit fallback recorded)", resp.Provider)
	}
	if other.calls != 1 {
		t.Fatalf("expected fallback adapter to be called once, got %d", other.calls)
	}
}
