package swarm

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

type budgetUsageProvider struct {
	mu    sync.Mutex
	text  string
	used  int
	calls int
}

func (p *budgetUsageProvider) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	return &cognitive.InferResponse{Text: p.text, ModelUsed: "qwen3:14b", PromptTokens: p.used - 100, CompletionTokens: 100, TokensUsed: p.used}, nil
}

func (p *budgetUsageProvider) Probe(context.Context) (bool, error) { return true, nil }

func budgetAgentRouter(provider cognitive.LLMProvider, spec protocol.TokenBudgetPolicySpec) *cognitive.Router {
	return &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles:  map[string]string{"chat": "local"},
			Providers: map[string]cognitive.ProviderConfig{"local": {Type: "openai_compatible", Enabled: true, ModelID: "qwen3:14b", MaxOutputTokens: 2048}},
		},
		Adapters: map[string]cognitive.LLMProvider{"local": provider},
		Budgets:  cognitive.NewBudgetGovernor(spec, nil),
	}
}

func assertHonestBudgetStop(t *testing.T, result ProcessResult) {
	t.Helper()
	if result.Availability == nil || result.Availability.Available || result.Availability.Code != cognitive.TokenBudgetExhaustedCode {
		t.Fatalf("availability = %+v, want token_budget_exhausted", result.Availability)
	}
	if result.ExecutionStatus != ExecutionStatusStoppedBudget {
		t.Fatalf("execution status = %q, want stopped_budget", result.ExecutionStatus)
	}
	if !strings.Contains(result.Availability.Summary, "reached its token budget") {
		t.Fatalf("summary = %q", result.Availability.Summary)
	}
	payload := map[string]any{}
	if err := json.Unmarshal(teamAgentResponsePayload(result), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["state"] != "degraded" || payload["degradation_state"] != cognitive.TokenBudgetExhaustedCode || payload["execution_status"] != ExecutionStatusStoppedBudget {
		t.Fatalf("team payload = %v, want degraded stopped_budget", payload)
	}
	for _, word := range []string{"completed", "verified"} {
		if strings.Contains(strings.ToLower(result.Text), word) || payload["state"] == word {
			t.Fatalf("budget stop claims %q: %q", word, result.Text)
		}
	}
}

func TestProcessMessage_ExecutionBudgetStopsRecoveryWithoutProviderCall(t *testing.T) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Agent = map[string]protocol.TokenBudgetLimits{"admin": {PerExecution: 2048}}
	provider := &budgetUsageProvider{text: "", used: 1900}
	agent := NewAgent(context.Background(), protocol.AgentManifest{ID: "admin", Role: "admin"}, "admin-core", nil, budgetAgentRouter(provider, spec), nil)

	result := agent.processMessageStructured("summarize the workspace", nil)
	assertHonestBudgetStop(t, result)
	if provider.calls != 1 {
		t.Fatalf("provider calls = %d, want 1 (recovery refused below 256 remaining)", provider.calls)
	}
	if result.Partial || result.Text != "" {
		t.Fatalf("no text was produced, so nothing may be returned: %+v", result)
	}
}

func TestProcessMessage_TeamDayBudgetRefusesInitialInference(t *testing.T) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Team = map[string]protocol.TokenBudgetLimits{"delivery": {PerExecution: 2048, PerRun: 2048, PerTeamDay: 2048}}
	provider := &budgetUsageProvider{text: "First answer.", used: 1900}
	router := budgetAgentRouter(provider, spec)
	agent := NewAgent(context.Background(), protocol.AgentManifest{ID: "coder", Role: "coder"}, "delivery", nil, router, nil)

	first := agent.processMessageStructured("first ask", nil)
	if first.Availability != nil || first.Text != "First answer." || first.ExecutionStatus != "" {
		t.Fatalf("first execution = %+v", first)
	}
	second := agent.processMessageStructured("second ask", nil)
	assertHonestBudgetStop(t, second)
	if provider.calls != 1 {
		t.Fatalf("provider calls = %d, want 1", provider.calls)
	}
}

func TestBudgetStopResult_LabelsProducedTextPartial(t *testing.T) {
	stop := &cognitive.TokenBudgetExhaustedError{Scope: protocol.TokenBudgetScopeRun, Used: 4096, Limit: 4096}
	result := budgetStopResult(stop, "Section one drafted", "chat", "local", "m", nil)
	if !result.Partial || result.Text != "Section one drafted" || result.ExecutionStatus != ExecutionStatusStoppedBudget {
		t.Fatalf("result = %+v, want partial text with stopped_budget", result)
	}
	if !strings.Contains(result.Availability.Summary, "partial") {
		t.Fatalf("summary must say the text is partial: %q", result.Availability.Summary)
	}
}
