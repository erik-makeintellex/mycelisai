package server

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/internal/memory"
)

// Inception recipe handler unit tests. Owner-scoped reads and writes are
// proven on real PostgreSQL in memlanes3_inception_realdb_test.go
// (MEM-LANES-3); these cover the contract bundle and request validation.

// withInception wires sqlmock-backed inception and memory stores. The
// validation tests below fail before any query, so no expectation is set.
func withInception(t *testing.T) func(*AdminServer) {
	t.Helper()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.Inception = inception.NewStore(db)
		s.Mem = memory.NewServiceWithDB(db)
	}
}

func TestHandleInceptionContracts_HappyPath(t *testing.T) {
	s := newTestServer()

	mux := setupMux(t, "GET /api/v1/inception/contracts", s.HandleInceptionContracts)
	rr := doRequest(t, mux, "GET", "/api/v1/inception/contracts", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	paths, ok := data["allowed_paths"].([]any)
	if !ok {
		t.Fatalf("expected allowed_paths array, got %T", data["allowed_paths"])
	}
	if len(paths) != 4 {
		t.Fatalf("expected 4 allowed paths, got %d", len(paths))
	}
	lifetimes, ok := data["allowed_lifetimes"].([]any)
	if !ok {
		t.Fatalf("expected allowed_lifetimes array, got %T", data["allowed_lifetimes"])
	}
	if len(lifetimes) != 3 {
		t.Fatalf("expected 3 allowed lifetimes, got %d", len(lifetimes))
	}
	if _, ok := data["decision_frame"].(map[string]any); !ok {
		t.Fatalf("expected decision_frame object, got %T", data["decision_frame"])
	}
	if _, ok := data["heartbeat_budget"].(map[string]any); !ok {
		t.Fatalf("expected heartbeat_budget object, got %T", data["heartbeat_budget"])
	}
	if _, ok := data["universal_invoke"].(map[string]any); !ok {
		t.Fatalf("expected universal_invoke object, got %T", data["universal_invoke"])
	}
}

func TestHandleListInceptionRecipes_NilStore(t *testing.T) {
	s := newTestServer() // no Inception wired
	mux := setupMux(t, "GET /api/v1/inception/recipes", s.HandleListInceptionRecipes)
	rr := doAuthenticatedRequestAs(t, mux, "GET", "/api/v1/inception/recipes", "", memoryUser("u-inc", "inc", "operator"))
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleSearchInceptionRecipes_MissingQuery(t *testing.T) {
	s := newTestServer(withInception(t))
	mux := setupMux(t, "GET /api/v1/inception/recipes/search", s.HandleSearchInceptionRecipes)
	rr := doAuthenticatedRequestAs(t, mux, "GET", "/api/v1/inception/recipes/search", "", memoryUser("u-inc", "inc", "operator"))
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCreateInceptionRecipe_Validation(t *testing.T) {
	s := newTestServer(withInception(t))
	mux := setupMux(t, "POST /api/v1/inception/recipes", s.HandleCreateInceptionRecipe)
	who := memoryUser("u-inc", "inc", "operator")
	for name, body := range map[string]string{
		"missing fields":     `{"category": "research"}`,
		"unknown visibility": `{"category":"r","title":"t","intent_pattern":"p","visibility":"public"}`,
		"invalid json":       `{`,
	} {
		if rr := doAuthenticatedRequestAs(t, mux, "POST", "/api/v1/inception/recipes", body, who); rr.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", name, rr.Code)
		}
	}
}

func TestHandleUpdateRecipeQuality_InvalidScoreRange(t *testing.T) {
	s := newTestServer(withInception(t))
	mux := setupMux(t, "PATCH /api/v1/inception/recipes/{id}/quality", s.HandleUpdateRecipeQuality)
	for _, body := range []string{`{"score": 1.5}`, `{"score": -0.1}`} {
		rr := doAuthenticatedRequestAs(t, mux, "PATCH", "/api/v1/inception/recipes/r-10/quality", body, memoryUser("u-inc", "inc", "operator"))
		assertStatus(t, rr, http.StatusBadRequest)
	}
}
