package server

import (
	"net/http"
	"strings"
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

// findServiceDetail returns the "detail" string of the named row, or "" if
// the row is missing (callers assert presence separately when it matters).
func findServiceDetail(data []any, name string) (string, bool) {
	for _, item := range data {
		svc, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if svc["name"] == name {
			detail, _ := svc["detail"].(string)
			return detail, true
		}
	}
	return "", false
}

// C1 (S6e security-QA): GET /api/v1/services/status must never leak a
// provider endpoint or model id to a caller who is not a root admin holding
// cognitive:read/cognitive:write, matching the F2 narrowing already applied
// to /cognitive/status and /brains. A full-view admin still gets the model
// and endpoint in the ollama row's detail.
func TestHandleServicesStatus_OllamaDetailNarrowedForNonAdmins(t *testing.T) {
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

	callers := []struct {
		name     string
		identity *RequestIdentity
	}{
		{"anonymous", nil},
		{"standard user", &RequestIdentity{UserID: "u-1", Role: "user", Scopes: []string{"soma:work"}}},
		{"admin without cognitive scope", &RequestIdentity{UserID: "u-2", Role: "admin", Scopes: []string{"config_documents:write"}}},
	}
	for _, caller := range callers {
		t.Run(caller.name, func(t *testing.T) {
			var rr = doAuthenticatedRequestAs(t, mux, http.MethodGet, "/api/v1/services/status", "", caller.identity)
			assertStatus(t, rr, http.StatusOK)
			var resp map[string]any
			assertJSON(t, rr, &resp)
			data := resp["data"].([]any)
			for _, name := range []string{"ollama", "cognitive"} {
				detail, ok := findServiceDetail(data, name)
				if !ok {
					continue
				}
				if strings.Contains(detail, "://") || strings.Contains(strings.ToLower(detail), "qwen2.5-coder") {
					t.Fatalf("%s row leaked endpoint/model to %s: %q", name, caller.name, detail)
				}
			}
		})
	}

	rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/services/status", "")
	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)
	detail, ok := findServiceDetail(data, "ollama")
	if !ok || !strings.Contains(detail, "qwen2.5-coder:7b") || !strings.Contains(detail, "http://localhost:11434/v1") {
		t.Fatalf("full-view admin should still see model and endpoint, got %q (found=%v)", detail, ok)
	}
}
