package runs

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// ── CreateRun ──────────────────────────────────────────────────────

func TestCreateRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	mock.ExpectExec("INSERT INTO mission_runs").
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := m.CreateRun(context.Background(), "mission-abc")
	if err != nil {
		t.Fatalf("CreateRun error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty run ID")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestCreateRun_EmptyMissionID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	_, err = m.CreateRun(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty mission_id")
	}
}

func TestCreateRun_NilDB(t *testing.T) {
	m := &Manager{db: nil}
	_, err := m.CreateRun(context.Background(), "m-1")
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── CreateChildRun ─────────────────────────────────────────────────

func TestCreateChildRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	mock.ExpectExec("INSERT INTO mission_runs").
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := m.CreateChildRun(context.Background(), "m-1", "parent-run-1", 1)
	if err != nil {
		t.Fatalf("CreateChildRun error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty child run ID")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestCreateChildRun_NilDB(t *testing.T) {
	m := &Manager{db: nil}
	_, err := m.CreateChildRun(context.Background(), "m-1", "parent-1", 1)
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── UpdateRunStatus ────────────────────────────────────────────────

func TestUpdateRunStatus_Running(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	// Non-terminal status: no completed_at update
	mock.ExpectExec("UPDATE mission_runs SET status").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := m.UpdateRunStatus(context.Background(), "run-1", StatusRunning); err != nil {
		t.Fatalf("UpdateRunStatus error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestUpdateRunStatus_Completed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	// Terminal status: includes completed_at
	mock.ExpectExec("UPDATE mission_runs SET status").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := m.UpdateRunStatus(context.Background(), "run-1", StatusCompleted); err != nil {
		t.Fatalf("UpdateRunStatus(completed) error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestUpdateRunStatus_Failed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	mock.ExpectExec("UPDATE mission_runs SET status").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := m.UpdateRunStatus(context.Background(), "run-1", StatusFailed); err != nil {
		t.Fatalf("UpdateRunStatus(failed) error: %v", err)
	}
}

func TestUpdateRunStatus_NilDB(t *testing.T) {
	m := &Manager{db: nil}
	if err := m.UpdateRunStatus(context.Background(), "run-1", StatusCompleted); err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── GetRun ─────────────────────────────────────────────────────────

func TestGetRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	runID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	missionID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM mission_runs WHERE").
		WithArgs(runID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "mission_id", "tenant_id", "status", "run_depth",
			"parent_run_id", "started_at", "completed_at",
		}).AddRow(runID, missionID, "default", StatusRunning, 0, "", now, nil))

	run, err := m.GetRun(context.Background(), runID)
	if err != nil {
		t.Fatalf("GetRun error: %v", err)
	}
	if run.ID != runID {
		t.Errorf("expected ID %s, got %s", runID, run.ID)
	}
	if run.MissionID != missionID {
		t.Errorf("expected mission_id %s, got %s", missionID, run.MissionID)
	}
	if run.Status != StatusRunning {
		t.Errorf("expected status %q, got %q", StatusRunning, run.Status)
	}
	if run.CompletedAt != nil {
		t.Error("expected nil completed_at for running run")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestGetRun_WithCompletedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	runID := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM mission_runs WHERE").
		WithArgs(runID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "mission_id", "tenant_id", "status", "run_depth",
			"parent_run_id", "started_at", "completed_at",
		}).AddRow(runID, "m-1", "default", StatusCompleted, 0, "", now, now))

	run, err := m.GetRun(context.Background(), runID)
	if err != nil {
		t.Fatalf("GetRun error: %v", err)
	}
	if run.CompletedAt == nil {
		t.Error("expected non-nil completed_at for completed run")
	}
}

func TestGetRun_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	mock.ExpectQuery("SELECT .+ FROM mission_runs WHERE").
		WillReturnRows(sqlmock.NewRows(nil))

	_, err = m.GetRun(context.Background(), "nonexistent-run-id")
	if err == nil {
		t.Error("expected not-found error")
	}
}

func TestGetRun_NilDB(t *testing.T) {
	m := &Manager{db: nil}
	_, err := m.GetRun(context.Background(), "run-1")
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── ListRunsForMission ─────────────────────────────────────────────

func TestListRunsForMission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM mission_runs").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "mission_id", "tenant_id", "status", "run_depth",
			"parent_run_id", "started_at", "completed_at",
		}).
			AddRow("run-2", "m-1", "default", StatusRunning, 0, "", now, nil).
			AddRow("run-1", "m-1", "default", StatusCompleted, 0, "", now, now))

	runs, err := m.ListRunsForMission(context.Background(), "m-1", 10)
	if err != nil {
		t.Fatalf("ListRunsForMission error: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
	if runs[0].ID != "run-2" {
		t.Errorf("expected newest run first, got %s", runs[0].ID)
	}
}

func TestListRunsForMission_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	mock.ExpectQuery("SELECT .+ FROM mission_runs").
		WillReturnRows(sqlmock.NewRows(nil))

	runs, err := m.ListRunsForMission(context.Background(), "m-99", 10)
	if err != nil {
		t.Fatalf("ListRunsForMission error: %v", err)
	}
	if runs == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestListRunsForMission_NilDB(t *testing.T) {
	m := &Manager{db: nil}
	_, err := m.ListRunsForMission(context.Background(), "m-1", 10)
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

func TestListRunsForMission_DefaultLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	// limit <= 0 should be clamped to 20
	mock.ExpectQuery("SELECT .+ FROM mission_runs").
		WillReturnRows(sqlmock.NewRows(nil))

	if _, err := m.ListRunsForMission(context.Background(), "m-1", 0); err != nil {
		t.Fatalf("ListRunsForMission error: %v", err)
	}
}

// ── ListRecentRuns ─────────────────────────────────────────────────

func TestListRecentRuns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM mission_runs").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "mission_id", "tenant_id", "status", "run_depth",
			"parent_run_id", "started_at", "completed_at",
		}).
			AddRow("run-3", "m-1", "default", StatusRunning, 0, "", now, nil).
			AddRow("run-2", "m-2", "default", StatusCompleted, 0, "", now, now).
			AddRow("run-1", "m-1", "default", StatusFailed, 0, "", now, now))

	runs, err := m.ListRecentRuns(context.Background(), "default", 10)
	if err != nil {
		t.Fatalf("ListRecentRuns error: %v", err)
	}
	if len(runs) != 3 {
		t.Errorf("expected 3 runs, got %d", len(runs))
	}
	// Newest first
	if runs[0].ID != "run-3" {
		t.Errorf("expected newest run first, got %s", runs[0].ID)
	}
	// Active run has nil completed_at
	if runs[0].CompletedAt != nil {
		t.Error("expected nil completed_at for running run")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestListRecentRuns_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	mock.ExpectQuery("SELECT .+ FROM mission_runs").
		WillReturnRows(sqlmock.NewRows(nil))

	runs, err := m.ListRecentRuns(context.Background(), "default", 20)
	if err != nil {
		t.Fatalf("ListRecentRuns error: %v", err)
	}
	if runs == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestListRecentRuns_NilDB(t *testing.T) {
	m := &Manager{db: nil}
	_, err := m.ListRecentRuns(context.Background(), "default", 20)
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

func TestListRecentRuns_DefaultLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	m := NewManager(db)

	// limit <= 0 should be clamped to 20, empty tenant → "default"
	mock.ExpectQuery("SELECT .+ FROM mission_runs").
		WillReturnRows(sqlmock.NewRows(nil))

	if _, err := m.ListRecentRuns(context.Background(), "", 0); err != nil {
		t.Fatalf("ListRecentRuns error: %v", err)
	}
}
