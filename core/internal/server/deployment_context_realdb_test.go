package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/memory"
)

// Real PostgreSQL proof. MYCELIS_MEMORY_TEST_DSN must point at a disposable
// database with 001_current_schema.sql installed. Skip is not proof.
func openMemoryServerTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("MYCELIS_MEMORY_TEST_DSN")
	if dsn == "" {
		t.Skip("requires disposable MYCELIS_MEMORY_TEST_DSN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	cleanup := func() {
		_, _ = db.Exec(`DELETE FROM context_vectors WHERE metadata->>'knowledge_store' = 'governed_context_store'`)
		_, _ = db.Exec(`DELETE FROM artifacts WHERE metadata->>'knowledge_store' = 'governed_context_store'`)
	}
	cleanup()
	t.Cleanup(cleanup)
	return db
}

type failingEmbedChatProvider struct{}

func (failingEmbedChatProvider) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "ok", Provider: "vllm", ModelUsed: "qwen-coder"}, nil
}
func (failingEmbedChatProvider) Probe(context.Context) (bool, error) { return true, nil }
func (failingEmbedChatProvider) Embed(context.Context, string, string) ([]float64, error) {
	return nil, errors.New("embedding failed")
}

// chatOnlyBrain mirrors the retained stack: a chat model whose adapter
// answers Embed with a failure on every call.
func chatOnlyBrain() *cognitive.Router {
	return &cognitive.Router{
		Config:   &cognitive.BrainConfig{Profiles: map[string]string{"chat": "vllm"}, Providers: map[string]cognitive.ProviderConfig{"vllm": {Enabled: true}}},
		Adapters: map[string]cognitive.LLMProvider{"vllm": failingEmbedChatProvider{}},
	}
}

func memoryTestServer(db *sql.DB, brain *cognitive.Router) *AdminServer {
	return &AdminServer{DB: db, Artifacts: artifacts.NewService(db, os.TempDir()), Mem: memory.NewServiceWithDB(db), Cognitive: brain}
}

const bakerySaveBody = `{"title":"Juniper & Rye Bakery","knowledge_class":"company_knowledge","source_kind":"user_note","visibility":"global",
"content":"Weekend special: blueberry-lavender scones, 3 for $10, Saturday and Sunday. Sign-off line: See you at the counter."}`

func TestDeploymentContextRealDB_SaveWithoutEmbeddingsIsDurableAndHonest(t *testing.T) {
	db := openMemoryServerTestDB(t)
	s := memoryTestServer(db, chatOnlyBrain())
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodPost, "/api/v1/memory/deployment-context", bakerySaveBody)
	assertStatus(t, rr, http.StatusCreated)
	var saved map[string]any
	assertJSON(t, rr, &saved)
	if saved["embedding_status"] != "pending" || saved["vector_count"] != float64(0) {
		t.Fatalf("save must report pending embeddings honestly: %+v", saved)
	}
	modes, _ := saved["retrieval_modes"].([]any)
	if len(modes) != 1 || modes[0] != "keyword" || saved["status_message"] == "" {
		t.Fatalf("save must say keyword-only recall: %+v", saved)
	}

	var artifacts, chunks, orphans int
	_ = db.QueryRow(`SELECT count(*) FROM artifacts WHERE metadata->>'knowledge_store'='governed_context_store'`).Scan(&artifacts)
	_ = db.QueryRow(`SELECT count(*) FROM context_vectors WHERE metadata->>'artifact_id' = $1 AND embedding IS NULL`, saved["artifact_id"]).Scan(&chunks)
	_ = db.QueryRow(`SELECT count(*) FROM artifacts a WHERE a.metadata->>'knowledge_store'='governed_context_store'
		AND NOT EXISTS (SELECT 1 FROM context_vectors v WHERE v.metadata->>'artifact_id' = a.id::text)`).Scan(&orphans)
	if artifacts != 1 || chunks != 1 || orphans != 0 {
		t.Fatalf("durable save: artifacts=%d pendingChunks=%d orphans=%d", artifacts, chunks, orphans)
	}

	list := doAuthenticatedRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodGet, "/api/v1/memory/deployment-context", "")
	assertStatus(t, list, http.StatusOK)
	var listed map[string]any
	assertJSON(t, list, &listed)
	entries, _ := listed["entries"].([]any)
	if listed["semantic_search"] != "unavailable" || len(entries) != 1 {
		t.Fatalf("GET must list the entry and report semantic search unavailable: %+v", listed)
	}
	if entry, _ := entries[0].(map[string]any); entry["embedding_status"] != "pending" {
		t.Fatalf("entry must carry its embedding status: %+v", entry)
	}
}

func TestDeploymentContextBackfillRealDB_AdminScopedAndAudited(t *testing.T) {
	db := openMemoryServerTestDB(t)
	s := memoryTestServer(db, chatOnlyBrain())
	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodPost, "/api/v1/memory/deployment-context", bakerySaveBody)
	assertStatus(t, rr, http.StatusCreated)

	handler := http.HandlerFunc(s.HandleDeploymentContextBackfill)
	path := "/api/v1/memory/deployment-context/backfill"
	operator := localAdminIdentityForTest()
	operator.Role, operator.Scopes = "operator", []string{"*"}
	assertStatus(t, doAuthenticatedRequestAs(t, handler, http.MethodPost, path, "", operator), http.StatusForbidden)
	unscoped := localAdminIdentityForTest()
	unscoped.Scopes = []string{"groups:read"}
	assertStatus(t, doAuthenticatedRequestAs(t, handler, http.MethodPost, path, "", unscoped), http.StatusForbidden)
	assertStatus(t, doRequest(t, handler, http.MethodPost, path, ""), http.StatusUnauthorized)

	scoped := localAdminIdentityForTest()
	scoped.Scopes = []string{"memory:write"}
	down := doAuthenticatedRequestAs(t, handler, http.MethodPost, path, "", scoped)
	assertStatus(t, down, http.StatusOK)
	var downBody map[string]any
	assertJSON(t, down, &downBody)
	data, _ := downBody["data"].(map[string]any)
	if data["status"] != "unavailable" || data["embedded"] != float64(0) || data["remaining"] != float64(1) {
		t.Fatalf("backfill without an engine must report unavailable honestly: %+v", downBody)
	}

	vec := make([]float64, 768)
	vec[0] = 1
	s.Cognitive = newDeploymentContextBrainWithVector(vec)
	up := doAuthenticatedRequestAs(t, handler, http.MethodPost, path, "", scoped)
	assertStatus(t, up, http.StatusOK)
	var upBody map[string]any
	assertJSON(t, up, &upBody)
	data, _ = upBody["data"].(map[string]any)
	if data["status"] != "complete" || data["embedded"] != float64(1) || data["remaining"] != float64(0) {
		t.Fatalf("backfill with an engine must embed the pending row: %+v", upBody)
	}
	var audits int
	_ = db.QueryRow(`SELECT count(*) FROM log_entries WHERE level='audit' AND context->>'action' = 'deployment_context_backfill'`).Scan(&audits)
	if audits < 2 {
		t.Fatalf("each backfill run must be audited, found %d", audits)
	}
}
