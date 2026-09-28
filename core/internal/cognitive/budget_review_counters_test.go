package cognitive

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func b1raNoon() *b1raClock {
	return &b1raClock{t: time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)}
}

func b1raDurable(ledger *b1raLedger, at time.Time, corr InferenceCorrelation, tokens int) {
	ledger.mem.entries = append(ledger.mem.entries, TokenLedgerEntry{OccurredAt: at, RunID: corr.RunID, TeamID: corr.TeamID, AgentID: corr.AgentID,
		ExecutionKind: ExecutionKindAgentTurn, TotalTokens: tokens, UsageReported: true, Outcome: TokenLedgerOutcomeCharged})
}

var b1raLocal = BudgetSubject{Class: protocol.TokenBudgetClassLocalLarge}

// F4: a ledger read that fails on first touch must not pin the counter at a
// since-restart zero; once the ledger recovers the durable total counts.
func TestB1RA_LedgerBlipRecoversDurableDayTotal(t *testing.T) {
	clock := b1raNoon()
	ledger := &b1raLedger{failReads: true}
	corr := InferenceCorrelation{TeamID: "t"}
	b1raDurable(ledger, clock.Now().Add(-time.Hour), corr, 1_999_900)
	adapter := &usageAdapter{prompt: 100, complete: 100}
	r := b1raRouter(adapter, protocol.DefaultTokenBudgetPolicySpec(), ledger)
	r.Budgets.now = clock.Now
	limits := r.Budgets.Limits(b1raLocal)
	if err := b1raInfer(r, corr); err != nil {
		t.Fatalf("call during outage: %v", err)
	}
	if u := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "t", limits); u.Used != 200 || u.Period != protocol.TokenBudgetPeriodSinceRestart {
		t.Fatalf("outage usage = %+v, want since_restart 200", u)
	}
	ledger.set(false, false)
	reads := ledger.readCount()
	if u := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "t", limits); u.Period != protocol.TokenBudgetPeriodSinceRestart || ledger.readCount() != reads {
		t.Fatalf("retry inside the rate limit: usage=%+v reads=%d->%d", u, reads, ledger.readCount())
	}
	clock.Advance(31 * time.Second)
	u := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "t", limits)
	// The outage call's append succeeded, so durable already holds it: 1_999_900+200, never +400.
	if u.Used != 2_000_100 || u.Period != protocol.TokenBudgetPeriodUTCDay {
		t.Fatalf("recovered usage = %+v, want utc_day 2000100 (durable, not double-counted)", u)
	}
	err := b1raInfer(r, corr)
	if stop := AsTokenBudgetExhausted(err); stop == nil || stop.Scope != protocol.TokenBudgetScopeTeamDay {
		t.Fatalf("err = %v, want team_day stop from the durable total", err)
	}
	if adapter.calls != 1 {
		t.Fatalf("provider calls = %d, want 1", adapter.calls)
	}
}

// F4: loads retry at most once per interval per counter, and a recovered load
// keeps in-memory charges whose appends failed (durable + unappended).
func TestB1RA_LedgerRetryIsRateLimitedAndKeepsUnappendedCharges(t *testing.T) {
	clock := b1raNoon()
	ledger := &b1raLedger{failReads: true, failAppends: true}
	corr := InferenceCorrelation{TeamID: "t"}
	b1raDurable(ledger, clock.Now().Add(-time.Hour), corr, 150)
	r := b1raRouter(&usageAdapter{prompt: 100, complete: 100}, protocol.DefaultTokenBudgetPolicySpec(), ledger)
	r.Budgets.now = clock.Now
	for i := 0; i < 3; i++ {
		if err := b1raInfer(r, corr); err != nil {
			t.Fatal(err)
		}
		clock.Advance(5 * time.Second)
	}
	if got := ledger.readCount(); got != 1 {
		t.Fatalf("ledger reads = %d within 30s, want 1", got)
	}
	clock.Advance(31 * time.Second)
	if err := b1raInfer(r, corr); err != nil {
		t.Fatal(err)
	}
	if got := ledger.readCount(); got != 2 {
		t.Fatalf("ledger reads = %d after the interval, want 2", got)
	}
	ledger.set(false, false)
	clock.Advance(31 * time.Second)
	u := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "t", r.Budgets.Limits(b1raLocal))
	// B1R-C: the unappended charges count on top of the durable total
	// (was max(durable 150, in-memory 800) = 800, which lost the 150).
	if u.Used != 950 || u.Period != protocol.TokenBudgetPeriodUTCDay {
		t.Fatalf("usage = %+v, want utc_day durable 150 + unappended 800 = 950", u)
	}
}

