package cognitive

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// B1R-C: asserting ports of the B1R-QA attack probes (zz_qa_budget_attack_test.go).

// b1rcAdapter reports prompt=b1rcPrompt and completion=opts.MaxTokens. err is
// returned together with the response; noUsage drops the usage fields.
type b1rcAdapter struct {
	mu      sync.Mutex
	prompt  int
	err     error
	noUsage bool
	gate    chan struct{}
	maxes   []int
}

func (a *b1rcAdapter) Infer(_ context.Context, _ string, opts InferOptions) (*InferResponse, error) {
	a.mu.Lock()
	a.maxes = append(a.maxes, opts.MaxTokens)
	gate := a.gate
	a.mu.Unlock()
	if gate != nil {
		<-gate
	}
	resp := &InferResponse{Text: "partial", PromptTokens: a.prompt, CompletionTokens: opts.MaxTokens, TokensUsed: a.prompt + opts.MaxTokens}
	if a.noUsage {
		resp.PromptTokens, resp.CompletionTokens, resp.TokensUsed = 0, 0, 0
	}
	return resp, a.err
}

func (a *b1rcAdapter) Probe(context.Context) (bool, error) { return true, nil }

func (a *b1rcAdapter) admitted() []int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]int(nil), a.maxes...)
}

// b1rcState sums reservations and pins over every live counter.
func b1rcState(g *BudgetGovernor) (reserved, pins int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, c := range g.counters {
		reserved += c.reserved
		pins += c.pins
	}
	return reserved, pins
}

type b1rcPanicLedger struct{ b1raLedger }

func (l *b1rcPanicLedger) Append(context.Context, TokenLedgerEntry) error {
	panic("b1rc ledger append panic")
}

// LOW (QA TestZZQA_AppendPanicPinsCountersForever, pins=6): a panicking ledger
// append must not leave counters pinned (unevictable) or reservations held.
// The panic is handled as a failed append: the charge stays counted in memory
// and the periods are labelled since_restart until the ledger is readable.
func TestB1RC_LedgerAppendPanicDrainsPinsAndKeepsCharge(t *testing.T) {
	r := b1raRouter(&b1rcAdapter{prompt: 10}, protocol.DefaultTokenBudgetPolicySpec(), &b1rcPanicLedger{})
	corr := InferenceCorrelation{TeamID: "T", AgentID: "A", RunID: "R"}
	err := func() (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("panic escaped: %v", p)
			}
		}()
		return b1raInfer(r, corr)
	}()
	if err != nil {
		t.Fatalf("call = %v", err)
	}
	if reserved, pins := b1rcState(r.Budgets); reserved != 0 || pins != 0 {
		t.Fatalf("after append panic reserved=%d pins=%d, want 0/0", reserved, pins)
	}
	u := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "T", r.Budgets.Limits(b1raLocal))
	if u.Used != 2058 || u.Period != protocol.TokenBudgetPeriodSinceRestart {
		t.Fatalf("usage after append panic = %+v, want used 2058 labelled since_restart", u)
	}
}

// LOW (QA TestZZQA_ReservationLeakPaths "resp+err"): a response that arrives
// with an error still charges the provider-reported usage, releases the
// reservation and returns the error. Without reported usage nothing is
// charged (the reservation is only released).
func TestB1RC_ResponseWithErrorChargesReportedUsage(t *testing.T) {
	cases := []struct {
		name    string
		noUsage bool
		want    int
	}{{"usage reported", false, 2058}, {"no usage reported", true, 0}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ledger := &b1raLedger{}
			r := b1raRouter(&b1rcAdapter{prompt: 10, err: errors.New("stream cut after usage"), noUsage: tc.noUsage},
				protocol.DefaultTokenBudgetPolicySpec(), ledger)
			corr := InferenceCorrelation{TeamID: "T", AgentID: "A", RunID: "R"}
			meter := NewExecutionMeter(ExecutionKindAgentTurn, corr)
			resp, err := r.InferWithContract(WithExecutionMeter(context.Background(), meter), InferRequest{Profile: "chat", Prompt: "p", Correlation: corr})
			if !errors.Is(err, ErrAIEngineUnavailable) || resp != nil {
				t.Fatalf("resp=%v err=%v, want the provider error and no response", resp, err)
			}
			if reserved, pins := b1rcState(r.Budgets); reserved != 0 || pins != 0 || meter.reserved != 0 {
				t.Fatalf("reserved=%d pins=%d meter.reserved=%d, want 0", reserved, pins, meter.reserved)
			}
			for _, scope := range []struct{ scope, ref string }{{protocol.TokenBudgetScopeTeamDay, "T"}, {protocol.TokenBudgetScopeAgentDay, "A"}, {protocol.TokenBudgetScopeRun, "R"}} {
				if u := r.Budgets.Usage(context.Background(), "", scope.scope, scope.ref, r.Budgets.Limits(b1raLocal)); u.Used != tc.want {
					t.Fatalf("%s used = %d, want %d", scope.scope, u.Used, tc.want)
				}
			}
			if meter.Used() != tc.want {
				t.Fatalf("meter used = %d, want %d", meter.Used(), tc.want)
			}
			restarted := NewBudgetGovernor(protocol.DefaultTokenBudgetPolicySpec(), ledger)
			if u := restarted.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "T", r.Budgets.Limits(b1raLocal)); u.Used != tc.want {
				t.Fatalf("durable team used = %d, want %d", u.Used, tc.want)
			}
		})
	}
}

