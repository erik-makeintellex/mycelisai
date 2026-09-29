package swarm

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
)

// MEM-LANES-2 (MEMLANES-QA C1, C2 and the LOWs). Real PostgreSQL
// (MYCELIS_MEMORY_TEST_DSN). Markers start with "MemLanes" so memlanesSeed
// cleans them up.

func memlanes2Recall(t *testing.T, reg *InternalToolRegistry, label string, access RecallAccess, query string) string {
	t.Helper()
	args := func() map[string]any { return map[string]any{"query": query, "limit": float64(20)} }
	recall := memlanesCall(t, label+" recall", reg.handleRecall, access, args())
	search := memlanesCall(t, label+" search", reg.handleSearchMemory, access, args())
	convo := memlanesCall(t, label+" search conversations", reg.handleSearchMemory, access,
		map[string]any{"query": query, "types": []any{"conversation"}, "limit": float64(20)})
	return recall + "\n" + search + "\n" + convo
}

// C1: during a user turn the model's visibility is a request, not authority.
// Org-wide needs the M2 authority (root admin with memory:write); a team
// share needs proven membership of that team; otherwise it stays private and
// the tool result says so.
func TestMemLanes2RealDB_ModelChosenGlobal(t *testing.T) {
	w := memlanesSeed(t)
	router, _ := memlanesRouter("MemLanes2 okapi summary 7201.", true)
	reg := w.registry(router)
	const aGlobal, bTeam, rootGlobal = "MemLanes2 okapi ada global 7101.", "MemLanes2 okapi bob team 7102.", "MemLanes2 okapi org fact 7103."

	clamped := func(label, out string) {
		t.Helper()
		if !strings.Contains(out, "private") {
			t.Errorf("%s: the result must say the entry was kept private: %q", label, out)
		}
	}
	clamped("A remember global", memlanesCall(t, "A remember", reg.handleRemember, w.accA,
		map[string]any{"category": "fact", "content": aGlobal, "visibility": "global", "org_wide_write": true, "owner_user_id": w.userB}))
	// B is not a proven member of admin-core, so a team share stays private.
	clamped("B remember team", memlanesCall(t, "B remember", reg.handleRemember, w.accB,
		map[string]any{"category": "fact", "content": bTeam, "visibility": "team"}))
	clamped("A summarize global", memlanesCall(t, "A summarize", reg.handleSummarizeConversation, w.accA,
		map[string]any{"messages": "user: plan the okapi trip", "visibility": "global"}))

	root := RecallAccess{User: true, OrgWideWrite: true, Reader: memory.GovernedReader{UserID: uuid.NewString(), Label: "root-lanes2"}}
	if out := memlanesCall(t, "root remember", reg.handleRemember, root,
		map[string]any{"category": "fact", "content": rootGlobal, "visibility": "global"}); strings.Contains(out, "Kept private") {
		t.Errorf("a root admin with memory:write may share org-wide: %q", out)
	}

	for _, tc := range []struct {
		name       string
		access     RecallAccess
		want, deny []string
	}{
		{"A", w.accA, []string{"7101", "7201", "7103"}, []string{"7102"}},
		{"B", w.accB, []string{"7102", "7103"}, []string{"7101", "7201"}},
		{"team member C", w.accC, []string{"7103"}, []string{"7101", "7102", "7201"}},
		{"no user", RecallAccess{}, []string{"7103"}, []string{"7101", "7102", "7201"}},
	} {
		memlanesCheck(t, tc.name, memlanes2Recall(t, reg, tc.name, tc.access, "okapi"), tc.want, tc.deny)
	}
}

// C2: a write with no user on the shared channels (signal.latest.*, and
// lead.shared) comes only from Core's bus writer, never from a model tool,
// so nothing a model writes there reaches another user or a no-user prompt.
func TestMemLanes2RealDB_NoUserWritesReachOtherTurns(t *testing.T) {
	w := memlanesSeed(t)
	_, nc := startTestNATS(t)
	router, _ := memlanesRouter("ok", false)
	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc, Brain: router, Mem: memory.NewServiceWithDB(w.db), DB: w.db})
	const signalCh = "signal.latest.swarm.team.memlanes2.signal.status"
	noUser := sruToolCtx(RecallAccess{})

	if _, err := reg.handleTempMemoryWrite(noUser, map[string]any{"channel": "lead.shared", "content": "MemLanes2 no-user lead 7302."}); err == nil {
		t.Error("a no-user model write to lead.shared must be refused")
	}
	if _, err := reg.handleTempMemoryWrite(noUser, map[string]any{"channel": signalCh, "content": "MemLanes2 no-user signal 7301."}); err == nil {
		t.Error("a no-user model write to a signal checkpoint channel must be refused")
	}
	for _, args := range []map[string]any{
		{"subject": "swarm.team.memlanes2.signal.status", "message": "MemLanes2 no-user published 7303."},
		{"subject": "swarm.team.memlanes2.signal.status", "message": "MemLanes2 no-user lead via signal 7304.", "channel_key": "lead.shared"},
	} {
		out, err := reg.handlePublishSignal(noUser, args)
		if err != nil {
			t.Fatalf("publish_signal: %v", err)
		}
		if strings.Contains(out, "Latest channel checkpoint") || !strings.Contains(out, "not stored") {
			t.Errorf("a no-user publish must say its checkpoint was not stored: %q", out)
		}
	}

	var stored int
	_ = w.db.QueryRow(`SELECT count(*) FROM temp_memory_channels WHERE content LIKE '%MemLanes2 no-user%'`).Scan(&stored)
	if stored != 0 {
		t.Errorf("%d no-user model rows were stored on shared channels", stored)
	}
	signalOut := memlanesCall(t, "B read_signals", reg.handleReadSignals, w.accB, map[string]any{"latest_only": true, "channel_key": signalCh})
	memlanesCheck(t, "B read_signals", signalOut, nil, []string{"7301", "7303"})
	leadRouter, provider := memlanesRouter("ok", false)
	somaAgent(leadRouter, w.db).processMessageStructured("What is next for the okapi launch?", nil)
	memlanesCheck(t, "no-user lead prompt", provider.allPrompts(), nil, []string{"7302", "7304"})
}

