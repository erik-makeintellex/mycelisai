package protocol

import "testing"

func TestWorkLifecycleForKind(t *testing.T) {
	tests := []struct {
		kind     string
		stop     string
		retry    string
		recovery string
	}{
		{"one_shot", WorkStopCancelActive, WorkRetryRerun, WorkRecoverReviseAndRerun},
		{"scheduled", WorkStopDisableSchedule, WorkRetryRunOnceNow, WorkRecoverRepairSchedule},
		{"service", WorkStopService, WorkRetryRestartService, WorkRecoverInspectAndRestart},
		{"project", WorkStopPauseProject, WorkRetryResumeProject, WorkRecoverRestoreCheckpoint},
		{"self_extension", WorkStopRollbackExtension, WorkRetryRebuildExtension, WorkRecoverRollbackAndReview},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			contract := WorkLifecycleForKind(tt.kind)
			if contract.StopAction != tt.stop || contract.RetryAction != tt.retry || contract.RecoveryAction != tt.recovery {
				t.Fatalf("unexpected lifecycle for %s: %#v", tt.kind, contract)
			}
			if contract.ControlSummary == "" {
				t.Fatalf("expected operator control summary for %s", tt.kind)
			}
		})
	}
}

func TestNormalizeWorkIntentAddsLifecycleWithoutOverwritingExplicitContract(t *testing.T) {
	normalized := NormalizeWorkIntent(&WorkIntent{Kind: " Scheduled ", Objective: "  Review incidents  "})
	if normalized.Kind != "scheduled" || normalized.Objective != "Review incidents" {
		t.Fatalf("expected normalized fields, got %#v", normalized)
	}
	if normalized.Lifecycle == nil || normalized.Lifecycle.StopAction != WorkStopDisableSchedule {
		t.Fatalf("expected scheduled lifecycle, got %#v", normalized.Lifecycle)
	}

	explicit := NormalizeWorkIntent(&WorkIntent{
		Kind: "service",
		Lifecycle: &WorkLifecycleContract{
			StopAction: " custom_stop ", ControlSummary: " Keep the operator override. ",
		},
	})
	if explicit.Lifecycle.StopAction != "custom_stop" || explicit.Lifecycle.ControlSummary != "Keep the operator override." {
		t.Fatalf("expected explicit lifecycle to be retained, got %#v", explicit.Lifecycle)
	}
}
