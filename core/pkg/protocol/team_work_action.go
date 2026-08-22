package protocol

import (
	"fmt"
	"strings"
)

type TeamWorkAction string

const (
	TeamWorkActionStartWork             TeamWorkAction = "start_work"
	TeamWorkActionPause                 TeamWorkAction = "pause"
	TeamWorkActionResume                TeamWorkAction = "resume"
	TeamWorkActionArchive               TeamWorkAction = "archive"
	TeamWorkActionSteer                 TeamWorkAction = "steer"
	TeamWorkActionRecover               TeamWorkAction = "recover"
	TeamWorkActionVerifyExternalOutcome TeamWorkAction = "verify_external_outcome"

	TeamWorkDegradationExternalMutationUnknown              = "external_mutation_outcome_unknown"
	TeamWorkDegradationExternalMutationVerifiedNotCommitted = "external_mutation_verified_not_committed"
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
	if action == TeamWorkActionVerifyExternalOutcome {
		if !WorkIntentHasExternalMutation(item.WorkIntent) {
			return item.State, fmt.Errorf("verify_external_outcome requires external mutation work")
		}
		if item.DegradationState != TeamWorkDegradationExternalMutationUnknown &&
			item.WorkIntent.SideEffect.SideEffectState != WorkSideEffectUnknown {
			return item.State, fmt.Errorf("verify_external_outcome requires an unknown external mutation outcome")
		}
		return item.State, nil
	}
	if reason := externalMutationControlBlockReason(item); reason != "" && action != TeamWorkActionSteer && action != TeamWorkActionArchive {
		return item.State, fmt.Errorf("%s before %s", reason, action)
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

func externalMutationControlBlockReason(item TeamWorkItem) string {
	if !WorkIntentHasExternalMutation(item.WorkIntent) {
		return ""
	}
	if item.DegradationState == TeamWorkDegradationExternalMutationVerifiedNotCommitted ||
		item.WorkIntent.SideEffect.SideEffectState == WorkSideEffectVerifiedNotCommitted {
		return "verified not-committed external mutation requires a new governed Soma proposal; generic control is unavailable"
	}
	if item.DegradationState == TeamWorkDegradationExternalMutationUnknown ||
		item.WorkIntent.SideEffect.SideEffectState == WorkSideEffectUnknown {
		return "external mutation outcome must be verified through Soma"
	}
	return ""
}

// ApplyExternalOutcomeVerification applies a complete server-attributed
// verification to an external-mutation work item.
func ApplyExternalOutcomeVerification(item TeamWorkItem, verification WorkExternalOutcomeVerification) (TeamWorkItem, error) {
	item = NormalizeTeamWorkItem(item)
	if _, err := ApplyTeamWorkAction(item, TeamWorkActionVerifyExternalOutcome); err != nil {
		return item, err
	}
	verification.Result = NormalizeWorkExternalOutcomeResult(verification.Result)
	verification.ActorRef = strings.TrimSpace(verification.ActorRef)
	verification.Summary = strings.TrimSpace(verification.Summary)
	verification.EvidenceRefs = dedupeStrings(compactStrings(verification.EvidenceRefs))
	if verification.Result == "" {
		return item, fmt.Errorf("result must be committed, not_committed, or still_unknown")
	}
	if verification.ActorRef == "" {
		return item, fmt.Errorf("actor_ref is required")
	}
	if verification.Summary == "" {
		return item, fmt.Errorf("summary is required")
	}
	if verification.RecordedAt.IsZero() {
		return item, fmt.Errorf("recorded_at is required")
	}

	sideEffect := *item.WorkIntent.SideEffect
	sideEffect.Verification = &verification
	item.WorkIntent.SideEffect = &sideEffect
	switch verification.Result {
	case WorkExternalOutcomeCommitted:
		sideEffect.SideEffectState = WorkSideEffectCommitted
		item.State = TeamWorkStateOutputReady
		item.NeedsOperator = false
		item.DegradationState = ""
		item.RecoveryOptions = nil
	case WorkExternalOutcomeNotCommitted:
		sideEffect.SideEffectState = WorkSideEffectVerifiedNotCommitted
		item.State = TeamWorkStateDegraded
		item.NeedsOperator = true
		item.DegradationState = TeamWorkDegradationExternalMutationVerifiedNotCommitted
		item.RecoveryOptions = []string{"Ask Soma to create a new governed proposal before attempting this external mutation again."}
	case WorkExternalOutcomeStillUnknown:
		sideEffect.SideEffectState = WorkSideEffectUnknown
		item.State = TeamWorkStateDegraded
		item.NeedsOperator = true
		item.DegradationState = TeamWorkDegradationExternalMutationUnknown
		item.RecoveryOptions = []string{"Verify the external system outcome before trusting completion; archive this work if verification cannot continue."}
	}
	return NormalizeTeamWorkItem(item), nil
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