// Condition 1(b) (QA TestZZQA_RecoveryMaxDropsOutageCharge): read and append
// both fail, then the ledger recovers. The charge whose append failed must
// stay counted on top of the durable total (was max(): 100000, not 102058).
func TestB1RC_RecoveryKeepsChargeWhoseAppendFailed(t *testing.T) {
	clock := b1raNoon()
	ledger := &b1raLedger{failReads: true, failAppends: true}
	corr := InferenceCorrelation{TeamID: "T"}
	b1raDurable(ledger, clock.Now().Add(-time.Hour), corr, 100000)
	r := b1raRouter(&b1rcAdapter{prompt: 10}, protocol.DefaultTokenBudgetPolicySpec(), ledger)
	r.Budgets.now = clock.Now
	if err := b1raInfer(r, corr); err != nil {
		t.Fatal(err)
	}
	ledger.set(false, false)
	clock.Advance(31 * time.Second)
	u := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "T", r.Budgets.Limits(b1raLocal))
	if u.Used != 102058 || u.Period != protocol.TokenBudgetPeriodUTCDay {
		t.Fatalf("recovered usage = %+v, want used 102058 labelled utc_day", u)
	}
	// A later persisted charge is counted once.
	if err := b1raInfer(r, corr); err != nil {
		t.Fatal(err)
	}
	if u := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "T", r.Budgets.Limits(b1raLocal)); u.Used != 104116 {
		t.Fatalf("after a persisted charge used = %d, want 104116", u.Used)
	}
}

// Condition 1(a) (QA TestZZQA_OverlappingScopesConcurrent): prompts are not
// reserved. The admitted output clamps never exceed the limit, but with
// concurrent meters the scope can overshoot by the prompts of every admitted
// in-flight call. This pins the documented bound.
func TestB1RC_ConcurrentOvershootIsBoundedByInFlightPrompts(t *testing.T) {
	const teamLimit, callers, prompt = 10000, 40, 5000
	adapter := &b1rcAdapter{prompt: prompt, gate: make(chan struct{})}
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Team = map[string]protocol.TokenBudgetLimits{"T": {PerTeamDay: teamLimit}}
	r := b1raRouter(adapter, spec, &b1raLedger{})
	var wg sync.WaitGroup
	var mu sync.Mutex
	refused := 0
	for i := 0; i < callers; i++ {
		corr := InferenceCorrelation{TeamID: "T", AgentID: fmt.Sprintf("a%d", i%7), RunID: fmt.Sprintf("r%d", i%3)}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if AsTokenBudgetExhausted(b1raInfer(r, corr)) != nil {
				mu.Lock()
				refused++
				mu.Unlock()
			}
		}()
	}
	for deadline := time.Now().Add(5 * time.Second); ; {
		mu.Lock()
		done := refused+len(adapter.admitted()) == callers
		mu.Unlock()
		if done || time.Now().After(deadline) {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	close(adapter.gate)
	wg.Wait()
	admitted, sum := adapter.admitted(), 0
	for _, clamp := range admitted {
		sum += clamp
	}
	used := r.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "T", r.Budgets.Limits(BudgetSubject{TeamID: "T", Class: b1raLocal.Class})).Used
	if sum > teamLimit || len(admitted) < 2 {
		t.Fatalf("admitted=%d sum(clamps)=%d, want >=2 calls within %d", len(admitted), sum, teamLimit)
	}
	if overshoot := used - teamLimit; overshoot > len(admitted)*prompt || used != sum+len(admitted)*prompt {
		t.Fatalf("used=%d with %d admitted calls: overshoot must be exactly the in-flight prompts", used, len(admitted))
	}
}
