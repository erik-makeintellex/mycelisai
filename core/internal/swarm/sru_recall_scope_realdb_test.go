package swarm

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/deploymentcontext"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/searchcap"
	"github.com/mycelis/core/pkg/protocol"
)

// SRU (HIGH, from MEM-LIST-QA): Soma recall was scoped by agent and team,
// never by user, so user B's chat recalled user A's private, reflection and
// admin-core team entries into the prompt, context_sources and the memory
// tools, and promote read any entry by id. Real PostgreSQL
// (MYCELIS_MEMORY_TEST_DSN).

const sruAsk = "What is our quokka weekend plan?"

type sruWorld struct {
	db                     *sql.DB
	router                 *cognitive.Router
	provider               *retainedStackProvider
	userA, userB, userC    string
	adaPromoteSource       string
	accessA, accessB, accC RecallAccess
}

// sruText is the distinctive content of each seeded entry.
var sruText = map[string]string{
	"ada-private":    "Quokka weekend plan: Ada private ledger balance 4410.",
	"ada-reflection": "Quokka weekend plan reflection: Ada prefers dawn deliveries.",
	"team":           "Quokka weekend plan for the admin-core team: rota seven.",
	"org":            "Quokka weekend plan: the shop opens at nine.",
	"bob-private":    "Quokka weekend plan: Bob private reminder eighty-eight.",
}

func sruSeed(t *testing.T) *sruWorld {
	t.Helper()
	db := openSwarmMemoryTestDB(t)
	provider := &retainedStackProvider{reply: "Here is the quokka weekend plan."}
	w := &sruWorld{db: db, provider: provider, router: retainedStackRouter(provider),
		userA: uuid.NewString(), userB: uuid.NewString(), userC: uuid.NewString()}
	teamKey := memory.GovernedTeamKey("default", "admin-core")
	w.accessA = RecallAccess{User: true, Reader: memory.GovernedReader{UserID: w.userA, Label: "ada-sru", TeamKeys: []string{teamKey}}}
	w.accessB = RecallAccess{User: true, Reader: memory.GovernedReader{UserID: w.userB, Label: "bob-sru"}}
	w.accC = RecallAccess{User: true, Reader: memory.GovernedReader{UserID: w.userC, Label: "cy-sru", TeamKeys: []string{teamKey}}}
	owner := func(id string) map[string]any { return map[string]any{"owner_user_id": id} }
	save := func(req deploymentcontext.IngestRequest) { saveContext(t, db, w.router, req) }
	save(deploymentcontext.IngestRequest{KnowledgeClass: "user_private_context", Title: "SRU Ada Private Ledger", Content: sruText["ada-private"], UserLabel: "ada-sru", ExtraMetadata: owner(w.userA)})
	save(deploymentcontext.IngestRequest{KnowledgeClass: "reflection_synthesis", Title: "SRU Ada Reflection", Content: sruText["ada-reflection"], UserLabel: "ada-sru", ExtraMetadata: owner(w.userA)})
	save(deploymentcontext.IngestRequest{KnowledgeClass: "customer_context", Title: "SRU Admin Core Team Note", TeamID: "admin-core", Visibility: "team", Content: sruText["team"], UserLabel: "ada-sru", ExtraMetadata: owner(w.userA)})
	save(deploymentcontext.IngestRequest{KnowledgeClass: "company_knowledge", Title: "SRU Org Quokka Hours", Visibility: "global", Content: sruText["org"], UserLabel: "ada-sru", ExtraMetadata: owner(w.userA)})
	save(deploymentcontext.IngestRequest{KnowledgeClass: "user_private_context", Title: "SRU Bob Private Note", Content: sruText["bob-private"], UserLabel: "bob-sru", ExtraMetadata: owner(w.userB)})
	save(deploymentcontext.IngestRequest{KnowledgeClass: "customer_context", Title: "SRU Ada Promote Source", Visibility: "private", Content: "Quokka supplier contract terms for Ada only.", UserLabel: "ada-sru", ExtraMetadata: owner(w.userA)})
	if err := db.QueryRow(`SELECT id::text FROM artifacts WHERE title = 'SRU Ada Promote Source'`).Scan(&w.adaPromoteSource); err != nil {
		t.Fatalf("promote source: %v", err)
	}
	return w
}

// sruCheck fails when text is missing a wanted entry or holds an unwanted one.
func sruCheck(t *testing.T, label, text string, want, deny []string) {
	t.Helper()
	for _, key := range want {
		if !strings.Contains(text, sruText[key]) {
			t.Errorf("%s: missing %s entry", label, key)
		}
	}
	for _, key := range deny {
		if strings.Contains(text, sruText[key]) {
			t.Errorf("%s: leaked %s entry", label, key)
		}
	}
}

