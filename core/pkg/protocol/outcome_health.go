package protocol

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
	if item.NeedsOperator || len(item.RecoveryOptions) > 0 {
		return OutcomeHealthBlocked
	}
	switch item.State {
	case TeamWorkStateArchived:
		return OutcomeHealthArchived
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
	if len(item.BlockedBy) > 0 {
		return OutcomeHealthBlocked
	}
	switch item.State {
	case TeamWorkStateArchived:
		return OutcomeHealthArchived
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
	switch status {
	case "pending":
		return OutcomeHealthWaiting
	case "running":
		return OutcomeHealthRunning
	case "completed":
		return OutcomeHealthCompleted
	case "failed":
		return OutcomeHealthBlocked
	default:
		return OutcomeHealthHealthy
	}
}
