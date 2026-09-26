package swarm

import (
	"strings"
	"sync"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

func routingSnapshotTestRouter() *cognitive.Router {
	return &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Providers: map[string]cognitive.ProviderConfig{
				"mock":   {Type: "mock", Enabled: true, ModelID: "contract-test", Endpoint: "http://127.0.0.1:9/v1", MaxOutputTokens: 321},
				"backup": {Type: "mock", Enabled: true, ModelID: "backup-model"},
				"off":    {Type: "mock", Enabled: false, ModelID: "off-model"},
			},
			Profiles:         map[string]string{"chat": "mock", "coder": "off"},
			ProfileFallbacks: map[string][]string{"coder": {"backup"}},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"mock": &boundedInferenceProvider{response: "done"}, "backup": &boundedInferenceProvider{}, "off": &boundedInferenceProvider{},
		},
	}
}

func TestInferenceLogConfig_UsesLockedEffectiveBinding(t *testing.T) {
	agent := &Agent{brain: routingSnapshotTestRouter()}
	cases := []struct {
		req              cognitive.InferRequest
		provider, model  string
		maxOutputAtLeast int
	}{
		{cognitive.InferRequest{Profile: "chat"}, "mock", "contract-test", 321},
		{cognitive.InferRequest{Profile: "coder"}, "backup", "backup-model", 1}, // fallback the router executes
		{cognitive.InferRequest{Provider: "off"}, "off", "off-model", 1},        // explicit provider wins
		{cognitive.InferRequest{Profile: "unknown"}, "auto", "unknown", 0},
		{cognitive.InferRequest{}, "auto", "unknown", 0},
	}
	for _, tc := range cases {
		provider, model, maxOutput := agent.inferenceLogConfig(tc.req)
		if provider != tc.provider || model != tc.model || maxOutput < tc.maxOutputAtLeast {
			t.Fatalf("%#v: got (%s, %s, %d), want (%s, %s, >=%d)", tc.req, provider, model, maxOutput, tc.provider, tc.model, tc.maxOutputAtLeast)
		}
	}
	if provider, model, _ := (&Agent{}).inferenceLogConfig(cognitive.InferRequest{Profile: "chat"}); provider != "auto" || model != "unknown" {
		t.Fatalf("nil brain: got (%s, %s)", provider, model)
	}
}

func TestWriteCognitiveStatus_ReadsSnapshot(t *testing.T) {
	var sb strings.Builder
	(&InternalToolRegistry{brain: routingSnapshotTestRouter()}).writeCognitiveStatus(&sb)
	out := sb.String()
	if !strings.Contains(out, "**mock** (mock): model=`contract-test`") || !strings.Contains(out, "chat→mock") {
		t.Fatalf("cognitive status missing routing: %q", out)
	}
	sb.Reset()
	(&InternalToolRegistry{}).writeCognitiveStatus(&sb)
	if !strings.Contains(sb.String(), "Cognitive engine offline") {
		t.Fatalf("offline brain: %q", sb.String())
	}
}

// TestSwarmRoutingReads_RaceFreeUnderMutation is the swarm half of the S6d
// race regression: -race is the assertion.
func TestSwarmRoutingReads_RaceFreeUnderMutation(t *testing.T) {
	router := routingSnapshotTestRouter()
	agent := &Agent{brain: router}
	registry := &InternalToolRegistry{brain: router}
	var wg sync.WaitGroup
	run := func(fn func(i int)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				fn(i)
			}
		}()
	}
	run(func(i int) {
		target := []string{"mock", "backup"}[i%2]
		_ = router.SetProfileOverrides(map[string]string{"chat": target}, cognitive.ProfileOriginRuntime)
	})
	run(func(int) { router.ClearProfileOverride("chat") })
	run(func(i int) {
		cfg, _ := router.ProviderSnapshot("off")
		cfg.Enabled = i%2 == 0
		router.StoreProviderConfig("off", cfg)
		router.StoreProviderConfig("new", cognitive.ProviderConfig{Type: "mock", ModelID: "n"})
	})
	run(func(int) {
		agent.inferenceLogConfig(cognitive.InferRequest{Profile: "chat"})
		agent.inferenceLogConfig(cognitive.InferRequest{Profile: "coder"})
		agent.inferenceLogConfig(cognitive.InferRequest{Provider: "off"})
	})
	run(func(int) {
		var sb strings.Builder
		registry.writeCognitiveStatus(&sb)
	})
	wg.Wait()
}
