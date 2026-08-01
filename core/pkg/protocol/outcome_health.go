package protocol

import "strings"

// OutcomeHealthState is the user-facing operational state for an Outcome.
// Stored runtime statuses remain source-specific; this projection gives UI and
// API clients one stable trust/recovery vocabulary.
type OutcomeHealthState string

const (
	OutcomeHealthHealthy   OutcomeHealthState = "healthy"
	OutcomeHealthWaiting   OutcomeHealthState = "waiting"
	OutcomeHealthRunning   OutcomeHealthState = "running"
	OutcomeHealthDegraded  OutcomeHealthState = "degraded"
	OutcomeHealthBlocked   OutcomeHealthState = "blocked"
	OutcomeHealthCompleted OutcomeHealthState = "completed"
	OutcomeHealthArchived  OutcomeHealthState = "archived"
)

func OutcomeHealthForProject(item OutcomeProject) OutcomeHealthState {
	if item.Status == OutcomeProjectStatusArchived {
		return OutcomeHealthArchived
	}
	if len(item.RecoveryRefs) > 0 || item.Status == OutcomeProjectStatusNeedsAttention {
		return OutcomeHealthDegraded
	}
	if item.Status == OutcomeProjectStatusOutputReady || len(item.OutputRefs) > 0 {
		return OutcomeHealthCompleted
	}
	return OutcomeHealthHealthy
}

func OutcomeHealthForTeamWork(item TeamWorkItem) OutcomeHealthState {
	if item.State == TeamWorkStateArchived {
		return OutcomeHealthArchived
	}
	if item.NeedsOperator {
		return OutcomeHealthBlocked
	}
	switch item.State {
	case TeamWorkStateOutputReady:
		return OutcomeHealthCompleted
	case TeamWorkStateNeedsOperator:
		return OutcomeHealthBlocked
	case TeamWorkStateDegraded:
		return OutcomeHealthDegraded
	case TeamWorkStateRunning, TeamWorkStateReviewing:
		return OutcomeHealthRunning
	case TeamWorkStateNew, TeamWorkStateBriefed, TeamWorkStateQueued, TeamWorkStatePaused:
		return OutcomeHealthWaiting
	default:
		return OutcomeHealthHealthy
	}
}

func OutcomeHealthForTeamStatusEvent(item TeamStatusEvent) OutcomeHealthState {
	if item.State == TeamWorkStateArchived {
		return OutcomeHealthArchived
	}
	if len(item.BlockedBy) > 0 {
		return OutcomeHealthBlocked
	}
	switch item.State {
	case TeamWorkStateOutputReady:
		return OutcomeHealthCompleted
	case TeamWorkStateNeedsOperator:
		return OutcomeHealthBlocked
	case TeamWorkStateDegraded:
		return OutcomeHealthDegraded
	case TeamWorkStateRunning, TeamWorkStateReviewing:
		return OutcomeHealthRunning
	case TeamWorkStateNew, TeamWorkStateBriefed, TeamWorkStateQueued, TeamWorkStatePaused:
		return OutcomeHealthWaiting
	default:
		return OutcomeHealthHealthy
	}
}

func OutcomeHealthForRunStatus(status string) OutcomeHealthState {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "archived":
		return OutcomeHealthArchived
	case "pending", "queued", "new", "briefed", "paused":
		return OutcomeHealthWaiting
	case "running", "reviewing":
		return OutcomeHealthRunning
	case "completed", "succeeded", "success", "output_ready":
		return OutcomeHealthCompleted
	case "degraded", "needs_attention":
		return OutcomeHealthDegraded
	case "failed", "blocked", "needs_operator", "cancelled", "canceled":
		return OutcomeHealthBlocked
	default:
		return OutcomeHealthHealthy
	}
}

// AggregateOutcomeHealth returns the most actionable state across an Outcome.
// Archived only wins when every supplied state is archived.
func AggregateOutcomeHealth(states ...OutcomeHealthState) OutcomeHealthState {
	if len(states) == 0 {
		return OutcomeHealthHealthy
	}
	allArchived := true
	seen := make(map[OutcomeHealthState]bool, len(states))
	for _, state := range states {
		seen[state] = true
		if state != OutcomeHealthArchived {
			allArchived = false
		}
	}
	if allArchived {
		return OutcomeHealthArchived
	}
	for _, state := range []OutcomeHealthState{
		OutcomeHealthBlocked,
		OutcomeHealthDegraded,
		OutcomeHealthRunning,
		OutcomeHealthCompleted,
		OutcomeHealthWaiting,
		OutcomeHealthHealthy,
	} {
		if seen[state] {
			return state
		}
	}
	return OutcomeHealthHealthy
}
