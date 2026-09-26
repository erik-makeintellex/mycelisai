package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	approverTestToken = "11111111-1111-1111-1111-111111111111"
	approverTestProof = "22222222-2222-2222-2222-222222222222"
)

// approverTestMinter minted the fixture token (A2b Q3 proposer binding).
var approverTestMinter = standardUserIdentity().UserID

const confirmTokenTxQuery = "(?s)SELECT intent_proof_id, consumed, expires_at, COALESCE\\(purpose.+FROM confirm_tokens WHERE token = \\$1"

func confirmTokenTxRow(consumed bool, mintedBy string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at", "purpose", "binding_digest", "minted_by"}).
		AddRow(approverTestProof, consumed, time.Now().Add(time.Hour), tokenPurposeChatAction, "", mintedBy)
}

func approverTestScope(approval *protocol.ApprovalPolicy) protocol.ScopeValidation {
	return protocol.ScopeValidation{
		Tools: []string{"write_file"},
		PlannedToolCalls: []protocol.PlannedToolCall{{
			Name:      "write_file",
			Arguments: map[string]any{"path": "output/confirmed.txt", "content": "hello world"},
		}},
		CapabilityIDs: []string{"file_output"},
		Approval:      approval,
	}
}

func postureRaisedApproval() *protocol.ApprovalPolicy {
	return applyPostureApprovalFloor(nil, postureIntent("delivery-posture-governed-enterprise"), nil, []string{"write_file"})
}

func capabilityRiskApproval() *protocol.ApprovalPolicy {
	return &protocol.ApprovalPolicy{ApprovalRequired: true, ApprovalReason: "capability_risk", ApprovalMode: "required",
		CapabilityRisk: "high", ApprovalSteps: []string{"operator_review", "future_role_gate"}}
}

// expectTokenAndScope covers the reads/writes up to the approver gate.
func expectTokenAndScope(t *testing.T, mock sqlmock.Sqlmock, scope protocol.ScopeValidation) {
	t.Helper()
	scopeJSON, err := json.Marshal(scope)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery(confirmTokenTxQuery).
		WithArgs(uuid.MustParse(approverTestToken)).
		WillReturnRows(confirmTokenTxRow(false, approverTestMinter))
	mock.ExpectExec("UPDATE confirm_tokens SET consumed = TRUE").
		WithArgs(sqlmock.AnyArg(), uuid.MustParse(approverTestToken)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT scope_validation FROM intent_proofs WHERE id = \\$1").
		WithArgs(uuid.MustParse(approverTestProof)).
		WillReturnRows(sqlmock.NewRows([]string{"scope_validation"}).AddRow(scopeJSON))
}

// expectConfirmSuccess mirrors the inline write_file success path and
// captures every audit context.
func expectConfirmSuccess(t *testing.T, mock sqlmock.Sqlmock, scope protocol.ScopeValidation, audits *[]string) {
	t.Helper()
	mock.MatchExpectationsInOrder(false)
	expectTokenAndScope(t, mock, scope)
	audit := func() {
		mock.ExpectExec("INSERT INTO log_entries").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), auditContextCapture{audits}).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectExec("INSERT INTO mission_runs").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT template_id, resolved_intent").
		WillReturnRows(sqlmock.NewRows([]string{"template_id", "resolved_intent", "audit_event_id"}).AddRow(string(protocol.TemplateChatToProposal), "chat-action", ""))
	mock.ExpectQuery("INSERT INTO execution_contracts").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("33333333-3333-3333-3333-333333333333"))
	audit()
	audit()
	mock.ExpectExec("UPDATE mission_runs SET status").WithArgs(runs.StatusCompleted, sqlmock.AnyArg(), runs.StatusFailed).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO mission_events").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE intent_proofs SET status = 'confirmed'").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	audit()
	audit()
	mock.ExpectQuery("INSERT INTO proof_artifacts").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("44444444-4444-4444-4444-444444444444"))
	mock.ExpectExec("UPDATE execution_contracts").WillReturnResult(sqlmock.NewResult(0, 1))
}

