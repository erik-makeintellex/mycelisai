package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/artifacts"
)

var artTestColumns = []string{
	"id", "mission_id", "team_id", "agent_id", "trace_id", "artifact_type",
	"title", "content_type", "content", "file_path", "file_size_bytes",
	"metadata", "trust_score", "status", "created_at",
}

func TestHandleListArtifacts_Recent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	now := time.Now()
	rows := sqlmock.NewRows(artTestColumns).
		AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", nil, nil, "agent-1", nil, "code",
			"main.go", "text/x-go", "package main", nil, nil,
			[]byte(`{}`), 0.9, "approved", now)

	mock.ExpectQuery("SELECT .+ FROM artifacts ORDER BY created_at DESC").
		WithArgs(50).
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/api/v1/artifacts", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleListArtifacts)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var result []artifacts.Artifact
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 artifact, got %d", len(result))
	}
	if result[0].Title != "main.go" {
		t.Errorf("Expected title 'main.go', got %q", result[0].Title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestHandleListArtifacts_ByMission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	missionID := "11111111-1111-1111-1111-111111111111"
	rows := sqlmock.NewRows(artTestColumns)
	mock.ExpectQuery("SELECT .+ FROM artifacts WHERE mission_id").
		WithArgs(sqlmock.AnyArg(), 50).
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/api/v1/artifacts?mission_id="+missionID, nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleListArtifacts)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestHandleListArtifacts_ByAgent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	rows := sqlmock.NewRows(artTestColumns)
	mock.ExpectQuery("SELECT .+ FROM artifacts WHERE agent_id").
		WithArgs("scanner-1", 10).
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/api/v1/artifacts?agent_id=scanner-1&limit=10", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleListArtifacts)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestHandleListArtifacts_InvalidMissionID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	req, _ := http.NewRequest("GET", "/api/v1/artifacts?mission_id=not-a-uuid", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleListArtifacts)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleListArtifacts_NotInitialized(t *testing.T) {
	s := &AdminServer{Artifacts: nil}

	req, _ := http.NewRequest("GET", "/api/v1/artifacts", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleListArtifacts)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", rr.Code)
	}
}

func TestHandleStoreArtifact(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	now := time.Now()
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			"agent-writer-1", sqlmock.AnyArg(),
			artifacts.ArtifactType("document"), "Summary Report", "text/markdown",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), "pending",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", now))

	body := `{
		"agent_id": "agent-writer-1",
		"artifact_type": "document",
		"title": "Summary Report",
		"content_type": "text/markdown",
		"content": "# Summary\n\nAll tasks completed."
	}`
	req, _ := http.NewRequest("POST", "/api/v1/artifacts", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleStoreArtifact)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var result artifacts.Artifact
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if result.Title != "Summary Report" {
		t.Errorf("Expected title 'Summary Report', got %q", result.Title)
	}
	if result.Status != "pending" {
		t.Errorf("Expected status 'pending', got %q", result.Status)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestHandleStoreArtifact_ValidationErrors(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	tests := []struct {
		name string
		body string
		want int
	}{
		{"missing agent_id", `{"artifact_type":"code","title":"test"}`, http.StatusBadRequest},
		{"missing artifact_type", `{"agent_id":"a1","title":"test"}`, http.StatusBadRequest},
		{"missing title", `{"agent_id":"a1","artifact_type":"code"}`, http.StatusBadRequest},
		{"invalid JSON", `{broken`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/v1/artifacts", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(s.handleStoreArtifact)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.want {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.want, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleUpdateArtifactStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	artID := "cccccccc-cccc-cccc-cccc-cccccccccccc"

	mock.ExpectExec("UPDATE artifacts SET status").
		WithArgs("approved", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Use ServeMux to get path value routing
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v1/artifacts/{id}/status", s.handleUpdateArtifactStatus)

	body := `{"status":"approved"}`
	req, _ := http.NewRequest("PUT", "/api/v1/artifacts/"+artID+"/status", strings.NewReader(body))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if result["status"] != "approved" {
		t.Errorf("Expected status 'approved', got %q", result["status"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func TestHandleUpdateArtifactStatus_MissingStatus(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v1/artifacts/{id}/status", s.handleUpdateArtifactStatus)

	body := `{}`
	req, _ := http.NewRequest("PUT", "/api/v1/artifacts/cccccccc-cccc-cccc-cccc-cccccccccccc/status", strings.NewReader(body))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleUpdateArtifactStatus_InvalidID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v1/artifacts/{id}/status", s.handleUpdateArtifactStatus)

	body := `{"status":"approved"}`
	req, _ := http.NewRequest("PUT", "/api/v1/artifacts/not-a-uuid/status", strings.NewReader(body))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleGetArtifact_NotInitialized(t *testing.T) {
	s := &AdminServer{Artifacts: nil}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/artifacts/{id}", s.handleGetArtifact)

	req, _ := http.NewRequest("GET", "/api/v1/artifacts/cccccccc-cccc-cccc-cccc-cccccccccccc", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", rr.Code)
	}
}
