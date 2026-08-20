package protocol

import (
	"strings"
	"testing"
	"time"
)

func TestApplyTeamWorkAction_RequiresExternalMutationVerification(t *testing.T) {
	item := NormalizeTeamWorkItem(TeamWorkItem{
		TeamID:           "delivery-team",
		Objective:        "Update the customer system",
		ExecutionShape:   TeamExecutionShapeDeliverable,
		State:            TeamWorkStateDegraded,
		DegradationState: "external_mutation_outcome_unknown",
		WorkIntent: &WorkIntent{SideEffect: &WorkSideEffectContract{
			EffectKind:      WorkEffectExternalMutation,
			RetrySafety:     WorkRetrySafe,
			IdempotencyKey:  "customer-update-1",
			SideEffectState: WorkSideEffectUnknown,
		}},
	})

	for _, action := range []TeamWorkAction{
		TeamWorkActionRecover,
		TeamWorkActionPause,
		TeamWorkActionResume,
		TeamWorkActionStartWork,
	} {
		t.Run(string(action), func(t *testing.T) {
			got, err := ApplyTeamWorkAction(item, action)
			if err == nil || !strings.Contains(err.Error(), "must be verified through Soma") {
				t.Fatalf("ApplyTeamWorkAction error = %v", err)
			}
			if got != TeamWorkStateDegraded {
				t.Fatalf("state = %q, want degraded", got)
			}
		})
	}

	for _, action := range []TeamWorkAction{TeamWorkActionSteer, TeamWorkActionArchive} {
		t.Run(string(action), func(t *testing.T) {
			if _, err := ApplyTeamWorkAction(item, action); err != nil {
				t.Fatalf("ApplyTeamWorkAction(%s): %v", action, err)
			}
		})
	}
	if _, err := ApplyTeamWorkAction(item, TeamWorkActionVerifyExternalOutcome); err != nil {
		t.Fatalf("verify_external_outcome: %v", err)
	}
}

func TestApplyExternalOutcomeVerification(t *testing.T) {
	recordedAt := time.Date(2026, time.August, 20, 12, 30, 0, 0, time.UTC)
	base := NormalizeTeamWorkItem(TeamWorkItem{
		TeamID:           "delivery-team",
		Objective:        "Update the customer system",
		ExecutionShape:   TeamExecutionShapeDeliverable,
		State:            TeamWorkStateDegraded,
		NeedsOperator:    true,
		DegradationState: TeamWorkDegradationExternalMutationUnknown,
		WorkIntent: &WorkIntent{SideEffect: &WorkSideEffectContract{
			EffectKind: WorkEffectExternalMutation, RetrySafety: WorkRetrySafe,
			IdempotencyKey: "customer-update-1", SideEffectState: WorkSideEffectUnknown,
		}},
	})
	cases := []struct {
		name            string
		result          string
		wantState       TeamWorkState
		wantSideEffect  string
		wantDegradation string
		wantOperator    bool
	}{
		{"committed", WorkExternalOutcomeCommitted, TeamWorkStateOutputReady, WorkSideEffectCommitted, "", false},
		{"not committed", WorkExternalOutcomeNotCommitted, TeamWorkStateDegraded, WorkSideEffectVerifiedNotCommitted, TeamWorkDegradationExternalMutationVerifiedNotCommitted, true},
		{"still unknown", WorkExternalOutcomeStillUnknown, TeamWorkStateDegraded, WorkSideEffectUnknown, TeamWorkDegradationExternalMutationUnknown, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ApplyExternalOutcomeVerification(base, WorkExternalOutcomeVerification{
				Result: tc.result, ActorRef: " operator@example.com ", Summary: " Checked the customer record. ",
				EvidenceRefs: []string{" external://customer/42 ", "external://customer/42"}, RecordedAt: recordedAt,
			})
			if err != nil {
				t.Fatalf("ApplyExternalOutcomeVerification: %v", err)
			}
			if got.State != tc.wantState || got.WorkIntent.SideEffect.SideEffectState != tc.wantSideEffect {
				t.Fatalf("state/side effect = %q/%q, want %q/%q", got.State, got.WorkIntent.SideEffect.SideEffectState, tc.wantState, tc.wantSideEffect)
			}
			if got.DegradationState != tc.wantDegradation || got.NeedsOperator != tc.wantOperator {
				t.Fatalf("degradation/operator = %q/%t, want %q/%t", got.DegradationState, got.NeedsOperator, tc.wantDegradation, tc.wantOperator)
			}
			verification := got.WorkIntent.SideEffect.Verification
			if verification == nil || verification.ActorRef != "operator@example.com" || verification.Summary != "Checked the customer record." {
				t.Fatalf("verification = %#v", verification)
			}
			if len(verification.EvidenceRefs) != 1 || verification.RecordedAt != recordedAt {
				t.Fatalf("verification evidence/time = %#v/%v", verification.EvidenceRefs, verification.RecordedAt)
			}
		})
	}
}

