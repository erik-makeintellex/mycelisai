package cognitive

import (
	"context"
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
// Reserved is the output allowance held by in-flight calls; Remaining is
// limit - used - reserved (never negative).
type BudgetUsage struct {
	Scope         string     `json:"scope"`
	Ref           string     `json:"ref"`
	Used          int        `json:"used"`
	Limit         int        `json:"limit"`
	Remaining     int        `json:"remaining"`
	Reserved      int        `json:"reserved"`
	Warn          bool       `json:"warn"`
	WarnPct       int        `json:"warn_pct"`
	Period        string     `json:"period"`
	ResetsAt      *time.Time `json:"resets_at,omitempty"`
	UsageReported bool       `json:"usage_reported"`
}

// periodCounter is one run or day scope. All fields are guarded by g.mu.
type periodCounter struct {
	used       int
	reserved   int // output clamps of admitted, unsettled calls
	unreported bool
	source     string
	warned     bool
	loaded     bool      // durable total applied, or no ledger to load from
	lastLoad   time.Time // last ledger read attempt (retry rate limit)
	lastTouch  time.Time
	day        string // UTC day for day scopes, "" for run scopes
	pins       int    // in-flight calls holding this counter; pinned counters are never evicted
}

// BudgetGovernor resolves limits, reserves each admitted call's output clamp
// against every applicable scope, and settles provider-reported usage. Period
// totals load from the ledger into bounded write-through counters; while the
// ledger is unreadable they are since-restart counters, say so, and retry.
type BudgetGovernor struct {
	mu          sync.Mutex
	policy      protocol.TokenBudgetPolicySpec
	ledger      BudgetLedger
	counters    map[string]*periodCounter
	warn        func(BudgetWarning)
	now         func() time.Time
	maxCounters int
	sweptDay    string
	sweptAt     time.Time
}

func NewBudgetGovernor(spec protocol.TokenBudgetPolicySpec, ledger BudgetLedger) *BudgetGovernor {
	return &BudgetGovernor{policy: cloneBudgetSpec(spec), ledger: ledger, counters: map[string]*periodCounter{},
		now: time.Now, maxCounters: budgetMaxCounters}
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
	counter    *periodCounter // pinned by periodScopes
}

// budgetCharge carries one admitted call from preflight to settlement.
type budgetCharge struct {
	meter               *ExecutionMeter
	correlation         InferenceCorrelation
	subject             BudgetSubject
	limits              protocol.TokenBudgetLimits
	providerID, modelID string
	reservation         int
	scopes              []budgetScope // the counters holding the reservation
	settled             bool          // guarded by g.mu
}

// periodScopes returns the pinned run/team-day/agent-day counters for one
// call; the caller must unpin them (unpinLocked) when done.
func (g *BudgetGovernor) periodScopes(ctx context.Context, meter *ExecutionMeter, correlation InferenceCorrelation, limits protocol.TokenBudgetLimits) []budgetScope {
	if meter.kind == ExecutionKindSystem {
		return nil
	}
	var scopes []budgetScope
	add := func(scope, ref string, limit int) {
		if ref == "" {
			return
		}
		scopes = append(scopes, budgetScope{scope, ref, limit, g.counter(ctx, scope, ref)})
	}
	add(protocol.TokenBudgetScopeRun, correlation.RunID, limits.PerRun)
	add(protocol.TokenBudgetScopeTeamDay, correlation.TeamID, limits.PerTeamDay)
	add(protocol.TokenBudgetScopeAgentDay, correlation.AgentID, limits.PerAgentDay)
	return scopes
}

func (g *BudgetGovernor) unpinLocked(scopes []budgetScope) {
	for _, s := range scopes {
		s.counter.pins--
	}
}

// remainingLocked returns the tightest scope net of in-flight reservations.
// Caller holds g.mu and meter.mu.
func (g *BudgetGovernor) remainingLocked(meter *ExecutionMeter, limits protocol.TokenBudgetLimits, scopes []budgetScope) (int, *TokenBudgetExhaustedError) {
	best := limits.PerExecution - meter.used - meter.reserved
	stop := &TokenBudgetExhaustedError{Scope: protocol.TokenBudgetScopeExecution, Ref: meter.id, Used: meter.used, Reserved: meter.reserved, Limit: limits.PerExecution}
	for _, s := range scopes {
		if left := s.limit - s.counter.used - s.counter.reserved; left < best {
			best = left
			stop = &TokenBudgetExhaustedError{Scope: s.scope, Ref: s.ref, Used: s.counter.used, Reserved: s.counter.reserved, Limit: s.limit}
			if s.scope != protocol.TokenBudgetScopeRun {
				stop.ResetsAt = g.nextUTCDay()
			}
		}
	}
	return best, stop
}

// Headroom reports the tokens left for one more call without reserving.
func (g *BudgetGovernor) Headroom(ctx context.Context, meter *ExecutionMeter, correlation InferenceCorrelation, subject BudgetSubject) (int, *TokenBudgetExhaustedError) {
	limits := g.Limits(subject)
	scopes := g.periodScopes(ctx, meter, correlation, limits)
	g.mu.Lock()
	defer g.mu.Unlock()
	defer g.unpinLocked(scopes)
	meter.mu.Lock()
	defer meter.mu.Unlock()
	return g.remainingLocked(meter, limits, scopes)
}
