package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

// ════════════════════════════════════════════════════════════════════
// SERVICES — HandleServicesStatus (GET /api/v1/services/status)
// ════════════════════════════════════════════════════════════════════

func TestHandleServicesStatus_AllOffline(t *testing.T) {
	// No DB, no Cognitive, no Reactive → everything offline/degraded
	s := newTestServer()

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")

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
	// Should have 7 services: nats, postgres, cognitive, ollama, reactive, comms, groups_bus
	if len(data) != 7 {
		t.Fatalf("expected 7 services, got %d", len(data))
	}

	// All should be offline since nothing is wired
	statusMap := map[string]string{}
	for _, item := range data {
		svc := item.(map[string]any)
		statusMap[svc["name"].(string)] = svc["status"].(string)
	}

	if statusMap["nats"] != "offline" {
		t.Errorf("expected nats=offline, got %v", statusMap["nats"])
	}
	if statusMap["postgres"] != "offline" {
		t.Errorf("expected postgres=offline, got %v", statusMap["postgres"])
	}
	if statusMap["cognitive"] != "offline" {
		t.Errorf("expected cognitive=offline, got %v", statusMap["cognitive"])
	}
	if statusMap["ollama"] != "offline" {
		t.Errorf("expected ollama=offline, got %v", statusMap["ollama"])
	}
	if statusMap["reactive"] != "offline" {
		t.Errorf("expected reactive=offline, got %v", statusMap["reactive"])
	}
	if statusMap["comms"] != "offline" {
		t.Errorf("expected comms=offline, got %v", statusMap["comms"])
	}
	if statusMap["groups_bus"] != "offline" {
		t.Errorf("expected groups_bus=offline, got %v", statusMap["groups_bus"])
	}
}

func TestHandleServicesStatus_CognitiveDegraded(t *testing.T) {
	// Cognitive wired but all providers disabled → degraded
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: false},
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

	// Find cognitive entry
	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "cognitive" {
			if svc["status"] != "degraded" {
				t.Errorf("expected cognitive=degraded (no enabled providers), got %v", svc["status"])
			}
			return
		}
	}
	t.Error("cognitive service entry not found in response")
}

func TestHandleServicesStatus_CognitiveOnline(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
			"vllm":   {Type: "openai_compatible", Enabled: true},
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
		if svc["name"] == "cognitive" {
			if svc["status"] != "online" {
				t.Errorf("expected cognitive=online, got %v", svc["status"])
			}
			detail := svc["detail"].(string)
			if detail != "2/2 providers enabled" {
				t.Errorf("expected detail '2/2 providers enabled', got %v", detail)
			}
			return
		}
	}
	t.Error("cognitive service entry not found")
}
