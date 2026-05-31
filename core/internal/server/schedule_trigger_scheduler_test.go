package server

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/registry"
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

func TestProposeScheduleRuleHandoff_PersistsTrustRefsWithoutRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := triggers.NewStore(db)
	s := &AdminServer{DB: db, Registry: &registry.Service{DB: db}, Triggers: store}
	dueAt := time.Date(2026, 5, 30, 10, 0, 0, 0, time.UTC)
	rule := triggers.TriggerRule{
		ID:                      "11111111-1111-1111-1111-111111111111",
		Name:                    "Daily proof review",
		TriggerKind:             "schedule",
		TargetMissionID:         "mission-review",
		ScheduleIntervalSeconds: 900,
		ProofExpectations:       "Visible proof",
		RecoveryBehavior:        "Pause and inspect",
		NextRunAt:               &dueAt,
	}

	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(triggerExecColumns))
	mock.ExpectExec("INSERT INTO intent_proofs").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("INSERT INTO execution_contracts").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("22222222-2222-2222-2222-222222222222"))
	mock.ExpectExec("UPDATE trigger_rules").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO trigger_executions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if !s.proposeScheduleRuleHandoff(t.Context(), rule, dueAt.Add(time.Minute)) {
		t.Fatal("expected handoff proposal to be recorded")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestProposeScheduleRuleHandoff_ReusesExistingHandoff(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := triggers.NewStore(db)
	s := &AdminServer{DB: db, Triggers: store}
	dueAt := time.Date(2026, 5, 30, 10, 0, 0, 0, time.UTC)
	rule := triggers.TriggerRule{
		ID:                      "11111111-1111-1111-1111-111111111111",
		TriggerKind:             "schedule",
		TargetMissionID:         "mission-review",
		ScheduleIntervalSeconds: 900,
		NextRunAt:               &dueAt,
	}

	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(triggerExecColumns).
			AddRow("exec-1", rule.ID, scheduleRuleHandoffKey(rule.ID, dueAt), "", "proposed", "",
				scheduleRuleHandoffKey(rule.ID, dueAt), "33333333-3333-3333-3333-333333333333",
				"22222222-2222-2222-2222-222222222222", "awaiting_approval",
				[]byte(`{"autonomous_execution":false}`), dueAt))
	mock.ExpectExec("UPDATE trigger_rules").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if !s.proposeScheduleRuleHandoff(t.Context(), rule, dueAt.Add(time.Minute)) {
		t.Fatal("expected existing handoff to be reused")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestProposeScheduleRuleHandoff_CompletesPartialExistingHandoff(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := triggers.NewStore(db)
	s := &AdminServer{DB: db, Registry: &registry.Service{DB: db}, Triggers: store}
	dueAt := time.Date(2026, 5, 30, 10, 0, 0, 0, time.UTC)
	rule := triggers.TriggerRule{
		ID:                      "11111111-1111-1111-1111-111111111111",
		TriggerKind:             "schedule",
		TargetMissionID:         "mission-review",
		ScheduleIntervalSeconds: 900,
		NextRunAt:               &dueAt,
	}

	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(triggerExecColumns).
			AddRow("44444444-4444-4444-4444-444444444444", rule.ID, scheduleRuleHandoffKey(rule.ID, dueAt), "", "proposed", "",
				scheduleRuleHandoffKey(rule.ID, dueAt), "", "", "recorded", []byte(`{}`), dueAt))
	mock.ExpectExec("INSERT INTO intent_proofs").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("INSERT INTO execution_contracts").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("22222222-2222-2222-2222-222222222222"))
	mock.ExpectExec("UPDATE trigger_executions").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE trigger_rules").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if !s.proposeScheduleRuleHandoff(t.Context(), rule, dueAt.Add(time.Minute)) {
		t.Fatal("expected partial handoff to be completed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
