package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// writeFileDraftCorrelation charges proposal-time drafting to Soma (agent
// admin), so it counts against the admin agent's day budget.
var writeFileDraftCorrelation = cognitive.InferenceCorrelation{AgentID: "admin"}

// writeFileDraftBudgetPreflight refuses the drafting pass before any
// inference when the drafts' output reservations (n x provider
// max_output_tokens) do not fit the remaining budget. Nothing is proposed.
func (s *AdminServer) writeFileDraftBudgetPreflight(ctx context.Context, planned []protocol.PlannedToolCall, targets []int) *writeFileDraftBlocker {
	if s == nil || s.Cognitive == nil || len(targets) == 0 {
		return nil
	}
	fit, need := 0, 0
	var tightest *cognitive.TokenBudgetExhaustedError
	for _, i := range targets {
		profile := writeFileDraftProfile(s.Cognitive, firstNonEmptyString(planned[i].Arguments["path"]))
		remaining, stop, enforced := s.Cognitive.BudgetHeadroom(ctx, cognitive.InferRequest{Profile: profile, Correlation: writeFileDraftCorrelation})
		if !enforced {
			return nil
		}
		_, cfg, _ := s.Cognitive.ProfileProviderSnapshot(profile)
		need += cognitive.NormalizeProviderTokenDefaults(cfg).MaxOutputTokens
		if remaining >= cognitive.MinBudgetHeadroom && need <= remaining {
			fit++
			continue
		}
		if tightest == nil {
			tightest = stop
		}
	}
	if tightest == nil {
		return nil
	}
	return draftBudgetBlocker(tightest, fmt.Sprintf(
		"This work stopped because it reached its token budget: Soma can draft %d of %d files within it, so nothing was proposed.", fit, len(targets)))
}

func draftBudgetBlocker(stop *cognitive.TokenBudgetExhaustedError, summary string) *writeFileDraftBlocker {
	return &writeFileDraftBlocker{Status: http.StatusTooManyRequests, Budget: stop, Availability: cognitive.ExecutionAvailability{
		Available: false, Code: cognitive.TokenBudgetExhaustedCode, Summary: summary,
		RecommendedAction: tokenBudgetUserAction,
	}}
}