func confirmAs(t *testing.T, s *AdminServer, identity *RequestIdentity) *httptest.ResponseRecorder {
	body := `{"confirm_token":"` + approverTestToken + `"}`
	if identity == nil {
		return doRequest(t, http.HandlerFunc(s.HandleConfirmAction), http.MethodPost, "/api/v1/intent/confirm-action", body)
	}
	return doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleConfirmAction), http.MethodPost, "/api/v1/intent/confirm-action", body, identity)
}

func TestConfirmActionPostureApprovalRequiresApprover(t *testing.T) {
	// Posture-raised on its own, and posture-matched on top of a capability-risk approval.
	raisedOverCapability := applyPostureApprovalFloor(capabilityRiskApproval(), postureIntent("delivery-posture-governed-enterprise"), nil, []string{"write_file"})
	for approvalName, approval := range map[string]*protocol.ApprovalPolicy{"posture": postureRaisedApproval(), "posture-over-capability": raisedOverCapability} {
		for name, identity := range map[string]*RequestIdentity{
			"anonymous":           nil,
			"standard-user":       standardUserIdentity(),
			"operator-with-all":   {UserID: "u-op", Role: "operator", Scopes: []string{"*"}},
			"admin-without-scope": adminWithScopes(scopeGovernanceRead, scopeGovernanceWrite),
		} {
			t.Run(approvalName+"/"+name, func(t *testing.T) {
				workspace := t.TempDir()
				t.Setenv("MYCELIS_WORKSPACE", workspace)
				dbOpt, mock := withDB(t)
				s := newTestServer(dbOpt)
				expectTokenAndScope(t, mock, approverTestScope(approval))
				mock.ExpectRollback() // token consumption is undone: the token stays valid
				rr := confirmAs(t, s, identity)
				want := http.StatusForbidden
				if identity == nil {
					want = http.StatusUnauthorized
				}
				assertStatus(t, rr, want)
				var resp struct {
					OK    bool
					Error string
					Data  struct {
						Code              string                    `json:"code"`
						RecommendedAction string                    `json:"recommended_action"`
						ExecutionSummary  protocol.ExecutionSummary `json:"execution_summary"`
					}
				}
				assertJSON(t, rr, &resp)
				deg := resp.Data.ExecutionSummary.AuditRecovery.Degradation
				if resp.OK || resp.Data.Code != "approver_required" || resp.Data.RecommendedAction == "" || deg == nil || deg.Code != "approver_required" || deg.SafeContinuation == "" {
					t.Fatalf("expected normalized approver blocker, got %s", rr.Body.String())
				}
				if !strings.Contains(strings.ToLower(resp.Error), "admin approval") {
					t.Fatalf("blocker must say it needs admin approval: %q", resp.Error)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("denial must stop before any run/contract/audit write and roll back: %v", err)
				}
				if _, err := os.Stat(filepath.Join(workspace, "output", "confirmed.txt")); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("denied confirm executed the planned tool")
				}
			})
		}
	}
}

func TestConfirmActionPostureApprovalAllowsScopedAdmin(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	var audits []string
	expectConfirmSuccess(t, mock, approverTestScope(postureRaisedApproval()), &audits)
	rr := confirmAs(t, s, adminWithScopes("approvals:*"))
	assertStatus(t, rr, http.StatusOK)
	if !strings.Contains(strings.Join(audits, "\n"), `"approval_authority":"approvals:decide"`) {
		t.Fatalf("confirm audit must record approval_authority=approvals:decide: %v", audits)
	}
	if _, err := os.Stat(filepath.Join(workspace, "output", "confirmed.txt")); err != nil {
		t.Fatalf("approved confirm should execute: %v", err)
	}
}

