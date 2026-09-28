package cognitive

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// usageAdapter reports exact provider usage and counts calls.
type usageAdapter struct {
	mu       sync.Mutex
	prompt   int
	complete int
	calls    int
	lastMax  int
}

func (a *usageAdapter) Infer(_ context.Context, _ string, opts InferOptions) (*InferResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.calls++
	a.lastMax = opts.MaxTokens
	return &InferResponse{Text: "partial work", ModelUsed: "m", PromptTokens: a.prompt, CompletionTokens: a.complete, TokensUsed: a.prompt + a.complete}, nil
}

func (a *usageAdapter) Probe(context.Context) (bool, error) { return true, nil }

// memoryLedger is a test ledger; failing=true simulates the DB being down.
type memoryLedger struct {
	mu      sync.Mutex
	entries []TokenLedgerEntry
	failing bool
}

func (l *memoryLedger) PeriodTotal(_ context.Context, tenant, scope, ref string, since time.Time) (int, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.failing {
		return 0, false, errors.New("db down")
	}
	total, unreported := 0, false
	for _, e := range l.entries {
		match := (scope == protocol.TokenBudgetScopeRun && e.RunID == ref) ||
			(scope == protocol.TokenBudgetScopeTeamDay && e.TeamID == ref && !e.OccurredAt.Before(since)) ||
			(scope == protocol.TokenBudgetScopeAgentDay && e.AgentID == ref && !e.OccurredAt.Before(since))
		if match && BudgetTenant(e.TenantID) == BudgetTenant(tenant) && e.Outcome == TokenLedgerOutcomeCharged && e.ExecutionKind != ExecutionKindSystem {
			total += e.TotalTokens
			unreported = unreported || !e.UsageReported
		}
	}
	return total, unreported, nil
}

func (l *memoryLedger) Append(_ context.Context, entry TokenLedgerEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.failing {
		return errors.New("db down")
	}
	l.entries = append(l.entries, entry)
	return nil
}

func (l *memoryLedger) snapshot() []TokenLedgerEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]TokenLedgerEntry(nil), l.entries...)
}

func budgetRouter(adapter LLMProvider, spec protocol.TokenBudgetPolicySpec, ledger BudgetLedger) *Router {
	return &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{"local": {Type: "openai_compatible", ModelID: "qwen3:14b", Enabled: true, MaxOutputTokens: 2048, DataBoundary: DataBoundaryLocalOnly}},
			Profiles:  map[string]string{"chat": "local"},
		},
		Adapters: map[string]LLMProvider{"local": adapter},
		Budgets:  NewBudgetGovernor(spec, ledger),
	}
}

func teamCorrelation() InferenceCorrelation {
	return InferenceCorrelation{RunID: "run-1", TeamID: "team-a", AgentID: "coder"}
}

func TestInferWithContract_BudgetStopsBeforeProviderWhenRemainingBelowFloor(t *testing.T) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Team = map[string]protocol.TokenBudgetLimits{"team-a": {PerExecution: 2000}}
	adapter := &usageAdapter{prompt: 600, complete: 200}
	ledger := &memoryLedger{}
	r := budgetRouter(adapter, spec, ledger)
	meter := NewExecutionMeter(ExecutionKindAgentTurn, teamCorrelation())
	ctx := WithExecutionMeter(context.Background(), meter)
	req := InferRequest{Profile: "chat", Prompt: "go", Correlation: teamCorrelation()}

	for i := 0; i < 2; i++ {
		if _, err := r.InferWithContract(ctx, req); err != nil {
			t.Fatalf("call %d: %v", i+1, err)
		}
	}
	// 1600 used, 400 remaining: still above the 256 floor, clamped to 400.
	if _, err := r.InferWithContract(ctx, req); err != nil {
		t.Fatalf("call 3: %v", err)
	}
	if adapter.lastMax != 400 {
		t.Fatalf("MaxTokens = %d, want clamp to remaining 400", adapter.lastMax)
	}
	_, err := r.InferWithContract(ctx, req)
	var stop *TokenBudgetExhaustedError
	if !errors.Is(err, ErrTokenBudgetExhausted) || !errors.As(err, &stop) {
		t.Fatalf("err = %v, want token budget exhausted", err)
	}
	if stop.Scope != protocol.TokenBudgetScopeExecution || stop.Used != 2400 || stop.Limit != 2000 || stop.Code() != "token_budget_exhausted" {
		t.Fatalf("stop = %+v", stop)
	}
	if adapter.calls != 3 {
		t.Fatalf("provider calls = %d, want 3 (no call once remaining < 256)", adapter.calls)
	}
	if meter.Stop() == nil || meter.Used() != 2400 {
		t.Fatalf("meter used=%d stop=%v", meter.Used(), meter.Stop())
	}
	entries := ledger.snapshot()
	if len(entries) != 4 || entries[3].Outcome != TokenLedgerOutcomeRefused || entries[3].TotalTokens != 0 {
		t.Fatalf("ledger = %+v, want 3 charged + 1 refused", entries)
	}
	for _, e := range entries[:3] {
		if e.ExecutionID != meter.ID() || e.ExecutionKind != ExecutionKindAgentTurn || e.TeamID != "team-a" || e.BudgetClass != protocol.TokenBudgetClassLocalLarge || !e.UsageReported || e.TotalTokens != 800 {
			t.Fatalf("ledger entry = %+v", e)
		}
	}
}

