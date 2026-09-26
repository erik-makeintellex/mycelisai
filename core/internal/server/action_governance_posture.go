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
// snapshot. With no loaded policy (guard nil), posture-scoped work requires
// approval. Approval is not role-gated yet (future_role_gate).
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
	raised := protocol.ApprovalPolicy{ApprovalSteps: []string{"operator_review", "future_role_gate"}}
	if approval != nil {
		raised = *approval
	}
	if !raised.ApprovalRequired {
		raised.ApprovalReason = approvalReasonOutcomePosture
	}
	raised.ApprovalRequired = true
	raised.ApprovalMode = "required"
	return &raised
}
