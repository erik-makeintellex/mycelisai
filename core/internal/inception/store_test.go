package inception

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func TestCreateRecipe_HappyPath(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery("INSERT INTO inception_recipes").
		WithArgs(
			"default",
			"deployment",
			"Deploy FastAPI",
			"deploy * to *",
			"{}",
			"deploy my api",
			"",
			"",
			"",
			"admin",
			pq.Array([]string{"infra", "deploy"}),
			0.0,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("recipe-001"))

	id, err := store.CreateRecipe(context.Background(), Recipe{
		TenantID:      "default",
		Category:      "deployment",
		Title:         "Deploy FastAPI",
		IntentPattern: "deploy * to *",
		ExamplePrompt: "deploy my api",
		AgentID:       "admin",
		Tags:          []string{"infra", "deploy"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "recipe-001" {
		t.Errorf("expected recipe-001, got %s", id)
	}
	assertExpectationsMet(t, mock)
}

func TestCreateRecipe_NilDB(t *testing.T) {
	store := NewStore(nil)

	_, err := store.CreateRecipe(context.Background(), Recipe{
		Category:      "test",
		Title:         "Test Recipe",
		IntentPattern: "test *",
	})
	if err == nil {
		t.Fatal("expected error for nil DB")
	}
	if err.Error() != "inception store: database not available" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCreateRecipe_DBError(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery("INSERT INTO inception_recipes").
		WillReturnError(fmt.Errorf("connection refused"))

	_, err := store.CreateRecipe(context.Background(), Recipe{
		TenantID:      "default",
		Category:      "test",
		Title:         "Test Recipe",
		IntentPattern: "test *",
		AgentID:       "admin",
		Tags:          []string{},
	})
	if err == nil {
		t.Fatal("expected error from DB failure")
	}
	if err.Error() != "inception store: insert failed: connection refused" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateRecipe_DefaultTenantID(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery("INSERT INTO inception_recipes").
		WithArgs(
			"default",
			"coding",
			"Scaffold project",
			"scaffold *",
			"{}",
			"",
			"",
			"",
			"",
			"admin",
			pq.Array([]string{}),
			0.0,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("recipe-002"))

	id, err := store.CreateRecipe(context.Background(), Recipe{
		Category:      "coding",
		Title:         "Scaffold project",
		IntentPattern: "scaffold *",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty id")
	}
	assertExpectationsMet(t, mock)
}