func TestInferWithContract_RunAndDayCapsSpanExecutions(t *testing.T) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Team = map[string]protocol.TokenBudgetLimits{"team-a": {PerRun: 2048, PerExecution: 2048, PerTeamDay: 3072}}
	adapter := &usageAdapter{prompt: 900, complete: 100}
	r := budgetRouter(adapter, spec, &memoryLedger{})
	req := InferRequest{Profile: "chat", Prompt: "go", Correlation: teamCorrelation()}
	for i := 0; i < 2; i++ {
		ctx := WithExecutionMeter(context.Background(), NewExecutionMeter(ExecutionKindAgentTurn, teamCorrelation()))
		if _, err := r.InferWithContract(ctx, req); err != nil {
			t.Fatalf("execution %d: %v", i+1, err)
		}
	}
	ctx := WithExecutionMeter(context.Background(), NewExecutionMeter(ExecutionKindAgentTurn, teamCorrelation()))
	_, err := r.InferWithContract(ctx, req)
	var stop *TokenBudgetExhaustedError
	if !errors.As(err, &stop) || stop.Scope != protocol.TokenBudgetScopeRun {
		t.Fatalf("err = %v, want run scope stop", err)
	}
	// A new run on the same team is still bounded by the team day.
	next := InferenceCorrelation{RunID: "run-2", TeamID: "team-a", AgentID: "coder"}
	for i := 0; i < 2; i++ {
		ctx = WithExecutionMeter(context.Background(), NewExecutionMeter(ExecutionKindAgentTurn, next))
		_, err = r.InferWithContract(ctx, InferRequest{Profile: "chat", Prompt: "go", Correlation: next})
	}
	if !errors.As(err, &stop) || stop.Scope != protocol.TokenBudgetScopeTeamDay || stop.ResetsAt.IsZero() {
		t.Fatalf("err = %v, want team_day stop with reset time", err)
	}
}

func TestInferWithContract_WarnsOnceAtEightyPercent(t *testing.T) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Agent = map[string]protocol.TokenBudgetLimits{"coder": {PerExecution: 10000}}
	adapter := &usageAdapter{prompt: 2000, complete: 500}
	r := budgetRouter(adapter, spec, &memoryLedger{})
	var mu sync.Mutex
	var warnings []BudgetWarning
	r.Budgets.SetWarningSink(func(w BudgetWarning) { mu.Lock(); warnings = append(warnings, w); mu.Unlock() })
	meter := NewExecutionMeter(ExecutionKindAgentTurn, teamCorrelation())
	ctx := WithExecutionMeter(context.Background(), meter)
	for i := 0; i < 3; i++ {
		if _, err := r.InferWithContract(ctx, InferRequest{Profile: "chat", Prompt: "go", Correlation: teamCorrelation()}); err != nil {
			t.Fatalf("call %d: %v", i+1, err)
		}
		if i == 2 && len(warnings) != 0 {
			t.Fatalf("warned at 75%%: %+v", warnings)
		}
	}
	for i := 0; i < 1; i++ {
		_, _ = r.InferWithContract(ctx, InferRequest{Profile: "chat", Prompt: "go", Correlation: teamCorrelation()})
	}
	mu.Lock()
	defer mu.Unlock()
	if len(warnings) != 1 || warnings[0].Scope != protocol.TokenBudgetScopeExecution || warnings[0].Used != 10000 || warnings[0].Limit != 10000 || warnings[0].ExecutionID != meter.ID() {
		t.Fatalf("warnings = %+v, want exactly one execution warning", warnings)
	}
}

