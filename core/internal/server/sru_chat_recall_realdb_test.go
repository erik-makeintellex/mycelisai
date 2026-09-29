package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/searchcap"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// SRU (HIGH): the real Soma chat path. POST /api/v1/chat as a signed-in user
// reaches the in-process Soma agent over NATS, and ambient recall,
// context_sources and the reply follow that user's MEM-LIST read rule.
// Real PostgreSQL (MYCELIS_MEMORY_TEST_DSN) and an embedded NATS server.

type sruChatProvider struct {
	mu      sync.Mutex
	prompts []string
}

func (p *sruChatProvider) Infer(_ context.Context, prompt string, opts cognitive.InferOptions) (*cognitive.InferResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, message := range opts.Messages {
		prompt += "\n" + message.Content
	}
	p.prompts = append(p.prompts, prompt)
	return &cognitive.InferResponse{Text: "Here is the quokka weekend plan.", Provider: "vllm", ModelUsed: "qwen-coder"}, nil
}
func (p *sruChatProvider) Probe(context.Context) (bool, error) { return true, nil }
func (p *sruChatProvider) Embed(context.Context, string, string) ([]float64, error) {
	return nil, context.Canceled
}

// take returns and clears the prompts seen so far.
func (p *sruChatProvider) take() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := strings.Join(p.prompts, "\n")
	p.prompts = nil
	return out
}

var sruChatText = map[string]string{
	"ada-private":    "Quokka weekend plan: Ada private ledger balance 4410.",
	"ada-reflection": "Quokka weekend plan reflection: Ada prefers dawn deliveries.",
	"team":           "Quokka weekend plan for the admin-core team: rota seven.",
	"org":            "Quokka weekend plan: the shop opens at nine.",
	"bob-private":    "Quokka weekend plan: Bob private reminder eighty-eight.",
}

type sruChatWorld struct {
	s                  *AdminServer
	h                  http.Handler
	provider           *sruChatProvider
	ada, bob, cy, root *RequestIdentity
	adaSource          string
}

func sruChatSeed(t *testing.T) *sruChatWorld {
	t.Helper()
	db := openMemoryServerTestDB(t)
	provider := &sruChatProvider{}
	router := &cognitive.Router{
		Config:   &cognitive.BrainConfig{Profiles: map[string]string{"chat": "vllm"}, Providers: map[string]cognitive.ProviderConfig{"vllm": {Type: "mock", Enabled: true, ModelID: "qwen-coder"}}},
		Adapters: map[string]cognitive.LLMProvider{"vllm": provider},
	}
	s := memoryTestServer(db, router)
	withNATS(t)(s)
	run := uuid.NewString()[:8]
	_, _ = db.Exec(`DELETE FROM runtime_team_manifests WHERE team_id = 'admin-core'`)
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM runtime_team_manifests WHERE team_id = 'admin-core'`) })
	// ada and cy are proven members of admin-core; bob is not.
	ids := b1rcSeedTenantTeam(t, db, "default", "sru-a-"+run, "admin-core")
	var cyID string
	if err := db.QueryRow(`INSERT INTO users (id, username, account_id) VALUES (gen_random_uuid(), $1, $2) RETURNING id`, "u-sru-c-"+run, ids.account).Scan(&cyID); err != nil {
		t.Fatalf("seed cy: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO org_memberships (account_id, user_id, group_id, role_id) VALUES ($1,$2,$3,$4)`, ids.account, cyID, ids.group, ids.role); err != nil {
		t.Fatalf("seed cy membership: %v", err)
	}
	w := &sruChatWorld{s: s, provider: provider,
		ada:  memoryUser(ids.user, "u-sru-a-"+run, "operator", "memory:read"),
		bob:  memoryUser(uuid.NewString(), "bob-sru-"+run, "operator", "memory:read"),
		cy:   memoryUser(cyID, "u-sru-c-"+run, "operator", "memory:read"),
		root: memoryUser(uuid.NewString(), "root-sru-"+run, "admin", "memory:write")}
	mux := http.NewServeMux()
	mux.Handle("/api/v1/memory/", lifecycleMux(s))
	mux.HandleFunc("/api/v1/chat", s.HandleChat)
	w.h = mux

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	soma := swarm.NewAgent(ctx, protocol.AgentManifest{ID: "admin", Role: "admin", Provider: "vllm"}, "admin-core", s.NC, router, nil)
	soma.SetInternalTools(swarm.NewInternalToolRegistry(swarm.InternalToolDeps{Brain: router, Mem: s.Mem, DB: db}))
	soma.Start()

	saveAs(t, w.h, w.ada, `{"title":"SRU Ada Private Ledger","knowledge_class":"user_private_context","content":"`+sruChatText["ada-private"]+`"}`)
	saveAs(t, w.h, w.ada, `{"title":"SRU Ada Reflection","knowledge_class":"reflection_synthesis","content":"`+sruChatText["ada-reflection"]+`"}`)
	saveAs(t, w.h, w.ada, `{"title":"SRU Admin Core Team Note","visibility":"team","team_id":"admin-core","content":"`+sruChatText["team"]+`"}`)
	saveAs(t, w.h, w.ada, `{"title":"SRU Org Quokka Hours","knowledge_class":"company_knowledge","visibility":"global","content":"`+sruChatText["org"]+`"}`)
	saveAs(t, w.h, w.bob, `{"title":"SRU Bob Private Note","knowledge_class":"user_private_context","content":"`+sruChatText["bob-private"]+`"}`)
	w.adaSource = saveAs(t, w.h, w.ada, `{"title":"SRU Ada Promote Source","visibility":"private","content":"Quokka supplier contract terms for Ada only."}`)
	return w
}

