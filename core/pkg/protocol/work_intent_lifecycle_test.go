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
		OutputContract: &WorkOutputContract{
			Shape: " App_Package ", PrimaryDeliverable: " dist/app.zip ",
			Retention: " User_Deliverable ", LaunchHint: " Open index.html ",
			Validation: []string{" playable ", "", "playable", " audio "},
		},
		Lifecycle: &WorkLifecycleContract{
			StopAction: " custom_stop ", ControlSummary: " Keep the operator override. ",
		},
	})
	if explicit.Lifecycle.StopAction != "custom_stop" || explicit.Lifecycle.ControlSummary != "Keep the operator override." {
		t.Fatalf("expected explicit lifecycle to be retained, got %#v", explicit.Lifecycle)
	}
	if explicit.Lifecycle.RetryAction != WorkRetryRestartService || explicit.Lifecycle.RecoveryAction != WorkRecoverInspectAndRestart {
		t.Fatalf("expected missing lifecycle controls to use service defaults, got %#v", explicit.Lifecycle)
	}
	if explicit.OutputContract.Shape != "app_package" || explicit.OutputContract.Retention != "user_deliverable" {
		t.Fatalf("expected output contract enums to be normalized, got %#v", explicit.OutputContract)
	}
	if explicit.OutputContract.PrimaryDeliverable != "dist/app.zip" || explicit.OutputContract.LaunchHint != "Open index.html" {
		t.Fatalf("expected output contract copy to be trimmed, got %#v", explicit.OutputContract)
	}
	if len(explicit.OutputContract.Validation) != 2 || explicit.OutputContract.Validation[1] != "audio" {
		t.Fatalf("expected compact validation requirements, got %#v", explicit.OutputContract.Validation)
	}
}
