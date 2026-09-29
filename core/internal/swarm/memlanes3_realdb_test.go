package swarm

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
)

// MEM-LANES-3 (MEMLANES-QA round 2). Real PostgreSQL
// (MYCELIS_MEMORY_TEST_DSN). Markers start with "MemLanes" so memlanesSeed
// cleans the vector, temp and memory rows.

func memlanes3Registry(t *testing.T, w *memlanesWorld, router *cognitive.Router) *InternalToolRegistry {
	t.Helper()
	t.Cleanup(func() {
		_, _ = w.db.Exec(`DELETE FROM artifacts WHERE title LIKE 'MemLanes3%'`)
		_, _ = w.db.Exec(`DELETE FROM inception_recipes WHERE title LIKE 'MemLanes3%'`)
	})
	return NewInternalToolRegistry(InternalToolDeps{Brain: router, Mem: memory.NewServiceWithDB(w.db), DB: w.db, Inception: inception.NewStore(w.db)})
}

// Item 1: load_deployment_context governs visibility like remember. The
// model's customer_context default (global), "global", "GLOBAL" and "public"
// stay private for an ordinary user; org-wide needs root admin with
// memory:write, a team share needs proven membership.
func TestMemLanes3RealDB_DeploymentContextVisibilityIsGoverned(t *testing.T) {
	w := memlanesSeed(t)
	router, _ := memlanesRouter("ok", true)
	reg := memlanes3Registry(t, w, router)
	save := func(label string, access RecallAccess, marker string, vis any) string {
		t.Helper()
		args := map[string]any{"title": "MemLanes3 " + label, "content": "MemLanes wallaby customer note " + marker + ".", "knowledge_class": "customer_context"}
		if vis != nil {
			args["visibility"] = vis
		}
		return memlanesCall(t, label, reg.handleLoadDeploymentContext, access, args)
	}
	for i, vis := range []any{nil, "global", "GLOBAL", "public"} {
		out := save(fmt.Sprintf("Ada %d", i), w.accA, fmt.Sprintf("830%d", i), vis)
		if !strings.Contains(out, "Kept private") || !strings.Contains(out, `"visibility":"private"`) {
			t.Errorf("A customer_context %v must be kept private and say so: %s", vis, out)
		}
	}
	if out := save("Bob Team", w.accB, "8310", "team"); !strings.Contains(out, "Kept private") {
		t.Errorf("a non-member's team share must be kept private: %s", out)
	}
	save("Ada Team", w.accA, "8311", "team")
	root := RecallAccess{User: true, OrgWideWrite: true, Reader: memory.GovernedReader{UserID: uuid.NewString(), Label: "root-lanes3"}}
	if out := save("Root Org", root, "8312", "global"); strings.Contains(out, "Kept private") {
		t.Errorf("a root admin with memory:write may share org-wide: %s", out)
	}
	for _, tc := range []struct {
		name       string
		access     RecallAccess
		want, deny []string
	}{
		{"A", w.accA, []string{"8300", "8301", "8302", "8303", "8311", "8312"}, []string{"8310"}},
		{"B", w.accB, []string{"8310", "8312"}, []string{"8300", "8301", "8302", "8303", "8311"}},
		{"team member C", w.accC, []string{"8311", "8312"}, []string{"8300", "8301", "8302", "8303", "8310"}},
		{"no user", RecallAccess{}, []string{"8312"}, []string{"8300", "8301", "8302", "8303", "8310", "8311"}},
	} {
		memlanesCheck(t, tc.name, memlanes2Recall(t, reg, tc.name, tc.access, "wallaby customer note"), tc.want, tc.deny)
	}
}

