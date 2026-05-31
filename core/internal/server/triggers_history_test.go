package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// ── GET /api/v1/triggers/{id}/history ─────────────────────────────

func TestHandleTriggerHistory(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(triggerExecColumns).
			AddRow("exec-1", "r-1", "ev-1", "run-1", "fired", "", "", "", "", "", []byte(`{}`), now).
			AddRow("exec-2", "r-1", "ev-2", "", "skipped", "cooldown", "", "", "", "", []byte(`{}`), now))

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

func TestHandleTriggerHistory_ReturnsScheduleHandoffRefs(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(triggerExecColumns).
			AddRow("exec-1", "r-1", "schedule:r-1:due", "", "proposed", "",
				"schedule:r-1:due", "11111111-1111-1111-1111-111111111111",
				"22222222-2222-2222-2222-222222222222", "awaiting_approval",
				[]byte(`{"autonomous_execution":false}`), now))

	mux := setupMux(t, "GET /api/v1/triggers/{id}/history", s.HandleTriggerHistory)
	rr := doRequest(t, mux, "GET", "/api/v1/triggers/r-1/history", "")

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].([]any)
	first := data[0].(map[string]any)
	if first["handoff_key"] != "schedule:r-1:due" || first["proposal_status"] != "awaiting_approval" {
		t.Fatalf("missing handoff refs: %#v", first)
	}
	if first["run_id"] != nil {
		t.Fatalf("schedule handoff unexpectedly returned run_id: %#v", first["run_id"])
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
