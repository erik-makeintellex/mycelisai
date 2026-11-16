package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

func TestServiceAddressBook_EnvPrecedence(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("MYCELIS_DEPLOYMENT_ENV", "")
	t.Setenv("MYCELIS_K8S_CLUSTER_MODE", "")

	t.Setenv("NATS_URL", "nats://legacy-nats:4222")
	t.Setenv("MYCELIS_SERVICE_NATS_ADDRESS", "nats://canonical-nats:4222")
	t.Setenv("DB_HOST", "legacy-pg")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("MYCELIS_SERVICE_DB_ADDRESS", "canonical-pg:5432")
	t.Setenv("MYCELIS_PROVIDER_OLLAMA_ENDPOINT", "http://legacy-ollama:11434/v1")
	t.Setenv("MYCELIS_K8S_TEXT_ENDPOINT", "http://k8s-ollama:11434/v1")
	t.Setenv("MYCELIS_SERVICE_TEXT_AI_ADDRESS", "http://canonical-text:11434/v1")
	t.Setenv("MYCELIS_MEDIA_ENDPOINT", "http://legacy-media:8001/v1")
	t.Setenv("MYCELIS_SERVICE_MEDIA_AI_ADDRESS", "http://canonical-media:8001/v1")
	t.Setenv("MYCELIS_SEARXNG_ENDPOINT", "http://legacy-searxng:8080")
	t.Setenv("MYCELIS_SERVICE_SEARCH_ADDRESS", "http://canonical-search:8080")
	t.Setenv("MYCELIS_SERVICE_MCP_GATEWAY_ADDRESS", "http://canonical-mcp:9090")
	t.Setenv("CORS_ORIGIN", "http://legacy-ui:3000")
	t.Setenv("MYCELIS_SERVICE_CORS_ORIGIN", "http://canonical-ui:3000")

	book := ResolveServiceAddressBook(nil)
	if book.ClusterMode {
		t.Fatal("expected ClusterMode=false")
	}

	expected := map[string]string{
		"nats":        "nats://canonical-nats:4222",
		"postgres":    "canonical-pg:5432",
		"text_ai":     "http://canonical-text:11434/v1",
		"media_ai":    "http://canonical-media:8001/v1",
		"search":      "http://canonical-search:8080",
		"mcp_gateway": "http://canonical-mcp:9090",
		"cors_origin": "http://canonical-ui:3000",
	}
	for name, wantAddr := range expected {
		if got := book.AddressFor(name); got != wantAddr {
			t.Errorf("%s address = %q, want %q", name, got, wantAddr)
		}
	}
}

func TestServiceAddressBook_ClusterLoopbackDetection(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "10.96.0.1")
	t.Setenv("MYCELIS_SERVICE_NATS_ADDRESS", "nats://localhost:4222")
	t.Setenv("MYCELIS_SERVICE_DB_ADDRESS", "127.0.0.1:5432")
	t.Setenv("MYCELIS_SERVICE_TEXT_AI_ADDRESS", "http://127.0.0.1:11434/v1")
	t.Setenv("MYCELIS_SERVICE_MEDIA_AI_ADDRESS", "")
	t.Setenv("MYCELIS_MEDIA_ENDPOINT", "")
	t.Setenv("MYCELIS_K8S_MEDIA_ENDPOINT", "")

	book := ResolveServiceAddressBook(nil)
	if !book.ClusterMode {
		t.Fatal("expected ClusterMode=true when KUBERNETES_SERVICE_HOST is set")
	}
	if len(book.Warnings) != 3 {
		t.Fatalf("expected 3 loopback warnings (nats, postgres, text_ai), got %d: %v", len(book.Warnings), book.Warnings)
	}

	for _, name := range []string{"nats", "postgres", "text_ai"} {
		entry, ok := book.EntryFor(name)
		if !ok {
			t.Fatalf("missing entry for %s", name)
		}
		if !strings.Contains(entry.LoopbackWarning, "uses loopback inside a Kubernetes pod") {
			t.Errorf("expected loopback warning on %s, got %q", name, entry.LoopbackWarning)
		}
	}
}

func TestServiceAddressBook_ClusterNonLoopbackAddressesZeroWarnings(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "10.96.0.1")
	t.Setenv("MYCELIS_SERVICE_NATS_ADDRESS", "nats://mycelis-core-nats:4222")
	t.Setenv("MYCELIS_SERVICE_DB_ADDRESS", "mycelis-core-postgresql:5432")
	t.Setenv("MYCELIS_SERVICE_TEXT_AI_ADDRESS", "http://host.rancher-desktop.internal:11434/v1")
	t.Setenv("MYCELIS_SERVICE_MEDIA_AI_ADDRESS", "http://mycelis-media:8001/v1")
	t.Setenv("MYCELIS_SERVICE_SEARCH_ADDRESS", "http://mycelis-searxng:8080")
	t.Setenv("MYCELIS_SERVICE_MCP_GATEWAY_ADDRESS", "http://mycelis-mcp-gateway:9090")

	book := ResolveServiceAddressBook(nil)
	if !book.ClusterMode {
		t.Fatal("expected ClusterMode=true")
	}
	if len(book.Warnings) != 0 {
		t.Fatalf("expected 0 warnings for cluster DNS / host.rancher-desktop.internal addresses, got: %v", book.Warnings)
	}
}

func TestHandleServicesStatus_IncludesResolvedServiceAddresses(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "10.96.0.1")
	t.Setenv("MYCELIS_SERVICE_NATS_ADDRESS", "nats://mycelis-core-nats:4222")
	t.Setenv("MYCELIS_SERVICE_DB_ADDRESS", "mycelis-core-postgresql:5432")
	t.Setenv("MYCELIS_SERVICE_TEXT_AI_ADDRESS", "")
	t.Setenv("MYCELIS_K8S_TEXT_ENDPOINT", "")
	t.Setenv("MYCELIS_PROVIDER_OLLAMA_ENDPOINT", "")

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

	byName := map[string]map[string]any{}
	for _, item := range data {
		svc := item.(map[string]any)
		byName[svc["name"].(string)] = svc
	}

	if got := byName["nats"]["address"]; got != "nats://mycelis-core-nats:4222" {
		t.Errorf("nats address = %v, want nats://mycelis-core-nats:4222", got)
	}
	if got := byName["postgres"]["address"]; got != "mycelis-core-postgresql:5432" {
		t.Errorf("postgres address = %v, want mycelis-core-postgresql:5432", got)
	}
	if got := byName["ollama"]["address"]; got != "http://localhost:11434/v1" {
		t.Errorf("ollama address = %v, want http://localhost:11434/v1", got)
	}
	if warn, _ := byName["ollama"]["warning"].(string); !strings.Contains(warn, "uses loopback inside a Kubernetes pod") {
		t.Errorf("expected cluster loopback warning on ollama service status, got %q", warn)
	}
}
