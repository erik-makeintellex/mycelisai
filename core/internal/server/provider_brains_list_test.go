package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

// ════════════════════════════════════════════════════════════════════
// BRAINS — HandleListBrains (GET /api/v1/brains)
// ════════════════════════════════════════════════════════════════════

func TestHandleListBrains_HappyPath(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
			"vllm":   {Type: "openai_compatible", Endpoint: "http://localhost:8000/v1", ModelID: "mixtral", Enabled: false},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "GET /api/v1/brains", s.HandleListBrains)
	rr := doRequest(t, mux, "GET", "/api/v1/brains", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 2 {
		t.Errorf("expected 2 brains, got %d", len(data))
	}

	// Check that disabled provider has status "disabled"
	found := false
	for _, item := range data {
		brain := item.(map[string]any)
		if brain["id"] == "vllm" {
			found = true
			if brain["status"] != "disabled" {
				t.Errorf("expected vllm status=disabled, got %v", brain["status"])
			}
		}
	}
	if !found {
		t.Error("expected to find vllm in response")
	}
}

func TestHandleListBrains_NilCognitive(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/brains", s.HandleListBrains)
	rr := doRequest(t, mux, "GET", "/api/v1/brains", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 0 {
		t.Errorf("expected 0 brains for nil cognitive, got %d", len(data))
	}
}

func TestHandleListBrains_DefaultsForEmptyFields(t *testing.T) {
	// Provider with empty location/boundary/policy should get defaults
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"bare": {Type: "openai_compatible", ModelID: "test", Enabled: false},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "GET /api/v1/brains", s.HandleListBrains)
	rr := doRequest(t, mux, "GET", "/api/v1/brains", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)
	brain := data[0].(map[string]any)
	if brain["location"] != "local" {
		t.Errorf("expected default location=local, got %v", brain["location"])
	}
	if brain["data_boundary"] != "local_only" {
		t.Errorf("expected default data_boundary=local_only, got %v", brain["data_boundary"])
	}
	if brain["usage_policy"] != "local_first" {
		t.Errorf("expected default usage_policy=local_first, got %v", brain["usage_policy"])
	}
	if brain["token_budget_profile"] != "standard" {
		t.Errorf("expected default token_budget_profile=standard, got %v", brain["token_budget_profile"])
	}
	if brain["max_output_tokens"] != float64(1024) {
		t.Errorf("expected default max_output_tokens=1024, got %v", brain["max_output_tokens"])
	}
}
