package events

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

// ── Emit ───────────────────────────────────────────────────────────

func TestEmit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db, nil) // nil NATS = degraded mode; events still persist

	mock.ExpectExec("INSERT INTO mission_events").
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := s.Emit(
		context.Background(),
		"run-aaaa-1111",
		protocol.EventMissionStarted,
		protocol.SeverityInfo,
		"soma", "admin-core",
		map[string]interface{}{"mission_id": "m-1"},
	)
	if err != nil {
		t.Fatalf("Emit error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty event ID")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestEmit_EmptyPayload(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db, nil)

	mock.ExpectExec("INSERT INTO mission_events").
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := s.Emit(
		context.Background(),
		"run-bbbb-2222",
		protocol.EventToolInvoked,
		protocol.SeverityInfo,
		"coder", "council-core",
		nil, // nil payload → marshals as "{}"
	)
	if err != nil {
		t.Fatalf("Emit (nil payload) error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty event ID")
	}
}

func TestEmit_EmptyRunID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db, nil)

	_, err = s.Emit(
		context.Background(),
		"", // empty run_id — must be rejected
		protocol.EventMissionStarted,
		protocol.SeverityInfo,
		"", "", nil,
	)
	if err == nil {
		t.Error("expected error for empty run_id")
	}
}

func TestEmit_NilDB(t *testing.T) {
	s := &Store{db: nil, nc: nil}

	_, err := s.Emit(
		context.Background(),
		"run-cccc-3333",
		protocol.EventMissionStarted,
		protocol.SeverityInfo,
		"", "", nil,
	)
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── GetRunTimeline ─────────────────────────────────────────────────

func TestGetRunTimeline(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db, nil)

	runID := "run-dddd-4444"
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM mission_events").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "run_id", "tenant_id", "event_type", "severity",
			"source_agent", "source_team", "payload", "audit_event_id", "emitted_at",
		}).
			AddRow("ev-1", runID, "default", "mission.started", "info",
				"soma", "admin-core", `{"mission_id":"m-1"}`, "", now).
			AddRow("ev-2", runID, "default", "tool.invoked", "info",
				"coder", "council-core", `{"tool":"read_file"}`, "", now))

	events, err := s.GetRunTimeline(context.Background(), runID)
	if err != nil {
		t.Fatalf("GetRunTimeline error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].EventType != protocol.EventMissionStarted {
		t.Errorf("expected %q, got %q", protocol.EventMissionStarted, events[0].EventType)
	}
	if events[0].SourceAgent != "soma" {
		t.Errorf("expected source_agent 'soma', got %q", events[0].SourceAgent)
	}
	if events[0].Payload["mission_id"] != "m-1" {
		t.Errorf("expected payload.mission_id='m-1', got %v", events[0].Payload["mission_id"])
	}
	if events[1].EventType != protocol.EventToolInvoked {
		t.Errorf("expected %q, got %q", protocol.EventToolInvoked, events[1].EventType)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestGetRunTimeline_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db, nil)

	mock.ExpectQuery("SELECT .+ FROM mission_events").
		WillReturnRows(sqlmock.NewRows(nil))

	events, err := s.GetRunTimeline(context.Background(), "run-no-events")
	if err != nil {
		t.Fatalf("GetRunTimeline error: %v", err)
	}
	if events == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestGetRunTimeline_EmptyPayload(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db, nil)

	now := time.Now()
	// Payload is "{}" — should not be parsed (len == 2)
	mock.ExpectQuery("SELECT .+ FROM mission_events").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "run_id", "tenant_id", "event_type", "severity",
			"source_agent", "source_team", "payload", "audit_event_id", "emitted_at",
		}).AddRow("ev-x", "run-1", "default", "agent.started", "debug",
			"", "", `{}`, "", now))

	events, err := s.GetRunTimeline(context.Background(), "run-1")
	if err != nil {
		t.Fatalf("GetRunTimeline error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	// Payload should be nil (skipped because len <= 2)
	if events[0].Payload != nil {
		t.Errorf("expected nil payload for '{}', got %v", events[0].Payload)
	}
}

func TestGetRunTimeline_NilDB(t *testing.T) {
	s := &Store{db: nil, nc: nil}
	_, err := s.GetRunTimeline(context.Background(), "run-1")
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── summarizePayload ───────────────────────────────────────────────

func TestSummarizePayload_Empty(t *testing.T) {
	result := summarizePayload(nil)
	if result != nil {
		t.Errorf("expected nil for empty payload, got %v", result)
	}
}

func TestSummarizePayload_WithKeys(t *testing.T) {
	result := summarizePayload(map[string]interface{}{
		"tool":    "read_file",
		"run_id":  "run-1",
		"success": true,
	})
	if result == nil {
		t.Fatal("expected non-nil summary")
	}
	keys, ok := result["keys"].([]string)
	if !ok {
		t.Fatalf("expected []string keys, got %T", result["keys"])
	}
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}
