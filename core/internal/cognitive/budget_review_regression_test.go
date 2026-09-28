package cognitive

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// b1raAdapter reports prompt=b1raPrompt and completion=opts.MaxTokens. With a
// gate it blocks until the gate closes; err/panicMsg make the call fail.
type b1raAdapter struct {
	mu       sync.Mutex
	calls    int
	lastMax  int
	prompt   int
	gate     chan struct{}
	err      error
	panicMsg string
}

func (a *b1raAdapter) Infer(_ context.Context, _ string, opts InferOptions) (*InferResponse, error) {
	a.mu.Lock()
	a.calls++
	a.lastMax = opts.MaxTokens
	gate, err, panicMsg := a.gate, a.err, a.panicMsg
	a.mu.Unlock()
	if gate != nil {
		<-gate
	}
	if panicMsg != "" {
		panic(panicMsg)
	}
	if err != nil {
		return nil, err
	}
	return &InferResponse{Text: "x", PromptTokens: a.prompt, CompletionTokens: opts.MaxTokens, TokensUsed: a.prompt + opts.MaxTokens}, nil
}

func (a *b1raAdapter) Probe(context.Context) (bool, error) { return true, nil }

func (a *b1raAdapter) snapshot() (int, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.calls, a.lastMax
}

// b1raLedger fails reads and appends independently and counts reads.
type b1raLedger struct {
	mu          sync.Mutex
	mem         memoryLedger
	failReads   bool
	failAppends bool
	reads       int
}

func (l *b1raLedger) PeriodTotal(ctx context.Context, tenant, scope, ref string, since time.Time) (int, bool, error) {
	l.mu.Lock()
	l.reads++
	fail := l.failReads
	l.mu.Unlock()
	if fail {
		return 0, false, errors.New("db down")
	}
	return l.mem.PeriodTotal(ctx, tenant, scope, ref, since)
}

func (l *b1raLedger) Append(ctx context.Context, entry TokenLedgerEntry) error {
	l.mu.Lock()
	fail := l.failAppends
	l.mu.Unlock()
	if fail {
		return errors.New("db down")
	}
	return l.mem.Append(ctx, entry)
}

func (l *b1raLedger) set(failReads, failAppends bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failReads, l.failAppends = failReads, failAppends
}

func (l *b1raLedger) readCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.reads
}

type b1raClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *b1raClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *b1raClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

func b1raRouter(adapter LLMProvider, spec protocol.TokenBudgetPolicySpec, ledger BudgetLedger) *Router {
	return &Router{Config: &BrainConfig{Profiles: map[string]string{"chat": "local"},
		Providers: map[string]ProviderConfig{"local": {Type: "openai_compatible", ModelID: "m", Enabled: true, MaxOutputTokens: 2048}}},
		Adapters: map[string]LLMProvider{"local": adapter}, Budgets: NewBudgetGovernor(spec, ledger)}
}

func b1raInfer(r *Router, corr InferenceCorrelation) error {
	ctx := WithExecutionMeter(context.Background(), NewExecutionMeter(ExecutionKindAgentTurn, corr))
	_, err := r.InferWithContract(ctx, InferRequest{Profile: "chat", Prompt: "p", Correlation: corr})
	return err
}

