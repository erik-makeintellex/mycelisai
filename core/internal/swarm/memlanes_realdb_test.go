package swarm

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/memory"
)

// MEM-LANES (HIGH, from SRU-QA): remembered facts, conversation summaries and
// lead temp-memory channels were scoped by agent and team, never by user, so
// user B's Soma recalled user A's facts, summaries and checkpoints. Real
// PostgreSQL (MYCELIS_MEMORY_TEST_DSN).

// memlanesEmbedProvider is a chat model with a working embedding engine, so
// conversation summaries get vectors and ambient summary recall runs.
type memlanesEmbedProvider struct{ *retainedStackProvider }

func (memlanesEmbedProvider) Embed(context.Context, string, string) ([]float64, error) {
	vec := make([]float64, cognitive.EmbeddingDimensions)
	for i := range vec {
		vec[i] = 0.01
	}
	return vec, nil
}

func memlanesRouter(reply string, embed bool) (*cognitive.Router, *retainedStackProvider) {
	base := &retainedStackProvider{reply: reply}
	var adapter cognitive.LLMProvider = base
	if embed {
		adapter = memlanesEmbedProvider{base}
	}
	return &cognitive.Router{
		Config:   &cognitive.BrainConfig{Profiles: map[string]string{"chat": "vllm"}, Providers: map[string]cognitive.ProviderConfig{"vllm": {Type: "mock", Enabled: true, ModelID: "qwen-coder"}}},
		Adapters: map[string]cognitive.LLMProvider{"vllm": adapter},
	}, base
}

type memlanesWorld struct {
	db                  *sql.DB
	userA, userB, userC string
	accA, accB, accC    RecallAccess
}

func memlanesSeed(t *testing.T) *memlanesWorld {
	t.Helper()
	db := openSwarmMemoryTestDB(t)
	clean := func() {
		_, _ = db.Exec(`DELETE FROM agent_memories WHERE content LIKE 'MemLanes%'`)
		_, _ = db.Exec(`DELETE FROM context_vectors WHERE content LIKE '%MemLanes%'`)
		_, _ = db.Exec(`DELETE FROM conversation_summaries WHERE summary LIKE '%MemLanes%'`)
		_, _ = db.Exec(`DELETE FROM temp_memory_channels WHERE content LIKE '%MemLanes%'`)
	}
	clean()
	t.Cleanup(clean)
	w := &memlanesWorld{db: db, userA: uuid.NewString(), userB: uuid.NewString(), userC: uuid.NewString()}
	teamKey := memory.GovernedTeamKey("default", "admin-core")
	w.accA = RecallAccess{User: true, Reader: memory.GovernedReader{UserID: w.userA, Label: "ada-lanes", TeamKeys: []string{teamKey}}}
	w.accB = RecallAccess{User: true, Reader: memory.GovernedReader{UserID: w.userB, Label: "bob-lanes"}}
	w.accC = RecallAccess{User: true, Reader: memory.GovernedReader{UserID: w.userC, Label: "cy-lanes", TeamKeys: []string{teamKey}}}
	return w
}

func (w *memlanesWorld) registry(router *cognitive.Router) *InternalToolRegistry {
	return NewInternalToolRegistry(InternalToolDeps{Brain: router, Mem: memory.NewServiceWithDB(w.db), DB: w.db})
}

// memlanesCheck fails when text misses a wanted marker or holds a denied one.
func memlanesCheck(t *testing.T, label, text string, want, deny []string) {
	t.Helper()
	for _, marker := range want {
		if !strings.Contains(text, marker) {
			t.Errorf("%s: missing %q", label, marker)
		}
	}
	for _, marker := range deny {
		if strings.Contains(text, marker) {
			t.Errorf("%s: leaked %q", label, marker)
		}
	}
}

func memlanesCall(t *testing.T, label string, call func(context.Context, map[string]any) (string, error), access RecallAccess, args map[string]any) string {
	t.Helper()
	out, err := call(sruToolCtx(access), args)
	if err != nil {
		t.Fatalf("%s: %v", label, err)
	}
	return out
}

