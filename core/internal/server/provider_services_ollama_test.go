package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

func TestHandleServicesStatus_OllamaDegradedWhenMissingProvider(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"vllm": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"vllm": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)

	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "ollama" {
			if svc["status"] != "degraded" {
				t.Errorf("expected ollama=degraded when provider missing, got %v", svc["status"])
			}
			return
		}
	}
	t.Error("ollama service entry not found")
}

func TestHandleServicesStatus_OllamaDegradedWhenDisabled(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: false},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)

	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "ollama" {
			if svc["status"] != "degraded" {
				t.Errorf("expected ollama=degraded when provider disabled, got %v", svc["status"])
			}
			return
		}
	}
	t.Error("ollama service entry not found")
}

func TestHandleServicesStatus_OllamaDegradedWhenAdapterMissing(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)

	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "ollama" {
			if svc["status"] != "degraded" {
				t.Errorf("expected ollama=degraded when adapter missing, got %v", svc["status"])
			}
			return
		}
	}
	t.Error("ollama service entry not found")
}

func TestHandleServicesStatus_OllamaOnline(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)

	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "ollama" {
			if svc["status"] != "online" {
				t.Errorf("expected ollama=online, got %v", svc["status"])
			}
			detail, _ := svc["detail"].(string)
			if detail == "" {
				t.Errorf("expected non-empty detail for ollama online status")
			}
			return
		}
	}
	t.Error("ollama service entry not found")
}
