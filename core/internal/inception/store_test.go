package inception

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

// recipeColumns returns the column list used by all SELECT queries in the store.
func recipeColumns() []string {
	return []string{
		"id", "tenant_id", "category", "title", "intent_pattern",
		"parameters", "example_prompt", "outcome_shape",
		"source_run_id", "source_session_id",
		"agent_id", "tags", "quality_score", "usage_count",
		"created_at", "updated_at",
	}
}

// ════════════════════════════════════════════════════════════════════
// CreateRecipe
// ════════════════════════════════════════════════════════════════════

func TestCreateRecipe_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery("INSERT INTO inception_recipes").
		WithArgs(
			"default",        // tenant_id
			"deployment",     // category
			"Deploy FastAPI", // title
			"deploy * to *",  // intent_pattern
			"{}",             // parameters JSON
			"deploy my api",  // example_prompt (NULLIF)
			"",               // outcome_shape (NULLIF)
			"",               // source_run_id (NULLIF)
			"",               // source_session_id (NULLIF)
			"admin",          // agent_id
			pq.Array([]string{"infra", "deploy"}), // tags
			0.0, // quality_score
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
	if id == "" {
		t.Error("expected non-empty id")
	}
	if id != "recipe-001" {
		t.Errorf("expected recipe-001, got %s", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery("INSERT INTO inception_recipes").
		WillReturnError(fmt.Errorf("connection refused"))

	_, err = store.CreateRecipe(context.Background(), Recipe{
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	// Expect tenant_id = "default" even though we pass empty string
	mock.ExpectQuery("INSERT INTO inception_recipes").
		WithArgs(
			"default",          // tenant_id — defaulted from ""
			"coding",           // category
			"Scaffold project", // title
			"scaffold *",       // intent_pattern
			"{}",               // parameters JSON
			"",                 // example_prompt
			"",                 // outcome_shape
			"",                 // source_run_id
			"",                 // source_session_id
			"admin",            // agent_id — defaulted from ""
			pq.Array([]string{}), // tags — defaulted from nil
			0.0,                // quality_score
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("recipe-002"))

	id, err := store.CreateRecipe(context.Background(), Recipe{
		TenantID:      "", // empty — should default to "default"
		Category:      "coding",
		Title:         "Scaffold project",
		IntentPattern: "scaffold *",
		// AgentID empty — should default to "admin"
		// Tags nil — should default to []string{}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// ════════════════════════════════════════════════════════════════════
// GetRecipe
// ════════════════════════════════════════════════════════════════════

func TestGetRecipe_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)
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
	if recipe.ID != "recipe-001" {
		t.Errorf("expected id=recipe-001, got %s", recipe.ID)
	}
	if recipe.TenantID != "default" {
		t.Errorf("expected tenant_id=default, got %s", recipe.TenantID)
	}
	if recipe.Category != "deployment" {
		t.Errorf("expected category=deployment, got %s", recipe.Category)
	}
	if recipe.Title != "Deploy FastAPI" {
		t.Errorf("expected title=Deploy FastAPI, got %s", recipe.Title)
	}
	if recipe.IntentPattern != "deploy * to *" {
		t.Errorf("expected intent_pattern=deploy * to *, got %s", recipe.IntentPattern)
	}
	if recipe.Parameters["target"] != "k8s" {
		t.Errorf("expected parameters.target=k8s, got %v", recipe.Parameters["target"])
	}
	if recipe.ExamplePrompt != "deploy my api" {
		t.Errorf("expected example_prompt=deploy my api, got %s", recipe.ExamplePrompt)
	}
	if recipe.OutcomeShape != "success_report" {
		t.Errorf("expected outcome_shape=success_report, got %s", recipe.OutcomeShape)
	}
	if recipe.SourceRunID != "run-100" {
		t.Errorf("expected source_run_id=run-100, got %s", recipe.SourceRunID)
	}
	if recipe.SourceSessionID != "sess-100" {
		t.Errorf("expected source_session_id=sess-100, got %s", recipe.SourceSessionID)
	}
	if recipe.AgentID != "admin" {
		t.Errorf("expected agent_id=admin, got %s", recipe.AgentID)
	}
	if len(recipe.Tags) != 2 || recipe.Tags[0] != "infra" || recipe.Tags[1] != "deploy" {
		t.Errorf("expected tags=[infra, deploy], got %v", recipe.Tags)
	}
	if recipe.QualityScore != 0.85 {
		t.Errorf("expected quality_score=0.85, got %f", recipe.QualityScore)
	}
	if recipe.UsageCount != 12 {
		t.Errorf("expected usage_count=12, got %d", recipe.UsageCount)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetRecipe_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

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

// ════════════════════════════════════════════════════════════════════
// ListRecipes
// ════════════════════════════════════════════════════════════════════

func TestListRecipes_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)
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

	// No category/agent filters → no WHERE clause, just LIMIT $1
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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestListRecipes_WithCategoryFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)
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

	// category filter → WHERE category = $1, LIMIT $2
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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
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

// ════════════════════════════════════════════════════════════════════
// SearchByTitle
// ════════════════════════════════════════════════════════════════════

func TestSearchByTitle_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)
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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestSearchByTitle_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	rows := sqlmock.NewRows(recipeColumns()) // no rows
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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// ════════════════════════════════════════════════════════════════════
// IncrementUsage
// ════════════════════════════════════════════════════════════════════

func TestIncrementUsage_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectExec("UPDATE inception_recipes SET usage_count = usage_count \\+ 1").
		WithArgs("recipe-001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.IncrementUsage(context.Background(), "recipe-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestIncrementUsage_NilDB(t *testing.T) {
	store := NewStore(nil)

	// IncrementUsage with nil DB returns nil (graceful degradation)
	err := store.IncrementUsage(context.Background(), "recipe-001")
	if err != nil {
		t.Errorf("expected nil error for nil DB, got: %v", err)
	}
}

// ════════════════════════════════════════════════════════════════════
// UpdateQuality
// ════════════════════════════════════════════════════════════════════

func TestUpdateQuality_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectExec("UPDATE inception_recipes SET quality_score = \\$1").
		WithArgs(0.95, "recipe-001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.UpdateQuality(context.Background(), "recipe-001", 0.95)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
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
