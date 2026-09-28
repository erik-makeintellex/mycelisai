package cognitive

import (
	"context"
	"fmt"
	"log"

	"github.com/mycelis/core/pkg/protocol"
)

// preflight admits or refuses one call and returns the output clamp. The
// remaining check and the reservation of the clamp against the meter and
// every period counter happen under one g.mu hold, so concurrent callers can
// never be admitted past a limit on the same scope (no TOCTOU). An admitted
// call must end in exactly one charge or release.
func (g *BudgetGovernor) preflight(ctx context.Context, meter *ExecutionMeter, correlation InferenceCorrelation, subject BudgetSubject, providerID, modelID string, maxOutput int) (*budgetCharge, int, error) {
	limits := g.Limits(subject)
	scopes := g.periodScopes(ctx, meter, correlation, limits)
	g.mu.Lock()
	meter.mu.Lock()
	remaining, stop := g.remainingLocked(meter, limits, scopes)
	admitted := remaining >= MinBudgetHeadroom
	clamp := maxOutput
	if clamp <= 0 || clamp > remaining {
		clamp = remaining
	}
	if admitted {
		meter.reserved += clamp
		for _, s := range scopes {
			s.counter.reserved += clamp
		}
	} else {
		meter.stop = stop
		g.unpinLocked(scopes)
	}
	meter.mu.Unlock()
	g.mu.Unlock()
	charge := &budgetCharge{meter: meter, correlation: correlation, subject: subject, limits: limits, providerID: providerID, modelID: modelID}
	if !admitted {
		g.append(ctx, charge, TokenLedgerEntry{Outcome: TokenLedgerOutcomeRefused, UsageReported: true}, nil)
		return nil, 0, stop
	}
	charge.reservation, charge.scopes = clamp, scopes
	return charge, clamp, nil
}

// releaseLocked returns an unsettled reservation. Caller holds g.mu.
func (g *BudgetGovernor) releaseLocked(c *budgetCharge) {
	c.settled = true
	c.meter.mu.Lock()
	c.meter.reserved -= c.reservation
	c.meter.mu.Unlock()
	for _, s := range c.scopes {
		s.counter.reserved -= c.reservation
	}
}

// release drops the reservation of a call that was not settled (provider
// error, nil response, panic). It is a no-op after charge.
func (g *BudgetGovernor) release(c *budgetCharge) {
	if c == nil {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if c.settled {
		return
	}
	g.releaseLocked(c)
	g.unpinLocked(c.scopes)
}

// charge settles one admitted call: it releases the reservation and charges
// provider-reported usage; with none reported it charges the clamped
// reservation (a configured bound, not a text estimate) and marks the row
// usage_reported=false (D6).
func (g *BudgetGovernor) charge(ctx context.Context, c *budgetCharge, resp *InferResponse) {
	g.settle(ctx, c, resp, true)
}

// chargeReported settles a call whose adapter returned a response together
// with an error: provider-reported usage is charged; without any, the call is
// only released (nothing is charged).
func (g *BudgetGovernor) chargeReported(ctx context.Context, c *budgetCharge, resp *InferResponse) {
	g.settle(ctx, c, resp, false)
}

func reportedTotal(resp *InferResponse) int {
	if resp.TokensUsed > 0 {
		return resp.TokensUsed
	}
	return resp.PromptTokens + resp.CompletionTokens
}

func (g *BudgetGovernor) settle(ctx context.Context, c *budgetCharge, resp *InferResponse, chargeUnreported bool) {
	if c == nil || resp == nil {
		return
	}
	entry := TokenLedgerEntry{Outcome: TokenLedgerOutcomeCharged, PromptTokens: resp.PromptTokens, CompletionTokens: resp.CompletionTokens}
	entry.TotalTokens = reportedTotal(resp)
	entry.UsageReported = entry.TotalTokens > 0
	if !entry.UsageReported && !chargeUnreported {
		g.release(c)
		return
	}
	if !entry.UsageReported {
		entry.TotalTokens, entry.PromptTokens, entry.CompletionTokens = c.reservation, 0, 0
	}
	if resp.ModelUsed != "" {
		c.modelID = resp.ModelUsed
	}
	// Charge the current period's counters (a call admitted before midnight
	// is charged to the day its ledger row lands in).
	scopes := g.periodScopes(ctx, c.meter, c.correlation, c.limits)
	var warnings []BudgetWarning
	g.mu.Lock()
	if c.settled {
		g.unpinLocked(scopes)
		g.mu.Unlock()
		return
	}
	g.releaseLocked(c)
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
	// Counters stay pinned until the ledger row lands, so an eviction can
	// never reload a durable total that is missing this charge. The unpin is
	// deferred so the pins drain whatever the append does.
	func() {
		defer func() {
			g.mu.Lock()
			g.unpinLocked(c.scopes)
			g.unpinLocked(scopes)
			g.mu.Unlock()
		}()
		g.append(ctx, c, entry, scopes)
	}()
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

func (g *BudgetGovernor) append(ctx context.Context, c *budgetCharge, entry TokenLedgerEntry, charged []budgetScope) {
	if g.ledger == nil {
		return
	}
	entry.TenantID = BudgetTenant(c.correlation.TenantID)
	entry.OccurredAt = g.now().UTC()
	entry.ExecutionID, entry.ExecutionKind = c.meter.id, c.meter.kind
	entry.RunID, entry.TeamID, entry.AgentID = c.correlation.RunID, c.correlation.TeamID, c.correlation.AgentID
	entry.ProviderID, entry.ModelID, entry.BudgetClass = c.providerID, c.modelID, c.subject.Class
	if err := g.appendSafely(ctx, entry); err != nil {
		// The in-memory counters stay exact but the durable record is now
		// incomplete: remember the unpersisted charge, label these periods
		// since restart, and let the next reload (durable + unpersisted vs
		// in-memory) restore the day label.
		log.Printf("WARN: token usage ledger append failed (execution %s): %v", entry.ExecutionID, err)
		g.mu.Lock()
		now := g.now()
		for _, s := range charged {
			s.counter.unpersisted += entry.TotalTokens
			s.counter.source, s.counter.loaded, s.counter.lastLoad = protocol.TokenBudgetPeriodSinceRestart, false, now
		}
		g.mu.Unlock()
	}
}

// appendSafely treats a panicking ledger as a failed append.
func (g *BudgetGovernor) appendSafely(ctx context.Context, entry TokenLedgerEntry) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("token usage ledger: append panicked: %v", p)
		}
	}()
	return g.ledger.Append(context.WithoutCancel(ctx), entry)
}
