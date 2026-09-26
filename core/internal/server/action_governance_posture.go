package server

import (
	"strings"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

// approvalReasonOutcomePosture marks a decision raised by a posture group.
const approvalReasonOutcomePosture = "outcome_posture"

// applyPostureApprovalFloor raises the approval decision when the work was
// shaped by an Outcome Template whose posture group requires approval. It is
// monotone: it can only set approval to required and never clears or lowers
// anything. The posture id comes only from the server-compiled WorkIntent
// snapshot. With no loaded (nil or degraded) policy, posture-scoped work
// requires approval. A raised policy carries approvalStepRoleGate, so only a
// root admin with approvals:decide can confirm it (requireApprover).
func applyPostureApprovalFloor(
	approval *protocol.ApprovalPolicy,
	intent *protocol.WorkIntent,
	guard *governance.Guard,
	tools []string,
) *protocol.ApprovalPolicy {
	if intent == nil || intent.OutcomeTemplateSnapshot == nil {
		return approval
	}
	postureID := strings.TrimSpace(intent.OutcomeTemplateSnapshot.ID)
	if postureID == "" {
		return approval
	}
	required := true
	if guard != nil {
		var capabilityIDs []string
		if approval != nil {
			capabilityIDs = approval.CapabilityIDs
		}
		required, _ = guard.PostureRequiresApproval(postureID, tools, capabilityIDs)
	}
	if !required {
		return approval
	}
	raised := protocol.ApprovalPolicy{ApprovalSteps: []string{"operator_review"}}
	if approval != nil {
		raised = *approval
	}
	raised.ApprovalSteps = withRoleGateStep(raised.ApprovalSteps)
	if !raised.ApprovalRequired {
		raised.ApprovalReason = approvalReasonOutcomePosture
	}
	raised.ApprovalRequired = true
	raised.ApprovalMode = "required"
	return &raised
}

// approvalStepRoleGate marks an approval that only an approver may confirm.
const approvalStepRoleGate = "role_gate:" + scopeApprovalsDecide

// withRoleGateStep replaces the placeholder future_role_gate step with the
// real role gate (or appends it) without mutating the caller's slice.
func withRoleGateStep(steps []string) []string {
	out := make([]string, 0, len(steps)+1)
	for _, step := range steps {
		if step != "future_role_gate" && step != approvalStepRoleGate {
			out = append(out, step)
		}
	}
	return append(out, approvalStepRoleGate)
}
