package server

import (
	"context"
	"database/sql/driver"
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

type metadataContains map[string]any

func (m metadataContains) Match(v driver.Value) bool {
	var raw []byte
	switch value := v.(type) {
	case []byte:
		raw = value
	case string:
		raw = []byte(value)
	default:
		return false
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		return false
	}
	for key, want := range m {
		if !metadataValueContains(got[key], want) {
			return false
		}
	}
	return true
}

func metadataValueContains(got any, want any) bool {
	switch want := want.(type) {
	case string:
		return strings.TrimSpace(strings.ToLower(stringValue(got))) == strings.TrimSpace(strings.ToLower(want))
	case []string:
		values := normalizeStringSliceValue(got)
		if len(values) == 0 {
			return false
		}
		have := map[string]struct{}{}
		for _, value := range values {
			have[strings.ToLower(strings.TrimSpace(value))] = struct{}{}
		}
		for _, item := range want {
			if _, ok := have[strings.ToLower(strings.TrimSpace(item))]; !ok {
				return false
			}
		}
		return true
	case []any:
		expected := make([]string, 0, len(want))
		for _, item := range want {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				expected = append(expected, text)
			}
		}
		return metadataValueContains(got, expected)
	default:
		return false
	}
}

func normalizeStringSliceValue(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func stringValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

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

func TestHandleDeploymentContext_PostNormalizesAdminSomaDefaults(t *testing.T) {
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
			artifacts.ArtifactType("document"), "Soma Identity Guidance", "text/markdown",
			"Keep shared Soma behavior steady and concise.", sqlmock.AnyArg(), sqlmock.AnyArg(),
			metadataContains{
				"knowledge_class":    "soma_operating_context",
				"source_kind":        "user_note",
				"source_label":       "admin guidance",
				"visibility":         "global",
				"sensitivity_class":  "restricted",
				"trust_class":        "trusted_internal",
				"soma_context_kind":  "identity",
				"output_specificity": "balanced",
				"tags":               []string{"shared-output-specificity", "soma-operating-context"},
			},
			sqlmock.AnyArg(), "approved",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("cccccccc-cccc-cccc-cccc-cccccccccccc", now))
	mock.ExpectExec("INSERT INTO context_vectors").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"knowledge_class": "soma_operating_context",
		"title": "Soma Identity Guidance",
		"content": "Keep shared Soma behavior steady and concise.",
		"source_kind": "user_document",
		"visibility": "private",
		"sensitivity_class": "role_scoped",
		"trust_class": "user_provided",
		"soma_context_kind": "identity",
		"output_specificity": "balanced"
	}`
	rr := doRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodPost, "/api/v1/memory/deployment-context", body)
	assertStatus(t, rr, http.StatusCreated)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["knowledge_class"] != "soma_operating_context" {
		t.Fatalf("expected soma_operating_context knowledge class, got %+v", resp)
	}
	if resp["source_kind"] != "user_note" {
		t.Fatalf("expected source_kind=user_note, got %+v", resp)
	}
	if resp["source_label"] != "admin guidance" {
		t.Fatalf("expected admin guidance source label, got %+v", resp)
	}
	if resp["visibility"] != "global" {
		t.Fatalf("expected visibility=global, got %+v", resp)
	}
	if resp["sensitivity_class"] != "restricted" {
		t.Fatalf("expected sensitivity_class=restricted, got %+v", resp)
	}
	if resp["trust_class"] != "trusted_internal" {
		t.Fatalf("expected trust_class=trusted_internal, got %+v", resp)
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
