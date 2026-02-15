package server

import (
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
