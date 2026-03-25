package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

func TestCognitiveMatrix(t *testing.T) {
	// 1. Setup Mock AdminServer with Cognitive Engine
	cfg := &cognitive.BrainConfig{
		Profiles: map[string]string{"test": "mock"},
		Providers: map[string]cognitive.ProviderConfig{
			"mock": {Type: "mock", ModelID: "gpt-4"},
		},
	}

	// Create a partial AdminServer just for this handler
	s := &AdminServer{
		Cognitive: &cognitive.Router{
			Config: cfg,
		},
	}

	// 2. Create Request
	req, _ := http.NewRequest("GET", "/api/v1/cognitive/matrix", nil)
	rr := httptest.NewRecorder()

	// 3. Invoke Handler
	handler := http.HandlerFunc(s.HandleCognitiveConfig)
	handler.ServeHTTP(rr, req)

	// 4. Assertions
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var resp cognitive.BrainConfig
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if val, ok := resp.Profiles["test"]; !ok || val != "mock" {
		t.Errorf("unexpected profile config: %v", resp.Profiles)
	}
}

func TestHandleChat_RequiresAvailableCognitiveEngine(t *testing.T) {
	s := &AdminServer{
		Cognitive: &cognitive.Router{
			Config: &cognitive.BrainConfig{
				Providers: map[string]cognitive.ProviderConfig{
					"ollama": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b"},
				},
				Profiles: map[string]string{
					"chat": "ollama",
				},
			},
		},
	}

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"hello"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusServiceUnavailable)
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["ok"] != false {
		t.Fatalf("ok = %v, want false", resp["ok"])
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["code"] != cognitive.ExecutionProviderDisabled {
		t.Fatalf("code = %v, want %s", data["code"], cognitive.ExecutionProviderDisabled)
	}
	if data["setup_required"] != true {
		t.Fatalf("setup_required = %v, want true", data["setup_required"])
	}
}
