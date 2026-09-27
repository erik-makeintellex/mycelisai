package swarm

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/deploymentcontext"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Real PostgreSQL acceptance. MYCELIS_MEMORY_TEST_DSN must point at a
// disposable database with 001_current_schema.sql installed. Skip is not proof.
func openSwarmMemoryTestDB(t *testing.T) *sql.DB {
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

// retainedStackProvider is the fake router adapter: a chat model that fails
// every embedding call, like the retained vLLM coder.
type retainedStackProvider struct {
	mu      sync.Mutex
	reply   string
	prompts []string
}

func (p *retainedStackProvider) Infer(_ context.Context, prompt string, opts cognitive.InferOptions) (*cognitive.InferResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, message := range opts.Messages {
		prompt += "\n" + message.Content
	}
	p.prompts = append(p.prompts, prompt)
	return &cognitive.InferResponse{Text: p.reply, Provider: "vllm", ModelUsed: "qwen-coder"}, nil
}
func (p *retainedStackProvider) Probe(context.Context) (bool, error) { return true, nil }
func (p *retainedStackProvider) Embed(context.Context, string, string) ([]float64, error) {
	return nil, errors.New("embedding failed")
}

func (p *retainedStackProvider) allPrompts() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return strings.Join(p.prompts, "\n")
}

func retainedStackRouter(p *retainedStackProvider) *cognitive.Router {
	return &cognitive.Router{
		Config:   &cognitive.BrainConfig{Profiles: map[string]string{"chat": "vllm"}, Providers: map[string]cognitive.ProviderConfig{"vllm": {Type: "mock", Enabled: true, ModelID: "qwen-coder"}}},
		Adapters: map[string]cognitive.LLMProvider{"vllm": p},
	}
}

func saveContext(t *testing.T, db *sql.DB, router *cognitive.Router, req deploymentcontext.IngestRequest) {
	t.Helper()
	svc := deploymentcontext.NewService(&artifacts.Service{DB: db}, memory.NewServiceWithDB(db), router)
	if _, err := svc.Ingest(context.Background(), req); err != nil {
		t.Fatalf("save %q: %v", req.Title, err)
	}
}

const (
	bakeryTitle = "Juniper & Rye Bakery"
	bakeryText  = "Weekend special: blueberry-lavender scones, 3 for $10, Saturday and Sunday. Sign-off line: See you at the counter."
	promoAsk    = "Write a two-line promo for our weekend special"
)

func seedBakeryAndScopedOutSources(t *testing.T, db *sql.DB, router *cognitive.Router) {
	saveContext(t, db, router, deploymentcontext.IngestRequest{KnowledgeClass: "company_knowledge", Title: bakeryTitle, Content: bakeryText, SourceKind: "user_note", Visibility: "global"})
	saveContext(t, db, router, deploymentcontext.IngestRequest{KnowledgeClass: "user_private_context", Title: "Competitor Watch", AgentID: "admin",
		Content: "Weekend special idea: undercut the rival with four muffins for $9.", TargetGoalSets: []string{"competitor-watch"}})
	saveContext(t, db, router, deploymentcontext.IngestRequest{KnowledgeClass: "customer_context", Title: "Team B Pricing", TeamID: "team-b",
		Visibility: "team", Content: "Weekend special for team B wholesale: croissant crates at cost."})
}

func somaAgent(router *cognitive.Router, db *sql.DB) *Agent {
	agent := NewAgent(context.Background(), protocol.AgentManifest{ID: "admin", Role: "admin", Provider: "vllm"}, "admin-core", nil, router, &countingToolExecutor{serverID: InternalServerID})
	agent.SetInternalTools(NewInternalToolRegistry(InternalToolDeps{Brain: router, Mem: memory.NewServiceWithDB(db), DB: db}))
	return agent
}

func findSource(sources []protocol.ContextSourceRef, title string) *protocol.ContextSourceRef {
	for i := range sources {
		if sources[i].Title == title {
			return &sources[i]
		}
	}
	return nil
}

