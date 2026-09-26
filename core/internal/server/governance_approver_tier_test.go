package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

// A2b item 1: every approval reason x risk x cost at the 5.0 ceiling.
func TestApproverTierMatrix(t *testing.T) {
	reasons := []string{"auto_approve", "capability_risk", "external_data_use", "cost", "escalation_preference", approvalReasonOutcomePosture}
	for _, reason := range reasons {
		for _, risk := range []string{"low", "medium", "high", "critical", "high-risk"} {
			for _, cost := range []float64{0, 5.0, 5.01} {
				for _, required := range []bool{false, true} {
					approval := &protocol.ApprovalPolicy{ApprovalRequired: required, ApprovalReason: reason, CapabilityRisk: risk, EstimatedCost: cost}
					got, _ := approverTier(&protocol.ScopeValidation{Approval: approval})
					want := approverTierAuto
					switch {
					case reason == approvalReasonOutcomePosture, risk != "low" && risk != "medium", cost > approverCostCeiling:
						want = approverTierApprover
					case required:
						want = approverTierSelf
					}
					if got != want {
						t.Errorf("reason=%s risk=%s cost=%v required=%v: tier %d, want %d", reason, risk, cost, required, got, want)
					}
				}
			}
		}
	}
	if tier, _ := approverTier(nil); tier != approverTierAuto {
		t.Fatal("nil scope is tier 0")
	}
	gated := &protocol.ApprovalPolicy{ApprovalSteps: []string{approvalStepRoleGate}}
	if tier, _ := approverTier(&protocol.ScopeValidation{Approval: gated}); tier != approverTierApprover {
		t.Fatal("a role-gate step is always tier 2")
	}
	// Scope-level cost counts too (stored proofs carry it on both).
	if tier, _ := approverTier(&protocol.ScopeValidation{EstimatedCost: 7, Approval: &protocol.ApprovalPolicy{ApprovalRequired: true}}); tier != approverTierApprover {
		t.Fatal("scope estimated cost above the ceiling is tier 2")
	}
}

func TestApplyApproverTierIsMonotone(t *testing.T) {
	if applyApproverTier(nil) != nil {
		t.Fatal("nil stays nil")
	}
	medium := &protocol.ApprovalPolicy{ApprovalRequired: true, ApprovalReason: "capability_risk", CapabilityRisk: "medium", ApprovalSteps: []string{"operator_review", "future_role_gate"}}
	if got := applyApproverTier(medium); got != medium {
		t.Fatal("tier-1 approvals are returned unchanged")
	}
	high := capabilityRiskApproval()
	got := applyApproverTier(high)
	if got == high || !got.ApprovalRequired || got.RequiredApproverRole != "admin" || !strings.Contains(strings.Join(got.ApprovalSteps, ","), approvalStepRoleGate) {
		t.Fatalf("tier-2 approval must gain the role gate and admin role: %+v", got)
	}
	if high.RequiredApproverRole == "admin" || strings.Contains(strings.Join(high.ApprovalSteps, ","), approvalStepRoleGate) {
		t.Fatal("input must not be mutated")
	}
	if again := applyApproverTier(got); again.ApprovalReason != got.ApprovalReason || !requiresApprover(&protocol.ScopeValidation{Approval: again}) {
		t.Fatal("re-applying keeps tier 2")
	}
}

// Q2-A: an admin approver may confirm their own tier-2 proposal; the audit
// records self_approved=true.
func TestConfirmActionAdminSelfApprovalIsAudited(t *testing.T) {
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	admin := adminWithScopes(scopeApprovalsDecide)
	previous := approverTestMinter
	approverTestMinter = admin.UserID
	t.Cleanup(func() { approverTestMinter = previous })
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	var audits []string
	expectConfirmSuccess(t, mock, approverTestScope(capabilityRiskApproval()), &audits)
	assertStatus(t, confirmAs(t, s, admin), http.StatusOK)
	joined := strings.Join(audits, "\n")
	for _, want := range []string{`"approval_authority":"approvals:decide"`, `"approval_tier":2`, `"self_approved":true`} {
		if !strings.Contains(joined, want) {
			t.Fatalf("audit must record %s: %v", want, audits)
		}
	}
}
