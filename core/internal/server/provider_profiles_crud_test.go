package server

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

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
