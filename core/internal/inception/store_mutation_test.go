package inception

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestIncrementUsage_HappyPath(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectExec("UPDATE inception_recipes SET usage_count = usage_count \\+ 1").
		WithArgs("recipe-001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := store.IncrementUsage(context.Background(), "recipe-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertExpectationsMet(t, mock)
}

func TestIncrementUsage_NilDB(t *testing.T) {
	store := NewStore(nil)

	err := store.IncrementUsage(context.Background(), "recipe-001")
	if err != nil {
		t.Errorf("expected nil error for nil DB, got: %v", err)
	}
}

func TestUpdateQuality_HappyPath(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectExec("UPDATE inception_recipes SET quality_score = \\$1").
		WithArgs(0.95, "recipe-001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := store.UpdateQuality(context.Background(), "recipe-001", 0.95)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertExpectationsMet(t, mock)
}

func TestUpdateQuality_NilDB(t *testing.T) {
	store := NewStore(nil)

	err := store.UpdateQuality(context.Background(), "recipe-001", 0.5)
	if err == nil {
		t.Fatal("expected error for nil DB")
	}
	if err.Error() != "inception store: database not available" {
		t.Errorf("unexpected error message: %v", err)
	}
}
