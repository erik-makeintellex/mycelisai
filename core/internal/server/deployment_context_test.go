package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/memory"
)

type fakeDeploymentContextProvider struct{}

func (fakeDeploymentContextProvider) Infer(_ context.Context, _ string, _ cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "ok", ModelUsed: "stub", Provider: "stub"}, nil
}

func (fakeDeploymentContextProvider) Probe(_ context.Context) (bool, error) {
	return true, nil
}

func (fakeDeploymentContextProvider) Embed(_ context.Context, _ string, _ string) ([]float64, error) {
	return []float64{0.11, 0.22}, nil
}

func newDeploymentContextBrain() *cognitive.Router {
	return &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Providers: map[string]cognitive.ProviderConfig{
				"stub": {Enabled: true, ModelID: "stub-model"},
			},
			Profiles: map[string]string{
				"chat":  "stub",
				"embed": "stub",
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"stub": fakeDeploymentContextProvider{},
		},
	}
}

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

func TestHandleDeploymentContext_GetListsEntries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{Artifacts: artSvc}

	meta, _ := json.Marshal(map[string]any{
		"knowledge_store":   "governed_context_store",
		"knowledge_class":   "company_knowledge",
		"source_label":      "operator brief",
		"source_kind":       "user_document",
		"visibility":        "global",
		"sensitivity_class": "role_scoped",
		"trust_class":       "user_provided",
		"chunk_count":       2,
		"vector_count":      2,
		"content_length":    42,
	})
	rows := sqlmock.NewRows([]string{"id", "title", "content", "metadata", "created_at"}).
		AddRow("ctx-1", "Deployment Brief", "Service topology and MCP security settings.", meta, time.Now())
	mock.ExpectQuery("SELECT id::text,\\s+title,\\s+content,\\s+metadata,\\s+created_at\\s+FROM artifacts").
		WithArgs(12).
		WillReturnRows(rows)

	rr := doRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodGet, "/api/v1/memory/deployment-context?limit=12", "")
	assertStatus(t, rr, http.StatusOK)

	var resp struct {
		Entries []map[string]any `json:"entries"`
		Count   int              `json:"count"`
	}
	assertJSON(t, rr, &resp)
	if resp.Count != 1 || len(resp.Entries) != 1 {
		t.Fatalf("expected one entry, got %+v", resp)
	}
	if !strings.Contains(resp.Entries[0]["content_preview"].(string), "MCP security") {
		t.Fatalf("expected content preview in response, got %+v", resp.Entries[0])
	}
	if resp.Entries[0]["knowledge_class"] != "company_knowledge" {
		t.Fatalf("expected company_knowledge entry, got %+v", resp.Entries[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestHandleDeploymentContext_ServiceUnavailableWithoutDependencies(t *testing.T) {
	s := newTestServer()

	rr := doRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodPost, "/api/v1/memory/deployment-context", `{"title":"x","content":"y"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}