// C2: a council consult made during A's turn runs under A's access, so the
// council's reads and writes are A's; B's consult and a no-user consult are
// not A's.
func TestMemLanes2RealDB_CouncilConsultCarriesTheTurnUser(t *testing.T) {
	w := memlanesSeed(t)
	_, nc := startTestNATS(t)
	router, _ := memlanesRouter("ok", false)
	mem := memory.NewServiceWithDB(w.db)
	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc, Brain: router, Mem: mem, DB: w.db})
	memlanesCall(t, "A temp write", reg.handleTempMemoryWrite, w.accA, map[string]any{"channel": "lead.shared", "content": "MemLanes2 ada checkpoint 7310."})

	councilRouter, provider := memlanesRouter("Council answer.", false)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	council := NewAgent(ctx, protocol.AgentManifest{ID: "council-memlanes2", Role: "architect", Provider: "vllm"}, "council-core", nc, councilRouter, nil)
	council.SetInternalTools(NewInternalToolRegistry(InternalToolDeps{Brain: councilRouter, Mem: mem, DB: w.db}))
	council.Start()
	_ = nc.Flush()

	for _, tc := range []struct {
		name   string
		access RecallAccess
		sees   bool
	}{{"A", w.accA, true}, {"B", w.accB, false}, {"no user", RecallAccess{}, false}} {
		provider.mu.Lock()
		provider.prompts = nil
		provider.mu.Unlock()
		out := memlanesCall(t, tc.name+" consult", reg.handleConsultCouncil, tc.access, map[string]any{"member": "council-memlanes2", "question": "What is next?"})
		if !strings.Contains(out, "Council answer.") {
			t.Fatalf("%s consult: %q", tc.name, out)
		}
		if tc.sees {
			memlanesCheck(t, tc.name+" council prompt", provider.allPrompts(), []string{"7310"}, nil)
		} else {
			memlanesCheck(t, tc.name+" council prompt", provider.allPrompts(), nil, []string{"7310"})
		}
	}
}

// LOW: a memory tool call with no invocation context is not org-wide.
func TestMemLanes2RealDB_NoInvocationContextDefaultsGlobal(t *testing.T) {
	if got := resolveMemoryScope(context.Background(), map[string]any{}).Visibility; got == "global" {
		t.Errorf("no invocation context defaulted to %q", got)
	}
	w := memlanesSeed(t)
	router, _ := memlanesRouter("ok", false)
	reg := w.registry(router)
	for _, args := range []map[string]any{
		{"category": "fact", "content": "MemLanes2 tapir no-context 7401."},
		{"category": "fact", "content": "MemLanes2 tapir no-context global 7402.", "visibility": "global"},
	} {
		if _, err := reg.handleRemember(context.Background(), args); err != nil {
			t.Fatalf("remember: %v", err)
		}
	}
	for name, access := range map[string]RecallAccess{"A": w.accA, "B": w.accB, "no user": {}} {
		memlanesCheck(t, name, memlanes2Recall(t, reg, name, access, "tapir"), nil, []string{"7401", "7402"})
	}
}

// C2: Core's bus checkpoint of an MCP tool call made in A's turn holds A's
// tool arguments and result, so it is A's; with no user it stays system bus
// state every user can read; an unverified turn keeps none.
func TestMemLanes2RealDB_BusCheckpointFollowsTheTurn(t *testing.T) {
	w := memlanesSeed(t)
	_, nc := startTestNATS(t)
	router, _ := memlanesRouter("ok", false)
	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc, Brain: router, Mem: memory.NewServiceWithDB(w.db), DB: w.db})
	agent := NewAgent(context.Background(), protocol.AgentManifest{ID: "memlanes2-worker", Role: "worker"}, "memlanes2-bus", nc, router, nil)
	agent.SetInternalTools(reg)
	const ch = "signal.latest.swarm.team.memlanes2-bus.signal.status"
	read := func(label string, access RecallAccess) string {
		return memlanesCall(t, label, reg.handleReadSignals, access, map[string]any{"latest_only": true, "channel_key": ch})
	}
	bus := func(access RecallAccess, marker string) {
		agent.publishToolBusSignal(access, protocol.PayloadKindStatus, protocol.SourceKindMCP, map[string]any{"state": "invoked", "tool": marker})
	}
	bus(RecallAccess{}, "MemLanes2 bus system 7601")
	memlanesCheck(t, "B reads system bus state", read("B", w.accB), []string{"7601"}, nil)
	bus(w.accA, "MemLanes2 bus ada 7602")
	memlanesCheck(t, "A reads own bus checkpoint", read("A", w.accA), []string{"7602"}, nil)
	memlanesCheck(t, "B after A's turn", read("B", w.accB), []string{"7601"}, []string{"7602"})
	bus(RecallAccess{Unavailable: true}, "MemLanes2 bus unverified 7603")
	var n int
	_ = w.db.QueryRow(`SELECT count(*) FROM temp_memory_channels WHERE content LIKE '%7603%'`).Scan(&n)
	if n != 0 {
		t.Errorf("an unverified turn stored %d bus checkpoints", n)
	}
}