// chat posts one Soma chat turn as who and returns the reply payload.
func (w *sruChatWorld) chat(t *testing.T, who *RequestIdentity) protocol.ChatResponsePayload {
	t.Helper()
	rr := doAuthenticatedRequestAs(t, w.h, http.MethodPost, "/api/v1/chat", `{"messages":[{"role":"user","content":"What is our quokka weekend plan?"}]}`, who)
	assertStatus(t, rr, http.StatusOK)
	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	raw, _ := json.Marshal(resp.Data)
	var envelope protocol.CTSEnvelope
	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(raw, &envelope); err != nil || json.Unmarshal(envelope.Payload, &payload) != nil {
		t.Fatalf("decode chat reply: %s", rr.Body.String())
	}
	return payload
}

func sruChatCheck(t *testing.T, label, text string, want, deny []string) {
	t.Helper()
	for _, key := range want {
		if !strings.Contains(text, sruChatText[key]) {
			t.Errorf("%s: missing %s entry", label, key)
		}
	}
	for _, key := range deny {
		if strings.Contains(text, sruChatText[key]) {
			t.Errorf("%s: leaked %s entry", label, key)
		}
	}
}

func sruChatTitles(payload protocol.ChatResponsePayload) string {
	titles := []string{}
	for _, source := range payload.ContextSources {
		titles = append(titles, source.Title)
	}
	return strings.Join(titles, "|")
}

func TestSRUChatRealDB_EachUserRecallsOnlyWhatTheyMayRead(t *testing.T) {
	w := sruChatSeed(t)
	for _, tc := range []struct {
		name       string
		who        *RequestIdentity
		want, deny []string
	}{
		{"bob", w.bob, []string{"org", "bob-private"}, []string{"ada-private", "ada-reflection", "team"}},
		{"ada", w.ada, []string{"org", "ada-private", "ada-reflection", "team"}, []string{"bob-private"}},
		{"cy (team member)", w.cy, []string{"org", "team"}, []string{"ada-private", "ada-reflection", "bob-private"}},
		{"root admin", w.root, []string{"org"}, []string{"ada-private", "ada-reflection", "team", "bob-private"}},
	} {
		w.provider.take()
		payload := w.chat(t, tc.who)
		sruChatCheck(t, tc.name+" prompt", w.provider.take(), tc.want, tc.deny)
		titles := sruChatTitles(payload)
		for _, key := range tc.deny {
			if title := sruChatTitle[key]; strings.Contains(titles, title) {
				t.Errorf("%s context_sources carried %q: %s", tc.name, title, titles)
			}
		}
		for _, key := range tc.want {
			if title := sruChatTitle[key]; !strings.Contains(titles, title) {
				t.Errorf("%s context_sources missed %q: %s", tc.name, title, titles)
			}
		}
	}
}

var sruChatTitle = map[string]string{"ada-private": "SRU Ada Private Ledger", "ada-reflection": "SRU Ada Reflection",
	"team": "SRU Admin Core Team Note", "org": "SRU Org Quokka Hours", "bob-private": "SRU Bob Private Note"}

