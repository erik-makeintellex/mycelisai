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

func TestHandleCognitiveStatus_ExposesTypedMediaProviderContract(t *testing.T) {
	mediaEnabled := true
	mediaHealth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer mediaHealth.Close()

	s := &AdminServer{
		Cognitive: &cognitive.Router{
			Config: &cognitive.BrainConfig{
				Media: &cognitive.MediaConfig{
					Provider: cognitive.MediaProviderConfig{
						ProviderID:   "media-local",
						Type:         "openai_compatible",
						Endpoint:     mediaHealth.URL + "/v1",
						ModelID:      "stable-diffusion-xl",
						Location:     "local",
						DataBoundary: "local_only",
						UsagePolicy:  "local_first",
						Enabled:      &mediaEnabled,
					},
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cognitive/status", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleCognitiveStatus).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	media, ok := resp["media"].(map[string]any)
	if !ok {
		t.Fatalf("expected media status object, got %T", resp["media"])
	}
	if media["status"] != "online" {
		t.Fatalf("media status = %v, want online", media["status"])
	}
	if media["provider_id"] != "media-local" {
		t.Fatalf("media provider_id = %v, want media-local", media["provider_id"])
	}
	if media["provider_type"] != "openai_compatible" {
		t.Fatalf("media provider_type = %v, want openai_compatible", media["provider_type"])
	}
	if media["location"] != "local" {
		t.Fatalf("media location = %v, want local", media["location"])
	}
	if media["data_boundary"] != "local_only" {
		t.Fatalf("media data_boundary = %v, want local_only", media["data_boundary"])
	}
	if media["usage_policy"] != "local_first" {
		t.Fatalf("media usage_policy = %v, want local_first", media["usage_policy"])
	}
	if media["configured"] != true {
		t.Fatalf("media configured = %v, want true", media["configured"])
	}
	if media["enabled"] != true {
		t.Fatalf("media enabled = %v, want true", media["enabled"])
	}
}

func TestHandleCognitiveStatus_DoesNotExposeTextEnabledFalse(t *testing.T) {
	s := &AdminServer{
		Cognitive: &cognitive.Router{
			Config: &cognitive.BrainConfig{},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cognitive/status", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleCognitiveStatus).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	text, ok := resp["text"].(map[string]any)
	if !ok {
		t.Fatalf("expected text status object, got %T", resp["text"])
	}
	if _, ok := text["enabled"]; ok {
		t.Fatalf("text status should not expose media-only enabled flag: %#v", text)
	}
}

func TestHandleCognitiveStatus_HostedMediaProviderReportsConfiguredWithoutHealthProbe(t *testing.T) {
	mediaEnabled := true
	s := &AdminServer{
		Cognitive: &cognitive.Router{
			Config: &cognitive.BrainConfig{
				Media: &cognitive.MediaConfig{
					Provider: cognitive.MediaProviderConfig{
						ProviderID: "replicate",
						Type:       "hosted_api",
						Endpoint:   "https://api.replicate.example/v1",
						ModelID:    "sdxl",
						Enabled:    &mediaEnabled,
					},
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cognitive/status", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleCognitiveStatus).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	media, ok := resp["media"].(map[string]any)
	if !ok {
		t.Fatalf("expected media status object, got %T", resp["media"])
	}
	if media["status"] != "configured" {
		t.Fatalf("media status = %v, want configured", media["status"])
	}
	if media["location"] != "remote" {
		t.Fatalf("media location = %v, want remote", media["location"])
	}
}

func TestHandleCognitiveStatus_MediaProviderCanBeExplicitlyDisabled(t *testing.T) {
	mediaEnabled := false
	s := &AdminServer{
		Cognitive: &cognitive.Router{
			Config: &cognitive.BrainConfig{
				Media: &cognitive.MediaConfig{
					Provider: cognitive.MediaProviderConfig{
						Endpoint: "http://127.0.0.1:8001/v1",
						ModelID:  "stable-diffusion-xl",
						Enabled:  &mediaEnabled,
					},
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cognitive/status", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleCognitiveStatus).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	media, ok := resp["media"].(map[string]any)
	if !ok {
		t.Fatalf("expected media status object, got %T", resp["media"])
	}
	if media["status"] != "disabled" {
		t.Fatalf("media status = %v, want disabled", media["status"])
	}
	if media["enabled"] != false {
		t.Fatalf("media enabled = %v, want false", media["enabled"])
	}
}