// F1: concurrent executions sharing a period scope cannot all pass preflight;
// the admitted reservations never exceed the limit.
func TestB1RA_ConcurrentReservationNeverOverspendsPeriodScopes(t *testing.T) {
	cases := []struct {
		name, scope, ref string
		limit, admitted  int
		corr             InferenceCorrelation
		override         func(*protocol.TokenBudgetPolicySpec, int)
	}{
		{"team_day", protocol.TokenBudgetScopeTeamDay, "t", 4096, 2, InferenceCorrelation{TeamID: "t"},
			func(s *protocol.TokenBudgetPolicySpec, l int) {
				s.Overrides.Team = map[string]protocol.TokenBudgetLimits{"t": {PerExecution: l, PerRun: l, PerTeamDay: l}}
			}},
		{"team_day_uneven", protocol.TokenBudgetScopeTeamDay, "t", 5000, 3, InferenceCorrelation{TeamID: "t"},
			func(s *protocol.TokenBudgetPolicySpec, l int) {
				s.Overrides.Team = map[string]protocol.TokenBudgetLimits{"t": {PerExecution: l, PerRun: l, PerTeamDay: l}}
			}},
		{"agent_day", protocol.TokenBudgetScopeAgentDay, "a", 4096, 2, InferenceCorrelation{AgentID: "a"},
			func(s *protocol.TokenBudgetPolicySpec, l int) {
				s.Overrides.Agent = map[string]protocol.TokenBudgetLimits{"a": {PerExecution: l, PerRun: l, PerAgentDay: l}}
			}},
		{"run", protocol.TokenBudgetScopeRun, "r", 4096, 2, InferenceCorrelation{RunID: "r"},
			func(s *protocol.TokenBudgetPolicySpec, l int) {
				s.Overrides.Class = map[string]protocol.TokenBudgetLimits{protocol.TokenBudgetClassLocalLarge: {PerExecution: l, PerRun: l}}
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec := protocol.DefaultTokenBudgetPolicySpec()
			tc.override(&spec, tc.limit)
			adapter := &b1raAdapter{gate: make(chan struct{})}
			r := b1raRouter(adapter, spec, &memoryLedger{})
			limits := r.Budgets.Limits(BudgetSubject{AgentID: tc.corr.AgentID, TeamID: tc.corr.TeamID, Class: protocol.TokenBudgetClassLocalLarge})
			const n = 10
			var wg sync.WaitGroup
			errs := make(chan error, n)
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func() { defer wg.Done(); errs <- b1raInfer(r, tc.corr) }()
			}
			// Every refused caller returns while the admitted ones are gated.
			for i := 0; i < n-tc.admitted; i++ {
				select {
				case err := <-errs:
					if stop := AsTokenBudgetExhausted(err); stop == nil || stop.Scope != tc.scope {
						t.Fatalf("err = %v, want %s stop", err, tc.scope)
					}
				case <-time.After(5 * time.Second):
					t.Fatalf("only %d refusals before timeout", i)
				}
			}
			inflight := r.Budgets.Usage(context.Background(), "", tc.scope, tc.ref, limits)
			if calls, _ := adapter.snapshot(); calls != tc.admitted || inflight.Reserved != tc.limit || inflight.Remaining != 0 || inflight.Used != 0 {
				t.Fatalf("in flight: calls=%d usage=%+v, want %d calls reserving the whole limit", calls, inflight, tc.admitted)
			}
			close(adapter.gate)
			wg.Wait()
			close(errs)
			for err := range errs {
				if err != nil {
					t.Fatalf("admitted call failed: %v", err)
				}
			}
			after := r.Budgets.Usage(context.Background(), "", tc.scope, tc.ref, limits)
			if after.Used != tc.limit || after.Reserved != 0 || after.Remaining != 0 {
				t.Fatalf("settled usage = %+v, want used=%d (never above the limit), reserved=0", after, tc.limit)
			}
		})
	}
}

// Required behavior 1: a failed or panicking call releases its reservation
// and charges nothing; the next caller sees the full allowance again.
func TestB1RA_ReservationReleasedOnErrorAndPanic(t *testing.T) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Team = map[string]protocol.TokenBudgetLimits{"t": {PerExecution: 4096, PerRun: 4096, PerTeamDay: 4096}}
	ledger := &memoryLedger{}
	adapter := &b1raAdapter{err: errors.New("provider down")}
	r := b1raRouter(adapter, spec, ledger)
	corr := InferenceCorrelation{TeamID: "t"}
	limits := r.Budgets.Limits(BudgetSubject{TeamID: "t", Class: protocol.TokenBudgetClassLocalLarge})
	for i := 0; i < 3; i++ {
		if err := b1raInfer(r, corr); err == nil || AsTokenBudgetExhausted(err) != nil {
			t.Fatalf("call %d: err = %v, want provider failure (not a budget stop)", i+1, err)
		}
	}
	adapter.mu.Lock()
	adapter.err, adapter.panicMsg = nil, "adapter panic"
	adapter.mu.Unlock()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("adapter panic did not propagate")
			}
		}()
		_ = b1raInfer(r, corr)
	}()
	usage := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "t", limits)
	if usage.Reserved != 0 || usage.Used != 0 || usage.Remaining != 4096 {
		t.Fatalf("usage after failures = %+v, want nothing reserved or charged", usage)
	}
	if len(ledger.snapshot()) != 0 {
		t.Fatalf("failed calls wrote ledger rows: %+v", ledger.snapshot())
	}
	adapter.mu.Lock()
	adapter.panicMsg = ""
	adapter.mu.Unlock()
	for i := 0; i < 2; i++ {
		if err := b1raInfer(r, corr); err != nil {
			t.Fatalf("call after release %d: %v", i+1, err)
		}
		if _, lastMax := adapter.snapshot(); lastMax != 2048 {
			t.Fatalf("clamp = %d, want the full 2048 after release", lastMax)
		}
	}
	if err := b1raInfer(r, corr); AsTokenBudgetExhausted(err) == nil {
		t.Fatalf("err = %v, want team_day stop once actual usage fills the day", err)
	}
}

