package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/events"
	"github.com/mycelis/core/internal/runs"
)

// ── Local test helpers ─────────────────────────────────────────────

// withEventsStore wires a real events.Store (backed by sqlmock) onto the server.
// NATS is nil — events persist to DB only (degraded mode, no panic).
func withEventsStore(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock (events): %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.Events = events.NewStore(db, nil)
	}, mock
}

// withRunsManager wires a real runs.Manager (backed by sqlmock) onto the server.
func withRunsManager(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock (runs): %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.Runs = runs.NewManager(db)
	}, mock
}

// ── GET /api/v1/runs/{id}/events ───────────────────────────────────

func TestHandleGetRunEvents(t *testing.T) {
	evOpt, mock := withEventsStore(t)
	s := newTestServer(evOpt)

	runID := "aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM mission_events").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "run_id", "tenant_id", "event_type", "severity",
			"source_agent", "source_team", "payload", "audit_event_id", "emitted_at",
		}).
			AddRow("ev-1", runID, "default", "mission.started", "info",
				"soma", "admin-core", `{"mission_id":"m-1"}`, "", now).
			AddRow("ev-2", runID, "default", "tool.invoked", "info",
				"coder", "council-core", `{}`, "", now))

	mux := setupMux(t, "GET /api/v1/runs/{id}/events", s.handleGetRunEvents)
	rr := doRequest(t, mux, "GET", "/api/v1/runs/"+runID+"/events", "")

	assertStatus(t, rr, http.StatusOK)

	var timeline []map[string]any
	assertJSON(t, rr, &timeline)
	if len(timeline) != 2 {
		t.Errorf("expected 2 events, got %d", len(timeline))
	}
	if timeline[0]["event_type"] != "mission.started" {
		t.Errorf("expected event_type 'mission.started', got %v", timeline[0]["event_type"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleGetRunEvents_Empty(t *testing.T) {
	evOpt, mock := withEventsStore(t)
	s := newTestServer(evOpt)

	mock.ExpectQuery("SELECT .+ FROM mission_events").
		WillReturnRows(sqlmock.NewRows(nil))

	mux := setupMux(t, "GET /api/v1/runs/{id}/events", s.handleGetRunEvents)
	rr := doRequest(t, mux, "GET", "/api/v1/runs/run-no-events/events", "")

	assertStatus(t, rr, http.StatusOK)

	var timeline []map[string]any
	assertJSON(t, rr, &timeline)
	if len(timeline) != 0 {
		t.Errorf("expected empty array, got %d events", len(timeline))
	}
}

func TestHandleGetRunEvents_NilStore(t *testing.T) {
	s := newTestServer() // no Events wired → s.Events == nil
	mux := setupMux(t, "GET /api/v1/runs/{id}/events", s.handleGetRunEvents)
	rr := doRequest(t, mux, "GET", "/api/v1/runs/run-1/events", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── GET /api/v1/runs/{id}/chain ────────────────────────────────────

func TestHandleGetRunChain(t *testing.T) {
	runsOpt, mock := withRunsManager(t)
	s := newTestServer(runsOpt)

	runID := "bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	missionID := "cccc3333-cccc-cccc-cccc-cccccccccccc"
	now := time.Now()

	// GetRun query
	mock.ExpectQuery("SELECT .+ FROM mission_runs WHERE").
		WithArgs(runID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "mission_id", "tenant_id", "status", "run_depth",
			"parent_run_id", "started_at", "completed_at",
		}).AddRow(runID, missionID, "default", "running", 0, "", now, nil))

	// ListRunsForMission query
	mock.ExpectQuery("SELECT .+ FROM mission_runs").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "mission_id", "tenant_id", "status", "run_depth",
			"parent_run_id", "started_at", "completed_at",
		}).
			AddRow(runID, missionID, "default", "running", 0, "", now, nil).
			AddRow("prev-run-id", missionID, "default", "completed", 0, "", now, now))

	mux := setupMux(t, "GET /api/v1/runs/{id}/chain", s.handleGetRunChain)
	rr := doRequest(t, mux, "GET", "/api/v1/runs/"+runID+"/chain", "")

	assertStatus(t, rr, http.StatusOK)

	var chain map[string]any
	assertJSON(t, rr, &chain)
	if chain["run_id"] != runID {
		t.Errorf("expected run_id=%s, got %v", runID, chain["run_id"])
	}
	if chain["mission_id"] != missionID {
		t.Errorf("expected mission_id=%s, got %v", missionID, chain["mission_id"])
	}
	chainRuns, ok := chain["chain"].([]any)
	if !ok {
		t.Fatalf("expected chain array, got %T", chain["chain"])
	}
	if len(chainRuns) != 2 {
		t.Errorf("expected 2 runs in chain, got %d", len(chainRuns))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestHandleGetRunChain_NotFound(t *testing.T) {
	runsOpt, mock := withRunsManager(t)
	s := newTestServer(runsOpt)

	// GetRun returns no rows → not found
	mock.ExpectQuery("SELECT .+ FROM mission_runs WHERE").
		WillReturnRows(sqlmock.NewRows(nil))

	mux := setupMux(t, "GET /api/v1/runs/{id}/chain", s.handleGetRunChain)
	rr := doRequest(t, mux, "GET", "/api/v1/runs/nonexistent-run/chain", "")

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleGetRunChain_NilStore(t *testing.T) {
	s := newTestServer() // no Runs wired → s.Runs == nil
	mux := setupMux(t, "GET /api/v1/runs/{id}/chain", s.handleGetRunChain)
	rr := doRequest(t, mux, "GET", "/api/v1/runs/run-1/chain", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}
