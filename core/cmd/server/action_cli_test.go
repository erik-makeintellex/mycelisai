package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestResolveActionURL(t *testing.T) {
	got, err := resolveActionURL("http://localhost:8081", "/api/v1/services/status")
	if err != nil {
		t.Fatalf("resolveActionURL: %v", err)
	}
	want := "http://localhost:8081/api/v1/services/status"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	abs, err := resolveActionURL("http://localhost:8081", "https://example.com/api")
	if err != nil {
		t.Fatalf("resolveActionURL abs: %v", err)
	}
	if abs != "https://example.com/api" {
		t.Fatalf("got %q, want %q", abs, "https://example.com/api")
	}
}

func TestMergeActionCLIConfig(t *testing.T) {
	dst := ActionCLIConfig{
		APIBaseURL:     "http://localhost:8081",
		APIKey:         "old",
		TimeoutSeconds: 10,
		Headers:        map[string]string{"X-Test": "1"},
	}
	src := ActionCLIConfig{
		APIBaseURL:     "http://remote:8081",
		APIKey:         "new",
		TimeoutSeconds: 30,
		Headers:        map[string]string{"X-Other": "2"},
	}
	mergeActionCLIConfig(&dst, src)

	if dst.APIBaseURL != "http://remote:8081" {
		t.Fatalf("api_base_url = %q", dst.APIBaseURL)
	}
	if dst.APIKey != "new" {
		t.Fatalf("api_key = %q", dst.APIKey)
	}
	if dst.TimeoutSeconds != 30 {
		t.Fatalf("timeout_seconds = %d", dst.TimeoutSeconds)
	}
	if dst.Headers["X-Test"] != "1" || dst.Headers["X-Other"] != "2" {
		t.Fatalf("headers merge failed: %#v", dst.Headers)
	}
}

func TestLoadActionCLIConfigFromPaths_Precedence(t *testing.T) {
	tmp := t.TempDir()
	low := filepath.Join(tmp, "low.yaml")
	high := filepath.Join(tmp, "high.yaml")

	if err := os.WriteFile(low, []byte("api_base_url: http://low:8081\napi_key: lowkey\ntimeout_seconds: 5\nheaders:\n  X-Env: low\n"), 0o644); err != nil {
		t.Fatalf("write low: %v", err)
	}
	if err := os.WriteFile(high, []byte("api_base_url: http://high:8081\napi_key: highkey\nheaders:\n  X-Env: high\n"), 0o644); err != nil {
		t.Fatalf("write high: %v", err)
	}

	cfg, loaded, err := loadActionCLIConfigFromPaths([]string{low, high}, defaultActionCLIConfig())
	if err != nil {
		t.Fatalf("loadActionCLIConfigFromPaths: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("loaded = %d, want 2", len(loaded))
	}
	if cfg.APIBaseURL != "http://high:8081" {
		t.Fatalf("api_base_url = %q, want high override", cfg.APIBaseURL)
	}
	if cfg.APIKey != "highkey" {
		t.Fatalf("api_key = %q, want highkey", cfg.APIKey)
	}
	if cfg.TimeoutSeconds != 5 {
		t.Fatalf("timeout_seconds = %d, want 5", cfg.TimeoutSeconds)
	}
	if cfg.Headers["X-Env"] != "high" {
		t.Fatalf("header precedence failed: %#v", cfg.Headers)
	}
}

func TestDiscoverActionConfigPathsWith_HomeAndOverrideLast(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "XDG_CONFIG_HOME":
			return "/xdg"
		case "MYCELIS_CONFIG":
			return "/custom/final.yaml"
		default:
			return ""
		}
	}
	getwd := func() (string, error) { return "/cwd", nil }
	paths := discoverActionConfigPathsWith(getenv, getwd, "/home/alice")

	wantContains := []string{
		filepath.Join("/cwd", "mycelis.yaml"),
		filepath.Join("/xdg", "mycelis", "config.yaml"),
		filepath.Join("/home/alice", ".config", "mycelis", "config.yaml"),
		filepath.Join("/home/alice", ".mycelis", "config.yaml"),
		"/custom/final.yaml",
	}
	for _, p := range wantContains {
		if !slices.Contains(paths, p) {
			t.Fatalf("expected paths to contain %q", p)
		}
	}
	if paths[len(paths)-1] != "/custom/final.yaml" {
		t.Fatalf("expected override path last, got %q", paths[len(paths)-1])
	}
}

func TestParseActionShellLine(t *testing.T) {
	method, target, body, handled, err := parseActionShellLine("status")
	if err != nil || !handled {
		t.Fatalf("status parse failed: err=%v handled=%v", err, handled)
	}
	if method != "GET" || target != "/api/v1/services/status" || body != "" {
		t.Fatalf("unexpected status parse: %s %s %s", method, target, body)
	}

	method, target, body, handled, err = parseActionShellLine("chat sentry verify system integrity")
	if err != nil || !handled {
		t.Fatalf("chat parse failed: err=%v handled=%v", err, handled)
	}
	if method != "POST" || target != "/api/v1/council/sentry/chat" {
		t.Fatalf("unexpected chat parse: %s %s", method, target)
	}
	if !strings.Contains(body, "verify system integrity") {
		t.Fatalf("unexpected chat body: %s", body)
	}
}

func TestExecuteActionRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/services/status" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	cfg := defaultActionCLIConfig()
	cfg.APIBaseURL = ts.URL
	status, body, err := executeActionRequest(cfg, "GET", "/api/v1/services/status", "")
	if err != nil {
		t.Fatalf("executeActionRequest: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(body, `"ok":true`) {
		t.Fatalf("body = %s", body)
	}
}

func TestParseProviderOverrideMap(t *testing.T) {
	got := parseProviderOverrideMap(`{"council-core":"ollama-local","council-architect":"vllm-west"}`)
	if got["council-core"] != "ollama-local" {
		t.Fatalf("unexpected map value: %#v", got)
	}
	if got["council-architect"] != "vllm-west" {
		t.Fatalf("unexpected map value: %#v", got)
	}

	none := parseProviderOverrideMap("not-json")
	if none != nil {
		t.Fatalf("expected nil on invalid json, got %#v", none)
	}
}