func TestMemLanesRealDB_RememberedFactsFollowTheRequestingUser(t *testing.T) {
	w := memlanesSeed(t)
	router, _ := memlanesRouter("ok", false)
	reg := w.registry(router)
	const aPrivate, aTeam, bFact = "MemLanes wombat vault code 7719.", "MemLanes wombat team rota 3302.", "MemLanes wombat bob note 5150."
	memlanesCall(t, "A remember", reg.handleRemember, w.accA, map[string]any{"category": "fact", "content": aPrivate})
	memlanesCall(t, "A remember team", reg.handleRemember, w.accA, map[string]any{"category": "fact", "content": aTeam, "visibility": "team"})
	memlanesCall(t, "B remember", reg.handleRemember, w.accB, map[string]any{"category": "fact", "content": bFact})
	args := func() map[string]any { return map[string]any{"query": "wombat", "limit": float64(10)} }
	for _, tc := range []struct {
		name       string
		access     RecallAccess
		want, deny []string
	}{
		{"A", w.accA, []string{aPrivate, aTeam}, []string{bFact}},
		{"B", w.accB, []string{bFact}, []string{aPrivate, aTeam}},
		{"team member C", w.accC, []string{aTeam}, []string{aPrivate, bFact}},
		{"no user", RecallAccess{}, nil, []string{aPrivate, aTeam, bFact}},
	} {
		memlanesCheck(t, tc.name+" recall", memlanesCall(t, tc.name+" recall", reg.handleRecall, tc.access, args()), tc.want, tc.deny)
		memlanesCheck(t, tc.name+" search_memory", memlanesCall(t, tc.name+" search", reg.handleSearchMemory, tc.access, args()), tc.want, tc.deny)
	}
	if _, err := reg.handleRemember(sruToolCtx(RecallAccess{Unavailable: true}), map[string]any{"category": "fact", "content": "MemLanes unverified 1111."}); err == nil {
		t.Error("remember must refuse when the requesting user could not be verified")
	}
	var unverified int
	_ = w.db.QueryRow(`SELECT count(*) FROM agent_memories WHERE content LIKE 'MemLanes unverified%'`).Scan(&unverified)
	if unverified != 0 {
		t.Fatalf("a refused remember wrote %d rows", unverified)
	}
}

// Rows saved before owners were recorded are not user-readable unless they
// are already org-wide.
func TestMemLanesRealDB_LegacyOwnerlessRowsAreNotUserReadable(t *testing.T) {
	w := memlanesSeed(t)
	router, _ := memlanesRouter("ok", false)
	reg := w.registry(router)
	for _, row := range []struct{ content, visibility string }{
		{"MemLanes legacy wombat team fact 8080.", "team"}, {"MemLanes legacy wombat private fact 8181.", "private"},
		{"MemLanes legacy wombat global fact 9090.", "global"}} {
		if _, err := w.db.Exec(`INSERT INTO agent_memories (category, content, tenant_id, team_id, agent_id, visibility)
			VALUES ('fact', $1, 'default', 'admin-core', 'admin', $2)`, row.content, row.visibility); err != nil {
			t.Fatalf("seed legacy fact: %v", err)
		}
	}
	if _, err := w.db.Exec(`INSERT INTO context_vectors (content, embedding, metadata) VALUES ($1, NULL,
		'{"type":"agent_memory","tenant_id":"default","team_id":"admin-core","agent_id":"admin","visibility":"team"}')`, "[fact] MemLanes legacy wombat vector fact 6060."); err != nil {
		t.Fatalf("seed legacy vector: %v", err)
	}
	for name, access := range map[string]RecallAccess{"A": w.accA, "B": w.accB, "C": w.accC, "no user": {}} {
		out := memlanesCall(t, name+" recall", reg.handleRecall, access, map[string]any{"query": "wombat", "limit": float64(10)})
		memlanesCheck(t, name+" legacy recall", out, []string{"9090"}, []string{"8080", "8181", "6060"})
	}
}

