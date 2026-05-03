package server

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

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
