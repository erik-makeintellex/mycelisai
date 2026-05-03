package inception

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockStore(t *testing.T) (*Store, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}

	return NewStore(db), mock, func() {
		_ = db.Close()
	}
}

func recipeColumns() []string {
	return []string{
		"id", "tenant_id", "category", "title", "intent_pattern",
		"parameters", "example_prompt", "outcome_shape",
		"source_run_id", "source_session_id",
		"agent_id", "tags", "quality_score", "usage_count",
		"created_at", "updated_at",
	}
}

func assertExpectationsMet(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func assertRecipe001(t *testing.T, recipe *Recipe) {
	t.Helper()

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
}
