package cognitive

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

const (
	TokenLedgerOutcomeCharged = "charged"
	TokenLedgerOutcomeRefused = "refused_exhausted"
)

// TokenLedgerEntry is one append-only token_usage_ledger row.
type TokenLedgerEntry struct {
	OccurredAt       time.Time
	ExecutionID      string
	ExecutionKind    string
	RunID            string
	TeamID           string
	AgentID          string
	ProviderID       string
	ModelID          string
	BudgetClass      string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	UsageReported    bool
	Outcome          string
}

// BudgetLedger is the durable period store. PeriodTotal sums charged,
// non-system rows for one scope since the period start (run: all rows) and
// reports whether any of them was an unreported-usage reservation.
type BudgetLedger interface {
	PeriodTotal(ctx context.Context, scope, ref string, since time.Time) (int, bool, error)
	Append(ctx context.Context, entry TokenLedgerEntry) error
}

// BudgetWarning is emitted once per scope per execution (execution scope) or
// per run/day counter when usage crosses warn_pct.
type BudgetWarning struct {
	Scope, Ref                          string
	Used, Limit, WarnPct                int
	ExecutionID, ExecutionKind          string
	RunID, TeamID, AgentID, BudgetClass string
}

// BudgetUsage is the read model served by GET /api/v1/cognitive/budgets/usage.
type BudgetUsage struct {
	Scope         string     `json:"scope"`
	Ref           string     `json:"ref"`
	Used          int        `json:"used"`
	Limit         int        `json:"limit"`
	Remaining     int        `json:"remaining"`
	Warn          bool       `json:"warn"`
	WarnPct       int        `json:"warn_pct"`
	Period        string     `json:"period"`
	ResetsAt      *time.Time `json:"resets_at,omitempty"`
	UsageReported bool       `json:"usage_reported"`
}

type periodCounter struct {
	used       int
	unreported bool
	source     string
	warned     bool
}

// BudgetGovernor resolves limits, refuses calls that would start below the
// headroom floor, and charges provider-reported usage. Period totals are read
// once per scope from the ledger and then kept as write-through counters; when
// the ledger is unavailable they are since-restart counters and say so.
type BudgetGovernor struct {
	mu       sync.Mutex
	policy   protocol.TokenBudgetPolicySpec
	ledger   BudgetLedger
	counters map[string]*periodCounter
	warn     func(BudgetWarning)
	now      func() time.Time
}

func NewBudgetGovernor(spec protocol.TokenBudgetPolicySpec, ledger BudgetLedger) *BudgetGovernor {
	return &BudgetGovernor{policy: cloneBudgetSpec(spec), ledger: ledger, counters: map[string]*periodCounter{}, now: time.Now}
}

// Policy returns a copy of the effective policy.
func (g *BudgetGovernor) Policy() protocol.TokenBudgetPolicySpec {
	g.mu.Lock()
	defer g.mu.Unlock()
	return cloneBudgetSpec(g.policy)
}

func (g *BudgetGovernor) SetPolicy(spec protocol.TokenBudgetPolicySpec) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.policy = cloneBudgetSpec(spec)
}

func (g *BudgetGovernor) SetWarningSink(sink func(BudgetWarning)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.warn = sink
}

// Durable reports whether period totals have a ledger behind them.
func (g *BudgetGovernor) Durable() bool { return g.ledger != nil }

func (g *BudgetGovernor) Limits(subject BudgetSubject) protocol.TokenBudgetLimits {
	g.mu.Lock()
	defer g.mu.Unlock()
	limits, _ := ResolveBudgetLimits(g.policy, subject)
	return limits
}

type budgetScope struct {
	scope, ref string
	limit      int
	counter    *periodCounter // nil for the execution scope
}

// budgetCharge carries one admitted call from preflight to charge.
type budgetCharge struct {
	meter               *ExecutionMeter
	correlation         InferenceCorrelation
	subject             BudgetSubject
	limits              protocol.TokenBudgetLimits
	providerID, modelID string
	reservation         int
	keys                []string
}

func (g *BudgetGovernor) periodScopes(ctx context.Context, meter *ExecutionMeter, correlation InferenceCorrelation, limits protocol.TokenBudgetLimits) ([]budgetScope, []string) {
	if meter.kind == ExecutionKindSystem {
		return nil, nil
	}
	var scopes []budgetScope
	var keys []string
	add := func(scope, ref string, limit int) {
		if ref == "" {
			return
		}
		key, counter := g.counter(ctx, scope, ref)
		scopes = append(scopes, budgetScope{scope, ref, limit, counter})
		keys = append(keys, key)
	}
	add(protocol.TokenBudgetScopeRun, correlation.RunID, limits.PerRun)
	add(protocol.TokenBudgetScopeTeamDay, correlation.TeamID, limits.PerTeamDay)
	add(protocol.TokenBudgetScopeAgentDay, correlation.AgentID, limits.PerAgentDay)
	return scopes, keys
}

// remainingLocked returns the tightest scope. Caller holds g.mu and meter.mu.
func (g *BudgetGovernor) remainingLocked(meter *ExecutionMeter, limits protocol.TokenBudgetLimits, scopes []budgetScope) (int, *TokenBudgetExhaustedError) {
	best := limits.PerExecution - meter.used
	stop := &TokenBudgetExhaustedError{Scope: protocol.TokenBudgetScopeExecution, Ref: meter.id, Used: meter.used, Limit: limits.PerExecution}
	for _, s := range scopes {
		if left := s.limit - s.counter.used; left < best {
			best = left
			stop = &TokenBudgetExhaustedError{Scope: s.scope, Ref: s.ref, Used: s.counter.used, Limit: s.limit}
			if s.scope != protocol.TokenBudgetScopeRun {
				stop.ResetsAt = g.nextUTCDay()
			}
		}
	}
	return best, stop
}

