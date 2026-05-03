package triggers

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

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
