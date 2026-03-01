package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

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
// For tests that only need Reactive == nil, just don't wire it.

// ════════════════════════════════════════════════════════════════════
// BRAINS — HandleAddBrain (POST /api/v1/brains)
// ════════════════════════════════════════════════════════════════════

func TestHandleAddBrain_HappyPath(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	body := `{
		"id": "vllm-local",
		"type": "openai_compatible",
		"endpoint": "http://localhost:8000/v1",
		"model_id": "mixtral",
		"enabled": true
	}`

	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", body)

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["id"] != "vllm-local" {
		t.Errorf("expected id=vllm-local, got %v", data["id"])
	}
	if data["type"] != "openai_compatible" {
		t.Errorf("expected type=openai_compatible, got %v", data["type"])
	}
	// Location should default to "local"
	if data["location"] != "local" {
		t.Errorf("expected location=local, got %v", data["location"])
	}
	if data["data_boundary"] != "local_only" {
		t.Errorf("expected data_boundary=local_only, got %v", data["data_boundary"])
	}
}

func TestHandleAddBrain_BadJSON(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", "not-json")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleAddBrain_MissingID(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	body := `{"type": "openai_compatible", "endpoint": "http://localhost:8000/v1"}`
	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", body)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleAddBrain_InvalidID(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	body := `{"id": "UPPER_CASE!", "type": "openai_compatible"}`
	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", body)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleAddBrain_MissingType(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	body := `{"id": "vllm-local"}`
	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", body)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleAddBrain_DuplicateID(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	body := `{"id": "ollama", "type": "openai_compatible"}`
	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", body)

	assertStatus(t, rr, http.StatusConflict)
}