func TestApplyExternalOutcomeVerification_AllowsObservationWithoutEvidenceRef(t *testing.T) {
	item := NormalizeTeamWorkItem(TeamWorkItem{
		TeamID: "delivery-team", Objective: "Update the customer system",
		ExecutionShape: TeamExecutionShapeDeliverable, State: TeamWorkStateDegraded,
		DegradationState: TeamWorkDegradationExternalMutationUnknown,
		WorkIntent: &WorkIntent{SideEffect: &WorkSideEffectContract{
			EffectKind: WorkEffectExternalMutation, SideEffectState: WorkSideEffectUnknown,
		}},
	})
	got, err := ApplyExternalOutcomeVerification(item, WorkExternalOutcomeVerification{
		Result: WorkExternalOutcomeStillUnknown, ActorRef: "operator",
		Summary: "The provider has no queryable receipt yet.", RecordedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("ApplyExternalOutcomeVerification: %v", err)
	}
	if got.WorkIntent.SideEffect.Verification == nil || len(got.WorkIntent.SideEffect.Verification.EvidenceRefs) != 0 {
		t.Fatalf("verification = %#v", got.WorkIntent.SideEffect.Verification)
	}
}

func TestApplyTeamWorkAction_RejectsVerificationBeforeOutcomeIsUnknown(t *testing.T) {
	item := NormalizeTeamWorkItem(TeamWorkItem{
		TeamID: "delivery-team", Objective: "Update the customer system",
		ExecutionShape: TeamExecutionShapeDeliverable, State: TeamWorkStateQueued,
		WorkIntent: &WorkIntent{SideEffect: &WorkSideEffectContract{
			EffectKind: WorkEffectExternalMutation, SideEffectState: WorkSideEffectNotStarted,
		}},
	})
	if _, err := ApplyTeamWorkAction(item, TeamWorkActionVerifyExternalOutcome); err == nil ||
		!strings.Contains(err.Error(), "requires an unknown") {
		t.Fatalf("ApplyTeamWorkAction error = %v", err)
	}
}

func TestApplyTeamWorkAction_VerifiedNotCommittedRequiresNewProposal(t *testing.T) {
	item := NormalizeTeamWorkItem(TeamWorkItem{
		TeamID: "delivery-team", Objective: "Update the customer system",
		ExecutionShape: TeamExecutionShapeDeliverable, State: TeamWorkStateDegraded,
		DegradationState: TeamWorkDegradationExternalMutationVerifiedNotCommitted,
		WorkIntent: &WorkIntent{SideEffect: &WorkSideEffectContract{
			EffectKind: WorkEffectExternalMutation, SideEffectState: WorkSideEffectVerifiedNotCommitted,
		}},
	})
	for _, action := range []TeamWorkAction{TeamWorkActionStartWork, TeamWorkActionPause, TeamWorkActionResume, TeamWorkActionRecover} {
		if _, err := ApplyTeamWorkAction(item, action); err == nil || !strings.Contains(err.Error(), "new governed Soma proposal") {
			t.Fatalf("ApplyTeamWorkAction(%s) error = %v", action, err)
		}
	}
	for _, action := range []TeamWorkAction{TeamWorkActionSteer, TeamWorkActionArchive} {
		if _, err := ApplyTeamWorkAction(item, action); err != nil {
			t.Fatalf("ApplyTeamWorkAction(%s): %v", action, err)
		}
	}
}

func TestWorkIntentAllowsIdempotentRetry_RequiresNotStartedSideEffect(t *testing.T) {
	intent := NormalizeWorkIntent(&WorkIntent{SideEffect: &WorkSideEffectContract{
		EffectKind: WorkEffectExternalMutation, RetrySafety: WorkRetrySafe,
		IdempotencyKey: "customer-update-1", SideEffectState: WorkSideEffectNotStarted,
	}})
	if !WorkIntentAllowsIdempotentRetry(intent) {
		t.Fatal("expected not-started idempotent mutation to permit retry planning")
	}
	intent.SideEffect.SideEffectState = WorkSideEffectVerifiedNotCommitted
	if WorkIntentAllowsIdempotentRetry(intent) {
		t.Fatal("verified-not-committed work must require a new governed proposal")
	}
	intent.SideEffect.SideEffectState = WorkSideEffectUnknown
	if WorkIntentAllowsIdempotentRetry(intent) {
		t.Fatal("unknown external outcome must not permit retry")
	}
}
