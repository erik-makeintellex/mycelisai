package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/triggers"
)

// ── Local test helpers ─────────────────────────────────────────────

var triggerRuleColumns = []string{
	"id", "tenant_id", "name", "description", "event_pattern",
	"condition", "target_mission_id", "mode", "cooldown_seconds",
	"max_depth", "max_active_runs", "is_active", "last_fired_at",
	"created_at", "updated_at",
}

var triggerExecColumns = []string{
	"id", "rule_id", "event_id", "run_id", "status", "skip_reason", "executed_at",
}

// withTriggerStore wires a real triggers.Store (backed by sqlmock) onto the server.
func withTriggerStore(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock (triggers): %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.Triggers = triggers.NewStore(db)
	}, mock
}

// ── GET /api/v1/triggers ──────────────────────────────────────────

func TestHandleListTriggers(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM trigger_rules").
		WillReturnRows(sqlmock.NewRows(triggerRuleColumns).
			AddRow("r-1", "default", "Rule A", "desc", "mission.completed",
				[]byte(`{}`), "m-target-1", "propose", 60, 5, 3, true, nil, now, now).
			AddRow("r-2", "default", "Rule B", "", "tool.completed",
				[]byte(`{}`), "m-target-2", "auto_execute", 120, 3, 1, false, nil, now, now))

	mux := setupMux(t, "GET /api/v1/triggers", s.HandleListTriggers)
	rr := doRequest(t, mux, "GET", "/api/v1/triggers", "")

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
		t.Errorf("expected 2 rules, got %d", len(data))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleListTriggers_Empty(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	mock.ExpectQuery("SELECT .+ FROM trigger_rules").
		WillReturnRows(sqlmock.NewRows(nil))

	mux := setupMux(t, "GET /api/v1/triggers", s.HandleListTriggers)
	rr := doRequest(t, mux, "GET", "/api/v1/triggers", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 0 {
		t.Errorf("expected empty array, got %d rules", len(data))
	}
}

func TestHandleListTriggers_NilStore(t *testing.T) {
	s := newTestServer() // no Triggers wired → returns empty array
	mux := setupMux(t, "GET /api/v1/triggers", s.HandleListTriggers)
	rr := doRequest(t, mux, "GET", "/api/v1/triggers", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 0 {
		t.Errorf("expected empty array for nil store, got %d", len(data))
	}
}

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

// ── GET /api/v1/triggers/{id}/history ─────────────────────────────

func TestHandleTriggerHistory(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(triggerExecColumns).
			AddRow("exec-1", "r-1", "ev-1", "run-1", "fired", "", now).
			AddRow("exec-2", "r-1", "ev-2", "", "skipped", "cooldown", now))

	mux := setupMux(t, "GET /api/v1/triggers/{id}/history", s.HandleTriggerHistory)
	rr := doRequest(t, mux, "GET", "/api/v1/triggers/r-1/history", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 2 {
		t.Errorf("expected 2 executions, got %d", len(data))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleTriggerHistory_Empty(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(nil))

	mux := setupMux(t, "GET /api/v1/triggers/{id}/history", s.HandleTriggerHistory)
	rr := doRequest(t, mux, "GET", "/api/v1/triggers/r-1/history", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 0 {
		t.Errorf("expected 0 executions, got %d", len(data))
	}
}

func TestHandleTriggerHistory_NilStore(t *testing.T) {
	s := newTestServer() // no Triggers wired → returns empty array
	mux := setupMux(t, "GET /api/v1/triggers/{id}/history", s.HandleTriggerHistory)
	rr := doRequest(t, mux, "GET", "/api/v1/triggers/r-1/history", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 0 {
		t.Errorf("expected empty array for nil store, got %d", len(data))
	}
}