func sruToolCtx(access RecallAccess) context.Context {
	return WithToolInvocationContext(context.Background(), ToolInvocationContext{AgentID: "admin", TeamID: "admin-core", AgentRole: "admin", Recall: access})
}

func (w *sruWorld) registry() *InternalToolRegistry {
	return NewInternalToolRegistry(InternalToolDeps{Brain: w.router, Mem: memory.NewServiceWithDB(w.db), DB: w.db})
}

func sruSourceTitles(sources []protocol.ContextSourceRef) string {
	titles := make([]string, 0, len(sources))
	for _, source := range sources {
		titles = append(titles, source.Title)
	}
	return strings.Join(titles, "|")
}

func TestSRUSomaRecallRealDB_OtherUserTurnHidesPrivateReflectionAndTeam(t *testing.T) {
	w := sruSeed(t)
	result := somaAgent(w.router, w.db).processUserTurn(w.accessB, sruAsk, nil)
	sruCheck(t, "B prompt", w.provider.allPrompts(), []string{"org", "bob-private"}, []string{"ada-private", "ada-reflection", "team"})
	titles := sruSourceTitles(result.ContextSources)
	for _, hidden := range []string{"SRU Ada Private Ledger", "SRU Ada Reflection", "SRU Admin Core Team Note"} {
		if strings.Contains(titles, hidden) {
			t.Errorf("context_sources gave B the title %q: %s", hidden, titles)
		}
	}
	if !strings.Contains(titles, "SRU Org Quokka Hours") || !strings.Contains(titles, "SRU Bob Private Note") {
		t.Errorf("B must still get org-wide and own entries as sources: %s", titles)
	}
}

func TestSRUSomaRecallRealDB_OwnerAndTeamMemberKeepTheirEntries(t *testing.T) {
	w := sruSeed(t)
	somaAgent(w.router, w.db).processUserTurn(w.accessA, sruAsk, nil)
	sruCheck(t, "A prompt", w.provider.allPrompts(), []string{"ada-private", "ada-reflection", "team", "org"}, []string{"bob-private"})

	member := &retainedStackProvider{reply: "ok"}
	router := retainedStackRouter(member)
	somaAgent(router, w.db).processUserTurn(w.accC, sruAsk, nil)
	sruCheck(t, "team member prompt", member.allPrompts(), []string{"team", "org"}, []string{"ada-private", "ada-reflection", "bob-private"})
}

func TestSRUSomaRecallRealDB_MemoryToolsFollowTheRequestingUser(t *testing.T) {
	w := sruSeed(t)
	registry := w.registry()
	args := func() map[string]any { return map[string]any{"query": "quokka weekend plan", "limit": float64(10)} }
	for _, tc := range []struct {
		name       string
		access     RecallAccess
		want, deny []string
	}{
		{"B", w.accessB, []string{"org", "bob-private"}, []string{"ada-private", "ada-reflection", "team"}},
		{"A", w.accessA, []string{"org", "ada-private", "ada-reflection", "team"}, []string{"bob-private"}},
		{"team member", w.accC, []string{"org", "team"}, []string{"ada-private", "ada-reflection", "bob-private"}},
		{"no user", RecallAccess{}, []string{"org"}, []string{"ada-private", "ada-reflection", "team", "bob-private"}},
	} {
		out, err := registry.handleSearchMemory(sruToolCtx(tc.access), args())
		if err != nil {
			t.Fatalf("%s search_memory: %v", tc.name, err)
		}
		sruCheck(t, tc.name+" search_memory", out, tc.want, tc.deny)
		out, err = registry.handleRecall(sruToolCtx(tc.access), args())
		if err != nil {
			t.Fatalf("%s recall: %v", tc.name, err)
		}
		sruCheck(t, tc.name+" recall", out, tc.want, tc.deny)
	}
}

func TestSRUSomaRecallRealDB_NoUserTurnReadsOrgWideOnly(t *testing.T) {
	w := sruSeed(t)
	result := somaAgent(w.router, w.db).processMessageStructured(sruAsk, nil)
	sruCheck(t, "no-user prompt", w.provider.allPrompts(), []string{"org"}, []string{"ada-private", "ada-reflection", "team", "bob-private"})
	if titles := sruSourceTitles(result.ContextSources); titles != "SRU Org Quokka Hours" {
		t.Fatalf("no-user sources must be org-wide only: %s", titles)
	}
}