// A NATS publisher without Core's turn token gets no user's scope: no token
// reads org-wide only, and a forged token withholds the governed lane.
func TestSRUChatRealDB_DirectRequestsWithoutCoreTokenGetNoUserScope(t *testing.T) {
	w := sruChatSeed(t)
	subject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, "admin")
	payload := []byte(`[{"role":"user","content":"What is our quokka weekend plan?"}]`)
	send := func(token string) swarm.ProcessResult {
		msg := nats.NewMsg(subject)
		msg.Data = payload
		if token != "" {
			msg.Header.Set(swarm.RecallTurnHeader, token)
		}
		reply, err := w.s.NC.RequestMsg(msg, 20*time.Second)
		if err != nil {
			t.Fatalf("direct request: %v", err)
		}
		var result swarm.ProcessResult
		if err := json.Unmarshal(reply.Data, &result); err != nil {
			t.Fatalf("decode: %s", reply.Data)
		}
		return result
	}
	w.provider.take()
	send("")
	sruChatCheck(t, "no token", w.provider.take(), []string{"org"}, []string{"ada-private", "ada-reflection", "team", "bob-private"})
	forged := send("forged-token")
	sruChatCheck(t, "forged token", w.provider.take(), nil, []string{"org", "ada-private", "ada-reflection", "team", "bob-private"})
	if len(forged.ContextSources) != 0 || !strings.Contains(forged.Text, "Saved memory context was unavailable") {
		t.Fatalf("a forged token must withhold memory and say so: %+v", forged)
	}
	// A real token is single use: replaying it, even on its own subject and
	// reply inbox (MEM-LANES binding), is a forged token.
	inbox := nats.NewInbox()
	token, release := swarm.RegisterRecallTurn(swarm.RecallAccess{User: true}, subject, inbox, time.Minute)
	defer release()
	bound := func() swarm.ProcessResult {
		msg := nats.NewMsg(subject)
		msg.Data, msg.Reply = payload, inbox
		msg.Header.Set(swarm.RecallTurnHeader, token)
		return memlanesDecode(t, "bound", memlanesSpend(t, w.s.NC, msg))
	}
	if first := bound(); strings.Contains(first.Text, "Saved memory context was unavailable") {
		t.Fatalf("the bound token must be honoured once: %+v", first)
	}
	w.provider.take()
	if replay := bound(); !strings.Contains(replay.Text, "Saved memory context was unavailable") {
		t.Fatalf("a replayed token must not be honoured: %+v", replay)
	}
}

// The confirm-action path scopes promote to the confirming user.
func TestSRUConfirmedPromoteRealDB_FollowsTheConfirmingUser(t *testing.T) {
	w := sruChatSeed(t)
	registry := swarm.NewInternalToolRegistry(swarm.InternalToolDeps{Brain: w.s.Cognitive, Mem: w.s.Mem, DB: w.s.DB})
	executor := swarm.NewCompositeToolExecutor(registry, nil)
	promote := func(who *RequestIdentity) (string, error) {
		ctx := context.Background()
		if who != nil {
			ctx = context.WithValue(ctx, ctxKeyIdentity, who)
		}
		toolCtx := w.s.withConfirmedRecallAccess(confirmedActionToolContext(ctx, "confirming-user", "", nil))
		return executor.CallTool(toolCtx, swarm.InternalServerID, "promote_deployment_context", map[string]any{"source_artifact_id": w.adaSource, "title": "SRU promoted"})
	}
	for name, who := range map[string]*RequestIdentity{"bob": w.bob, "cy": w.cy, "root admin": w.root, "no identity": nil} {
		if out, err := promote(who); err == nil || !strings.Contains(err.Error(), "Nothing was promoted") {
			t.Fatalf("promote of ada's private entry confirmed by %s must be refused: %s err=%v", name, out, err)
		}
	}
	if out, err := promote(w.ada); err != nil || !strings.Contains(out, "company_knowledge") {
		t.Fatalf("ada promotes her own entry: %s err=%v", out, err)
	}
}

// POST /api/v1/search over local sources reads the same governed store; its
// team_id and agent_id can only narrow what the caller may read.
func TestSRUSearchRouteRealDB_LocalSourcesFollowTheCaller(t *testing.T) {
	w := sruChatSeed(t)
	w.s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderLocalSources}, nil, w.s.Mem)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/search", w.s.HandleSearch)
	body := `{"query":"quokka weekend plan","source_scope":"local_sources","max_results":10,"team_id":"admin-core","agent_id":"admin"}`
	for _, tc := range []struct {
		name       string
		who        *RequestIdentity
		want, deny []string
	}{
		{"bob", w.bob, []string{"org", "bob-private"}, []string{"ada-private", "ada-reflection", "team"}},
		{"ada", w.ada, []string{"org", "ada-private", "team"}, []string{"bob-private"}},
		{"no identity", nil, []string{"org"}, []string{"ada-private", "ada-reflection", "team", "bob-private"}},
	} {
		rr := doRequest(t, mux, http.MethodPost, "/api/v1/search", body)
		if tc.who != nil {
			rr = doAuthenticatedRequestAs(t, mux, http.MethodPost, "/api/v1/search", body, tc.who)
		}
		assertStatus(t, rr, http.StatusOK)
		sruChatCheck(t, tc.name+" /api/v1/search", rr.Body.String(), tc.want, tc.deny)
	}
}