func TestSomaRecallRealDB_WeekendPromoUsesBakeryFactsWithoutEmbeddings(t *testing.T) {
	db := openSwarmMemoryTestDB(t)
	provider := &retainedStackProvider{reply: "Blueberry-lavender scones, 3 for $10 this Saturday and Sunday only.\nSee you at the counter!"}
	router := retainedStackRouter(provider)
	seedBakeryAndScopedOutSources(t, db, router)

	result := somaAgent(router, db).processMessageStructured(promoAsk, nil)

	prompt := provider.allPrompts()
	for _, fact := range []string{bakeryTitle, "blueberry-lavender scones", "See you at the counter"} {
		if !strings.Contains(prompt, fact) {
			t.Fatalf("model context is missing %q", fact)
		}
	}
	for _, leaked := range []string{"Competitor Watch", "undercut the rival", "Team B Pricing", "croissant crates"} {
		if strings.Contains(prompt, leaked) {
			t.Fatalf("scoped-out source leaked into the model context: %q", leaked)
		}
	}
	ref := findSource(result.ContextSources, bakeryTitle)
	if ref == nil || !ref.Used || ref.RetrievalMode != "keyword" || ref.KnowledgeClass != "company_knowledge" || ref.ArtifactID == "" {
		t.Fatalf("the answer overlaps the bakery source and must cite it as used: %+v", result.ContextSources)
	}
	if findSource(result.ContextSources, "Competitor Watch") != nil || findSource(result.ContextSources, "Team B Pricing") != nil {
		t.Fatalf("scoped-out sources must not be cited: %+v", result.ContextSources)
	}
}

func TestSomaRecallRealDB_NoOverlapIsConsultedNotUsed(t *testing.T) {
	db := openSwarmMemoryTestDB(t)
	provider := &retainedStackProvider{reply: "Come by this weekend for something sweet and warm from the oven."}
	router := retainedStackRouter(provider)
	seedBakeryAndScopedOutSources(t, db, router)
	result := somaAgent(router, db).processMessageStructured(promoAsk, nil)
	if ref := findSource(result.ContextSources, bakeryTitle); ref == nil || ref.Used {
		t.Fatalf("without overlap the source is consulted, never used: %+v", result.ContextSources)
	}
}

func TestSomaRecallRealDB_RequestEchoOverlapDoesNotCount(t *testing.T) {
	db := openSwarmMemoryTestDB(t)
	provider := &retainedStackProvider{reply: "Fresh bakes all weekend. See you at the counter"}
	router := retainedStackRouter(provider)
	seedBakeryAndScopedOutSources(t, db, router)
	result := somaAgent(router, db).processMessageStructured("Write a weekend special promo that ends with see you at the counter", nil)
	if ref := findSource(result.ContextSources, bakeryTitle); ref == nil || ref.Used {
		t.Fatalf("overlap that only echoes the request must not count as use: %+v", result.ContextSources)
	}
}

func TestWorkerRecallRealDB_NeverReceivesRestrictedOrOperatingClasses(t *testing.T) {
	db := openSwarmMemoryTestDB(t)
	provider := &retainedStackProvider{reply: "ok"}
	router := retainedStackRouter(provider)
	saveContext(t, db, router, deploymentcontext.IngestRequest{KnowledgeClass: "company_knowledge", Title: bakeryTitle, Content: bakeryText})
	saveContext(t, db, router, deploymentcontext.IngestRequest{KnowledgeClass: "soma_operating_context", Title: "Soma Tone Guide", Content: "Weekend special replies stay playful and short."})
	saveContext(t, db, router, deploymentcontext.IngestRequest{KnowledgeClass: "user_private_context", Title: "Owner Diary", Visibility: "global", Content: "Weekend special margins are thin this month."})
	saveContext(t, db, router, deploymentcontext.IngestRequest{KnowledgeClass: "customer_context", Title: "Team A Restricted", TeamID: "team-a", SensitivityClass: "restricted", Content: "Weekend special wholesale terms for team A."})

	registry := NewInternalToolRegistry(InternalToolDeps{Brain: router, Mem: memory.NewServiceWithDB(db), DB: db})
	text, injected := registry.BuildContextWithSources("worker-b", "team-b", "worker", nil, nil, promoAsk)
	sources := citeContextSources(injected, promoAsk, "")
	if !strings.Contains(text, bakeryTitle) || findSource(sources, bakeryTitle) == nil {
		t.Fatalf("worker must still receive company knowledge: %s", text)
	}
	for _, hidden := range []string{"Soma Tone Guide", "Owner Diary", "Team A Restricted"} {
		if strings.Contains(text, hidden) || findSource(sources, hidden) != nil {
			t.Fatalf("worker received %q", hidden)
		}
	}
}

