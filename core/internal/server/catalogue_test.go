package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/catalogue"
)

var catTestColumns = []string{
	"id", "name", "role", "system_prompt", "model",
	"tools", "inputs", "outputs",
	"verification_strategy", "verification_rubric", "validation_command",
	"created_at", "updated_at",
}

func TestHandleListCatalogue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	catSvc := catalogue.NewService(db)
	s := &AdminServer{Catalogue: catSvc}

	now := time.Now()
	rows := sqlmock.NewRows(catTestColumns).
		AddRow("11111111-1111-1111-1111-111111111111", "Test Agent", "cognitive",
			"You are a test agent", "qwen2.5",
			[]byte(`["read_file"]`), []byte(`[]`), []byte(`[]`),
			"semantic", []byte(`["Check output"]`), nil,
			now, now)

	mock.ExpectQuery("SELECT .+ FROM agent_catalogue").WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/api/v1/catalogue/agents", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleListCatalogue)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var agents []catalogue.AgentTemplate
	if err := json.NewDecoder(rr.Body).Decode(&agents); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(agents) != 1 {
		t.Fatalf("Expected 1 agent, got %d", len(agents))
	}
	if agents[0].Name != "Test Agent" {
		t.Errorf("Expected 'Test Agent', got %q", agents[0].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestHandleListCatalogue_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	catSvc := catalogue.NewService(db)
	s := &AdminServer{Catalogue: catSvc}

	mock.ExpectQuery("SELECT .+ FROM agent_catalogue").
		WillReturnRows(sqlmock.NewRows(catTestColumns))

	req, _ := http.NewRequest("GET", "/api/v1/catalogue/agents", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleListCatalogue)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var agents []catalogue.AgentTemplate
	if err := json.NewDecoder(rr.Body).Decode(&agents); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(agents) != 0 {
		t.Errorf("Expected empty array, got %d agents", len(agents))
	}
}

func TestHandleListCatalogue_NotInitialized(t *testing.T) {
	s := &AdminServer{Catalogue: nil}

	req, _ := http.NewRequest("GET", "/api/v1/catalogue/agents", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleListCatalogue)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", rr.Code)
	}
}

func TestHandleCreateCatalogue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	catSvc := catalogue.NewService(db)
	s := &AdminServer{Catalogue: catSvc}

	now := time.Now()
	mock.ExpectQuery("INSERT INTO agent_catalogue").
		WithArgs("New Agent", "sensory",
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("22222222-2222-2222-2222-222222222222", now, now))

	body := `{"name":"New Agent","role":"sensory","model":"qwen2.5"}`
	req, _ := http.NewRequest("POST", "/api/v1/catalogue/agents", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleCreateCatalogue)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var agent catalogue.AgentTemplate
	if err := json.NewDecoder(rr.Body).Decode(&agent); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if agent.Name != "New Agent" {
		t.Errorf("Expected 'New Agent', got %q", agent.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestHandleCreateCatalogue_MissingName(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	catSvc := catalogue.NewService(db)
	s := &AdminServer{Catalogue: catSvc}

	body := `{"role":"cognitive"}`
	req, _ := http.NewRequest("POST", "/api/v1/catalogue/agents", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleCreateCatalogue)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleCreateCatalogue_MissingRole(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	catSvc := catalogue.NewService(db)
	s := &AdminServer{Catalogue: catSvc}

	body := `{"name":"Agent Without Role"}`
	req, _ := http.NewRequest("POST", "/api/v1/catalogue/agents", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleCreateCatalogue)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleCreateCatalogue_InvalidJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	catSvc := catalogue.NewService(db)
	s := &AdminServer{Catalogue: catSvc}

	req, _ := http.NewRequest("POST", "/api/v1/catalogue/agents", strings.NewReader("not json"))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleCreateCatalogue)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleDeleteCatalogue_InvalidID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	catSvc := catalogue.NewService(db)
	s := &AdminServer{Catalogue: catSvc}

	// Use Go 1.22+ ServeMux with path value
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v1/catalogue/agents/{id}", s.handleDeleteCatalogue)

	req, _ := http.NewRequest("DELETE", "/api/v1/catalogue/agents/not-a-uuid", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}