func TestInferWithContract_MissingUsageChargesReservation(t *testing.T) {
	adapter := &captureAdapter{response: &InferResponse{Text: "no usage reported", ModelUsed: "m"}}
	ledger := &memoryLedger{}
	r := budgetRouter(adapter, protocol.DefaultTokenBudgetPolicySpec(), ledger)
	meter := NewExecutionMeter(ExecutionKindAgentTurn, teamCorrelation())
	resp, err := r.InferWithContract(WithExecutionMeter(context.Background(), meter), InferRequest{Profile: "chat", Prompt: "go", Correlation: teamCorrelation()})
	if err != nil {
		t.Fatal(err)
	}
	if resp.TokensUsed != 0 {
		t.Fatalf("response usage was estimated: %d", resp.TokensUsed)
	}
	entries := ledger.snapshot()
	if len(entries) != 1 || entries[0].UsageReported || entries[0].TotalTokens != 2048 || meter.Used() != 2048 {
		t.Fatalf("ledger = %+v used=%d, want reservation 2048 with usage_reported=false", entries, meter.Used())
	}
	usage := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "team-a", protocol.TokenBudgetLimits{PerTeamDay: 100000, WarnPct: 80})
	if usage.UsageReported || usage.Used != 2048 {
		t.Fatalf("usage = %+v, want at-least usage with usage_reported=false", usage)
	}
}

func TestInferWithContract_LedgerDownUsesSinceRestartCounters(t *testing.T) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Team = map[string]protocol.TokenBudgetLimits{"team-a": {PerExecution: 2048, PerRun: 2048, PerTeamDay: 2048}}
	adapter := &usageAdapter{prompt: 1000, complete: 800}
	r := budgetRouter(adapter, spec, &memoryLedger{failing: true})
	req := InferRequest{Profile: "chat", Prompt: "go", Correlation: InferenceCorrelation{TeamID: "team-a", AgentID: "coder"}}
	if _, err := r.InferWithContract(WithExecutionMeter(context.Background(), NewExecutionMeter(ExecutionKindAgentTurn, req.Correlation)), req); err != nil {
		t.Fatal(err)
	}
	_, err := r.InferWithContract(WithExecutionMeter(context.Background(), NewExecutionMeter(ExecutionKindAgentTurn, req.Correlation)), req)
	var stop *TokenBudgetExhaustedError
	if !errors.As(err, &stop) || stop.Scope != protocol.TokenBudgetScopeTeamDay {
		t.Fatalf("err = %v, want team_day stop on since-restart counters", err)
	}
	usage := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "team-a", protocol.TokenBudgetLimits{PerTeamDay: 2048, WarnPct: 80})
	if usage.Period != protocol.TokenBudgetPeriodSinceRestart || usage.Used != 1800 || usage.Remaining != 248 || !usage.Warn {
		t.Fatalf("usage = %+v, want since_restart 1800/2048", usage)
	}
	if adapter.calls != 1 {
		t.Fatalf("provider calls = %d, want 1", adapter.calls)
	}
}

func TestInferWithContract_SystemCallWithoutMeterUsesOnlyPerExecution(t *testing.T) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	adapter := &usageAdapter{prompt: 100, complete: 50}
	ledger := &memoryLedger{}
	r := budgetRouter(adapter, spec, ledger)
	if _, err := r.InferWithContract(context.Background(), InferRequest{Profile: "chat", Prompt: "archive"}); err != nil {
		t.Fatal(err)
	}
	entries := ledger.snapshot()
	if len(entries) != 1 || entries[0].ExecutionKind != ExecutionKindSystem || entries[0].ExecutionID == "" {
		t.Fatalf("ledger = %+v, want one system execution", entries)
	}
	// A correlated call without a meter is still charged to its period scopes.
	if _, err := r.InferWithContract(context.Background(), InferRequest{Profile: "chat", Prompt: "recall", Correlation: teamCorrelation()}); err != nil {
		t.Fatal(err)
	}
	usage := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "team-a", protocol.TokenBudgetLimits{PerTeamDay: 100000, WarnPct: 80})
	if usage.Used != 150 || usage.Period != protocol.TokenBudgetPeriodUTCDay {
		t.Fatalf("usage = %+v, want 150 charged to team day", usage)
	}
}
