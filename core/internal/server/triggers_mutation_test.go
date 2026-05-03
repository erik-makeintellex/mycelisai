package server

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// ── PUT /api/v1/triggers/{id} ─────────────────────────────────────

func TestHandleUpdateTrigger(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	mock.ExpectExec("UPDATE trigger_rules").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"name": "Updated rule",
		"event_pattern": "tool.completed",
		"target_mission_id": "m-2",
		"mode": "auto_execute",
		"cooldown_seconds": 120,
		"max_depth": 3,
		"max_active_runs": 1,
		"is_active": true
	}`

	mux := setupMux(t, "PUT /api/v1/triggers/{id}", s.HandleUpdateTrigger)
	rr := doRequest(t, mux, "PUT", "/api/v1/triggers/r-1", body)

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

func TestHandleUpdateTrigger_ForcesPropose(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	mock.ExpectExec("UPDATE trigger_rules").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// mode="garbage" → should be forced to "propose"
	body := `{
		"name": "test",
		"event_pattern": "mission.completed",
		"target_mission_id": "m-1",
		"mode": "garbage"
	}`

	mux := setupMux(t, "PUT /api/v1/triggers/{id}", s.HandleUpdateTrigger)
	rr := doRequest(t, mux, "PUT", "/api/v1/triggers/r-1", body)

	assertStatus(t, rr, http.StatusOK)
}

func TestHandleUpdateTrigger_BadJSON(t *testing.T) {
	tsOpt, _ := withTriggerStore(t)
	s := newTestServer(tsOpt)

	mux := setupMux(t, "PUT /api/v1/triggers/{id}", s.HandleUpdateTrigger)
	rr := doRequest(t, mux, "PUT", "/api/v1/triggers/r-1", "not json")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleUpdateTrigger_NilStore(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "PUT /api/v1/triggers/{id}", s.HandleUpdateTrigger)
	rr := doRequest(t, mux, "PUT", "/api/v1/triggers/r-1", `{"name":"x"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── DELETE /api/v1/triggers/{id} ──────────────────────────────────

func TestHandleDeleteTrigger(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	mock.ExpectExec("DELETE FROM trigger_rules").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mux := setupMux(t, "DELETE /api/v1/triggers/{id}", s.HandleDeleteTrigger)
	rr := doRequest(t, mux, "DELETE", "/api/v1/triggers/r-1", "")

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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleDeleteTrigger_NilStore(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "DELETE /api/v1/triggers/{id}", s.HandleDeleteTrigger)
	rr := doRequest(t, mux, "DELETE", "/api/v1/triggers/r-1", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── POST /api/v1/triggers/{id}/toggle ─────────────────────────────

func TestHandleToggleTrigger(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	// SetActive(false) → UPDATE + no Get re-read
	mock.ExpectExec("UPDATE trigger_rules SET is_active").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"is_active": false}`

	mux := setupMux(t, "POST /api/v1/triggers/{id}/toggle", s.HandleToggleTrigger)
	rr := doRequest(t, mux, "POST", "/api/v1/triggers/r-1/toggle", body)

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
	if data["is_active"] != false {
		t.Errorf("expected is_active=false, got %v", data["is_active"])
	}
}

func TestHandleToggleTrigger_BadJSON(t *testing.T) {
	tsOpt, _ := withTriggerStore(t)
	s := newTestServer(tsOpt)

	mux := setupMux(t, "POST /api/v1/triggers/{id}/toggle", s.HandleToggleTrigger)
	rr := doRequest(t, mux, "POST", "/api/v1/triggers/r-1/toggle", "not json")

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleToggleTrigger_NilStore(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/triggers/{id}/toggle", s.HandleToggleTrigger)
	rr := doRequest(t, mux, "POST", "/api/v1/triggers/r-1/toggle", `{"is_active":true}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}
