package runs

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

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
	if runs[0].ID != "run-3" {
		t.Errorf("expected newest run first, got %s", runs[0].ID)
	}
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

	mock.ExpectQuery("SELECT .+ FROM mission_runs").WillReturnRows(sqlmock.NewRows(nil))
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
	if _, err := m.ListRecentRuns(context.Background(), "default", 20); err == nil {
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

	mock.ExpectQuery("SELECT .+ FROM mission_runs").WillReturnRows(sqlmock.NewRows(nil))
	if _, err := m.ListRecentRuns(context.Background(), "", 0); err != nil {
		t.Fatalf("ListRecentRuns error: %v", err)
	}
}
