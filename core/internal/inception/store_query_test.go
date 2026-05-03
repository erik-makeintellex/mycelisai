package inception

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func TestGetRecipe_HappyPath(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows(recipeColumns()).
		AddRow(
			"recipe-001", "default", "deployment", "Deploy FastAPI",
			"deploy * to *", `{"target":"k8s"}`,
			"deploy my api", "success_report",
			"run-100", "sess-100",
			"admin", pq.Array([]string{"infra", "deploy"}),
			0.85, 12,
			now, now,
		)

	mock.ExpectQuery("SELECT .+ FROM inception_recipes WHERE id = \\$1").
		WithArgs("recipe-001").
		WillReturnRows(rows)

	recipe, err := store.GetRecipe(context.Background(), "recipe-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recipe == nil {
		t.Fatal("expected non-nil recipe")
	}
	assertRecipe001(t, recipe)
	assertExpectationsMet(t, mock)
}

func TestGetRecipe_NotFound(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery("SELECT .+ FROM inception_recipes WHERE id = \\$1").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	recipe, err := store.GetRecipe(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
	if recipe != nil {
		t.Errorf("expected nil recipe, got %+v", recipe)
	}
}

func TestGetRecipe_NilDB(t *testing.T) {
	store := NewStore(nil)

	_, err := store.GetRecipe(context.Background(), "recipe-001")
	if err == nil {
		t.Fatal("expected error for nil DB")
	}
	if err.Error() != "inception store: database not available" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestListRecipes_HappyPath(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows(recipeColumns()).
		AddRow(
			"recipe-001", "default", "deployment", "Deploy FastAPI",
			"deploy * to *", "{}",
			"deploy my api", "",
			"", "",
			"admin", pq.Array([]string{"infra"}),
			0.9, 5,
			now, now,
		).
		AddRow(
			"recipe-002", "default", "coding", "Scaffold Go service",
			"scaffold *", `{"lang":"go"}`,
			"scaffold a go service", "",
			"", "",
			"coder", pq.Array([]string{"go", "scaffold"}),
			0.7, 3,
			now, now,
		)

	mock.ExpectQuery("SELECT .+ FROM inception_recipes\\s+ORDER BY").
		WithArgs(20).
		WillReturnRows(rows)

	recipes, err := store.ListRecipes(context.Background(), "", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recipes) != 2 {
		t.Fatalf("expected 2 recipes, got %d", len(recipes))
	}
	if recipes[0].ID != "recipe-001" {
		t.Errorf("expected first recipe id=recipe-001, got %s", recipes[0].ID)
	}
	if recipes[1].Category != "coding" {
		t.Errorf("expected second recipe category=coding, got %s", recipes[1].Category)
	}
	assertExpectationsMet(t, mock)
}

func TestListRecipes_WithCategoryFilter(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows(recipeColumns()).
		AddRow(
			"recipe-001", "default", "deployment", "Deploy FastAPI",
			"deploy * to *", "{}",
			"deploy my api", "",
			"", "",
			"admin", pq.Array([]string{"infra"}),
			0.9, 5,
			now, now,
		)

	mock.ExpectQuery("SELECT .+ FROM inception_recipes WHERE category = \\$1").
		WithArgs("deployment", 10).
		WillReturnRows(rows)

	recipes, err := store.ListRecipes(context.Background(), "deployment", "", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recipes) != 1 {
		t.Fatalf("expected 1 recipe, got %d", len(recipes))
	}
	if recipes[0].Category != "deployment" {
		t.Errorf("expected category=deployment, got %s", recipes[0].Category)
	}
	assertExpectationsMet(t, mock)
}

func TestListRecipes_NilDB(t *testing.T) {
	store := NewStore(nil)

	_, err := store.ListRecipes(context.Background(), "", "", 10)
	if err == nil {
		t.Fatal("expected error for nil DB")
	}
	if err.Error() != "inception store: database not available" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSearchByTitle_HappyPath(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows(recipeColumns()).
		AddRow(
			"recipe-001", "default", "deployment", "Deploy FastAPI",
			"deploy * to *", "{}",
			"deploy my api", "",
			"", "",
			"admin", pq.Array([]string{"infra"}),
			0.9, 5,
			now, now,
		)

	mock.ExpectQuery("SELECT .+ FROM inception_recipes").
		WithArgs("Deploy", 10).
		WillReturnRows(rows)

	recipes, err := store.SearchByTitle(context.Background(), "Deploy", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recipes) != 1 {
		t.Fatalf("expected 1 recipe, got %d", len(recipes))
	}
	if recipes[0].Title != "Deploy FastAPI" {
		t.Errorf("expected title=Deploy FastAPI, got %s", recipes[0].Title)
	}
	assertExpectationsMet(t, mock)
}

func TestSearchByTitle_EmptyResult(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	rows := sqlmock.NewRows(recipeColumns())
	mock.ExpectQuery("SELECT .+ FROM inception_recipes").
		WithArgs("nonexistent-pattern", 10).
		WillReturnRows(rows)

	recipes, err := store.SearchByTitle(context.Background(), "nonexistent-pattern", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recipes == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(recipes) != 0 {
		t.Errorf("expected 0 recipes, got %d", len(recipes))
	}
	assertExpectationsMet(t, mock)
}
