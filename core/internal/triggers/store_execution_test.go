package triggers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// ── UpdateLastFired ───────────────────────────────────────────────

func TestUpdateLastFired(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	s.cache["r-1"] = &TriggerRule{ID: "r-1"}

	mock.ExpectExec("UPDATE trigger_rules SET last_fired_at").
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.UpdateLastFired(context.Background(), "r-1", now)

	s.mu.RLock()
	r := s.cache["r-1"]
	s.mu.RUnlock()

	if r.LastFiredAt == nil {
		t.Error("expected LastFiredAt to be set")
	}
	if !r.LastFiredAt.Equal(now) {
		t.Errorf("expected %v, got %v", now, *r.LastFiredAt)
	}
}

func TestUpdateLastFired_NilDB(t *testing.T) {
	// Should not panic
	s := &Store{db: nil, cache: make(map[string]*TriggerRule)}
	s.UpdateLastFired(context.Background(), "r-1", time.Now())
}

// ── LogExecution ──────────────────────────────────────────────────

func TestLogExecution(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectExec("INSERT INTO trigger_executions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	exec := &TriggerExecution{
		RuleID:  "r-1",
		EventID: "ev-1",
		RunID:   "run-1",
		Status:  "fired",
	}

	if err := s.LogExecution(context.Background(), exec); err != nil {
		t.Fatalf("LogExecution error: %v", err)
	}
	if exec.ID == "" {
		t.Error("expected non-empty execution ID")
	}
}

func TestLogExecution_Skipped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectExec("INSERT INTO trigger_executions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	exec := &TriggerExecution{
		RuleID:     "r-1",
		EventID:    "ev-2",
		Status:     "skipped",
		SkipReason: "cooldown: 30s since last fire, cooldown is 60s",
	}

	if err := s.LogExecution(context.Background(), exec); err != nil {
		t.Fatalf("LogExecution error: %v", err)
	}
}

func TestLogExecution_NilDB(t *testing.T) {
	s := &Store{db: nil, cache: make(map[string]*TriggerRule)}
	err := s.LogExecution(context.Background(), &TriggerExecution{RuleID: "r-1", EventID: "ev-1", Status: "fired"})
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── ListExecutions ────────────────────────────────────────────────

func TestListExecutions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(execColumns).
			AddRow("exec-1", "r-1", "ev-1", "run-1", "fired", "", now).
			AddRow("exec-2", "r-1", "ev-2", "", "skipped", "cooldown", now))

	execs, err := s.ListExecutions(context.Background(), "r-1", 20)
	if err != nil {
		t.Fatalf("ListExecutions error: %v", err)
	}
	if len(execs) != 2 {
		t.Errorf("expected 2 executions, got %d", len(execs))
	}
	if execs[0].Status != "fired" {
		t.Errorf("expected status 'fired', got %q", execs[0].Status)
	}
	if execs[1].Status != "skipped" {
		t.Errorf("expected status 'skipped', got %q", execs[1].Status)
	}
}

func TestListExecutions_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(nil))

	execs, err := s.ListExecutions(context.Background(), "r-1", 20)
	if err != nil {
		t.Fatalf("ListExecutions error: %v", err)
	}
	if execs == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(execs) != 0 {
		t.Errorf("expected 0 executions, got %d", len(execs))
	}
}

func TestListExecutions_DefaultLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(nil))

	// limit <= 0 should be clamped to 20
	if _, err := s.ListExecutions(context.Background(), "r-1", 0); err != nil {
		t.Fatalf("ListExecutions error: %v", err)
	}
}

func TestListExecutions_NilDB(t *testing.T) {
	s := &Store{db: nil, cache: make(map[string]*TriggerRule)}
	_, err := s.ListExecutions(context.Background(), "r-1", 20)
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── ActiveCount ───────────────────────────────────────────────────

func TestActiveCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	count, err := s.ActiveCount(context.Background(), "m-target")
	if err != nil {
		t.Fatalf("ActiveCount error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestActiveCount_Zero(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	count, err := s.ActiveCount(context.Background(), "m-target")
	if err != nil {
		t.Fatalf("ActiveCount error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestActiveCount_NilDB(t *testing.T) {
	s := &Store{db: nil, cache: make(map[string]*TriggerRule)}
	_, err := s.ActiveCount(context.Background(), "m-1")
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── Create with explicit condition ────────────────────────────────

func TestCreate_WithCondition(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	mock.ExpectQuery("INSERT INTO trigger_rules").
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
			AddRow(now, now))

	rule := &TriggerRule{
		Name:            "With Condition",
		EventPattern:    "tool.completed",
		TargetMissionID: "m-1",
		Condition:       json.RawMessage(`{"tool":"write_file"}`),
		Mode:            "auto_execute",
		CooldownSeconds: 120,
		MaxDepth:        3,
		MaxActiveRuns:   1,
		IsActive:        false,
	}

	if err := s.Create(context.Background(), rule); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if rule.Mode != "auto_execute" {
		t.Errorf("expected mode 'auto_execute', got %q", rule.Mode)
	}

	// Not in cache (isActive=false)
	s.mu.RLock()
	_, ok := s.cache[rule.ID]
	s.mu.RUnlock()
	if ok {
		t.Error("expected inactive rule not in cache")
	}
}
