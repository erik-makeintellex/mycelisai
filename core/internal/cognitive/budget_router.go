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
	correlation := req.Correlation
	if correlation == (InferenceCorrelation{}) {
		correlation = meter.correlation
	}
	subject := BudgetSubject{AgentID: correlation.AgentID, TeamID: correlation.TeamID, Profile: req.Profile, Class: ResolveBudgetClass(cfg)}
	charge, clamp, err := r.Budgets.preflight(ctx, meter, correlation, subject, providerID, cfg.ModelID, opts.MaxTokens)
	if err != nil {
		return nil, err
	}
	opts.MaxTokens = clamp
	return charge, nil
}

// budgetCharge records the provider-reported usage of one admitted call.
func (r *Router) budgetCharge(ctx context.Context, charge *budgetCharge, resp *InferResponse) {
	if r == nil || r.Budgets == nil || charge == nil || resp == nil {
		return
	}
	r.Budgets.charge(ctx, charge, resp)
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
	correlation := req.Correlation
	if correlation == (InferenceCorrelation{}) {
		correlation = meter.correlation
	}
	subject := BudgetSubject{AgentID: correlation.AgentID, TeamID: correlation.TeamID, Profile: req.Profile, Class: ResolveBudgetClass(cfg)}
	remaining, stop := r.Budgets.Headroom(ctx, meter, correlation, subject)
	return remaining, stop, true
}
