package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

// ════════════════════════════════════════════════════════════════════
// BRAINS — HandleDeleteBrain (DELETE /api/v1/brains/{id})
// ════════════════════════════════════════════════════════════════════

func TestHandleDeleteBrain_HappyPath(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
			"vllm":   {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
			"vllm":   &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "DELETE /api/v1/brains/{id}", s.HandleDeleteBrain)
	rr := doRequest(t, mux, "DELETE", "/api/v1/brains/vllm", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["deleted"] != true {
		t.Errorf("expected deleted=true, got %v", data["deleted"])
	}
}

func TestHandleDeleteBrain_NotFound(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "DELETE /api/v1/brains/{id}", s.HandleDeleteBrain)
	rr := doRequest(t, mux, "DELETE", "/api/v1/brains/ghost", "")

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleDeleteBrain_LastProvider(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "DELETE /api/v1/brains/{id}", s.HandleDeleteBrain)
	rr := doRequest(t, mux, "DELETE", "/api/v1/brains/ollama", "")

	assertStatus(t, rr, http.StatusConflict)
}

func TestHandleDeleteBrain_NilCognitive(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "DELETE /api/v1/brains/{id}", s.HandleDeleteBrain)
	rr := doRequest(t, mux, "DELETE", "/api/v1/brains/x", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ════════════════════════════════════════════════════════════════════
// BRAINS — HandleProbeBrain (POST /api/v1/brains/{id}/probe)
// ════════════════════════════════════════════════════════════════════

func TestHandleProbeBrain_Healthy(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "POST /api/v1/brains/{id}/probe", s.HandleProbeBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains/ollama/probe", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["alive"] != true {
		t.Errorf("expected alive=true, got %v", data["alive"])
	}
	if data["id"] != "ollama" {
		t.Errorf("expected id=ollama, got %v", data["id"])
	}
}

func TestHandleProbeBrain_Unhealthy(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: false},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "POST /api/v1/brains/{id}/probe", s.HandleProbeBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains/ollama/probe", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["alive"] != false {
		t.Errorf("expected alive=false, got %v", data["alive"])
	}
}

func TestHandleProbeBrain_NotFound(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "POST /api/v1/brains/{id}/probe", s.HandleProbeBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains/ghost/probe", "")

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleProbeBrain_NilCognitive(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/brains/{id}/probe", s.HandleProbeBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains/ollama/probe", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}
