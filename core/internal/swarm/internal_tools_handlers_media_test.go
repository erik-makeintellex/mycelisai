package swarm

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

func TestHandleGenerateImage_UsesLocalMediaProviderWithoutOpenAIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "should-not-be-used-for-local-media")

	var authHeader string
	var requestBody map[string]any
	server := newSwarmLocalHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/images/generations" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		authHeader = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"aW1hZ2U="}]}`))
	}))

	enabled := true
	registry := NewInternalToolRegistry(InternalToolDeps{
		Brain: &cognitive.Router{
			Config: &cognitive.BrainConfig{
				Media: &cognitive.MediaConfig{
					Provider: cognitive.MediaProviderConfig{
						ProviderID:   "pinokio-local",
						Type:         "openai_compatible",
						Endpoint:     server.URL + "/v1",
						ModelID:      "local-media",
						Location:     "local",
						DataBoundary: "local_only",
						UsagePolicy:  "local_first",
						AuthKeyEnv:   "",
						Enabled:      &enabled,
					},
				},
			},
		},
	})

	result, err := registry.handleGenerateImage(context.Background(), map[string]any{"prompt": "private launch graphic", "size": "768x512"})
	if err != nil {
		t.Fatalf("handleGenerateImage: %v", err)
	}
	if authHeader != "" {
		t.Fatalf("Authorization header = %q, want empty for local media", authHeader)
	}
	if requestBody["model"] != "local-media" {
		t.Fatalf("model = %v, want local-media", requestBody["model"])
	}
	if requestBody["response_format"] != "b64_json" {
		t.Fatalf("response_format = %v, want b64_json", requestBody["response_format"])
	}
	if !strings.Contains(result, "Image generated for") {
		t.Fatalf("expected generated image response, got %s", result)
	}
}

func TestHandleGenerateImage_UsesMediaProviderAuthKeyEnv(t *testing.T) {
	t.Setenv("HOSTED_MEDIA_API_KEY", "test-media-key")

	var authHeader string
	server := newSwarmLocalHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/images/generations" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		authHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"aW1hZ2U="}]}`))
	}))

	enabled := true
	registry := NewInternalToolRegistry(InternalToolDeps{
		Brain: &cognitive.Router{
			Config: &cognitive.BrainConfig{
				Media: &cognitive.MediaConfig{
					Provider: cognitive.MediaProviderConfig{
						ProviderID: "hosted-media",
						Type:       "hosted_api",
						Endpoint:   server.URL + "/v1",
						ModelID:    "sdxl",
						AuthKeyEnv: "HOSTED_MEDIA_API_KEY",
						Enabled:    &enabled,
					},
				},
			},
		},
	})

	result, err := registry.handleGenerateImage(context.Background(), map[string]any{"prompt": "launch graphic"})
	if err != nil {
		t.Fatalf("handleGenerateImage: %v", err)
	}
	if authHeader != "Bearer test-media-key" {
		t.Fatalf("Authorization header = %q, want bearer token", authHeader)
	}
	if !strings.Contains(result, "Image generated for") {
		t.Fatalf("expected generated image response, got %s", result)
	}
}

func newSwarmLocalHTTPTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("test http listen: %v", err)
	}
	srv := httptest.NewUnstartedServer(handler)
	srv.Listener = ln
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}