// F5: usage reads never allocate counters, whatever the ledger state.
func TestB1RA_UsageReadsNeverAllocateCounters(t *testing.T) {
	durable := &b1raLedger{}
	b1raDurable(durable, time.Now().UTC(), InferenceCorrelation{RunID: "known"}, 700)
	cases := []struct {
		name       string
		ledger     BudgetLedger
		wantPeriod string
	}{
		{"no ledger", nil, protocol.TokenBudgetPeriodSinceRestart},
		{"ledger", durable, protocol.TokenBudgetPeriodRun},
		{"ledger down", &b1raLedger{failReads: true}, protocol.TokenBudgetPeriodSinceRestart},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewBudgetGovernor(protocol.DefaultTokenBudgetPolicySpec(), tc.ledger)
			for i := 0; i < 5000; i++ {
				u := g.Usage(context.Background(), "", protocol.TokenBudgetScopeRun, fmt.Sprintf("r%d", i), protocol.TokenBudgetLimits{PerRun: 1024})
				if u.Used != 0 || u.Period != tc.wantPeriod || u.Remaining != 1024 {
					t.Fatalf("usage = %+v, want 0 with period %s", u, tc.wantPeriod)
				}
			}
			day := g.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "nobody", protocol.TokenBudgetLimits{PerTeamDay: 1024})
			if day.ResetsAt == nil {
				t.Fatalf("day usage without a counter lost resets_at: %+v", day)
			}
			if len(g.counters) != 0 {
				t.Fatalf("counters = %d after usage reads, want 0", len(g.counters))
			}
		})
	}
	g := NewBudgetGovernor(protocol.DefaultTokenBudgetPolicySpec(), durable)
	if u := g.Usage(context.Background(), "", protocol.TokenBudgetScopeRun, "known", protocol.TokenBudgetLimits{PerRun: 1024}); u.Used != 700 || u.Period != protocol.TokenBudgetPeriodRun || len(g.counters) != 0 {
		t.Fatalf("unknown-to-memory ref = %+v counters=%d, want the durable 700 without caching", u, len(g.counters))
	}
}

// F5: day counters from past days are dropped on roll-over.
func TestB1RA_DayRollOverDropsPastDayCounters(t *testing.T) {
	clock := b1raNoon()
	r := b1raRouter(&usageAdapter{prompt: 1, complete: 1}, protocol.DefaultTokenBudgetPolicySpec(), nil)
	r.Budgets.now = clock.Now
	for i := 0; i < 5; i++ {
		if err := b1raInfer(r, InferenceCorrelation{TeamID: fmt.Sprintf("t%d", i), AgentID: "a"}); err != nil {
			t.Fatal(err)
		}
	}
	clock.Advance(24 * time.Hour)
	if err := b1raInfer(r, InferenceCorrelation{TeamID: "t0"}); err != nil {
		t.Fatal(err)
	}
	today := clock.Now().Format("2006-01-02")
	if len(r.Budgets.counters) != 1 {
		t.Fatalf("counters = %d after roll-over, want only today's t0", len(r.Budgets.counters))
	}
	for key, c := range r.Budgets.counters {
		if c.day != today {
			t.Fatalf("past-day counter kept: %s", key)
		}
	}
}

// F5: idle run counters are evicted and reload from the ledger, so eviction
// loses no accuracy with a DB; a size cap bounds the map regardless.
func TestB1RA_RunCountersEvictIdleReloadAndCap(t *testing.T) {
	clock := b1raNoon()
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Class = map[string]protocol.TokenBudgetLimits{protocol.TokenBudgetClassLocalLarge: {PerExecution: 2048, PerRun: 2048}}
	adapter := &usageAdapter{prompt: 900, complete: 100}
	r := b1raRouter(adapter, spec, &b1raLedger{})
	r.Budgets.now = clock.Now
	for i := 0; i < 2; i++ {
		if err := b1raInfer(r, InferenceCorrelation{RunID: "r0"}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 1; i < 10; i++ {
		_ = b1raInfer(r, InferenceCorrelation{RunID: fmt.Sprintf("r%d", i)})
	}
	clock.Advance(61 * time.Minute)
	if err := b1raInfer(r, InferenceCorrelation{RunID: "fresh"}); err != nil {
		t.Fatal(err)
	}
	if n := len(r.Budgets.counters); n != 1 {
		t.Fatalf("counters = %d after the idle TTL, want 1", n)
	}
	err := b1raInfer(r, InferenceCorrelation{RunID: "r0"})
	if stop := AsTokenBudgetExhausted(err); stop == nil || stop.Scope != protocol.TokenBudgetScopeRun || stop.Used != 2000 {
		t.Fatalf("err = %v, want run stop at 2000 reloaded from the ledger", err)
	}

	capped := b1raRouter(&usageAdapter{prompt: 1, complete: 1}, protocol.DefaultTokenBudgetPolicySpec(), nil)
	capped.Budgets.maxCounters = 8
	for i := 0; i < 50; i++ {
		if err := b1raInfer(capped, InferenceCorrelation{RunID: fmt.Sprintf("c%d", i)}); err != nil {
			t.Fatal(err)
		}
	}
	if n := len(capped.Budgets.counters); n > 8 {
		t.Fatalf("counters = %d, want at most the cap 8", n)
	}
}
