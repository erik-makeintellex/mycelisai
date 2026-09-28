package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

// phe probe-failing adapter for F19: returns an unhealthy/error probe result so
// tests can assert HandleServicesStatus never reports "online" without probing.
type pheProbeFailingAdapter struct {
	err error
}

func (a *pheProbeFailingAdapter) Infer(_ context.Context, _ string, _ cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return nil, a.err
}

func (a *pheProbeFailingAdapter) Probe(_ context.Context) (bool, error) {
	return false, a.err
}

func pheOllamaStatus(t *testing.T, s *AdminServer) map[string]any {
	t.Helper()
	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)
	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "ollama" {
			return svc
		}
	}
	t.Fatal("ollama service entry not found")
	return nil
}

// TestPheServicesStatus_OllamaOfflineWhenProbeFails is the negative/adversarial
// case for F19: an enabled provider with an initialized adapter that fails its
// health probe must never be reported "online".
func TestPheServicesStatus_OllamaOfflineWhenProbeFails(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &pheProbeFailingAdapter{},
		},
	)
	s := newTestServer(cogOpt)

	svc := pheOllamaStatus(t, s)
	if svc["status"] != "offline" {
		t.Fatalf("expected ollama=offline when the probe fails, got %v", svc["status"])
	}
}

// TestPheServicesStatus_OllamaOnlineWhenProbeSucceeds is the positive case: a
// successful probe reports "online".
func TestPheServicesStatus_OllamaOnlineWhenProbeSucceeds(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	svc := pheOllamaStatus(t, s)
	if svc["status"] != "online" {
		t.Fatalf("expected ollama=online when the probe succeeds, got %v", svc["status"])
	}
}
