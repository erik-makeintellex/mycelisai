package swarm

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

var swarmTestHTTPPort int32 = 40000 + int32(time.Now().UnixNano()%5000)

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
	port := nextSwarmTestTCPPort(t, &swarmTestHTTPPort)
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("test http listen: %v", err)
	}
	srv := httptest.NewUnstartedServer(handler)
	srv.Listener = ln
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}
