package server

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/triggers"
)

func TestProposeScheduleRule_RecordsProposalAndNextRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := triggers.NewStore(db)
	now := time.Now()
	mock.ExpectExec("UPDATE trigger_rules").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO trigger_executions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	ok := proposeScheduleRule(t.Context(), store, triggers.TriggerRule{
		ID:                      "r-1",
		TriggerKind:             "schedule",
		ScheduleIntervalSeconds: 900,
	}, now)
	if !ok {
		t.Fatal("expected proposal to be recorded")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestProposeScheduleRule_InvalidIntervalSkips(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := triggers.NewStore(db)
	mock.ExpectExec("INSERT INTO trigger_executions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	ok := proposeScheduleRule(t.Context(), store, triggers.TriggerRule{ID: "r-1", TriggerKind: "schedule"}, time.Now())
	if ok {
		t.Fatal("expected invalid schedule to skip")
	}
}