func TestSRUSomaRecallRealDB_FailedLookupWithholdsGovernedLaneAndSaysSo(t *testing.T) {
	w := sruSeed(t)
	unavailable := RecallAccess{Unavailable: true}
	result := somaAgent(w.router, w.db).processUserTurn(unavailable, sruAsk, nil)
	sruCheck(t, "unavailable prompt", w.provider.allPrompts(), nil, []string{"org", "ada-private", "ada-reflection", "team", "bob-private"})
	if len(result.ContextSources) != 0 || !strings.Contains(result.Text, recallUnavailableNote) {
		t.Fatalf("an unavailable lane cites nothing and says so: sources=%v text=%q", result.ContextSources, result.Text)
	}
	for name, call := range map[string]func(context.Context, map[string]any) (string, error){
		"search_memory": w.registry().handleSearchMemory, "recall": w.registry().handleRecall,
	} {
		if out, err := call(sruToolCtx(unavailable), map[string]any{"query": "quokka weekend plan"}); err == nil || strings.Contains(out, "Quokka") {
			t.Fatalf("%s must refuse when memory access is unverified: %s err=%v", name, out, err)
		}
	}
}

// sruConfirmed is an operator-confirmed (non-planning) tool call.
func sruConfirmed(access RecallAccess) context.Context {
	return WithToolInvocationContext(context.Background(), ToolInvocationContext{AgentID: "admin", TeamID: "admin-core", UserLabel: "ada-sru", Recall: access})
}

func (w *sruWorld) promotedCount(t *testing.T) int {
	t.Helper()
	var n int
	if err := w.db.QueryRow(`SELECT count(*) FROM artifacts WHERE metadata->>'knowledge_class' = 'company_knowledge'
		AND metadata->>'promoted_from_artifact_id' = $1`, w.adaPromoteSource).Scan(&n); err != nil {
		t.Fatalf("count promoted: %v", err)
	}
	return n
}

func TestSRUPromoteRealDB_RequiresTheRequestingUserToReadTheSource(t *testing.T) {
	w := sruSeed(t)
	registry := w.registry()
	args := map[string]any{"source_artifact_id": w.adaPromoteSource, "title": "SRU promoted"}
	_, unknownErr := registry.handlePromoteDeploymentContext(sruConfirmed(w.accessB), map[string]any{"source_artifact_id": uuid.NewString()})
	for name, access := range map[string]RecallAccess{"B": w.accessB, "team member": w.accC, "no user": {}, "unavailable": {Unavailable: true}} {
		out, err := registry.handlePromoteDeploymentContext(sruConfirmed(access), args)
		if err == nil {
			t.Fatalf("promote of A's private entry by %s must be refused: %s", name, out)
		}
		if name == "B" && (unknownErr == nil || err.Error() != unknownErr.Error()) {
			t.Fatalf("an unreadable source must answer like an unknown id: %v vs %v", err, unknownErr)
		}
	}
	if n := w.promotedCount(t); n != 0 {
		t.Fatalf("refused promotes wrote %d entries", n)
	}
	if out, err := registry.handlePromoteDeploymentContext(sruConfirmed(w.accessA), args); err != nil || !strings.Contains(out, "company_knowledge") {
		t.Fatalf("A promotes their own entry: %s err=%v", out, err)
	}
	if n := w.promotedCount(t); n != 1 {
		t.Fatalf("owner promote wrote %d entries", n)
	}
}

// web_search over local sources recalls the same governed store, so it
// follows the same requesting-user rule.
func TestSRUWebSearchLocalSourcesRealDB_FollowTheRequestingUser(t *testing.T) {
	w := sruSeed(t)
	mem := memory.NewServiceWithDB(w.db)
	registry := NewInternalToolRegistry(InternalToolDeps{Brain: w.router, Mem: mem, DB: w.db,
		Search: searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderLocalSources}, nil, mem)})
	args := func() map[string]any {
		return map[string]any{"query": "quokka weekend plan", "source_scope": "local_sources", "max_results": float64(10)}
	}
	for _, tc := range []struct {
		name       string
		access     RecallAccess
		want, deny []string
	}{
		{"B", w.accessB, []string{"org", "bob-private"}, []string{"ada-private", "ada-reflection", "team"}},
		{"A", w.accessA, []string{"org", "ada-private", "team"}, []string{"bob-private"}},
		{"no user", RecallAccess{}, []string{"org"}, []string{"ada-private", "ada-reflection", "team", "bob-private"}},
	} {
		out, err := registry.handleWebSearch(sruToolCtx(tc.access), args())
		if err != nil {
			t.Fatalf("%s web_search: %v", tc.name, err)
		}
		sruCheck(t, tc.name+" web_search", out, tc.want, tc.deny)
	}
	if out, err := registry.handleWebSearch(sruToolCtx(RecallAccess{Unavailable: true}), args()); err == nil || strings.Contains(out, "Quokka") {
		t.Fatalf("web_search must refuse when memory access is unverified: %s err=%v", out, err)
	}
}
