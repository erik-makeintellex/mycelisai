package protocol

import "strings"

const (
	WorkStopCancelActive      = "cancel_active"
	WorkStopDisableSchedule   = "disable_schedule"
	WorkStopService           = "stop_service"
	WorkStopPauseProject      = "pause_project"
	WorkStopRollbackExtension = "rollback_extension"

	WorkRetryRerun            = "rerun"
	WorkRetryRunOnceNow       = "run_once_now"
	WorkRetryRestartService   = "restart_service"
	WorkRetryResumeProject    = "resume_project"
	WorkRetryRebuildExtension = "rebuild_extension"

	WorkRecoverReviseAndRerun    = "revise_and_rerun"
	WorkRecoverRepairSchedule    = "repair_schedule"
	WorkRecoverInspectAndRestart = "inspect_and_restart"
	WorkRecoverRestoreCheckpoint = "restore_checkpoint"
	WorkRecoverRollbackAndReview = "rollback_and_review"
)

// WorkLifecycleForKind returns the control contract for a normalized WorkIntent kind.
func WorkLifecycleForKind(kind string) *WorkLifecycleContract {
	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "scheduled":
		return lifecycle(WorkStopDisableSchedule, WorkRetryRunOnceNow, WorkRecoverRepairSchedule,
			"You can disable the schedule, run it once for verification, or repair its cadence and inputs.")
	case "service":
		return lifecycle(WorkStopService, WorkRetryRestartService, WorkRecoverInspectAndRestart,
			"You can stop the service, restart it, or inspect its last trusted state before recovery.")
	case "project":
		return lifecycle(WorkStopPauseProject, WorkRetryResumeProject, WorkRecoverRestoreCheckpoint,
			"You can pause new project work, resume it, or recover from retained outputs and checkpoints.")
	case "self_extension":
		return lifecycle(WorkStopRollbackExtension, WorkRetryRebuildExtension, WorkRecoverRollbackAndReview,
			"You can roll back the extension, rebuild it, or return to the last trusted Soma configuration.")
	default:
		return lifecycle(WorkStopCancelActive, WorkRetryRerun, WorkRecoverReviseAndRerun,
			"You can cancel while it is active, rerun it, or revise the request before recovery.")
	}
}

// NormalizeWorkIntent returns a trimmed copy with a complete lifecycle contract.
func NormalizeWorkIntent(raw *WorkIntent) *WorkIntent {
	if raw == nil {
		return nil
	}
	intent := *raw
	intent.Kind = strings.TrimSpace(strings.ToLower(intent.Kind))
	intent.Objective = strings.TrimSpace(intent.Objective)
	intent.Cadence = strings.TrimSpace(strings.ToLower(intent.Cadence))
	intent.ScheduleSummary = strings.TrimSpace(intent.ScheduleSummary)
	intent.RuntimePosture = strings.TrimSpace(intent.RuntimePosture)
	intent.TargetTeamID = strings.TrimSpace(intent.TargetTeamID)
	intent.BusScope = strings.TrimSpace(intent.BusScope)
	intent.ServiceRefs = compactStrings(intent.ServiceRefs)
	intent.NATSSubjects = compactStrings(intent.NATSSubjects)
	intent.ProjectRef = strings.TrimSpace(intent.ProjectRef)
	if intent.OutputContract != nil {
		outputCopy := *intent.OutputContract
		outputCopy.Shape = strings.TrimSpace(strings.ToLower(outputCopy.Shape))
		outputCopy.PrimaryDeliverable = strings.TrimSpace(outputCopy.PrimaryDeliverable)
		outputCopy.Retention = strings.TrimSpace(strings.ToLower(outputCopy.Retention))
		outputCopy.LaunchHint = strings.TrimSpace(outputCopy.LaunchHint)
		outputCopy.Validation = dedupeStrings(compactStrings(outputCopy.Validation))
		intent.OutputContract = &outputCopy
	}
	defaults := WorkLifecycleForKind(intent.Kind)
	if intent.Lifecycle == nil {
		intent.Lifecycle = defaults
	} else {
		lifecycleCopy := *intent.Lifecycle
		lifecycleCopy.StopAction = strings.TrimSpace(lifecycleCopy.StopAction)
		lifecycleCopy.RetryAction = strings.TrimSpace(lifecycleCopy.RetryAction)
		lifecycleCopy.RecoveryAction = strings.TrimSpace(lifecycleCopy.RecoveryAction)
		lifecycleCopy.ControlSummary = strings.TrimSpace(lifecycleCopy.ControlSummary)
		if lifecycleCopy.StopAction == "" {
			lifecycleCopy.StopAction = defaults.StopAction
		}
		if lifecycleCopy.RetryAction == "" {
			lifecycleCopy.RetryAction = defaults.RetryAction
		}
		if lifecycleCopy.RecoveryAction == "" {
			lifecycleCopy.RecoveryAction = defaults.RecoveryAction
		}
		if lifecycleCopy.ControlSummary == "" {
			lifecycleCopy.ControlSummary = defaults.ControlSummary
		}
		intent.Lifecycle = &lifecycleCopy
	}
	return &intent
}

func lifecycle(stop, retry, recovery, summary string) *WorkLifecycleContract {
	return &WorkLifecycleContract{
		StopAction: stop, RetryAction: retry, RecoveryAction: recovery, ControlSummary: summary,
	}
}
