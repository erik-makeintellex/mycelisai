package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

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
