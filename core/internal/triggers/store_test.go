package triggers

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

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
