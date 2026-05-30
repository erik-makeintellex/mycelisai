package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHandleListTriggers(t *testing.T) {
	tsOpt, mock := withTriggerStore(t)
	s := newTestServer(tsOpt)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM trigger_rules").
		WillReturnRows(sqlmock.NewRows(triggerRuleColumns).
			AddRow("r-1", "default", "Rule A", "desc", "event", "mission.completed",
				[]byte(`{}`), "m-target-1", "propose", 60, 5, 3, true, nil, 0, nil, "", "", now, now).
			AddRow("r-2", "default", "Rule B", "", "event", "tool.completed",
				[]byte(`{}`), "m-target-2", "auto_execute", 120, 3, 1, false, nil, 0, nil, "", "", now, now))

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
