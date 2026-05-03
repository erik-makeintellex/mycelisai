package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// ── POST /api/v1/triggers ─────────────────────────────────────────

func TestHandleCreateTrigger(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	now := time.Now()
	mock.ExpectQuery("INSERT INTO trigger_rules").
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
			AddRow(now, now))

	body := `{
		"name": "On completion",
		"event_pattern": "mission.completed",
		"target_mission_id": "m-target-1",
		"mode": "propose",
		"cooldown_seconds": 60,
		"max_depth": 5,
		"max_active_runs": 3,
		"is_active": true
	}`

	mux := setupMux(t, "POST /api/v1/triggers", s.HandleCreateTrigger)
	rr := doRequest(t, mux, "POST", "/api/v1/triggers", body)

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

func TestHandleCreateTrigger_DefaultMode(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	now := time.Now()
	mock.ExpectQuery("INSERT INTO trigger_rules").
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
			AddRow(now, now))

	// mode omitted → should default to "propose"
	body := `{
		"name": "No mode specified",
		"event_pattern": "mission.completed",
		"target_mission_id": "m-1"
	}`

	mux := setupMux(t, "POST /api/v1/triggers", s.HandleCreateTrigger)
	rr := doRequest(t, mux, "POST", "/api/v1/triggers", body)

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["mode"] != "propose" {
		t.Errorf("expected mode 'propose', got %v", data["mode"])
	}
}

func TestHandleCreateTrigger_MissingName(t *testing.T) {
	tsOpt, _ := withTriggerStore(t)
	s := newTestServer(tsOpt)

	body := `{"event_pattern": "mission.completed", "target_mission_id": "m-1"}`

	mux := setupMux(t, "POST /api/v1/triggers", s.HandleCreateTrigger)
	rr := doRequest(t, mux, "POST", "/api/v1/triggers", body)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCreateTrigger_MissingEventPattern(t *testing.T) {
	tsOpt, _ := withTriggerStore(t)
	s := newTestServer(tsOpt)

	body := `{"name": "test", "target_mission_id": "m-1"}`

	mux := setupMux(t, "POST /api/v1/triggers", s.HandleCreateTrigger)
	rr := doRequest(t, mux, "POST", "/api/v1/triggers", body)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCreateTrigger_MissingTarget(t *testing.T) {
	tsOpt, _ := withTriggerStore(t)
	s := newTestServer(tsOpt)

	body := `{"name": "test", "event_pattern": "mission.completed"}`

	mux := setupMux(t, "POST /api/v1/triggers", s.HandleCreateTrigger)
	rr := doRequest(t, mux, "POST", "/api/v1/triggers", body)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCreateTrigger_BadJSON(t *testing.T) {
	tsOpt, _ := withTriggerStore(t)
	s := newTestServer(tsOpt)

	mux := setupMux(t, "POST /api/v1/triggers", s.HandleCreateTrigger)
	rr := doRequest(t, mux, "POST", "/api/v1/triggers", "not json")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCreateTrigger_NilStore(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/triggers", s.HandleCreateTrigger)
	rr := doRequest(t, mux, "POST", "/api/v1/triggers", `{"name":"x","event_pattern":"y","target_mission_id":"z"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}