// A2b Q1-A: external-data, escalation, medium-risk and cost <= 5.0 approvals
// stay proposer-confirmable (tier 1); the proposer is recorded as the authority.
func TestConfirmActionSelfReviewTierStaysProposerConfirmable(t *testing.T) {
	for name, approval := range map[string]*protocol.ApprovalPolicy{
		"external-data": {ApprovalRequired: true, ApprovalReason: "external_data_use", CapabilityRisk: "low", ExternalDataUse: true},
		"escalation":    {ApprovalRequired: true, ApprovalReason: "escalation_preference", CapabilityRisk: "medium"},
		"medium-risk":   {ApprovalRequired: true, ApprovalReason: "capability_risk", CapabilityRisk: "medium"},
		"cost-at-5.0":   {ApprovalRequired: true, ApprovalReason: "cost", CapabilityRisk: "low", EstimatedCost: 5.0},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			var audits []string
			expectConfirmSuccess(t, mock, approverTestScope(approval), &audits)
			rr := confirmAs(t, s, standardUserIdentity())
			assertStatus(t, rr, http.StatusOK)
			joined := strings.Join(audits, "\n")
			for _, want := range []string{`"approval_authority":"proposer"`, `"approval_tier":1`, `"self_approved":true`} {
				if !strings.Contains(joined, want) {
					t.Fatalf("confirm audit must record %s: %v", want, audits)
				}
			}
		})
	}
}

// A2b Q1-A: high/critical capability risk and cost above 5.0 need an approver,
// even for proofs minted before A2b (no role-gate step). The standard user is
// refused and the token kept; an admin approver then confirms it.
func TestConfirmActionApproverTierForHighRiskAndCost(t *testing.T) {
	for name, approval := range map[string]*protocol.ApprovalPolicy{
		"high-risk-pre-a2b": capabilityRiskApproval(),
		"critical-risk":     {ApprovalRequired: true, ApprovalReason: "capability_risk", CapabilityRisk: "critical"},
		"cost-5.01":         {ApprovalRequired: true, ApprovalReason: "cost", CapabilityRisk: "low", EstimatedCost: 5.01},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			scope := approverTestScope(approval)
			expectTokenAndScope(t, mock, scope)
			mock.ExpectRollback()
			rr := confirmAs(t, s, standardUserIdentity())
			assertStatus(t, rr, http.StatusForbidden)
			if !strings.Contains(rr.Body.String(), "approver_required") {
				t.Fatalf("expected approver_required, got %s", rr.Body.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("refused confirm must roll back: %v", err)
			}
			var audits []string
			expectConfirmSuccess(t, mock, scope, &audits)
			assertStatus(t, confirmAs(t, s, adminWithScopes(scopeApprovalsDecide)), http.StatusOK)
			joined := strings.Join(audits, "\n")
			for _, want := range []string{`"approval_authority":"approvals:decide"`, `"approval_tier":2`, `"self_approved":false`} {
				if !strings.Contains(joined, want) {
					t.Fatalf("confirm audit must record %s: %v", want, audits)
				}
			}
		})
	}
}

func TestPostureFloorAddsRoleGateEvenWhenDegraded(t *testing.T) {
	degraded := governance.NewDegradedGuard(errors.New("missing policy"))
	for name, input := range map[string]*protocol.ApprovalPolicy{"none": nil, "capability": capabilityRiskApproval()} {
		got := applyPostureApprovalFloor(input, postureIntent("delivery-posture-lean"), degraded, nil)
		if got == nil || !got.ApprovalRequired || !requiresApprover(&protocol.ScopeValidation{Approval: got}) {
			t.Fatalf("%s: degraded posture work must be required and approver-gated: %#v", name, got)
		}
		for _, step := range got.ApprovalSteps {
			if step == "future_role_gate" {
				t.Fatalf("%s: placeholder step must be replaced: %v", name, got.ApprovalSteps)
			}
		}
	}
	if input := capabilityRiskApproval(); input.ApprovalSteps[1] != "future_role_gate" {
		t.Fatal("the posture floor must not mutate its input")
	}
}
