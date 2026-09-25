package swarm

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

type countingProviderStub struct{ calls *atomic.Int32 }

func (s countingProviderStub) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	s.calls.Add(1)
	return &cognitive.InferResponse{Text: "ok"}, nil
}

func (countingProviderStub) Probe(context.Context) (bool, error) { return true, nil }

func TestTeamNormalizeRuntimeProviderRoutingNeverRewritesNamedProviders(t *testing.T) {
	var ollamaCalls atomic.Int32
	brain := &cognitive.Router{
		Config: &cognitive.BrainConfig{Providers: map[string]cognitive.ProviderConfig{
			"ollama":           {Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local"},
			"local-ollama-dev": {Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local"},
		}},
		Adapters: map[string]cognitive.LLMProvider{"ollama": countingProviderStub{calls: &ollamaCalls}},
	}
	team := &Team{
		Manifest: &TeamManifest{
			ID: "admin-core", Name: "Soma", Provider: "local-ollama-dev",
			Members: []protocol.AgentManifest{
				{ID: "admin", Role: "admin"},
				{ID: "council-coder", Role: "coder", Provider: "local-ollama-dev"},
			},
		},
		brain: brain,
	}

	team.normalizeRuntimeProviderRouting()
	if team.Manifest.Provider != "local-ollama-dev" {
		t.Fatalf("team provider rewritten to %q", team.Manifest.Provider)
	}
	if team.Manifest.Members[0].Provider != "local-ollama-dev" {
		t.Fatalf("member without provider did not inherit team provider: %q", team.Manifest.Members[0].Provider)
	}
	if team.Manifest.Members[1].Provider != "local-ollama-dev" {
		t.Fatalf("named member provider rewritten to %q", team.Manifest.Members[1].Provider)
	}

	if got := brain.ExecutionAvailability("", "local-ollama-dev"); got.Available || got.Code != cognitive.ExecutionProviderDisabled {
		t.Fatalf("availability = %+v, want unavailable %s", got, cognitive.ExecutionProviderDisabled)
	}
	if _, err := brain.InferWithContract(context.Background(), cognitive.InferRequest{
		Provider: team.Manifest.Members[0].Provider, Prompt: "hello",
	}); err == nil {
		t.Fatal("inference against the disabled provider succeeded")
	}
	if n := ollamaCalls.Load(); n != 0 {
		t.Fatalf("%d calls silently reached ollama", n)
	}
}

// Even when cognitive would resolve a same-boundary fallback for the request,
// the manifest keeps the named provider; substitution stays per-request.
func TestTeamNormalizeRuntimeProviderRoutingIgnoresConfiguredFallbacks(t *testing.T) {
	var ollamaCalls atomic.Int32
	brain := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Providers: map[string]cognitive.ProviderConfig{
				"ollama":           {Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local"},
				"local-ollama-dev": {Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local"},
			},
			ProfileFallbacks: map[string][]string{"": {"ollama"}},
		},
		Adapters: map[string]cognitive.LLMProvider{"ollama": countingProviderStub{calls: &ollamaCalls}},
	}
	if got := brain.ExecutionAvailability("", "local-ollama-dev"); !got.FallbackApplied {
		t.Fatalf("fixture must resolve a same-boundary fallback: %+v", got)
	}
	team := &Team{
		Manifest: &TeamManifest{ID: "admin-core", Name: "Soma", Provider: "local-ollama-dev",
			Members: []protocol.AgentManifest{{ID: "coder", Role: "coder", Provider: "local-ollama-dev"}}},
		brain: brain,
	}
	team.normalizeRuntimeProviderRouting()
	if team.Manifest.Provider != "local-ollama-dev" || team.Manifest.Members[0].Provider != "local-ollama-dev" {
		t.Fatalf("manifest provider rewritten: %#v", team.Manifest)
	}
}
