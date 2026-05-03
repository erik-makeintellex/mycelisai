package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/memory"
)

func TestHandleDeploymentContext_PostStoresArtifactAndVectors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mem := memory.NewServiceWithDB(db)
	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{
		Artifacts: artSvc,
		Mem:       mem,
		Cognitive: newDeploymentContextBrain(),
	}

	now := time.Now()
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			"soma", sqlmock.AnyArg(),
			artifacts.ArtifactType("document"), "Deployment Brief", "text/markdown",
			"Service topology and MCP security settings.", sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), "approved",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", now))
	mock.ExpectExec("INSERT INTO context_vectors").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"title": "Deployment Brief",
		"content": "Service topology and MCP security settings.",
		"source_label": "operator brief",
		"visibility": "global",
		"sensitivity_class": "role_scoped",
		"trust_class": "user_provided",
		"tags": ["deployment","security"]
	}`
	rr := doRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodPost, "/api/v1/memory/deployment-context", body)
	assertStatus(t, rr, http.StatusCreated)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["artifact_id"] == "" {
		t.Fatalf("expected artifact_id in response: %+v", resp)
	}
	if resp["knowledge_class"] != "customer_context" {
		t.Fatalf("expected customer_context knowledge class, got %+v", resp)
	}
	if resp["vector_count"] != float64(1) {
		t.Fatalf("expected vector_count=1, got %+v", resp)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestHandleDeploymentContext_PostStoresAdminShapedSomaContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mem := memory.NewServiceWithDB(db)
	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{
		Artifacts: artSvc,
		Mem:       mem,
		Cognitive: newDeploymentContextBrain(),
	}

	now := time.Now()
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			"soma", sqlmock.AnyArg(),
			artifacts.ArtifactType("document"), "Soma Output Contract", "text/markdown",
			"Keep investor-facing responses concise and executive by default.", sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), "approved",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", now))
	mock.ExpectExec("INSERT INTO context_vectors").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"knowledge_class": "soma_operating_context",
		"title": "Soma Output Contract",
		"content": "Keep investor-facing responses concise and executive by default.",
		"soma_context_kind": "output_specificity",
		"output_specificity": "executive",
		"source_label": "root admin guidance"
	}`
	rr := doRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodPost, "/api/v1/memory/deployment-context", body)
	assertStatus(t, rr, http.StatusCreated)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["knowledge_class"] != "soma_operating_context" {
		t.Fatalf("expected soma_operating_context knowledge class, got %+v", resp)
	}
	if resp["vector_count"] != float64(1) {
		t.Fatalf("expected vector_count=1, got %+v", resp)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