// Required behavior 1: settlement charges actual usage, not the reservation.
func TestB1RA_SettleChargesActualUsageNotReservation(t *testing.T) {
	adapter := &usageAdapter{prompt: 40, complete: 60}
	r := b1raRouter(adapter, protocol.DefaultTokenBudgetPolicySpec(), &memoryLedger{})
	corr := InferenceCorrelation{RunID: "r", TeamID: "t", AgentID: "a"}
	meter := NewExecutionMeter(ExecutionKindAgentTurn, corr)
	if _, err := r.InferWithContract(WithExecutionMeter(context.Background(), meter), InferRequest{Profile: "chat", Prompt: "p", Correlation: corr}); err != nil {
		t.Fatal(err)
	}
	limits := r.Budgets.Limits(BudgetSubject{Class: protocol.TokenBudgetClassLocalLarge})
	for _, s := range []struct{ scope, ref string }{{protocol.TokenBudgetScopeRun, "r"}, {protocol.TokenBudgetScopeTeamDay, "t"}, {protocol.TokenBudgetScopeAgentDay, "a"}} {
		if u := r.Budgets.Usage(context.Background(), "", s.scope, s.ref, limits); u.Used != 100 || u.Reserved != 0 {
			t.Fatalf("%s usage = %+v, want used=100 reserved=0", s.scope, u)
		}
	}
	if meter.Used() != 100 {
		t.Fatalf("meter used = %d, want 100", meter.Used())
	}
	headroom, stop, ok := r.BudgetHeadroom(WithExecutionMeter(context.Background(), meter), InferRequest{Profile: "chat", Correlation: corr})
	if !ok || stop == nil || headroom != limits.PerExecution-100 {
		t.Fatalf("headroom = %d ok=%v, want per_execution-100 with no leaked reservation", headroom, ok)
	}
}

// F3 (by design, D4): one call's prompt may overshoot per_execution, but the
// next call on that meter is refused before the provider is called.
func TestB1RA_PromptOvershootRefusesNextCallBeforeProvider(t *testing.T) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Agent = map[string]protocol.TokenBudgetLimits{"a": {PerExecution: 2048}}
	adapter := &b1raAdapter{prompt: 50000}
	r := b1raRouter(adapter, spec, nil)
	corr := InferenceCorrelation{AgentID: "a"}
	meter := NewExecutionMeter(ExecutionKindAgentTurn, corr)
	req := InferRequest{Profile: "chat", Prompt: "p", Correlation: corr, Meter: meter}
	if _, err := r.InferWithContract(context.Background(), req); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if meter.Used() != 52048 {
		t.Fatalf("meter used = %d, want the documented one-prompt overshoot 52048", meter.Used())
	}
	_, err := r.InferWithContract(context.Background(), req)
	stop := AsTokenBudgetExhausted(err)
	if stop == nil || stop.Scope != protocol.TokenBudgetScopeExecution || stop.Used != 52048 || stop.Limit != 2048 {
		t.Fatalf("err = %v, want execution stop at 52048/2048", err)
	}
	if calls, _ := adapter.snapshot(); calls != 1 {
		t.Fatalf("provider calls = %d, want 1 (refused before the provider)", calls)
	}
}