// Item 2: a model tool with no user writes no temp channel at all; only
// Core's writers (the bus checkpoint, AutoSummarize) write system rows, and
// only rows they wrote reach a no-user prompt or a user's signal read.
func TestMemLanes3RealDB_NoUserModelToolWritesNoChannel(t *testing.T) {
	w := memlanesSeed(t)
	_, nc := startTestNATS(t)
	router, _ := memlanesRouter("MemLanes3 system continuity 7390.", false)
	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc, Brain: router, Mem: memory.NewServiceWithDB(w.db), DB: w.db})
	noUser := sruToolCtx(RecallAccess{})
	for _, c := range []struct{ ch, marker string }{
		{"lead.council-memlanes3", "7306"}, {"team.council-core.planning", "7307"}, {"Lead.Shared", "7305"}, {" lead.shared ", "7303"},
		{"SIGNAL.LATEST.memlanes3.x", "7308"}, {"interaction.contract", "7304"}, {"scratch.memlanes3", "7309"},
	} {
		if _, err := reg.handleTempMemoryWrite(noUser, map[string]any{"channel": c.ch, "content": "MemLanes3 no-user " + c.marker + "."}); err == nil {
			t.Errorf("a no-user model write to %q was accepted", c.ch)
		}
	}
	if _, err := reg.handlePublishSignal(noUser, map[string]any{"subject": "swarm.team.memlanes3.signal.status", "message": "MemLanes3 no-user private 7311.", "privacy_mode": "reference", "channel_key": "team.memlanes3.private"}); err == nil {
		t.Error("a no-user private-reference publish must be refused: its payload could not be kept")
	}
	// Legacy model-written system rows (no Core writer mark) reach nobody.
	if _, err := w.db.Exec(`INSERT INTO temp_memory_channels (tenant_id, channel_key, owner_agent_id, content, metadata) VALUES
		('default', 'lead.council-memlanes3', 'council-memlanes3', 'MemLanes3 legacy model row 7320.', '{"owner_class":"system"}'),
		('default', 'signal.latest.swarm.team.memlanes3.signal.status', 'x', 'MemLanes3 legacy signal 7321.', '{"owner_class":"system"}')`); err != nil {
		t.Fatalf("seed legacy rows: %v", err)
	}
	var stored int
	_ = w.db.QueryRow(`SELECT count(*) FROM temp_memory_channels WHERE content LIKE 'MemLanes3 no-user%'`).Scan(&stored)
	if stored != 0 {
		t.Errorf("%d no-user model rows were stored", stored)
	}
	// Core's AutoSummarize is a system writer: its no-user checkpoint reaches the no-user lead.
	history := make([]cognitive.ChatMessage, 15)
	for i := range history {
		history[i] = cognitive.ChatMessage{Role: "user", Content: "status"}
	}
	reg.AutoSummarize(context.Background(), RecallAccess{}, "council-memlanes3", "", history)
	prompter, provider := memlanesRouter("ok", false)
	council := NewAgent(context.Background(), protocol.AgentManifest{ID: "council-memlanes3", Role: "architect", Provider: "vllm"}, "council-core", nil, prompter, &countingToolExecutor{serverID: InternalServerID})
	council.SetInternalTools(NewInternalToolRegistry(InternalToolDeps{Brain: prompter, Mem: memory.NewServiceWithDB(w.db), DB: w.db}))
	council.processMessageStructured("What's next?", nil)
	memlanesCheck(t, "no-user lead prompt", provider.allPrompts(), []string{"7390"}, []string{"7303", "7304", "7305", "7306", "7307", "7309", "7320"})
	sig := memlanesCall(t, "B signals", reg.handleReadSignals, w.accB, map[string]any{"latest_only": true, "channel_key": "signal.latest.swarm.team.memlanes3.signal.status"})
	memlanesCheck(t, "B signal read", sig, nil, []string{"7321"})
}

// Items 4 and 5: store_inception_recipe works on the real schema, and a
// recipe follows its owner like a remembered fact.
func TestMemLanes3RealDB_InceptionRecipesFollowTheRequestingUser(t *testing.T) {
	w := memlanesSeed(t)
	router, _ := memlanesRouter("ok", false)
	reg := memlanes3Registry(t, w, router)
	store := func(label string, access RecallAccess, marker string, vis any) string {
		t.Helper()
		args := map[string]any{"category": "memlanes3", "title": "MemLanes3 quokka recipe " + marker, "intent_pattern": "MemLanes quokka medical history " + marker}
		if vis != nil {
			args["visibility"] = vis
		}
		return memlanesCall(t, label, reg.handleStoreInceptionRecipe, access, args)
	}
	store("A recipe", w.accA, "8401", nil)
	if out := store("A global recipe", w.accA, "8402", "global"); !strings.Contains(out, "Kept private") {
		t.Errorf("A's org-wide recipe must be kept private and say so: %s", out)
	}
	root := RecallAccess{User: true, OrgWideWrite: true, Reader: memory.GovernedReader{UserID: uuid.NewString(), Label: "root-lanes3"}}
	store("root recipe", root, "8403", "global")
	for _, tc := range []struct {
		name       string
		access     RecallAccess
		want, deny []string
	}{
		{"A", w.accA, []string{"8401", "8402", "8403"}, nil},
		{"B", w.accB, []string{"8403"}, []string{"8401", "8402"}},
		{"no user", RecallAccess{}, []string{"8403"}, []string{"8401", "8402"}},
	} {
		out := memlanesCall(t, tc.name, reg.handleRecallInceptionRecipes, tc.access, map[string]any{"query": "quokka", "limit": float64(20)})
		memlanesCheck(t, tc.name+" recipes", out, tc.want, tc.deny)
	}
	// The store keeps a real source run and session id (uuid columns).
	if _, err := inception.NewStore(w.db).CreateRecipe(context.Background(), inception.Recipe{Category: "memlanes3", Title: "MemLanes3 sourced recipe",
		IntentPattern: "MemLanes sourced", SourceRunID: uuid.NewString(), SourceSessionID: uuid.NewString()}); err != nil {
		t.Fatalf("store with source ids: %v", err)
	}
}
