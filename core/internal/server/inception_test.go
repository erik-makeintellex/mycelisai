package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/mycelis/core/internal/inception"
)

// ── Shared helpers for inception handler tests ───────────────────

// withInception creates an inception.Store backed by sqlmock.
func withInception(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	store := inception.NewStore(db)
	return func(s *AdminServer) {
		s.Inception = store
	}, mock
}

// recipeColumns returns the column names returned by inception_recipes SELECT queries.
func recipeColumns() []string {
	return []string{
		"id", "tenant_id", "category", "title", "intent_pattern", "parameters",
		"example_prompt", "outcome_shape",
		"source_run_id", "source_session_id",
		"agent_id", "tags", "quality_score", "usage_count",
		"created_at", "updated_at",
	}
}

// ════════════════════════════════════════════════════════════════════
// HandleListInceptionRecipes — GET /api/v1/inception/recipes
// ════════════════════════════════════════════════════════════════════

func TestHandleListInceptionRecipes_HappyPath(t *testing.T) {
	incOpt, mock := withInception(t)
	s := newTestServer(incOpt)
	now := time.Now()

	rows := sqlmock.NewRows(recipeColumns()).
		AddRow("r-1", "default", "research", "Web Search", "search for {topic}",
			`{"topic":"string"}`, "search for AI news", "summary",
			"", "", "admin", pq.Array([]string{"search", "web"}),
			0.85, 12, now, now).
		AddRow("r-2", "default", "coding", "Generate Tests", "write tests for {module}",
			`{}`, "write tests for auth", "code",
			"run-5", "", "coder", pq.Array([]string{"testing"}),
			0.72, 5, now, now)

	mock.ExpectQuery("SELECT .+ FROM inception_recipes").
		WillReturnRows(rows)

	mux := setupMux(t, "GET /api/v1/inception/recipes", s.HandleListInceptionRecipes)
	rr := doRequest(t, mux, "GET", "/api/v1/inception/recipes", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 2 {
		t.Errorf("expected 2 recipes, got %d", len(data))
	}

	first, ok := data[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected recipe object, got %T", data[0])
	}
	if first["id"] != "r-1" {
		t.Errorf("expected id=r-1, got %v", first["id"])
	}
	if first["category"] != "research" {
		t.Errorf("expected category=research, got %v", first["category"])
	}
	if first["title"] != "Web Search" {
		t.Errorf("expected title=Web Search, got %v", first["title"])
	}
}

func TestHandleListInceptionRecipes_NilStore(t *testing.T) {
	s := newTestServer() // no Inception wired

	mux := setupMux(t, "GET /api/v1/inception/recipes", s.HandleListInceptionRecipes)
	rr := doRequest(t, mux, "GET", "/api/v1/inception/recipes", "")

	assertStatus(t, rr, http.StatusServiceUnavailable)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != false {
		t.Errorf("expected ok=false, got %v", resp["ok"])
	}
	if resp["error"] == nil || resp["error"] == "" {
		t.Errorf("expected error message, got %v", resp["error"])
	}
}

// ════════════════════════════════════════════════════════════════════
// HandleSearchInceptionRecipes — GET /api/v1/inception/recipes/search
// ════════════════════════════════════════════════════════════════════

func TestHandleSearchInceptionRecipes_HappyPath(t *testing.T) {
	incOpt, mock := withInception(t)
	s := newTestServer(incOpt)
	now := time.Now()

	rows := sqlmock.NewRows(recipeColumns()).
		AddRow("r-3", "default", "analysis", "Data Analysis", "analyze {dataset}",
			`{"dataset":"string"}`, "analyze user metrics", "report",
			"", "", "admin", pq.Array([]string{"data", "analysis"}),
			0.90, 8, now, now)

	mock.ExpectQuery("SELECT .+ FROM inception_recipes").
		WithArgs("Data", 10).
		WillReturnRows(rows)

	mux := setupMux(t, "GET /api/v1/inception/recipes/search", s.HandleSearchInceptionRecipes)
	rr := doRequest(t, mux, "GET", "/api/v1/inception/recipes/search?q=Data", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(data) != 1 {
		t.Errorf("expected 1 recipe, got %d", len(data))
	}

	recipe, ok := data[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected recipe object, got %T", data[0])
	}
	if recipe["title"] != "Data Analysis" {
		t.Errorf("expected title=Data Analysis, got %v", recipe["title"])
	}
}

func TestHandleSearchInceptionRecipes_MissingQuery(t *testing.T) {
	incOpt, _ := withInception(t)
	s := newTestServer(incOpt)

	mux := setupMux(t, "GET /api/v1/inception/recipes/search", s.HandleSearchInceptionRecipes)
	rr := doRequest(t, mux, "GET", "/api/v1/inception/recipes/search", "")

	assertStatus(t, rr, http.StatusBadRequest)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != false {
		t.Errorf("expected ok=false, got %v", resp["ok"])
	}
}

// ════════════════════════════════════════════════════════════════════
// HandleGetInceptionRecipe — GET /api/v1/inception/recipes/{id}
// ════════════════════════════════════════════════════════════════════

func TestHandleGetInceptionRecipe_HappyPath(t *testing.T) {
	incOpt, mock := withInception(t)
	s := newTestServer(incOpt)
	now := time.Now()

	rows := sqlmock.NewRows(recipeColumns()).
		AddRow("r-10", "default", "deployment", "Deploy Service", "deploy {service} to {env}",
			`{"service":"string","env":"string"}`, "deploy auth to staging", "status",
			"run-42", "sess-7", "admin", pq.Array([]string{"deploy", "ops"}),
			0.95, 20, now, now)

	mock.ExpectQuery("SELECT .+ FROM inception_recipes WHERE id = \\$1").
		WithArgs("r-10").
		WillReturnRows(rows)

	mux := setupMux(t, "GET /api/v1/inception/recipes/{id}", s.HandleGetInceptionRecipe)
	rr := doRequest(t, mux, "GET", "/api/v1/inception/recipes/r-10", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["id"] != "r-10" {
		t.Errorf("expected id=r-10, got %v", data["id"])
	}
	if data["category"] != "deployment" {
		t.Errorf("expected category=deployment, got %v", data["category"])
	}
	if data["title"] != "Deploy Service" {
		t.Errorf("expected title=Deploy Service, got %v", data["title"])
	}
	if data["agent_id"] != "admin" {
		t.Errorf("expected agent_id=admin, got %v", data["agent_id"])
	}
}

func TestHandleGetInceptionRecipe_NotFound(t *testing.T) {
	incOpt, mock := withInception(t)
	s := newTestServer(incOpt)

	// Return empty rows so QueryRowContext.Scan returns sql.ErrNoRows
	rows := sqlmock.NewRows(recipeColumns())

	mock.ExpectQuery("SELECT .+ FROM inception_recipes WHERE id = \\$1").
		WithArgs("nonexistent").
		WillReturnRows(rows)

	mux := setupMux(t, "GET /api/v1/inception/recipes/{id}", s.HandleGetInceptionRecipe)
	rr := doRequest(t, mux, "GET", "/api/v1/inception/recipes/nonexistent", "")

	assertStatus(t, rr, http.StatusNotFound)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != false {
		t.Errorf("expected ok=false, got %v", resp["ok"])
	}
}

// ════════════════════════════════════════════════════════════════════
// HandleCreateInceptionRecipe — POST /api/v1/inception/recipes
// ════════════════════════════════════════════════════════════════════

func TestHandleCreateInceptionRecipe_HappyPath(t *testing.T) {
	incOpt, mock := withInception(t)
	s := newTestServer(incOpt)

	mock.ExpectQuery("INSERT INTO inception_recipes").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("new-recipe-id"))

	body := `{
		"category": "research",
		"title": "Literature Review",
		"intent_pattern": "review literature on {topic}",
		"parameters": {"topic": "string"},
		"example_prompt": "review literature on quantum computing",
		"outcome_shape": "summary",
		"agent_id": "architect",
		"tags": ["research", "literature"]
	}`

	mux := setupMux(t, "POST /api/v1/inception/recipes", s.HandleCreateInceptionRecipe)
	rr := doRequest(t, mux, "POST", "/api/v1/inception/recipes", body)

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["id"] != "new-recipe-id" {
		t.Errorf("expected id=new-recipe-id, got %v", data["id"])
	}
}

func TestHandleCreateInceptionRecipe_MissingFields(t *testing.T) {
	incOpt, _ := withInception(t)
	s := newTestServer(incOpt)

	// Missing title and intent_pattern
	body := `{"category": "research"}`

	mux := setupMux(t, "POST /api/v1/inception/recipes", s.HandleCreateInceptionRecipe)
	rr := doRequest(t, mux, "POST", "/api/v1/inception/recipes", body)

	assertStatus(t, rr, http.StatusBadRequest)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != false {
		t.Errorf("expected ok=false, got %v", resp["ok"])
	}
	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		t.Errorf("expected non-empty error message")
	}
}

// ════════════════════════════════════════════════════════════════════
// HandleUpdateRecipeQuality — PATCH /api/v1/inception/recipes/{id}/quality
// ════════════════════════════════════════════════════════════════════

func TestHandleUpdateRecipeQuality_HappyPath(t *testing.T) {
	incOpt, mock := withInception(t)
	s := newTestServer(incOpt)

	mock.ExpectExec("UPDATE inception_recipes SET quality_score").
		WithArgs(0.92, "r-10").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"score": 0.92}`

	mux := setupMux(t, "PATCH /api/v1/inception/recipes/{id}/quality", s.HandleUpdateRecipeQuality)
	rr := doRequest(t, mux, "PATCH", "/api/v1/inception/recipes/r-10/quality", body)

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["status"] != "updated" {
		t.Errorf("expected status=updated, got %v", data["status"])
	}
}

func TestHandleUpdateRecipeQuality_InvalidScoreRange(t *testing.T) {
	incOpt, _ := withInception(t)
	s := newTestServer(incOpt)

	tests := []struct {
		name string
		body string
	}{
		{"score too high", `{"score": 1.5}`},
		{"score negative", `{"score": -0.1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupMux(t, "PATCH /api/v1/inception/recipes/{id}/quality", s.HandleUpdateRecipeQuality)
			rr := doRequest(t, mux, "PATCH", "/api/v1/inception/recipes/r-10/quality", tt.body)

			assertStatus(t, rr, http.StatusBadRequest)

			var resp map[string]interface{}
			assertJSON(t, rr, &resp)
			if resp["ok"] != false {
				t.Errorf("expected ok=false, got %v", resp["ok"])
			}
		})
	}
}
