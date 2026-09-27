package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// draftUsageProvider drafts real text and reports exact usage per call.
type draftUsageProvider struct {
	mu    sync.Mutex
	used  int
	calls int
	corr  []cognitive.InferenceCorrelation
}

func (p *draftUsageProvider) Infer(_ context.Context, _ string, opts cognitive.InferOptions) (*cognitive.InferResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	p.corr = append(p.corr, opts.Correlation)
	return &cognitive.InferResponse{Text: teamBriefDraft, ModelUsed: "m", PromptTokens: p.used - 200, CompletionTokens: 200, TokensUsed: p.used}, nil
}

func (p *draftUsageProvider) Probe(context.Context) (bool, error) { return true, nil }

func draftBudgetRouter(provider cognitive.LLMProvider, perExecution int) *cognitive.Router {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Agent = map[string]protocol.TokenBudgetLimits{"admin": {PerExecution: perExecution}}
	return &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles:  map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{"mock": {Type: "mock", Enabled: true, ModelID: "test-model", MaxOutputTokens: 2048}},
		},
		Adapters: map[string]cognitive.LLMProvider{"mock": provider},
		Budgets:  cognitive.NewBudgetGovernor(spec, nil),
	}
}

func plannedDrafts(n int) []protocol.PlannedToolCall {
	calls := make([]protocol.PlannedToolCall, 0, n)
	for i := 0; i < n; i++ {
		calls = append(calls, protocol.PlannedToolCall{Name: "write_file", Arguments: map[string]any{"path": "notes/part-" + string(rune('a'+i)) + ".md"}})
	}
	return calls
}

func TestDraftMissingWriteFileContent_BudgetPreflightRefusesWithoutInference(t *testing.T) {
	provider := &draftUsageProvider{used: 1000}
	s := newTestServer()
	s.Cognitive = draftBudgetRouter(provider, 5000)
	planned := plannedDrafts(4)
	previews, blocker := s.draftMissingWriteFileContent(t.Context(), planned, teamBriefRequest)
	if blocker == nil || previews != nil || blocker.Budget == nil || blocker.Availability.Code != cognitive.TokenBudgetExhaustedCode {
		t.Fatalf("blocker=%+v previews=%+v, want token budget preflight refusal", blocker, previews)
	}
	if !strings.Contains(blocker.Availability.Summary, "can draft 2 of 4") || blocker.Status != http.StatusTooManyRequests {
		t.Fatalf("blocker = %+v, want 429 'can draft 2 of 4'", blocker)
	}
	if provider.calls != 0 {
		t.Fatalf("draft inferences = %d, want none", provider.calls)
	}
}

func TestDraftMissingWriteFileContent_MidPassExhaustionProposesNothing(t *testing.T) {
	provider := &draftUsageProvider{used: 3000}
	s := newTestServer()
	s.Cognitive = draftBudgetRouter(provider, 6144)
	planned := plannedDrafts(3)
	previews, blocker := s.draftMissingWriteFileContent(t.Context(), planned, teamBriefRequest)
	if blocker == nil || previews != nil || blocker.Budget == nil || blocker.Budget.Scope != protocol.TokenBudgetScopeExecution {
		t.Fatalf("blocker=%+v previews=%+v, want mid-pass execution stop", blocker, previews)
	}
	if provider.calls != 2 {
		t.Fatalf("draft inferences = %d, want 2 (third refused below the floor)", provider.calls)
	}
	for _, call := range planned {
		if _, has := call.Arguments["content"]; has {
			t.Fatalf("partial draft applied to %v", call.Arguments["path"])
		}
	}
	for _, corr := range provider.corr {
		if corr.AgentID != "admin" {
			t.Fatalf("draft correlation = %+v, want agent admin", corr)
		}
	}
}

func TestHandleChat_DraftBudgetBlockerIsHonest429(t *testing.T) {
	provider := &draftUsageProvider{used: 1000}
	s := newTestServer(withNATS(t))
	s.Cognitive = draftBudgetRouter(provider, 5000)
	respondAsAdminAgentForTest(t, s, map[string]any{"text": "Planning the notes.", "tools_used": []string{"write_file"}, "planned_tool_calls": plannedDrafts(5)})
	body, _ := json.Marshal(map[string]any{"messages": []chatRequestMessage{{Role: "user", Content: "Create five markdown notes for the launch."}}})
	for _, identity := range []*RequestIdentity{{UserID: "u-1", Role: "user"}, localAdminIdentityForTest()} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBuffer(body))
		req = req.WithContext(context.WithValue(req.Context(), ctxKeyIdentity, identity))
		rr := httptest.NewRecorder()
		http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)
		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("status = %d, want 429 body=%s", rr.Code, rr.Body.String())
		}
		var env struct {
			Error string            `json:"error"`
			Data  map[string]string `json:"data"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
			t.Fatal(err)
		}
		if env.Data["code"] != cognitive.TokenBudgetExhaustedCode || env.Data["scope"] != "execution" || env.Data["limit"] != "5000" || env.Data["used"] != "0" {
			t.Fatalf("data = %v", env.Data)
		}
		adminAction := strings.Contains(env.Data["recommended_action"], "/api/v1/cognitive/budgets/overrides")
		if adminAction != (identity.Role == "admin") {
			t.Fatalf("role %s recommended_action = %q", identity.Role, env.Data["recommended_action"])
		}
		for _, forbidden := range []string{"proposal", "confirm_token", "intent_proof_id"} {
			if strings.Contains(rr.Body.String(), forbidden) {
				t.Fatalf("budget blocker carries %q: %s", forbidden, rr.Body.String())
			}
		}
	}
	if provider.calls != 0 {
		t.Fatalf("draft inferences = %d, want none", provider.calls)
	}
}

func TestDraftPerTurnConstantIsRemoved(t *testing.T) {
	source, err := os.ReadFile("cognitive_write_file_draft.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "writeFileDraftMaxPerTurn") {
		t.Fatal("writeFileDraftMaxPerTurn must be replaced by the token budget")
	}
}
