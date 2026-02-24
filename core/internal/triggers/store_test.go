package triggers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// ── Column helpers ────────────────────────────────────────────────

var ruleColumns = []string{
	"id", "tenant_id", "name", "description", "event_pattern",
	"condition", "target_mission_id", "mode", "cooldown_seconds",
	"max_depth", "max_active_runs", "is_active", "last_fired_at",
	"created_at", "updated_at",
}

var execColumns = []string{
	"id", "rule_id", "event_id", "run_id", "status", "skip_reason", "executed_at",
}

// ── LoadActiveRules ───────────────────────────────────────────────

func TestLoadActiveRules(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM trigger_rules").
		WillReturnRows(sqlmock.NewRows(ruleColumns).
			AddRow("r-1", "default", "Rule A", "desc", "mission.completed",
				[]byte(`{}`), "m-target-1", "propose", 60, 5, 3, true, nil, now, now).
			AddRow("r-2", "default", "Rule B", "", "tool.completed",
				[]byte(`{}`), "m-target-2", "auto_execute", 120, 3, 1, true, now, now, now))

	if err := s.LoadActiveRules(context.Background()); err != nil {
		t.Fatalf("LoadActiveRules error: %v", err)
	}

	s.mu.RLock()
	count := len(s.cache)
	s.mu.RUnlock()

	if count != 2 {
		t.Errorf("expected 2 cached rules, got %d", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestLoadActiveRules_NilDB(t *testing.T) {
	s := &Store{db: nil, cache: make(map[string]*TriggerRule)}
	if err := s.LoadActiveRules(context.Background()); err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── MatchingRules ─────────────────────────────────────────────────

func TestMatchingRules(t *testing.T) {
	s := &Store{cache: make(map[string]*TriggerRule)}
	s.cache["r-1"] = &TriggerRule{ID: "r-1", EventPattern: "mission.completed", IsActive: true}
	s.cache["r-2"] = &TriggerRule{ID: "r-2", EventPattern: "tool.completed", IsActive: true}
	s.cache["r-3"] = &TriggerRule{ID: "r-3", EventPattern: "mission.completed", IsActive: false}

	matches := s.MatchingRules("mission.completed")
	if len(matches) != 1 {
		t.Errorf("expected 1 match, got %d", len(matches))
	}
	if matches[0].ID != "r-1" {
		t.Errorf("expected r-1, got %s", matches[0].ID)
	}
}

func TestMatchingRules_NoMatch(t *testing.T) {
	s := &Store{cache: make(map[string]*TriggerRule)}
	s.cache["r-1"] = &TriggerRule{ID: "r-1", EventPattern: "mission.completed", IsActive: true}

	matches := s.MatchingRules("tool.invoked")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

// ── ListAll ───────────────────────────────────────────────────────

func TestListAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM trigger_rules").
		WillReturnRows(sqlmock.NewRows(ruleColumns).
			AddRow("r-1", "default", "Rule A", "desc", "mission.completed",
				[]byte(`{}`), "m-1", "propose", 60, 5, 3, true, nil, now, now))

	rules, err := s.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll error: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Name != "Rule A" {
		t.Errorf("expected 'Rule A', got %q", rules[0].Name)
	}
}

func TestListAll_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectQuery("SELECT .+ FROM trigger_rules").
		WillReturnRows(sqlmock.NewRows(nil))

	rules, err := s.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll error: %v", err)
	}
	if rules == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(rules))
	}
}

func TestListAll_NilDB(t *testing.T) {
	s := &Store{db: nil, cache: make(map[string]*TriggerRule)}
	_, err := s.ListAll(context.Background())
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── Get ───────────────────────────────────────────────────────────

func TestGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM trigger_rules").
		WithArgs("r-1").
		WillReturnRows(sqlmock.NewRows(ruleColumns).
			AddRow("r-1", "default", "Rule A", "desc", "mission.completed",
				[]byte(`{}`), "m-1", "propose", 60, 5, 3, true, nil, now, now))

	rule, err := s.Get(context.Background(), "r-1")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if rule == nil {
		t.Fatal("expected non-nil rule")
	}
	if rule.ID != "r-1" {
		t.Errorf("expected ID 'r-1', got %q", rule.ID)
	}
	if rule.Mode != "propose" {
		t.Errorf("expected mode 'propose', got %q", rule.Mode)
	}
}

func TestGet_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectQuery("SELECT .+ FROM trigger_rules").
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows(nil))

	rule, err := s.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if rule != nil {
		t.Errorf("expected nil rule for not-found, got %+v", rule)
	}
}

func TestGet_NilDB(t *testing.T) {
	s := &Store{db: nil, cache: make(map[string]*TriggerRule)}
	_, err := s.Get(context.Background(), "r-1")
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── Create ────────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
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
		Name:            "Test Rule",
		EventPattern:    "mission.completed",
		TargetMissionID: "m-target",
		IsActive:        true,
	}

	if err := s.Create(context.Background(), rule); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if rule.ID == "" {
		t.Error("expected non-empty ID after create")
	}
	if rule.Mode != "propose" {
		t.Errorf("expected default mode 'propose', got %q", rule.Mode)
	}
	if rule.CooldownSeconds != 60 {
		t.Errorf("expected default cooldown 60, got %d", rule.CooldownSeconds)
	}
	if rule.MaxDepth != 5 {
		t.Errorf("expected default max_depth 5, got %d", rule.MaxDepth)
	}
	if rule.MaxActiveRuns != 3 {
		t.Errorf("expected default max_active_runs 3, got %d", rule.MaxActiveRuns)
	}

	// Should be in cache (isActive=true)
	s.mu.RLock()
	cached, ok := s.cache[rule.ID]
	s.mu.RUnlock()
	if !ok {
		t.Error("expected rule in cache after create")
	}
	if cached.Name != "Test Rule" {
		t.Errorf("cached name mismatch: %q", cached.Name)
	}
}

