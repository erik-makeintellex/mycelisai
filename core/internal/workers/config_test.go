package workers

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestParseEngineConfigDefaultsToCentral(t *testing.T) {
	cfg, err := ParseEngineConfig([]byte("text:\n  port: 8000\n"))
	if err != nil {
		t.Fatalf("ParseEngineConfig: %v", err)
	}
	if cfg.Backend != BackendCentral {
		t.Fatalf("backend = %q, want central", cfg.Backend)
	}
	backend, err := NewExecutionBackend(cfg, nil)
	if err != nil {
		t.Fatalf("NewExecutionBackend: %v", err)
	}
	if _, ok := backend.(*CentralBackend); !ok {
		t.Fatalf("backend = %T, want *CentralBackend", backend)
	}
}

func TestParseEngineConfigConstructsFrameworkRunsClientButExecutionSelectionFailsClosed(t *testing.T) {
	cfg, err := ParseEngineConfig([]byte(`
worker_runtime:
  backend: framework_runs
  base_url: https://workers.example.test
  api_key_secret_ref: env:MYCELIS_WORKER_API_KEY
  capabilities_endpoint: /v1/capabilities
  health_endpoint: /health
  preferred_protocol: runs_api
  approval_mode: mycelis_control_plane
  event_stream_mode: sse
  timeout_policy:
    connect_ms: 5000
    run_ms: 900000
    stream_ms: 900000
`))
	if err != nil {
		t.Fatalf("ParseEngineConfig: %v", err)
	}
	backend, err := NewBackend(cfg, EnvSecretResolver{})
	if err != nil {
		t.Fatalf("NewBackend: %v", err)
	}
	if _, ok := backend.(*FrameworkRunsBackend); !ok {
		t.Fatalf("backend = %T", backend)
	}
	if _, err := NewExecutionBackend(cfg, EnvSecretResolver{}); !errors.Is(err, ErrExternalExecutionUnavailable) {
		t.Fatalf("NewExecutionBackend error = %v", err)
	}
}

func TestFrameworkRunsConfigFailsClosed(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want string
	}{
		{"unknown backend", "worker_runtime:\n  backend: unsupported\n", "unsupported worker backend"},
		{"missing URL", "worker_runtime:\n  backend: framework_runs\n", "base_url"},
		{"raw secret", "worker_runtime:\n  backend: framework_runs\n  base_url: https://workers.example.test\n  api_key_secret_ref: raw-token\n", "secret_ref"},
		{"credential in URL", "worker_runtime:\n  backend: framework_runs\n  base_url: https://token@workers.example.test\n", "must not contain credentials"},
		{"unknown field", "worker_runtime:\n  backend: central\n  api_key: raw-token\n", "field api_key not found"},
		{"retired session strategy", "worker_runtime:\n  backend: central\n  session_key_strategy: org_project_run\n", "field session_key_strategy not found"},
		{"retired tool policy", "worker_runtime:\n  backend: central\n  tool_policy: {}\n", "field tool_policy not found"},
		{"retired fallback", "worker_runtime:\n  backend: central\n  fallback_backend: central\n", "field fallback_backend not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseEngineConfig([]byte(tt.yaml))
			if err == nil {
				_, err = NewBackend(cfg, EnvSecretResolver{})
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestEnvSecretResolverUsesReferenceOnly(t *testing.T) {
	t.Setenv("MYCELIS_TEST_WORKER_KEY", "resolved-secret")
	value, err := (EnvSecretResolver{}).ResolveSecret(context.Background(), "env:MYCELIS_TEST_WORKER_KEY")
	if err != nil || value != "resolved-secret" {
		t.Fatalf("ResolveSecret = %q, %v", value, err)
	}
	if _, err := (EnvSecretResolver{}).ResolveSecret(context.Background(), "resolved-secret"); err == nil {
		t.Fatal("expected raw secret to be rejected")
	}
}
