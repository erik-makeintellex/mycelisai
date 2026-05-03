package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/artifacts"
)

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
