package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

// ════════════════════════════════════════════════════════════════════
// BRAINS — HandleAddBrain (POST /api/v1/brains)
// ════════════════════════════════════════════════════════════════════

func TestHandleAddBrain_HappyPath(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	body := `{
		"id": "vllm-local",
		"type": "openai_compatible",
		"endpoint": "http://localhost:8000/v1",
		"model_id": "mixtral",
		"token_budget_profile": "extended",
		"max_output_tokens": 2048,
		"enabled": true
	}`

	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", body)

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
	if data["id"] != "vllm-local" {
		t.Errorf("expected id=vllm-local, got %v", data["id"])
	}
	if data["type"] != "openai_compatible" {
		t.Errorf("expected type=openai_compatible, got %v", data["type"])
	}
	// Location should default to "local"
	if data["location"] != "local" {
		t.Errorf("expected location=local, got %v", data["location"])
	}
	if data["data_boundary"] != "local_only" {
		t.Errorf("expected data_boundary=local_only, got %v", data["data_boundary"])
	}
	if data["token_budget_profile"] != "extended" {
		t.Errorf("expected token_budget_profile=extended, got %v", data["token_budget_profile"])
	}
	if data["max_output_tokens"] != float64(2048) {
		t.Errorf("expected max_output_tokens=2048, got %v", data["max_output_tokens"])
	}
}

func TestHandleAddBrain_BadJSON(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", "not-json")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleAddBrain_MissingID(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	body := `{"type": "openai_compatible", "endpoint": "http://localhost:8000/v1"}`
	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", body)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleAddBrain_InvalidID(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	body := `{"id": "UPPER_CASE!", "type": "openai_compatible"}`
	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", body)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleAddBrain_MissingType(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	body := `{"id": "vllm-local"}`
	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", body)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleAddBrain_DuplicateID(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	body := `{"id": "ollama", "type": "openai_compatible"}`
	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", body)

	assertStatus(t, rr, http.StatusConflict)
}

func TestHandleAddBrain_NilCognitive(t *testing.T) {
	s := newTestServer() // no cognitive
	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", `{"id":"x","type":"openai_compatible"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ════════════════════════════════════════════════════════════════════
// BRAINS — HandleUpdateBrain (PUT /api/v1/brains/{id})
// ════════════════════════════════════════════════════════════════════

func TestHandleUpdateBrain_HappyPath(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	body := `{
		"type": "openai_compatible",
		"endpoint": "http://localhost:11434/v1",
		"model_id": "qwen2.5:14b",
		"token_budget_profile": "deep",
		"max_output_tokens": 4096,
		"enabled": true
	}`
	mux := setupMux(t, "PUT /api/v1/brains/{id}", s.HandleUpdateBrain)
	rr := doRequest(t, mux, "PUT", "/api/v1/brains/ollama", body)

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
	if data["updated"] != true {
		t.Errorf("expected updated=true, got %v", data["updated"])
	}
	if got := s.Cognitive.Config.Providers["ollama"].MaxOutputTokens; got != 4096 {
		t.Errorf("expected updated max_output_tokens=4096, got %d", got)
	}
	if got := s.Cognitive.Config.Providers["ollama"].TokenBudgetProfile; got != "deep" {
		t.Errorf("expected updated token_budget_profile=deep, got %q", got)
	}
}

func TestHandleUpdateBrain_NotFound(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	body := `{"type": "openai_compatible"}`
	mux := setupMux(t, "PUT /api/v1/brains/{id}", s.HandleUpdateBrain)
	rr := doRequest(t, mux, "PUT", "/api/v1/brains/nonexistent", body)

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleUpdateBrain_BadJSON(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "PUT /api/v1/brains/{id}", s.HandleUpdateBrain)
	rr := doRequest(t, mux, "PUT", "/api/v1/brains/ollama", "not-json")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleUpdateBrain_NilCognitive(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "PUT /api/v1/brains/{id}", s.HandleUpdateBrain)
	rr := doRequest(t, mux, "PUT", "/api/v1/brains/ollama", `{"type":"openai"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}
