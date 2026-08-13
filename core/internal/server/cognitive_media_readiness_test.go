package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

func TestMediaGenerationPreflight_ForgeReady(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sdapi/v1/options" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)
	s := mediaReadinessTestServer(upstream.URL)

	if blocker := s.mediaGenerationPreflight(t.Context(), mediaGenerationTestCalls()); blocker != nil {
		t.Fatalf("unexpected blocker: %+v", blocker)
	}
}

func TestMediaGenerationPreflight_ForgeOpenWithoutAPI(t *testing.T) {
	upstream := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(upstream.Close)
	s := mediaReadinessTestServer(upstream.URL)

	blocker := s.mediaGenerationPreflight(t.Context(), mediaGenerationTestCalls())
	if blocker == nil {
		t.Fatal("expected Forge API readiness blocker")
	}
	if blocker.Summary != "Forge is open, but image generation is not enabled." {
		t.Fatalf("summary = %q", blocker.Summary)
	}
	if blocker.RecommendedAction == "" || blocker.SetupPath != "/resources?section=capabilities" {
		t.Fatalf("blocker = %+v, want recovery action and setup path", blocker)
	}
}

func mediaReadinessTestServer(endpoint string) *AdminServer {
	enabled := true
	return &AdminServer{Cognitive: &cognitive.Router{Config: &cognitive.BrainConfig{
		Media: &cognitive.MediaConfig{Provider: cognitive.MediaProviderConfig{
			ProviderID: "pinokio-forge", Type: cognitive.MediaProviderTypeForge,
			Endpoint: endpoint, ModelID: "forge-local", Enabled: &enabled,
		}},
	}}}
}

func mediaGenerationTestCalls() []protocol.PlannedToolCall {
	return []protocol.PlannedToolCall{{Name: "generate_image", Arguments: map[string]any{"prompt": "Soma portrait"}}}
}