func TestHandleAddBrain_NilCognitive(t *testing.T) {
	s := newTestServer() // no cognitive
	mux := setupMux(t, "POST /api/v1/brains", s.HandleAddBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains", `{"id":"x","type":"openai_compatible"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ════════════════════════════════════════════════════════════════════
// BRAINS — HandleUpdateBrain (PUT /api/v1/brains/{id})
// ════════════════════════════════════════════════════════════════════

func TestHandleUpdateBrain_HappyPath(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	body := `{
		"type": "openai_compatible",
		"endpoint": "http://localhost:11434/v1",
		"model_id": "qwen2.5:14b",
		"enabled": true
	}`
	mux := setupMux(t, "PUT /api/v1/brains/{id}", s.HandleUpdateBrain)
	rr := doRequest(t, mux, "PUT", "/api/v1/brains/ollama", body)

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["updated"] != true {
		t.Errorf("expected updated=true, got %v", data["updated"])
	}
}

func TestHandleUpdateBrain_NotFound(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	body := `{"type": "openai_compatible"}`
	mux := setupMux(t, "PUT /api/v1/brains/{id}", s.HandleUpdateBrain)
	rr := doRequest(t, mux, "PUT", "/api/v1/brains/nonexistent", body)

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleUpdateBrain_BadJSON(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "PUT /api/v1/brains/{id}", s.HandleUpdateBrain)
	rr := doRequest(t, mux, "PUT", "/api/v1/brains/ollama", "not-json")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleUpdateBrain_NilCognitive(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "PUT /api/v1/brains/{id}", s.HandleUpdateBrain)
	rr := doRequest(t, mux, "PUT", "/api/v1/brains/ollama", `{"type":"openai"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ════════════════════════════════════════════════════════════════════
// BRAINS — HandleDeleteBrain (DELETE /api/v1/brains/{id})
// ════════════════════════════════════════════════════════════════════

func TestHandleDeleteBrain_HappyPath(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
			"vllm":   {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
			"vllm":   &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "DELETE /api/v1/brains/{id}", s.HandleDeleteBrain)
	rr := doRequest(t, mux, "DELETE", "/api/v1/brains/vllm", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["deleted"] != true {
		t.Errorf("expected deleted=true, got %v", data["deleted"])
	}
}

func TestHandleDeleteBrain_NotFound(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "DELETE /api/v1/brains/{id}", s.HandleDeleteBrain)
	rr := doRequest(t, mux, "DELETE", "/api/v1/brains/ghost", "")

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleDeleteBrain_LastProvider(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "DELETE /api/v1/brains/{id}", s.HandleDeleteBrain)
	rr := doRequest(t, mux, "DELETE", "/api/v1/brains/ollama", "")

	assertStatus(t, rr, http.StatusConflict)
}

func TestHandleDeleteBrain_NilCognitive(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "DELETE /api/v1/brains/{id}", s.HandleDeleteBrain)
	rr := doRequest(t, mux, "DELETE", "/api/v1/brains/x", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ════════════════════════════════════════════════════════════════════
// BRAINS — HandleProbeBrain (POST /api/v1/brains/{id}/probe)
// ════════════════════════════════════════════════════════════════════

func TestHandleProbeBrain_Healthy(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "POST /api/v1/brains/{id}/probe", s.HandleProbeBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains/ollama/probe", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["alive"] != true {
		t.Errorf("expected alive=true, got %v", data["alive"])
	}
	if data["id"] != "ollama" {
		t.Errorf("expected id=ollama, got %v", data["id"])
	}
}

func TestHandleProbeBrain_Unhealthy(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: false},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "POST /api/v1/brains/{id}/probe", s.HandleProbeBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains/ollama/probe", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["alive"] != false {
		t.Errorf("expected alive=false, got %v", data["alive"])
	}
}

func TestHandleProbeBrain_NotFound(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "POST /api/v1/brains/{id}/probe", s.HandleProbeBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains/ghost/probe", "")

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleProbeBrain_NilCognitive(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/brains/{id}/probe", s.HandleProbeBrain)
	rr := doRequest(t, mux, "POST", "/api/v1/brains/ollama/probe", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ════════════════════════════════════════════════════════════════════
// BRAINS — HandleListBrains (GET /api/v1/brains)
// ════════════════════════════════════════════════════════════════════

func TestHandleListBrains_HappyPath(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
			"vllm":   {Type: "openai_compatible", Endpoint: "http://localhost:8000/v1", ModelID: "mixtral", Enabled: false},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "GET /api/v1/brains", s.HandleListBrains)
	rr := doRequest(t, mux, "GET", "/api/v1/brains", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 2 {
		t.Errorf("expected 2 brains, got %d", len(data))
	}

	// Check that disabled provider has status "disabled"
	found := false
	for _, item := range data {
		brain := item.(map[string]any)
		if brain["id"] == "vllm" {
			found = true
			if brain["status"] != "disabled" {
				t.Errorf("expected vllm status=disabled, got %v", brain["status"])
			}
		}
	}
	if !found {
		t.Error("expected to find vllm in response")
	}
}

func TestHandleListBrains_NilCognitive(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/brains", s.HandleListBrains)
	rr := doRequest(t, mux, "GET", "/api/v1/brains", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 0 {
		t.Errorf("expected 0 brains for nil cognitive, got %d", len(data))
	}
}

func TestHandleListBrains_DefaultsForEmptyFields(t *testing.T) {
	// Provider with empty location/boundary/policy should get defaults
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"bare": {Type: "openai_compatible", ModelID: "test", Enabled: false},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(cogOpt)

	mux := setupMux(t, "GET /api/v1/brains", s.HandleListBrains)
	rr := doRequest(t, mux, "GET", "/api/v1/brains", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)
	brain := data[0].(map[string]any)
	if brain["location"] != "local" {
		t.Errorf("expected default location=local, got %v", brain["location"])
	}
	if brain["data_boundary"] != "local_only" {
		t.Errorf("expected default data_boundary=local_only, got %v", brain["data_boundary"])
	}
	if brain["usage_policy"] != "local_first" {
		t.Errorf("expected default usage_policy=local_first, got %v", brain["usage_policy"])
	}
}

// ════════════════════════════════════════════════════════════════════
// PROFILES — HandleListMissionProfiles (GET /api/v1/mission-profiles)
// ════════════════════════════════════════════════════════════════════

var profileColumns = []string{
	"id", "name", "description", "role_providers", "subscriptions",
	"context_strategy", "auto_start", "is_active", "tenant_id",
	"created_at", "updated_at",
}

func TestHandleListMissionProfiles_HappyPath(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM mission_profiles").
		WillReturnRows(sqlmock.NewRows(profileColumns).
			AddRow("p-1", "Default Profile", "Primary", []byte(`{"chat":"ollama"}`), []byte(`[]`),
				"fresh", false, true, "default", now, now).
			AddRow("p-2", "Coding Profile", "", []byte(`{"coder":"vllm"}`), []byte(`[]`),
				"warm", false, false, "default", now, now))

	mux := setupMux(t, "GET /api/v1/mission-profiles", s.HandleListMissionProfiles)
	rr := doRequest(t, mux, "GET", "/api/v1/mission-profiles", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(data))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleListMissionProfiles_Empty(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM mission_profiles").
		WillReturnRows(sqlmock.NewRows(profileColumns))

	mux := setupMux(t, "GET /api/v1/mission-profiles", s.HandleListMissionProfiles)
	rr := doRequest(t, mux, "GET", "/api/v1/mission-profiles", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)
	if len(data) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(data))
	}
}

func TestHandleListMissionProfiles_NilDB(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/mission-profiles", s.HandleListMissionProfiles)
	rr := doRequest(t, mux, "GET", "/api/v1/mission-profiles", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleListMissionProfiles_DBError(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM mission_profiles").
		WillReturnError(fmt.Errorf("connection refused"))

	mux := setupMux(t, "GET /api/v1/mission-profiles", s.HandleListMissionProfiles)
	rr := doRequest(t, mux, "GET", "/api/v1/mission-profiles", "")

	assertStatus(t, rr, http.StatusInternalServerError)
}

// ════════════════════════════════════════════════════════════════════
// PROFILES — HandleCreateMissionProfile (POST /api/v1/mission-profiles)
// ════════════════════════════════════════════════════════════════════

func TestHandleCreateMissionProfile_HappyPath(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	now := time.Now()
	mock.ExpectQuery("INSERT INTO mission_profiles").
		WillReturnRows(sqlmock.NewRows(profileColumns).
			AddRow("p-new", "Research", "", []byte(`{"architect":"ollama"}`), []byte(`[]`),
				"fresh", false, false, "default", now, now))

	body := `{
		"name": "Research",
		"role_providers": {"architect":"ollama"},
		"context_strategy": "fresh"
	}`

	mux := setupMux(t, "POST /api/v1/mission-profiles", s.HandleCreateMissionProfile)
	rr := doRequest(t, mux, "POST", "/api/v1/mission-profiles", body)

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleCreateMissionProfile_MinimalBody(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	now := time.Now()
	// Minimal: only name provided; role_providers, subscriptions, context_strategy defaulted
	mock.ExpectQuery("INSERT INTO mission_profiles").
		WillReturnRows(sqlmock.NewRows(profileColumns).
			AddRow("p-min", "Just a name", "", []byte(`{}`), []byte(`[]`),
				"fresh", false, false, "default", now, now))

	body := `{"name": "Just a name"}`
	mux := setupMux(t, "POST /api/v1/mission-profiles", s.HandleCreateMissionProfile)
	rr := doRequest(t, mux, "POST", "/api/v1/mission-profiles", body)

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
}

func TestHandleCreateMissionProfile_MissingName(t *testing.T) {
	dbOpt, _ := withDirectDB(t)
	s := newTestServer(dbOpt)

	body := `{"role_providers": {"chat":"ollama"}}`
	mux := setupMux(t, "POST /api/v1/mission-profiles", s.HandleCreateMissionProfile)
	rr := doRequest(t, mux, "POST", "/api/v1/mission-profiles", body)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCreateMissionProfile_BadJSON(t *testing.T) {
	dbOpt, _ := withDirectDB(t)
	s := newTestServer(dbOpt)

	mux := setupMux(t, "POST /api/v1/mission-profiles", s.HandleCreateMissionProfile)
	rr := doRequest(t, mux, "POST", "/api/v1/mission-profiles", "not-json")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCreateMissionProfile_NilDB(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/mission-profiles", s.HandleCreateMissionProfile)
	rr := doRequest(t, mux, "POST", "/api/v1/mission-profiles", `{"name":"x"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleCreateMissionProfile_DBError(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("INSERT INTO mission_profiles").
		WillReturnError(fmt.Errorf("unique constraint violation"))

	body := `{"name": "Duplicate"}`
	mux := setupMux(t, "POST /api/v1/mission-profiles", s.HandleCreateMissionProfile)
	rr := doRequest(t, mux, "POST", "/api/v1/mission-profiles", body)

	assertStatus(t, rr, http.StatusInternalServerError)
}

// ════════════════════════════════════════════════════════════════════
// PROFILES — HandleUpdateMissionProfile (PUT /api/v1/mission-profiles/{id})
// ════════════════════════════════════════════════════════════════════

func TestHandleUpdateMissionProfile_HappyPath(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("UPDATE mission_profiles").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"name": "Updated Profile",
		"role_providers": {"coder":"vllm"},
		"context_strategy": "warm"
	}`

	mux := setupMux(t, "PUT /api/v1/mission-profiles/{id}", s.HandleUpdateMissionProfile)
	rr := doRequest(t, mux, "PUT", "/api/v1/mission-profiles/p-1", body)

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data := resp["data"].(map[string]any)
	if data["updated"] != true {
		t.Errorf("expected updated=true, got %v", data["updated"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleUpdateMissionProfile_NotFound(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("UPDATE mission_profiles").
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	body := `{"name": "Ghost"}`
	mux := setupMux(t, "PUT /api/v1/mission-profiles/{id}", s.HandleUpdateMissionProfile)
	rr := doRequest(t, mux, "PUT", "/api/v1/mission-profiles/ghost-id", body)

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleUpdateMissionProfile_BadJSON(t *testing.T) {
	dbOpt, _ := withDirectDB(t)
	s := newTestServer(dbOpt)

	mux := setupMux(t, "PUT /api/v1/mission-profiles/{id}", s.HandleUpdateMissionProfile)
	rr := doRequest(t, mux, "PUT", "/api/v1/mission-profiles/p-1", "bad")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleUpdateMissionProfile_NilDB(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "PUT /api/v1/mission-profiles/{id}", s.HandleUpdateMissionProfile)
	rr := doRequest(t, mux, "PUT", "/api/v1/mission-profiles/p-1", `{"name":"x"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ════════════════════════════════════════════════════════════════════
// PROFILES — HandleDeleteMissionProfile (DELETE /api/v1/mission-profiles/{id})
// ════════════════════════════════════════════════════════════════════

func TestHandleDeleteMissionProfile_HappyPath(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("DELETE FROM mission_profiles").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mux := setupMux(t, "DELETE /api/v1/mission-profiles/{id}", s.HandleDeleteMissionProfile)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mission-profiles/p-1", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data := resp["data"].(map[string]any)
	if data["deleted"] != true {
		t.Errorf("expected deleted=true, got %v", data["deleted"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleDeleteMissionProfile_NotFound(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("DELETE FROM mission_profiles").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mux := setupMux(t, "DELETE /api/v1/mission-profiles/{id}", s.HandleDeleteMissionProfile)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mission-profiles/ghost", "")

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleDeleteMissionProfile_NilDB(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "DELETE /api/v1/mission-profiles/{id}", s.HandleDeleteMissionProfile)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mission-profiles/p-1", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleDeleteMissionProfile_DBError(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("DELETE FROM mission_profiles").
		WillReturnError(fmt.Errorf("fk constraint"))

	mux := setupMux(t, "DELETE /api/v1/mission-profiles/{id}", s.HandleDeleteMissionProfile)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mission-profiles/p-1", "")

	assertStatus(t, rr, http.StatusInternalServerError)
}

// ════════════════════════════════════════════════════════════════════
// PROFILES — HandleActivateMissionProfile (POST /api/v1/mission-profiles/{id}/activate)
// ════════════════════════════════════════════════════════════════════

func TestHandleActivateMissionProfile_HappyPath(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: true},
		},
		map[string]cognitive.LLMProvider{},
	)
	s := newTestServer(dbOpt, cogOpt)

	now := time.Now()
	// Step 1: SELECT the profile
	mock.ExpectQuery("SELECT .+ FROM mission_profiles WHERE").
		WillReturnRows(sqlmock.NewRows(profileColumns).
			AddRow("p-1", "Default", sql.NullString{}, []byte(`{"chat":"ollama"}`), []byte(`[]`),
				"fresh", false, false, "default", now, now))

	// Step 2: Transaction — deactivate others + activate this one
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE mission_profiles SET is_active=false").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE mission_profiles SET is_active=true").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	mux := setupMux(t, "POST /api/v1/mission-profiles/{id}/activate", s.HandleActivateMissionProfile)
	rr := doRequest(t, mux, "POST", "/api/v1/mission-profiles/p-1/activate", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleActivateMissionProfile_NotFound(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM mission_profiles WHERE").
		WillReturnError(sql.ErrNoRows)

	mux := setupMux(t, "POST /api/v1/mission-profiles/{id}/activate", s.HandleActivateMissionProfile)
	rr := doRequest(t, mux, "POST", "/api/v1/mission-profiles/ghost/activate", "")

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleActivateMissionProfile_NilDB(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/mission-profiles/{id}/activate", s.HandleActivateMissionProfile)
	rr := doRequest(t, mux, "POST", "/api/v1/mission-profiles/p-1/activate", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleActivateMissionProfile_DBSelectError(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM mission_profiles WHERE").
		WillReturnError(fmt.Errorf("connection reset"))

	mux := setupMux(t, "POST /api/v1/mission-profiles/{id}/activate", s.HandleActivateMissionProfile)
	rr := doRequest(t, mux, "POST", "/api/v1/mission-profiles/p-1/activate", "")

	assertStatus(t, rr, http.StatusInternalServerError)
}

// ════════════════════════════════════════════════════════════════════
// CONTEXT — HandleCreateSnapshot (POST /api/v1/context/snapshot)
// ════════════════════════════════════════════════════════════════════

func TestHandleCreateSnapshot_HappyPath(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	now := time.Now()
	mock.ExpectQuery("INSERT INTO context_snapshots").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("snap-1", now))

	body := `{
		"name": "Pre-deploy snapshot",
		"messages": [{"role":"user","content":"hello"}],
		"run_state": {"phase":"idle"},
		"role_providers": {"chat":"ollama"}
	}`

	mux := setupMux(t, "POST /api/v1/context/snapshot", s.HandleCreateSnapshot)
	rr := doRequest(t, mux, "POST", "/api/v1/context/snapshot", body)

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data := resp["data"].(map[string]any)
	if data["id"] != "snap-1" {
		t.Errorf("expected id=snap-1, got %v", data["id"])
	}
	if data["name"] != "Pre-deploy snapshot" {
		t.Errorf("expected name='Pre-deploy snapshot', got %v", data["name"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleCreateSnapshot_DefaultsEmptyFields(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	now := time.Now()
	// Empty body: name gets auto-generated, messages/runState/roleProviders get defaults
	mock.ExpectQuery("INSERT INTO context_snapshots").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("snap-auto", now))

	body := `{}`
	mux := setupMux(t, "POST /api/v1/context/snapshot", s.HandleCreateSnapshot)
	rr := doRequest(t, mux, "POST", "/api/v1/context/snapshot", body)

	assertStatus(t, rr, http.StatusOK)
}

func TestHandleCreateSnapshot_BadJSON(t *testing.T) {
	dbOpt, _ := withDirectDB(t)
	s := newTestServer(dbOpt)

	mux := setupMux(t, "POST /api/v1/context/snapshot", s.HandleCreateSnapshot)
	rr := doRequest(t, mux, "POST", "/api/v1/context/snapshot", "not-json")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCreateSnapshot_NilDB(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/context/snapshot", s.HandleCreateSnapshot)
	rr := doRequest(t, mux, "POST", "/api/v1/context/snapshot", `{"name":"x"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleCreateSnapshot_DBError(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("INSERT INTO context_snapshots").
		WillReturnError(fmt.Errorf("disk full"))

	body := `{"name": "Will fail"}`
	mux := setupMux(t, "POST /api/v1/context/snapshot", s.HandleCreateSnapshot)
	rr := doRequest(t, mux, "POST", "/api/v1/context/snapshot", body)

	assertStatus(t, rr, http.StatusInternalServerError)
}

// ════════════════════════════════════════════════════════════════════
// CONTEXT — HandleListSnapshots (GET /api/v1/context/snapshots)
// ════════════════════════════════════════════════════════════════════

var snapshotListColumns = []string{
	"id", "name", "description", "source_profile", "tenant_id", "created_at",
}

func TestHandleListSnapshots_HappyPath(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM context_snapshots").
		WillReturnRows(sqlmock.NewRows(snapshotListColumns).
			AddRow("snap-1", "Checkpoint A", "desc", "p-1", "default", now).
			AddRow("snap-2", "Checkpoint B", "", "", "default", now))

	mux := setupMux(t, "GET /api/v1/context/snapshots", s.HandleListSnapshots)
	rr := doRequest(t, mux, "GET", "/api/v1/context/snapshots", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data := resp["data"].([]any)
	if len(data) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(data))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleListSnapshots_NilDB(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/context/snapshots", s.HandleListSnapshots)
	rr := doRequest(t, mux, "GET", "/api/v1/context/snapshots", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	// Nil DB should still return an empty array, not an error
	data := resp["data"].([]any)
	if len(data) != 0 {
		t.Errorf("expected 0 snapshots for nil DB, got %d", len(data))
	}
}

func TestHandleListSnapshots_DBError(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM context_snapshots").
		WillReturnError(fmt.Errorf("timeout"))

	mux := setupMux(t, "GET /api/v1/context/snapshots", s.HandleListSnapshots)
	rr := doRequest(t, mux, "GET", "/api/v1/context/snapshots", "")

	assertStatus(t, rr, http.StatusInternalServerError)
}

// ════════════════════════════════════════════════════════════════════
// CONTEXT — HandleGetSnapshot (GET /api/v1/context/snapshots/{id})
// ════════════════════════════════════════════════════════════════════

var snapshotFullColumns = []string{
	"id", "name", "description", "messages", "run_state", "role_providers",
	"source_profile", "tenant_id", "created_at",
}

func TestHandleGetSnapshot_HappyPath(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM context_snapshots WHERE").
		WillReturnRows(sqlmock.NewRows(snapshotFullColumns).
			AddRow("snap-1", "Checkpoint A",
				sql.NullString{String: "a description", Valid: true},
				[]byte(`[{"role":"user","content":"hi"}]`),
				[]byte(`{"phase":"idle"}`),
				[]byte(`{"chat":"ollama"}`),
				sql.NullString{String: "p-1", Valid: true},
				"default", now))

	mux := setupMux(t, "GET /api/v1/context/snapshots/{id}", s.HandleGetSnapshot)
	rr := doRequest(t, mux, "GET", "/api/v1/context/snapshots/snap-1", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data := resp["data"].(map[string]any)
	if data["id"] != "snap-1" {
		t.Errorf("expected id=snap-1, got %v", data["id"])
	}
	if data["description"] != "a description" {
		t.Errorf("expected description='a description', got %v", data["description"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleGetSnapshot_NotFound(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM context_snapshots WHERE").
		WillReturnError(sql.ErrNoRows)

	mux := setupMux(t, "GET /api/v1/context/snapshots/{id}", s.HandleGetSnapshot)
	rr := doRequest(t, mux, "GET", "/api/v1/context/snapshots/nonexistent", "")

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleGetSnapshot_NilDB(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/context/snapshots/{id}", s.HandleGetSnapshot)
	rr := doRequest(t, mux, "GET", "/api/v1/context/snapshots/snap-1", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleGetSnapshot_DBError(t *testing.T) {
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery("SELECT .+ FROM context_snapshots WHERE").
		WillReturnError(fmt.Errorf("connection lost"))

	mux := setupMux(t, "GET /api/v1/context/snapshots/{id}", s.HandleGetSnapshot)
	rr := doRequest(t, mux, "GET", "/api/v1/context/snapshots/snap-1", "")

	assertStatus(t, rr, http.StatusInternalServerError)
}

// ════════════════════════════════════════════════════════════════════
// SERVICES — HandleServicesStatus (GET /api/v1/services/status)
// ════════════════════════════════════════════════════════════════════

func TestHandleServicesStatus_AllOffline(t *testing.T) {
	// No DB, no Cognitive, no Reactive → everything offline/degraded
	s := newTestServer()

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	// Should have 6 services: nats, postgres, cognitive, ollama, reactive, comms
	if len(data) != 6 {
		t.Fatalf("expected 6 services, got %d", len(data))
	}

	// All should be offline since nothing is wired
	statusMap := map[string]string{}
	for _, item := range data {
		svc := item.(map[string]any)
		statusMap[svc["name"].(string)] = svc["status"].(string)
	}

	if statusMap["nats"] != "offline" {
		t.Errorf("expected nats=offline, got %v", statusMap["nats"])
	}
	if statusMap["postgres"] != "offline" {
		t.Errorf("expected postgres=offline, got %v", statusMap["postgres"])
	}
	if statusMap["cognitive"] != "offline" {
		t.Errorf("expected cognitive=offline, got %v", statusMap["cognitive"])
	}
	if statusMap["ollama"] != "offline" {
		t.Errorf("expected ollama=offline, got %v", statusMap["ollama"])
	}
	if statusMap["reactive"] != "offline" {
		t.Errorf("expected reactive=offline, got %v", statusMap["reactive"])
	}
	if statusMap["comms"] != "offline" {
		t.Errorf("expected comms=offline, got %v", statusMap["comms"])
	}
}

func TestHandleServicesStatus_CognitiveDegraded(t *testing.T) {
	// Cognitive wired but all providers disabled → degraded
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Enabled: false},
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

	// Find cognitive entry
	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "cognitive" {
			if svc["status"] != "degraded" {
				t.Errorf("expected cognitive=degraded (no enabled providers), got %v", svc["status"])
			}
			return
		}
	}
	t.Error("cognitive service entry not found in response")
}

func TestHandleServicesStatus_CognitiveOnline(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "openai_compatible", Endpoint: "http://localhost:11434/v1", ModelID: "qwen2.5-coder:7b", Enabled: true},
			"vllm":   {Type: "openai_compatible", Enabled: true},
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
		if svc["name"] == "cognitive" {
			if svc["status"] != "online" {
				t.Errorf("expected cognitive=online, got %v", svc["status"])
			}
			detail := svc["detail"].(string)
			if detail != "2/2 providers enabled" {
				t.Errorf("expected detail '2/2 providers enabled', got %v", detail)
			}
			return
		}
	}
	t.Error("cognitive service entry not found")
}

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

func TestHandleServicesStatus_PostgresOnline(t *testing.T) {
	dbOpt, mock := withDirectDBPing(t)
	s := newTestServer(dbOpt)

	mock.ExpectPing()

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)

	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "postgres" {
			if svc["status"] != "online" {
				t.Errorf("expected postgres=online, got %v", svc["status"])
			}
			return
		}
	}
	t.Error("postgres service entry not found")
}

func TestHandleServicesStatus_PostgresPingFails(t *testing.T) {
	dbOpt, mock := withDirectDBPing(t)
	s := newTestServer(dbOpt)

	mock.ExpectPing().WillReturnError(fmt.Errorf("connection refused"))

	mux := setupMux(t, "GET /api/v1/services/status", s.HandleServicesStatus)
	rr := doRequest(t, mux, "GET", "/api/v1/services/status", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)

	for _, item := range data {
		svc := item.(map[string]any)
		if svc["name"] == "postgres" {
			if svc["status"] != "offline" {
				t.Errorf("expected postgres=offline on ping failure, got %v", svc["status"])
			}
			return
		}
	}
	t.Error("postgres service entry not found")
}
