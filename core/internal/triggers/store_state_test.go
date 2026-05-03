package triggers

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

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
