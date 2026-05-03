package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

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
