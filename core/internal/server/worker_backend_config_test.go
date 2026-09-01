package server

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mycelis/core/internal/workers"
)

func TestLoadWorkerExecutionBackendDefaultsToCentral(t *testing.T) {
	path := filepath.Join(t.TempDir(), "engine.yaml")
	if err := os.WriteFile(path, []byte("text:\n  port: 8000\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	backend, err := loadWorkerExecutionBackend(path, nil)
	if err != nil {
		t.Fatalf("loadWorkerExecutionBackend: %v", err)
	}
	if _, ok := backend.(*workers.CentralBackend); !ok {
		t.Fatalf("backend = %T, want central", backend)
	}
}

func TestLoadWorkerExecutionBackendRejectsExternalSelectionUntilProjectionExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "engine.yaml")
	if err := os.WriteFile(path, []byte(`
worker_runtime:
  backend: framework_runs
  base_url: https://workers.example.test
  api_key_secret_ref: env:MYCELIS_WORKER_API_KEY
  preferred_protocol: runs_api
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	backend, err := loadWorkerExecutionBackend(path, workers.EnvSecretResolver{})
	if backend != nil || !errors.Is(err, workers.ErrExternalExecutionUnavailable) {
		t.Fatalf("backend/error = %T/%v", backend, err)
	}
}

func TestLoadWorkerExecutionBackendRejectsInvalidExplicitConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")
	backend, err := loadWorkerExecutionBackend(path, nil)
	if backend != nil || err == nil {
		t.Fatalf("backend/error = %T/%v", backend, err)
	}
}

func TestResolveWorkerEngineConfigPathFromRepositoryRoot(t *testing.T) {
	repoRoot := t.TempDir()
	configPath := filepath.Join(repoRoot, "cognitive", "config", "engine.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("worker_runtime:\n  backend: central\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })

	resolved, required, err := resolveWorkerEngineConfigPath("")
	if err != nil {
		t.Fatalf("resolveWorkerEngineConfigPath: %v", err)
	}
	if required || resolved != configPath {
		t.Fatalf("resolved/required = %q/%v, want %q/false", resolved, required, configPath)
	}
}