func TestSearchMemoryRealDB_ForgedScopeArgsDoNotWiden(t *testing.T) {
	db := openSwarmMemoryTestDB(t)
	router := retainedStackRouter(&retainedStackProvider{reply: "ok"})
	saveContext(t, db, router, deploymentcontext.IngestRequest{KnowledgeClass: "customer_context", Title: "Team A Plan", TeamID: "team-a", Visibility: "team", Content: "Weekend special rollout plan for team A."})
	registry := NewInternalToolRegistry(InternalToolDeps{Brain: router, Mem: memory.NewServiceWithDB(db), DB: db})
	ctx := WithToolInvocationContext(context.Background(), ToolInvocationContext{AgentID: "worker-b", TeamID: "team-b"})
	out, err := registry.handleSearchMemory(ctx, map[string]any{"query": "weekend special rollout", "team_id": "team-a", "agent_id": "lead-a"})
	if err != nil {
		t.Fatalf("keyword search_memory must work without embeddings: %v", err)
	}
	if strings.Contains(out, "Team A Plan") || strings.Contains(out, "rollout plan for team A") {
		t.Fatalf("forged team_id widened scope: %s", out)
	}
	own := WithToolInvocationContext(context.Background(), ToolInvocationContext{AgentID: "lead-a", TeamID: "team-a"})
	out, err = registry.handleSearchMemory(own, map[string]any{"query": "weekend special rollout"})
	if err != nil || !strings.Contains(out, "rollout plan for team A") || !strings.Contains(out, "keyword") {
		t.Fatalf("team A must find its own row by keyword: %s err=%v", out, err)
	}
}

func TestLoadDeploymentContextRealDB_OrgWideClassNeedsOperatorConfirmation(t *testing.T) {
	db := openSwarmMemoryTestDB(t)
	router := retainedStackRouter(&retainedStackProvider{reply: "ok"})
	registry := NewInternalToolRegistry(InternalToolDeps{Brain: router, Mem: memory.NewServiceWithDB(db), DB: db})
	args := func(class string) map[string]any {
		return map[string]any{"title": "Guidance " + class, "content": "Always greet customers warmly.", "knowledge_class": class, "visibility": "global"}
	}
	planning := WithToolInvocationContext(context.Background(), ToolInvocationContext{AgentID: "admin", TeamID: "admin-core", PlanningOnly: true})
	for _, ctx := range []context.Context{context.Background(), planning} {
		for _, class := range []string{"company_knowledge", "soma_operating_context"} {
			if _, err := registry.handleLoadDeploymentContext(ctx, args(class)); err == nil || !strings.Contains(err.Error(), "context_class_requires_approval") {
				t.Fatalf("unconfirmed %s write must be blocked, got %v", class, err)
			}
		}
	}
	var stored int
	_ = db.QueryRow(`SELECT count(*) FROM artifacts WHERE metadata->>'knowledge_store'='governed_context_store'`).Scan(&stored)
	if stored != 0 {
		t.Fatalf("blocked writes stored %d rows", stored)
	}
	if _, err := registry.handleLoadDeploymentContext(context.Background(), args("customer_context")); err != nil {
		t.Fatalf("customer context stays writable: %v", err)
	}
	confirmed := WithToolInvocationContext(context.Background(), ToolInvocationContext{AgentID: "admin", TeamID: "admin-core"})
	out, err := registry.handleLoadDeploymentContext(confirmed, args("company_knowledge"))
	if err != nil || !strings.Contains(out, "pending") {
		t.Fatalf("operator-confirmed company knowledge write must save with an honest status: %s err=%v", out, err)
	}
}
