package triggers

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreate_ScheduleRuleDefaultsToPropose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	mock.ExpectQuery("INSERT INTO trigger_rules").
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))

	rule := &TriggerRule{
		Name:                    "Hourly review",
		TriggerKind:             "schedule",
		TargetMissionID:         "mission-review",
		Mode:                    "auto_execute",
		ScheduleIntervalSeconds: 3600,
		IsActive:                true,
	}
	if err := s.Create(context.Background(), rule); err != nil {
		t.Fatalf("Create schedule error: %v", err)
	}
	if rule.Mode != "propose" {
		t.Fatalf("schedule mode = %q, want propose", rule.Mode)
	}
	if rule.EventPattern != "scheduler.due" {
		t.Fatalf("event pattern = %q, want scheduler.due", rule.EventPattern)
	}
	if rule.NextRunAt == nil {
		t.Fatal("expected next_run_at to be set")
	}
}

func TestListDueScheduleRules(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM trigger_rules").
		WithArgs(now, 10).
		WillReturnRows(sqlmock.NewRows(ruleColumns).
			AddRow("r-1", "default", "Hourly review", "", "schedule", "scheduler.due",
				[]byte(`{}`), "mission-review", "propose", 60, 5, 3, true, nil, 3600, now, "proof", "retry", now, now))

	rules, err := s.ListDueScheduleRules(context.Background(), now, 10)
	if err != nil {
		t.Fatalf("ListDueScheduleRules error: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("rules len = %d, want 1", len(rules))
	}
	if rules[0].TriggerKind != "schedule" || rules[0].ScheduleIntervalSeconds != 3600 {
		t.Fatalf("unexpected schedule rule: %+v", rules[0])
	}
}

func TestMarkScheduleProposed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	next := now.Add(time.Hour)
	s.cache["r-1"] = &TriggerRule{ID: "r-1", TriggerKind: "schedule"}
	mock.ExpectExec("UPDATE trigger_rules").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := s.MarkScheduleProposed(context.Background(), "r-1", now, next); err != nil {
		t.Fatalf("MarkScheduleProposed error: %v", err)
	}
	if got := s.cache["r-1"].NextRunAt; got == nil || !got.Equal(next) {
		t.Fatalf("next_run_at = %v, want %v", got, next)
	}
}
