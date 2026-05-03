package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/artifacts"
)

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
