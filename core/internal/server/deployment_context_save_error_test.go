package server

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/memory"
)

func TestHandleDeploymentContext_TransactionFailureIs5xxWithCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO artifacts").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", time.Now()))
	mock.ExpectQuery("INSERT INTO context_vectors").WillReturnError(errors.New("connection reset"))
	mock.ExpectRollback()

	s := &AdminServer{Artifacts: artifacts.NewService(db, t.TempDir()), Mem: memory.NewServiceWithDB(db), Cognitive: newDeploymentContextBrain()}
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodPost, "/api/v1/memory/deployment-context", bakerySaveBody)
	assertStatus(t, rr, http.StatusInternalServerError)
	var body map[string]any
	assertJSON(t, rr, &body)
	if body["code"] != "deployment_context_save_failed" || body["error"] == "" {
		t.Fatalf("save failure must be a normalized 5xx: %+v", body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHandleDeploymentContext_ValidationErrorStays400(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := &AdminServer{Artifacts: artifacts.NewService(db, t.TempDir()), Mem: memory.NewServiceWithDB(db), Cognitive: newDeploymentContextBrain()}
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodPost, "/api/v1/memory/deployment-context", `{"title":"","content":"x"}`)
	assertStatus(t, rr, http.StatusBadRequest)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("validation failure must not touch storage: %v", err)
	}
}