func TestMemLanesRealDB_ConversationSummariesFollowTheRequestingUser(t *testing.T) {
	w := memlanesSeed(t)
	const summary = "MemLanes platypus harbour summary 4242."
	routerA, _ := memlanesRouter(summary, true)
	if _, err := w.registry(routerA).handleSummarizeConversation(sruToolCtx(w.accA), map[string]any{"messages": "user: plan the platypus harbour"}); err != nil {
		t.Fatalf("A summarize: %v", err)
	}
	if _, err := w.registry(routerA).handleSummarizeConversation(sruToolCtx(RecallAccess{Unavailable: true}), map[string]any{"messages": "user: x"}); err == nil {
		t.Error("summarize must refuse when the requesting user could not be verified")
	}
	const ask = "What was the platypus harbour plan?"
	for _, tc := range []struct {
		name   string
		access RecallAccess
		sees   bool
	}{{"A", w.accA, true}, {"B", w.accB, false}, {"team member C", w.accC, false}, {"no user", RecallAccess{}, false}} {
		router, provider := memlanesRouter("ok", true)
		agent := somaAgent(router, w.db)
		if tc.access.User {
			agent.processUserTurn(tc.access, ask, nil)
		} else {
			agent.processMessageStructured(ask, nil)
		}
		search := memlanesCall(t, tc.name+" search", w.registry(router).handleSearchMemory, tc.access,
			map[string]any{"query": "platypus harbour", "types": []any{"conversation"}, "limit": float64(10)})
		if tc.sees {
			memlanesCheck(t, tc.name+" prompt", provider.allPrompts(), []string{"4242"}, nil)
			memlanesCheck(t, tc.name+" search_memory", search, []string{"4242"}, nil)
		} else {
			memlanesCheck(t, tc.name+" prompt", provider.allPrompts(), nil, []string{"4242"})
			memlanesCheck(t, tc.name+" search_memory", search, nil, []string{"4242"})
		}
	}
}

func (w *memlanesWorld) waitForTemp(t *testing.T, marker string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var n int
		_ = w.db.QueryRow(`SELECT count(*) FROM temp_memory_channels WHERE content LIKE '%' || $1 || '%'`, marker).Scan(&n)
		if n > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("no temp checkpoint holding %q was written", marker)
}

