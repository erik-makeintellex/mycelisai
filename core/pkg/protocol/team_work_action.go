package protocol

import (
	"fmt"
	"strings"
)

type TeamWorkAction string

const (
	TeamWorkActionStartWork TeamWorkAction = "start_work"
	TeamWorkActionPause     TeamWorkAction = "pause"
	TeamWorkActionResume    TeamWorkAction = "resume"
	TeamWorkActionArchive   TeamWorkAction = "archive"
	TeamWorkActionSteer     TeamWorkAction = "steer"
	TeamWorkActionRecover   TeamWorkAction = "recover"
)

func NormalizeTeamWorkAction(raw TeamWorkAction) TeamWorkAction {
	return TeamWorkAction(strings.TrimSpace(string(raw)))
}

func ApplyTeamWorkAction(item TeamWorkItem, action TeamWorkAction) (TeamWorkState, error) {
	action = NormalizeTeamWorkAction(action)
	if item.ExecutionShape == TeamExecutionShapeCreateTeam {
		return item.State, fmt.Errorf("create_team work cannot be controlled with %s; ask Soma to create a delegated work item", action)
	}
	if item.State == TeamWorkStateArchived {
		return item.State, fmt.Errorf("archived work cannot be changed")
	}

	switch action {
	case TeamWorkActionStartWork:
		return startTeamWork(item)
	case TeamWorkActionPause:
		return pauseTeamWork(item)
	case TeamWorkActionResume:
		return resumeTeamWork(item)
	case TeamWorkActionArchive:
		return TeamWorkStateArchived, nil
	case TeamWorkActionSteer:
		return item.State, nil
	case TeamWorkActionRecover:
		return recoverTeamWork(item)
	default:
		return item.State, fmt.Errorf("invalid team work action")
	}
}

func startTeamWork(item TeamWorkItem) (TeamWorkState, error) {
	switch item.State {
	case TeamWorkStateNew, TeamWorkStateBriefed, TeamWorkStateQueued:
		return TeamWorkStateRunning, nil
	default:
		return item.State, fmt.Errorf("start_work is not available from %s", item.State)
	}
}

func pauseTeamWork(item TeamWorkItem) (TeamWorkState, error) {
	switch item.State {
	case TeamWorkStateQueued, TeamWorkStateRunning, TeamWorkStateNeedsOperator, TeamWorkStateReviewing, TeamWorkStateDegraded:
		return TeamWorkStatePaused, nil
	default:
		return item.State, fmt.Errorf("pause is not available from %s", item.State)
	}
}

func resumeTeamWork(item TeamWorkItem) (TeamWorkState, error) {
	if item.State == TeamWorkStatePaused {
		return TeamWorkStateQueued, nil
	}
	return item.State, fmt.Errorf("resume is only available for paused work")
}

func recoverTeamWork(item TeamWorkItem) (TeamWorkState, error) {
	switch item.State {
	case TeamWorkStateDegraded, TeamWorkStateNeedsOperator:
		return TeamWorkStateQueued, nil
	default:
		return item.State, fmt.Errorf("recover is only available for degraded or operator-needed work")
	}
}
