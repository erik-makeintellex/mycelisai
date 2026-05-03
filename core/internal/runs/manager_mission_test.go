package runs

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

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

	mock.ExpectQuery("SELECT .+ FROM mission_runs").WillReturnRows(sqlmock.NewRows(nil))
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
	if _, err := m.ListRunsForMission(context.Background(), "m-1", 10); err == nil {
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

	mock.ExpectQuery("SELECT .+ FROM mission_runs").WillReturnRows(sqlmock.NewRows(nil))
	if _, err := m.ListRunsForMission(context.Background(), "m-1", 0); err != nil {
		t.Fatalf("ListRunsForMission error: %v", err)
	}
}