func TestMemLanesRealDB_TempChannelsFollowTheRequestingUser(t *testing.T) {
	w := memlanesSeed(t)
	router, _ := memlanesRouter("ok", false)
	reg := w.registry(router)
	const aTemp, legacy, signal = "MemLanes kiwi checkpoint 3131.", "MemLanes legacy kiwi temp 1212.", "MemLanes kiwi signal 4545."
	memlanesCall(t, "A temp write", reg.handleTempMemoryWrite, w.accA, map[string]any{"channel": "lead.shared", "content": aTemp})
	// MEM-LANES-2: a signal checkpoint reaches users only as system bus state;
	// a legacy row with neither owner key reaches no one.
	if _, err := w.db.Exec(`INSERT INTO temp_memory_channels (tenant_id, channel_key, owner_agent_id, content, metadata) VALUES
		('default', 'lead.shared', 'admin', $1, '{}'), ('default', 'signal.latest.swarm.team.memlanes.signal.status', 'memlanes', $2, '{"owner_class":"system"}'),
		('default', 'signal.latest.swarm.team.memlanes-legacy.signal.status', 'memlanes', 'MemLanes legacy kiwi signal 4646.', '{}')`, legacy, signal); err != nil {
		t.Fatalf("seed legacy temp: %v", err)
	}
	// A's own chat history becomes a continuity checkpoint in the team planning channel.
	aRouter, _ := memlanesRouter("MemLanes kiwi continuity 2727.", false)
	history := make([]cognitive.ChatMessage, 15)
	for i := range history {
		history[i] = cognitive.ChatMessage{Role: "user", Content: "kiwi planning"}
	}
	somaAgent(aRouter, w.db).processUserTurn(w.accA, "Keep planning the kiwi launch.", history)
	w.waitForTemp(t, "2727")

	read := func(access RecallAccess) string {
		return memlanesCall(t, "temp read", reg.handleTempMemoryRead, access, map[string]any{"channel": "lead.shared"})
	}
	memlanesCheck(t, "A temp_memory_read", read(w.accA), []string{"3131"}, []string{"1212"})
	for name, access := range map[string]RecallAccess{"B": w.accB, "C": w.accC, "no user": {}} {
		memlanesCheck(t, name+" temp_memory_read", read(access), nil, []string{"3131"})
	}
	for _, tc := range []struct {
		name   string
		access RecallAccess
		want   []string
		deny   []string
	}{
		{"A", w.accA, []string{"3131", "2727"}, []string{"1212"}},
		{"B", w.accB, nil, []string{"3131", "2727", "1212"}},
	} {
		prompterRouter, provider := memlanesRouter("ok", false)
		somaAgent(prompterRouter, w.db).processUserTurn(tc.access, "What is next for the kiwi launch?", nil)
		memlanesCheck(t, tc.name+" prompt", provider.allPrompts(), tc.want, tc.deny)
	}
	signalOut := memlanesCall(t, "B read_signals", reg.handleReadSignals, w.accB,
		map[string]any{"latest_only": true, "channel_key": "signal.latest.swarm.team.memlanes.signal.status"})
	memlanesCheck(t, "B read_signals", signalOut, []string{"4545"}, nil)
	legacyOut := memlanesCall(t, "B read_signals legacy", reg.handleReadSignals, w.accB,
		map[string]any{"latest_only": true, "channel_key": "signal.latest.swarm.team.memlanes-legacy.signal.status"})
	memlanesCheck(t, "B read_signals legacy", legacyOut, nil, []string{"4646"})
	memlanesCall(t, "B temp clear", reg.handleTempMemoryClear, w.accB, map[string]any{"channel": "lead.shared"})
	memlanesCheck(t, "A after B clear", read(w.accA), []string{"3131"}, nil)
	if _, err := reg.handleTempMemoryWrite(sruToolCtx(RecallAccess{Unavailable: true}), map[string]any{"channel": "lead.shared", "content": "MemLanes unverified 2222."}); err == nil {
		t.Fatal("temp_memory_write must refuse when the requesting user could not be verified")
	}
}

// Soma-tool saves carry the requesting user as owner; a no-user save is never
// readable through the legacy "soma" owner label.
func TestMemLanesRealDB_SomaToolSavesCarryTheRequestingUser(t *testing.T) {
	w := memlanesSeed(t)
	router, _ := memlanesRouter("ok", false)
	reg := w.registry(router)
	save := func(access RecallAccess, title string) string {
		memlanesCall(t, title, reg.handleLoadDeploymentContext, access, map[string]any{"title": title, "content": "MemLanes echidna note.", "knowledge_class": "customer_context", "visibility": "private"})
		var id, owner, loadedBy string
		if err := w.db.QueryRow(`SELECT id::text, COALESCE(metadata->>'owner_user_id',''), COALESCE(metadata->>'loaded_by','') FROM artifacts WHERE title = $1`, title).Scan(&id, &owner, &loadedBy); err != nil {
			t.Fatalf("read %s: %v", title, err)
		}
		if access.User && (owner != access.Reader.UserID || loadedBy != access.Reader.Label) {
			t.Errorf("%s: owner=%q loaded_by=%q, want the requesting user", title, owner, loadedBy)
		}
		return id
	}
	adaID := save(w.accA, "MemLanes Ada Echidna")
	systemID := save(RecallAccess{}, "MemLanes System Echidna")
	soma := memory.GovernedReader{UserID: uuid.NewString(), Label: "soma"}
	for name, tc := range map[string]struct {
		id     string
		reader memory.GovernedReader
		want   bool
	}{
		"A reads own": {adaID, w.accA.Reader, true}, "B cannot read A's": {adaID, w.accB.Reader, false},
		"identity named soma cannot read a no-user save": {systemID, soma, false},
	} {
		readable, err := memory.GovernedEntryReadable(context.Background(), w.db, tc.id, tc.reader)
		if err != nil || readable != tc.want {
			t.Errorf("%s: readable=%t err=%v", name, readable, err)
		}
	}
}
