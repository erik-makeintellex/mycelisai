package cognitive

import (
	"context"
	"log"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func (g *BudgetGovernor) dayStart() time.Time {
	now := g.now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func (g *BudgetGovernor) nextUTCDay() time.Time { return g.dayStart().Add(24 * time.Hour) }

// counter returns the write-through counter for one period scope, loading its
// durable total once. The ledger read runs outside g.mu.
func (g *BudgetGovernor) counter(ctx context.Context, scope, ref string) (string, *periodCounter) {
	since := time.Time{}
	key := scope + "|" + ref
	if scope != protocol.TokenBudgetScopeRun {
		since = g.dayStart()
		key += "|" + since.Format("2006-01-02")
	}
	g.mu.Lock()
	existing := g.counters[key]
	g.mu.Unlock()
	if existing != nil {
		return key, existing
	}
	loaded := &periodCounter{source: protocol.TokenBudgetPeriodSinceRestart}
	if g.ledger != nil {
		total, unreported, err := g.ledger.PeriodTotal(context.WithoutCancel(ctx), scope, ref, since)
		if err != nil {
			log.Printf("WARN: token usage ledger unavailable for %s %s; enforcing since-restart counters: %v", scope, ref, err)
		} else {
			loaded.used, loaded.unreported = total, unreported
			loaded.source = protocol.TokenBudgetPeriodUTCDay
			if scope == protocol.TokenBudgetScopeRun {
				loaded.source = protocol.TokenBudgetPeriodRun
			}
		}
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if existing := g.counters[key]; existing != nil {
		return key, existing
	}
	g.counters[key] = loaded
	return key, loaded
}

// Usage reports one period scope against the given limits.
func (g *BudgetGovernor) Usage(ctx context.Context, scope, ref string, limits protocol.TokenBudgetLimits) BudgetUsage {
	_, counter := g.counter(ctx, scope, ref)
	limit := limits.PerRun
	switch scope {
	case protocol.TokenBudgetScopeTeamDay:
		limit = limits.PerTeamDay
	case protocol.TokenBudgetScopeAgentDay:
		limit = limits.PerAgentDay
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	usage := BudgetUsage{Scope: scope, Ref: ref, Used: counter.used, Limit: limit, WarnPct: limits.WarnPct,
		Warn: crossed(counter.used, limit, limits.WarnPct), Period: counter.source, UsageReported: !counter.unreported}
	if usage.Remaining = limit - counter.used; usage.Remaining < 0 {
		usage.Remaining = 0
	}
	if scope != protocol.TokenBudgetScopeRun {
		resets := g.nextUTCDay()
		usage.ResetsAt = &resets
	}
	return usage
}

func cloneBudgetSpec(spec protocol.TokenBudgetPolicySpec) protocol.TokenBudgetPolicySpec {
	clone := func(in map[string]protocol.TokenBudgetLimits) map[string]protocol.TokenBudgetLimits {
		if in == nil {
			return nil
		}
		out := make(map[string]protocol.TokenBudgetLimits, len(in))
		for key, value := range in {
			out[key] = value
		}
		return out
	}
	spec.Classes = clone(spec.Classes)
	spec.Overrides.Agent = clone(spec.Overrides.Agent)
	spec.Overrides.Team = clone(spec.Overrides.Team)
	spec.Overrides.Profile = clone(spec.Overrides.Profile)
	spec.Overrides.Class = clone(spec.Overrides.Class)
	return spec
}