// Headroom reports the tokens left for one more call without charging.
func (g *BudgetGovernor) Headroom(ctx context.Context, meter *ExecutionMeter, correlation InferenceCorrelation, subject BudgetSubject) (int, *TokenBudgetExhaustedError) {
	limits := g.Limits(subject)
	scopes, _ := g.periodScopes(ctx, meter, correlation, limits)
	g.mu.Lock()
	defer g.mu.Unlock()
	meter.mu.Lock()
	defer meter.mu.Unlock()
	return g.remainingLocked(meter, limits, scopes)
}

// preflight admits or refuses one call and returns the output clamp.
func (g *BudgetGovernor) preflight(ctx context.Context, meter *ExecutionMeter, correlation InferenceCorrelation, subject BudgetSubject, providerID, modelID string, maxOutput int) (*budgetCharge, int, error) {
	limits := g.Limits(subject)
	scopes, keys := g.periodScopes(ctx, meter, correlation, limits)
	g.mu.Lock()
	meter.mu.Lock()
	remaining, stop := g.remainingLocked(meter, limits, scopes)
	if remaining < MinBudgetHeadroom {
		meter.stop = stop
	}
	meter.mu.Unlock()
	g.mu.Unlock()
	charge := &budgetCharge{meter: meter, correlation: correlation, subject: subject, limits: limits, providerID: providerID, modelID: modelID, keys: keys}
	if remaining < MinBudgetHeadroom {
		g.append(ctx, charge, TokenLedgerEntry{Outcome: TokenLedgerOutcomeRefused, UsageReported: true})
		return nil, 0, stop
	}
	clamp := maxOutput
	if clamp <= 0 || clamp > remaining {
		clamp = remaining
	}
	charge.reservation = clamp
	return charge, clamp, nil
}

// charge records provider-reported usage; with none reported it charges the
// clamped reservation (a configured bound, not a text estimate) and marks the
// row usage_reported=false.
func (g *BudgetGovernor) charge(ctx context.Context, c *budgetCharge, resp *InferResponse) {
	entry := TokenLedgerEntry{Outcome: TokenLedgerOutcomeCharged, PromptTokens: resp.PromptTokens, CompletionTokens: resp.CompletionTokens}
	entry.TotalTokens = resp.TokensUsed
	if entry.TotalTokens <= 0 {
		entry.TotalTokens = resp.PromptTokens + resp.CompletionTokens
	}
	entry.UsageReported = entry.TotalTokens > 0
	if !entry.UsageReported {
		entry.TotalTokens, entry.PromptTokens, entry.CompletionTokens = c.reservation, 0, 0
	}
	if resp.ModelUsed != "" {
		c.modelID = resp.ModelUsed
	}
	scopes, _ := g.periodScopes(ctx, c.meter, c.correlation, c.limits)
	var warnings []BudgetWarning
	g.mu.Lock()
	c.meter.mu.Lock()
	c.meter.used += entry.TotalTokens
	if crossed(c.meter.used, c.limits.PerExecution, c.limits.WarnPct) && !c.meter.warned {
		c.meter.warned = true
		warnings = append(warnings, g.warning(c, protocol.TokenBudgetScopeExecution, c.meter.id, c.meter.used, c.limits.PerExecution))
	}
	c.meter.mu.Unlock()
	for _, s := range scopes {
		s.counter.used += entry.TotalTokens
		s.counter.unreported = s.counter.unreported || !entry.UsageReported
		if crossed(s.counter.used, s.limit, c.limits.WarnPct) && !s.counter.warned {
			s.counter.warned = true
			warnings = append(warnings, g.warning(c, s.scope, s.ref, s.counter.used, s.limit))
		}
	}
	sink := g.warn
	g.mu.Unlock()
	g.append(ctx, c, entry)
	if sink != nil {
		for _, w := range warnings {
			sink(w)
		}
	}
}

func crossed(used, limit, warnPct int) bool {
	return limit > 0 && warnPct > 0 && used*100 >= limit*warnPct
}

func (g *BudgetGovernor) warning(c *budgetCharge, scope, ref string, used, limit int) BudgetWarning {
	return BudgetWarning{Scope: scope, Ref: ref, Used: used, Limit: limit, WarnPct: c.limits.WarnPct,
		ExecutionID: c.meter.id, ExecutionKind: c.meter.kind, RunID: c.correlation.RunID, TeamID: c.correlation.TeamID,
		AgentID: c.correlation.AgentID, BudgetClass: c.subject.Class}
}

func (g *BudgetGovernor) append(ctx context.Context, c *budgetCharge, entry TokenLedgerEntry) {
	if g.ledger == nil {
		return
	}
	entry.OccurredAt = g.now().UTC()
	entry.ExecutionID, entry.ExecutionKind = c.meter.id, c.meter.kind
	entry.RunID, entry.TeamID, entry.AgentID = c.correlation.RunID, c.correlation.TeamID, c.correlation.AgentID
	entry.ProviderID, entry.ModelID, entry.BudgetClass = c.providerID, c.modelID, c.subject.Class
	if err := g.ledger.Append(context.WithoutCancel(ctx), entry); err != nil {
		// The in-memory counters stay exact; the durable record is now
		// incomplete, so these periods are labelled since restart.
		log.Printf("WARN: token usage ledger append failed (execution %s): %v", entry.ExecutionID, err)
		g.mu.Lock()
		for _, key := range c.keys {
			if counter := g.counters[key]; counter != nil {
				counter.source = protocol.TokenBudgetPeriodSinceRestart
			}
		}
		g.mu.Unlock()
	}
}
