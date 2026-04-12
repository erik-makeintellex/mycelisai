package swarm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

func TestHandleGenerateImage_UsesMediaProviderAuthKeyEnv(t *testing.T) {
	t.Setenv("HOSTED_MEDIA_API_KEY", "test-media-key")

	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/images/generations" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		authHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"aW1hZ2U="}]}`))
	}))
	defer server.Close()

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