func TestCreate_ValidationErrors(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	tests := []struct {
		name string
		rule *TriggerRule
	}{
		{"missing name", &TriggerRule{EventPattern: "x", TargetMissionID: "m"}},
		{"missing event_pattern", &TriggerRule{Name: "x", TargetMissionID: "m"}},
		{"missing target_mission_id", &TriggerRule{Name: "x", EventPattern: "y"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.Create(context.Background(), tt.rule); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestCreate_DefaultCondition(t *testing.T) {
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
		Name:            "No Condition",
		EventPattern:    "mission.completed",
		TargetMissionID: "m-1",
	}

	if err := s.Create(context.Background(), rule); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if string(rule.Condition) != "{}" {
		t.Errorf("expected default condition '{}', got %q", string(rule.Condition))
	}
}

func TestCreate_NilDB(t *testing.T) {
	s := &Store{db: nil, cache: make(map[string]*TriggerRule)}
	err := s.Create(context.Background(), &TriggerRule{Name: "x", EventPattern: "y", TargetMissionID: "z"})
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── Update ────────────────────────────────────────────────────────

func TestUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectExec("UPDATE trigger_rules").
		WillReturnResult(sqlmock.NewResult(1, 1))

	rule := &TriggerRule{
		ID:              "r-1",
		Name:            "Updated",
		EventPattern:    "tool.completed",
		TargetMissionID: "m-2",
		Mode:            "propose",
		IsActive:        true,
	}

	if err := s.Update(context.Background(), rule); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	// Should be in cache
	s.mu.RLock()
	_, ok := s.cache["r-1"]
	s.mu.RUnlock()
	if !ok {
		t.Error("expected rule in cache after update with isActive=true")
	}
}

func TestUpdate_Deactivate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	// Pre-populate cache
	s.cache["r-1"] = &TriggerRule{ID: "r-1", IsActive: true}

	mock.ExpectExec("UPDATE trigger_rules").
		WillReturnResult(sqlmock.NewResult(1, 1))

	rule := &TriggerRule{
		ID:              "r-1",
		Name:            "Now Inactive",
		EventPattern:    "tool.completed",
		TargetMissionID: "m-2",
		Mode:            "propose",
		IsActive:        false,
	}

	if err := s.Update(context.Background(), rule); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	// Should be removed from cache
	s.mu.RLock()
	_, ok := s.cache["r-1"]
	s.mu.RUnlock()
	if ok {
		t.Error("expected rule removed from cache after deactivation")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectExec("UPDATE trigger_rules").
		WillReturnResult(sqlmock.NewResult(0, 0))

	rule := &TriggerRule{ID: "nonexistent", Name: "x", EventPattern: "y", TargetMissionID: "z", Mode: "propose"}
	if err := s.Update(context.Background(), rule); err == nil {
		t.Error("expected not-found error")
	}
}

// ── Delete ────────────────────────────────────────────────────────

func TestDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	// Pre-populate cache
	s.cache["r-1"] = &TriggerRule{ID: "r-1"}

	mock.ExpectExec("DELETE FROM trigger_rules").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := s.Delete(context.Background(), "r-1"); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	s.mu.RLock()
	_, ok := s.cache["r-1"]
	s.mu.RUnlock()
	if ok {
		t.Error("expected rule removed from cache after delete")
	}
}

func TestDelete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectExec("DELETE FROM trigger_rules").
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := s.Delete(context.Background(), "nonexistent"); err == nil {
		t.Error("expected not-found error")
	}
}

func TestDelete_NilDB(t *testing.T) {
	s := &Store{db: nil, cache: make(map[string]*TriggerRule)}
	if err := s.Delete(context.Background(), "r-1"); err == nil {
		t.Error("expected error for nil DB")
	}
}

// ── SetActive ─────────────────────────────────────────────────────

func TestSetActive_Deactivate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	s.cache["r-1"] = &TriggerRule{ID: "r-1", IsActive: true}

	mock.ExpectExec("UPDATE trigger_rules SET is_active").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := s.SetActive(context.Background(), "r-1", false); err != nil {
		t.Fatalf("SetActive(false) error: %v", err)
	}

	s.mu.RLock()
	_, ok := s.cache["r-1"]
	s.mu.RUnlock()
	if ok {
		t.Error("expected rule removed from cache on deactivation")
	}
}

func TestSetActive_Activate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()

	// SetActive(true) calls UPDATE, then Get() to re-read rule into cache
	mock.ExpectExec("UPDATE trigger_rules SET is_active").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT .+ FROM trigger_rules").
		WithArgs("r-1").
		WillReturnRows(sqlmock.NewRows(ruleColumns).
			AddRow("r-1", "default", "Rule A", "", "mission.completed",
				[]byte(`{}`), "m-1", "propose", 60, 5, 3, true, nil, now, now))

	if err := s.SetActive(context.Background(), "r-1", true); err != nil {
		t.Fatalf("SetActive(true) error: %v", err)
	}

	s.mu.RLock()
	_, ok := s.cache["r-1"]
	s.mu.RUnlock()
	if !ok {
		t.Error("expected rule in cache after activation")
	}
}

func TestSetActive_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectExec("UPDATE trigger_rules SET is_active").
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := s.SetActive(context.Background(), "nonexistent", true); err == nil {
		t.Error("expected not-found error")
	}
}

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
