package server

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

func shippedPostureGuard(t *testing.T) *governance.Guard {
	t.Helper()
	guard, err := governance.NewGuard("../../config/policy.yaml")
	if err != nil {
		t.Fatalf("load shipped policy: %v", err)
	}
	return guard
}

func postureIntent(id string) *protocol.WorkIntent {
	return &protocol.WorkIntent{Kind: "project", OutcomeTemplateSnapshot: &protocol.OutcomeTemplateSnapshot{ID: id, Version: "1.0.0", Digest: "sha256:x"}}
}

func autoAllowedApproval() *protocol.ApprovalPolicy {
	return &protocol.ApprovalPolicy{
		ApprovalMode: "auto_allowed", ApprovalReason: "auto_approve", CapabilityRisk: "low",
		CapabilityIDs: []string{"learning"}, ApprovalSteps: []string{"operator_review", "future_role_gate"},
	}
}

func TestPostureFloorGovernedEnterpriseRaisesAutoAllowedToRequired(t *testing.T) {
	approval := applyPostureApprovalFloor(autoAllowedApproval(), postureIntent("delivery-posture-governed-enterprise"), shippedPostureGuard(t), []string{"remember"})
	if !approval.ApprovalRequired || approval.ApprovalMode != "required" || approval.ApprovalReason != approvalReasonOutcomePosture {
		t.Fatalf("approval = %#v, want required/outcome_posture", approval)
	}
	if policyDecisionForApproval(approval) != "require_approval" || approvalStatusValue(approval) != "approval_required" {
		t.Fatalf("persisted decision = %q / %q", policyDecisionForApproval(approval), approvalStatusValue(approval))
	}
}

func TestPostureFloorKeepsExistingRequiredReason(t *testing.T) {
	cost := &protocol.ApprovalPolicy{ApprovalRequired: true, ApprovalMode: "required", ApprovalReason: "cost", EstimatedCost: 9}
	for _, posture := range []string{"delivery-posture-operations-desk", "delivery-posture-governed-enterprise"} {
		got := applyPostureApprovalFloor(cost, postureIntent(posture), shippedPostureGuard(t), []string{"write_file"})
		if !got.ApprovalRequired || got.ApprovalMode != "required" || got.ApprovalReason != "cost" {
			t.Fatalf("%s: approval = %#v, want unchanged required/cost", posture, got)
		}
	}
}

func TestPostureFloorLeavesUnmatchedWorkUnchanged(t *testing.T) {
	guard := shippedPostureGuard(t)
	cases := map[string]*protocol.WorkIntent{
		"no intent":       nil,
		"no snapshot":     {Kind: "project"},
		"unknown posture": postureIntent("some-operator-template"),
		"operations desk": postureIntent("delivery-posture-operations-desk"),
		"blank id":        postureIntent("  "),
	}
	for name, intent := range cases {
		original := autoAllowedApproval()
		got := applyPostureApprovalFloor(original, intent, guard, []string{"write_file"})
		if !reflect.DeepEqual(got, autoAllowedApproval()) {
			t.Fatalf("%s: approval = %#v, want golden unchanged", name, got)
		}
	}
}

func TestPostureFloorFailsClosedWithoutPolicy(t *testing.T) {
	got := applyPostureApprovalFloor(autoAllowedApproval(), postureIntent("delivery-posture-operations-desk"), nil, []string{"remember"})
	if !got.ApprovalRequired || got.ApprovalReason != approvalReasonOutcomePosture {
		t.Fatalf("nil guard approval = %#v, want required", got)
	}
	created := applyPostureApprovalFloor(nil, postureIntent("delivery-posture-operations-desk"), nil, nil)
	if created == nil || !created.ApprovalRequired || created.ApprovalMode != "required" {
		t.Fatalf("nil guard nil approval = %#v, want required", created)
	}
	if applyPostureApprovalFloor(nil, &protocol.WorkIntent{}, nil, nil) != nil {
		t.Fatal("no snapshot with nil guard must stay unchanged")
	}
}

// TestPostureFloorIsMonotone checks every input rank is preserved or raised.
func TestPostureFloorIsMonotone(t *testing.T) {
	guard := shippedPostureGuard(t)
	rank := func(p *protocol.ApprovalPolicy) int {
		switch {
		case p == nil:
			return 0
		case p.ApprovalRequired:
			return 3
		case p.ApprovalMode == "optional":
			return 2
		default:
			return 1
		}
	}
	inputs := []*protocol.ApprovalPolicy{
		nil, autoAllowedApproval(),
		{ApprovalMode: "optional", ApprovalReason: "capability_risk"},
		{ApprovalRequired: true, ApprovalMode: "required", ApprovalReason: "external_data_use"},
	}
	postures := []string{"delivery-posture-governed-enterprise", "delivery-posture-client-delivery-studio", "delivery-posture-product-delivery-team", "delivery-posture-operations-desk", "other"}
	for _, input := range inputs {
		for _, posture := range postures {
			for _, g := range []*governance.Guard{guard, nil} {
				var before *protocol.ApprovalPolicy
				if input != nil {
					copied := *input
					before = &copied
				}
				got := applyPostureApprovalFloor(before, postureIntent(posture), g, []string{"write_file", "delegate_task"})
				if rank(got) < rank(input) {
					t.Fatalf("posture %s lowered %#v to %#v", posture, input, got)
				}
			}
		}
	}
}

func TestUpdatePolicyRejectsNonStricterPostureGroups(t *testing.T) {
	t.Chdir(t.TempDir())
	s := newTestServer(withGuard(defaultTestPolicyConfig()))
	for _, body := range []string{
		`{"groups":[{"name":"p","targets":["posture:delivery-posture-governed-enterprise"],"rules":[{"intent":"^.*$","action":"ALLOW"}]}],"defaults":{"default_action":"ALLOW"}}`,
		`{"groups":[{"name":"p","targets":["posture:delivery-posture-governed-enterprise"],"rules":[{"intent":"^.*$","action":"DENY"}]}],"defaults":{"default_action":"ALLOW"}}`,
		`{"groups":[{"name":"p","targets":["posture:x-posture"],"rules":[{"intent":"^.*$","condition":"amount > 1","action":"REQUIRE_APPROVAL"}]}],"defaults":{"default_action":"ALLOW"}}`,
		`{"groups":[{"name":"p","targets":["posture:x-posture","team:admin-core"],"rules":[{"intent":"^.*$","action":"REQUIRE_APPROVAL"}]}],"defaults":{"default_action":"ALLOW"}}`,
	} {
		rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleUpdatePolicy), "PUT", "/api/v1/governance/policy", body)
		assertStatus(t, rr, http.StatusBadRequest)
	}
	if got := s.Guard.GetPolicyConfig(); len(got.Groups) != 1 || got.Groups[0].Name != "test-group" {
		t.Fatalf("rejected policy replaced in-memory config: %#v", got)
	}
}
