package cognitive

import (
	"context"
	"testing"
)

type startupProbeStub struct {
	probeCalls int
	healthy    bool
}

func (s *startupProbeStub) Infer(ctx context.Context, prompt string, opts InferOptions) (*InferResponse, error) {
	return &InferResponse{Text: "ok", Provider: "stub", ModelUsed: "stub"}, nil
}

func (s *startupProbeStub) Probe(ctx context.Context) (bool, error) {
	s.probeCalls++
	return s.healthy, nil
}

func TestStartupProbeProviderIDs_DefaultOllamaAndProfileRouted(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Profiles: map[string]string{
				"chat": "ollama",
			},
		},
		Adapters: map[string]LLMProvider{
			"ollama":   &startupProbeStub{healthy: true},
			"vllm":     &startupProbeStub{healthy: true},
			"lmstudio": &startupProbeStub{healthy: true},
		},
	}

	ids := r.startupProbeProviderIDs()
	if _, ok := ids["ollama"]; !ok {
		t.Fatal("expected ollama in startup probe set")
	}
	if _, ok := ids["vllm"]; ok {
		t.Fatal("did not expect vllm in startup probe set without profile routing")
	}
	if _, ok := ids["lmstudio"]; ok {
		t.Fatal("did not expect lmstudio in startup probe set without profile routing")
	}
}

func TestAutoConfigureStartup_ProbesOnlyScopedProviders(t *testing.T) {
	ollama := &startupProbeStub{healthy: true}
	vllm := &startupProbeStub{healthy: true}
	lmstudio := &startupProbeStub{healthy: true}

	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"ollama":   {ModelID: "qwen2.5-coder:7b"},
				"vllm":     {ModelID: "qwen2.5-coder"},
				"lmstudio": {ModelID: "default"},
			},
			Profiles: map[string]string{
				"chat": "ollama",
			},
		},
		Adapters: map[string]LLMProvider{
			"ollama":   ollama,
			"vllm":     vllm,
			"lmstudio": lmstudio,
		},
	}

	r.AutoConfigureStartup(context.Background())

	if ollama.probeCalls == 0 {
		t.Fatal("expected ollama to be probed during startup")
	}
	if vllm.probeCalls != 0 {
		t.Fatalf("expected vllm probeCalls=0, got %d", vllm.probeCalls)
	}
	if lmstudio.probeCalls != 0 {
		t.Fatalf("expected lmstudio probeCalls=0, got %d", lmstudio.probeCalls)
	}
}

func TestAutoConfigureStartup_ProbesAdditionalProfileProvider(t *testing.T) {
	ollama := &startupProbeStub{healthy: true}
	vllm := &startupProbeStub{healthy: true}

	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"ollama": {ModelID: "qwen2.5-coder:7b"},
				"vllm":   {ModelID: "qwen2.5-coder"},
			},
			Profiles: map[string]string{
				"chat":  "ollama",
				"coder": "vllm",
			},
		},
		Adapters: map[string]LLMProvider{
			"ollama": ollama,
			"vllm":   vllm,
		},
	}

	r.AutoConfigureStartup(context.Background())

	if ollama.probeCalls == 0 {
		t.Fatal("expected ollama to be probed")
	}
	if vllm.probeCalls == 0 {
		t.Fatal("expected vllm to be probed when explicitly profile-routed")
	}
}
