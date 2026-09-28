package cognitive

import "context"

// budgetPreflight is the single pre-call budget check. It resolves the limits
// for this call, refuses it (no provider call) when the tightest scope has
// less than MinBudgetHeadroom left, and otherwise clamps opts.MaxTokens to the
// remaining allowance. A nil governor means budgets are not enforced.
func (r *Router) budgetPreflight(ctx context.Context, req InferRequest, providerID string, cfg ProviderConfig, opts *InferOptions) (*budgetCharge, error) {
	if r == nil || r.Budgets == nil {
		return nil, nil
	}
	meter := meterFor(ctx, req)
	correlation := budgetCorrelation(req, meter)
	subject := BudgetSubject{AgentID: correlation.AgentID, TeamID: correlation.TeamID, Profile: req.Profile, Class: ResolveBudgetClass(cfg)}
	charge, clamp, err := r.Budgets.preflight(ctx, meter, correlation, subject, providerID, cfg.ModelID, opts.MaxTokens)
	if err != nil {
		return nil, err
	}
	opts.MaxTokens = clamp
	return charge, nil
}

// budgetCorrelation is the request correlation, or the meter's when the
// request names no run, team or agent (a tenant alone is not a scope).
func budgetCorrelation(req InferRequest, meter *ExecutionMeter) InferenceCorrelation {
	if c := req.Correlation; c.RunID != "" || c.TeamID != "" || c.AgentID != "" {
		return c
	}
	return meter.correlation
}

// budgetCharge records the usage of one admitted call. With a provider error
// only provider-reported usage is charged; otherwise the call is released.
func (r *Router) budgetCharge(ctx context.Context, charge *budgetCharge, resp *InferResponse, err error) {
	if r == nil || r.Budgets == nil || charge == nil || resp == nil {
		return
	}
	if err != nil {
		r.Budgets.chargeReported(ctx, charge, resp)
		return
	}
	r.Budgets.charge(ctx, charge, resp)
}

// budgetRelease returns the reservation of an admitted call that was never
// charged (provider error, nil response, panic). No-op after budgetCharge.
func (r *Router) budgetRelease(charge *budgetCharge) {
	if r == nil || r.Budgets == nil {
		return
	}
	r.Budgets.release(charge)
}

// BudgetHeadroom reports how many tokens one more call in this execution may
// use, without charging. ok=false when budgets are not enforced.
func (r *Router) BudgetHeadroom(ctx context.Context, req InferRequest) (int, *TokenBudgetExhaustedError, bool) {
	if r == nil || r.Budgets == nil {
		return 0, nil, false
	}
	r.mu.RLock()
	resolution := r.resolveExecutionProviderLocked(req.Profile, req.Provider)
	cfg := NormalizeProviderTokenDefaults(r.Config.Providers[resolution.ProviderID])
	r.mu.RUnlock()
	meter := meterFor(ctx, req)
	correlation := budgetCorrelation(req, meter)
	subject := BudgetSubject{AgentID: correlation.AgentID, TeamID: correlation.TeamID, Profile: req.Profile, Class: ResolveBudgetClass(cfg)}
	remaining, stop := r.Budgets.Headroom(ctx, meter, correlation, subject)
	return remaining, stop, true
}
