package swarm

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// ExecutionStatusStoppedBudget marks an execution that ended because it
// reached a token budget. It is never completed or verified.
const ExecutionStatusStoppedBudget = "stopped_budget"

// somaAgentID is the admin council member that runs the Soma turn.
const somaAgentID = "admin"

func (a *Agent) executionKind() string {
	if a.Manifest.ID == somaAgentID {
		return cognitive.ExecutionKindSomaTurn
	}
	return cognitive.ExecutionKindAgentTurn
}

func budgetStopSummary(stop *cognitive.TokenBudgetExhaustedError, partial bool) string {
	summary := fmt.Sprintf("This work stopped because it reached its token budget (%s: %d of %d tokens).", stop.Scope, stop.Used, stop.Limit)
	if partial {
		summary += " The text returned is partial and unfinished."
	}
	return summary
}

// budgetStopResult is the honest budget stop: any text the model already
// produced is returned labelled partial, the status is stopped_budget, and
// nothing is marked complete. Retained artifacts stay visible as evidence.
func budgetStopResult(stop *cognitive.TokenBudgetExhaustedError, text, profile, providerID, modelUsed string, artifacts []protocol.ChatArtifactRef) ProcessResult {
	text = stripToolCallJSON(text)
	partial := strings.TrimSpace(text) != ""
	if !partial {
		text = ""
	}
	action := "Retry with a smaller ask, or ask an admin to raise the token budget for this team or agent."
	if !stop.ResetsAt.IsZero() {
		action = fmt.Sprintf("The %s budget resets at %s UTC. %s", stop.Scope, stop.ResetsAt.UTC().Format("2006-01-02 15:04"), action)
	}
	return ProcessResult{
		Text:      text,
		Artifacts: artifacts,
		Availability: &cognitive.ExecutionAvailability{
			Available: false, Code: cognitive.TokenBudgetExhaustedCode, Summary: budgetStopSummary(stop, partial),
			RecommendedAction: action, Profile: profile, ProviderID: providerID, ModelID: modelUsed,
		},
		ProviderID: providerID, ModelUsed: modelUsed,
		ExecutionStatus: ExecutionStatusStoppedBudget,
		Partial:         partial,
	}
}
