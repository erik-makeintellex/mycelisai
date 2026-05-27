package server

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestRecoveryActionForUnavailableMediaCapabilityRetry(t *testing.T) {
	item := protocol.TeamWorkItem{
		WorkItemID:             "work-media-1",
		TeamID:                 "media-team",
		RunID:                  "run-123",
		Objective:              "Generate retained media output",
		ExecutionShape:         protocol.TeamExecutionShapeDeliverable,
		ExpectedProof:          []string{"Confirmed execution run", "Retained media output"},
		CapabilityRequirements: []string{"media_generation"},
		GovernancePosture:      protocol.ApprovalPostureRequired,
		State:                  protocol.TeamWorkStateDegraded,
		NeedsOperator:          true,
		DegradationState:       "media_provider_unavailable",
		Version:                "v1",
	}

	action := recoveryActionForTeamWorkItem(item)

	if action.ID != "retry_media_capability" {
		t.Fatalf("action.ID = %q, want retry_media_capability", action.ID)
	}
	if action.ApprovalPosture != protocol.ApprovalPostureRequired {
		t.Fatalf("approval posture = %q, want required", action.ApprovalPosture)
	}
	if action.CapabilityID != "media_generation" {
		t.Fatalf("capability = %q, want media_generation", action.CapabilityID)
	}
	if action.RetryTarget != "run-123" {
		t.Fatalf("retry target = %q, want run-123", action.RetryTarget)
	}
	if action.TargetState != protocol.TeamWorkStateQueued {
		t.Fatalf("target state = %q, want queued", action.TargetState)
	}
	if !containsRecoveryProof(action.ExpectedProof, "Media provider availability proof") {
		t.Fatalf("expected proof = %#v, want media provider availability proof", action.ExpectedProof)
	}
	if !strings.Contains(action.TrustedState, "media_provider_unavailable") {
		t.Fatalf("trusted state = %q, want degradation boundary", action.TrustedState)
	}
}

func TestRecoveryOptionFromActionDescribesRetryContract(t *testing.T) {
	action := recoveryAction{
		ID:              "retry_media_capability",
		Label:           "Retry media capability",
		ApprovalPosture: protocol.ApprovalPostureRequired,
		CapabilityID:    "media_output",
		RetryTarget:     "work-media-2",
		ExpectedProof:   []string{"Retained output reference", "Media provider availability proof"},
		TrustedState:    "Current work item state remains trusted.",
		TargetState:     protocol.TeamWorkStateQueued,
	}

	option := recoveryOptionFromAction(action)

	for _, want := range []string{
		"Retry media_output",
		"work-media-2",
		"after operator approval",
		"Retained output reference",
		"Media provider availability proof",
		"return to queued",
	} {
		if !strings.Contains(option, want) {
			t.Fatalf("option = %q, want substring %q", option, want)
		}
	}
}

func containsRecoveryProof(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
