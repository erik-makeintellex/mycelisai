package cognitive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func usageFixtureServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestAnthropicAdapter_ReadsReportedUsage(t *testing.T) {
	server := usageFixtureServer(t, `{"id":"msg_1","content":[{"type":"text","text":"hi"}],"usage":{"input_tokens":120,"output_tokens":30}}`)
	adapter, err := NewAnthropicAdapter(ProviderConfig{AuthKey: "test-key", Endpoint: server.URL, ModelID: "claude"})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := adapter.Infer(context.Background(), "hi", InferOptions{MaxTokens: 64})
	if err != nil {
		t.Fatal(err)
	}
	if resp.PromptTokens != 120 || resp.CompletionTokens != 30 || resp.TokensUsed != 150 || resp.UpstreamResponseID != "msg_1" {
		t.Fatalf("usage = %+v, want 120+30=150", resp)
	}
}

func TestGoogleAdapter_ReadsReportedUsage(t *testing.T) {
	server := usageFixtureServer(t, `{"candidates":[{"content":{"parts":[{"text":"hi"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":200,"candidatesTokenCount":40,"totalTokenCount":251}}`)
	adapter, err := NewGoogleAdapter(ProviderConfig{AuthKey: "test-key", Endpoint: server.URL, ModelID: "gemini"})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := adapter.Infer(context.Background(), "hi", InferOptions{MaxTokens: 64})
	if err != nil {
		t.Fatal(err)
	}
	// totalTokenCount is authoritative (it can include thinking tokens).
	if resp.PromptTokens != 200 || resp.CompletionTokens != 40 || resp.TokensUsed != 251 {
		t.Fatalf("usage = %+v, want 200/40 total 251", resp)
	}
}

func TestBudgetLedger_TotalsEqualAdapterReportedUsage(t *testing.T) {
	openaiServer := usageFixtureServer(t, `{"id":"c1","object":"chat.completion","model":"m","choices":[{"index":0,"message":{"role":"assistant","content":"a"},"finish_reason":"stop"}],"usage":{"prompt_tokens":11,"completion_tokens":7,"total_tokens":18}}`)
	anthropicServer := usageFixtureServer(t, `{"id":"msg_2","content":[{"type":"text","text":"b"}],"usage":{"input_tokens":21,"output_tokens":9}}`)
	googleServer := usageFixtureServer(t, `{"candidates":[{"content":{"parts":[{"text":"c"}]}}],"usageMetadata":{"promptTokenCount":31,"candidatesTokenCount":13,"totalTokenCount":44}}`)
	emptyGoogle := usageFixtureServer(t, `{"candidates":[{"content":{"parts":[{"text":"d"}]}}]}`)

	providers := map[string]ProviderConfig{
		"oa":      {Type: "openai_compatible", Endpoint: openaiServer.URL, ModelID: "m", Enabled: true, MaxOutputTokens: 512},
		"an":      {Type: "anthropic", AuthKey: "k", Endpoint: anthropicServer.URL, ModelID: "claude", Enabled: true, DataBoundary: DataBoundaryLeavesOrg, MaxOutputTokens: 512},
		"go":      {Type: "google", AuthKey: "k", Endpoint: googleServer.URL, ModelID: "gemini", Enabled: true, DataBoundary: DataBoundaryLeavesOrg, MaxOutputTokens: 512},
		"nousage": {Type: "google", AuthKey: "k", Endpoint: emptyGoogle.URL, ModelID: "gemini", Enabled: true, DataBoundary: DataBoundaryLeavesOrg, MaxOutputTokens: 300},
	}
	adapters := map[string]LLMProvider{}
	for id, cfg := range providers {
		var adapter LLMProvider
		var err error
		switch cfg.Type {
		case "anthropic":
			adapter, err = NewAnthropicAdapter(cfg)
		case "google":
			adapter, err = NewGoogleAdapter(cfg)
		default:
			adapter, err = NewOpenAIAdapter(cfg)
		}
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		adapters[id] = adapter
	}
	ledger := &memoryLedger{}
	r := &Router{Config: &BrainConfig{Providers: providers, Profiles: map[string]string{"chat": "oa"}}, Adapters: adapters,
		Budgets: NewBudgetGovernor(protocol.DefaultTokenBudgetPolicySpec(), ledger)}
	correlation := InferenceCorrelation{RunID: "run-u", TeamID: "team-u", AgentID: "agent-u"}
	ctx := WithExecutionMeter(context.Background(), NewExecutionMeter(ExecutionKindAgentTurn, correlation))
	for _, id := range []string{"oa", "an", "go", "nousage"} {
		if _, err := r.InferWithContract(ctx, InferRequest{Provider: id, Prompt: "x", Correlation: correlation}); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
	}
	entries := ledger.snapshot()
	want := map[string]struct {
		total    int
		reported bool
		class    string
	}{
		"oa": {18, true, protocol.TokenBudgetClassLocalLarge}, "an": {30, true, protocol.TokenBudgetClassHostedPremium},
		"go": {44, true, protocol.TokenBudgetClassHostedPremium}, "nousage": {300, false, protocol.TokenBudgetClassHostedPremium},
	}
	sum := 0
	for _, e := range entries {
		w := want[e.ProviderID]
		if e.TotalTokens != w.total || e.UsageReported != w.reported || e.BudgetClass != w.class {
			t.Fatalf("%s entry = %+v, want %+v", e.ProviderID, e, w)
		}
		sum += e.TotalTokens
	}
	if len(entries) != 4 || sum != 18+30+44+300 {
		t.Fatalf("ledger entries=%d sum=%d", len(entries), sum)
	}
	usage := r.Budgets.Usage(context.Background(), protocol.TokenBudgetScopeRun, "run-u", protocol.TokenBudgetLimits{PerRun: 100000, WarnPct: 80})
	if usage.Used != sum || usage.UsageReported {
		t.Fatalf("run usage = %+v, want %d with usage_reported=false (one reservation)", usage, sum)
	}
}
