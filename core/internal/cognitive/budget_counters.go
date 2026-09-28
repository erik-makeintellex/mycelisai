package cognitive

import (
	"context"
	"log"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

const (
	// budgetLedgerRetryInterval rate-limits durable reloads per counter.
	budgetLedgerRetryInterval = 30 * time.Second
	// budgetRunCounterIdleTTL evicts run counters nobody has touched.
	budgetRunCounterIdleTTL = time.Hour
	// budgetCounterSweepInterval rate-limits the sweep below the size cap.
	budgetCounterSweepInterval = time.Minute
	// budgetMaxCounters caps live counters; the least recently used
	// unpinned counter is evicted at the cap.
	budgetMaxCounters = 20000
)

func (g *BudgetGovernor) dayStart() time.Time {
	now := g.now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func (g *BudgetGovernor) nextUTCDay() time.Time { return g.dayStart().Add(24 * time.Hour) }

// counterKey names one tenant's period scope; day scopes key on the current
// UTC day. The tenant must already be normalized (BudgetTenant).
func (g *BudgetGovernor) counterKey(tenant, scope, ref string) (key string, since time.Time, day string) {
	key = tenant + "|" + scope + "|" + ref
	if scope == protocol.TokenBudgetScopeRun {
		return key, time.Time{}, ""
	}
	since = g.dayStart()
	day = since.Format("2006-01-02")
	return key + "|" + day, since, day
}

func budgetPeriodLabel(scope string) string {
	if scope == protocol.TokenBudgetScopeRun {
		return protocol.TokenBudgetPeriodRun
	}
	return protocol.TokenBudgetPeriodUTCDay
}

func (g *BudgetGovernor) readPeriod(ctx context.Context, tenant, scope, ref string, since time.Time) (int, bool, error) {
	total, unreported, err := g.ledger.PeriodTotal(context.WithoutCancel(ctx), tenant, scope, ref, since)
	if err != nil {
		log.Printf("WARN: token usage ledger unavailable for %s %s %s; enforcing since-restart counters: %v", tenant, scope, ref, err)
	}
	return total, unreported, err
}

// counter returns the pinned write-through counter for one tenant's period
// scope, creating it (with one durable read outside g.mu) on first touch. The
// caller must unpin it. An unloaded counter retries its durable read.
func (g *BudgetGovernor) counter(ctx context.Context, tenant, scope, ref string) *periodCounter {
	key, since, day := g.counterKey(tenant, scope, ref)
	g.mu.Lock()
	if existing := g.counters[key]; existing != nil {
		existing.pins++
		existing.lastTouch = g.now()
		g.mu.Unlock()
		g.reload(ctx, tenant, scope, ref, since, existing)
		return existing
	}
	g.mu.Unlock()
	fresh := &periodCounter{source: protocol.TokenBudgetPeriodSinceRestart, day: day, loaded: g.ledger == nil}
	if g.ledger != nil {
		fresh.lastLoad = g.now()
		if total, unreported, err := g.readPeriod(ctx, tenant, scope, ref, since); err == nil {
			fresh.used, fresh.unreported, fresh.source, fresh.loaded = total, unreported, budgetPeriodLabel(scope), true
		}
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()
	if existing := g.counters[key]; existing != nil {
		existing.pins++
		existing.lastTouch = now
		return existing
	}
	g.sweepLocked(now)
	fresh.pins, fresh.lastTouch = 1, now
	g.counters[key] = fresh
	return fresh
}

// reload retries the durable read of an unloaded counter, at most once per
// budgetLedgerRetryInterval. On success used becomes max(durable +
// unpersisted, in-memory): charges whose appends failed stay counted on top of
// the durable total, and nothing is counted twice.
func (g *BudgetGovernor) reload(ctx context.Context, tenant, scope, ref string, since time.Time, c *periodCounter) {
	if g.ledger == nil {
		return
	}
	g.mu.Lock()
	now := g.now()
	if c.loaded || now.Sub(c.lastLoad) < budgetLedgerRetryInterval {
		g.mu.Unlock()
		return
	}
	c.lastLoad = now
	g.mu.Unlock()
	total, unreported, err := g.readPeriod(ctx, tenant, scope, ref, since)
	if err != nil {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if total += c.unpersisted; total > c.used {
		c.used = total
	}
	c.unreported = c.unreported || unreported
	c.source, c.loaded = budgetPeriodLabel(scope), true
}

// sweepLocked bounds the counter map: past-day counters and idle run counters
// go, and at the size cap the least recently touched counter goes. Pinned
// counters stay. An evicted counter reloads from the ledger on its next
// touch; without a ledger its count restarts at zero. Caller holds g.mu.
func (g *BudgetGovernor) sweepLocked(now time.Time) {
	today := g.dayStart().Format("2006-01-02")
	full := len(g.counters) >= g.maxCounters
	if !full && today == g.sweptDay && now.Sub(g.sweptAt) < budgetCounterSweepInterval {
		return
	}
	g.sweptDay, g.sweptAt = today, now
	lru := ""
	var lruAt time.Time
	for key, c := range g.counters {
		if c.pins > 0 {
			continue
		}
		if (c.day != "" && c.day != today) || (c.day == "" && now.Sub(c.lastTouch) >= budgetRunCounterIdleTTL) {
			delete(g.counters, key)
			continue
		}
		if lru == "" || c.lastTouch.Before(lruAt) {
			lru, lruAt = key, c.lastTouch
		}
	}
	if len(g.counters) >= g.maxCounters && lru != "" {
		delete(g.counters, lru)
	}
}

// Usage reports one tenant's period scope against the given limits ("" is
// tenant "default"). It never creates a counter: an unknown ref is read from
// the ledger without caching (zero, labelled since_restart, when there is no
// readable ledger).
func (g *BudgetGovernor) Usage(ctx context.Context, tenant, scope, ref string, limits protocol.TokenBudgetLimits) BudgetUsage {
	tenant = BudgetTenant(tenant)
	key, since, _ := g.counterKey(tenant, scope, ref)
	limit := limits.PerRun
	switch scope {
	case protocol.TokenBudgetScopeTeamDay:
		limit = limits.PerTeamDay
	case protocol.TokenBudgetScopeAgentDay:
		limit = limits.PerAgentDay
	}
	usage := BudgetUsage{Scope: scope, Ref: ref, Limit: limit, WarnPct: limits.WarnPct, Period: protocol.TokenBudgetPeriodSinceRestart, UsageReported: true}
	g.mu.Lock()
	existing := g.counters[key]
	g.mu.Unlock()
	if existing != nil {
		g.reload(ctx, tenant, scope, ref, since, existing)
		g.mu.Lock()
		usage.Used, usage.Reserved, usage.Period, usage.UsageReported = existing.used, existing.reserved, existing.source, !existing.unreported
		g.mu.Unlock()
	} else if g.ledger != nil {
		if total, unreported, err := g.readPeriod(ctx, tenant, scope, ref, since); err == nil {
			usage.Used, usage.Period, usage.UsageReported = total, budgetPeriodLabel(scope), !unreported
		}
	}
	usage.Warn = crossed(usage.Used, limit, limits.WarnPct)
	if usage.Remaining = limit - usage.Used - usage.Reserved; usage.Remaining < 0 {
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
