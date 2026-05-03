package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

// ════════════════════════════════════════════════════════════════════
// Shared helpers for provider-CRUD handler tests
// ════════════════════════════════════════════════════════════════════

// stubAdapter implements cognitive.LLMProvider for test purposes.
type stubAdapter struct {
	healthy bool
}

func (s *stubAdapter) Infer(_ context.Context, _ string, _ cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "stub"}, nil
}

func (s *stubAdapter) Probe(_ context.Context) (bool, error) {
	return s.healthy, nil
}

// withCognitive builds a minimal cognitive.Router with the given providers/adapters
// and a temp ConfigPath for SaveConfig.
func withCognitive(t *testing.T, providers map[string]cognitive.ProviderConfig, adapters map[string]cognitive.LLMProvider) func(*AdminServer) {
	t.Helper()
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "cognitive.yaml")
	// Write a minimal seed file so SaveConfig can overwrite it
	os.WriteFile(cfgPath, []byte("# test\n"), 0644)

	profiles := map[string]string{"chat": "ollama"}
	r := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Providers: providers,
			Profiles:  profiles,
		},
		ConfigPath: cfgPath,
		Adapters:   adapters,
	}
	return func(s *AdminServer) {
		s.Cognitive = r
	}
}

// withDirectDB creates a sqlmock *sql.DB wired to s.DB (not s.Registry).
// Profiles and context snapshot handlers use s.DB directly.
func withDirectDB(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.DB = db
	}, mock
}

// withDirectDBPing creates a sqlmock with ping monitoring enabled (needed for
// HandleServicesStatus which calls DB.PingContext).
func withDirectDBPing(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock (ping): %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.DB = db
	}, mock
}

// stubReactive satisfies the HandleServicesStatus reactive checks.
// We cannot import reactive.Engine directly because it requires a NATS conn,
// so we leave Reactive nil in tests or set it via a thin wrapper where needed.
