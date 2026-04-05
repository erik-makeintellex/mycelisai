package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestBuildApprovalPolicy_AutoAllowedForLowRiskAction(t *testing.T) {
	profile := userGovernanceProfile{
		Role:                 "owner",
		CostSensitivity:      "low",
		ReviewStrictness:     "light",
		AutomationTolerance:  "aggressive",
		EscalationPreference: "notify",
	}

	policy := buildApprovalPolicy(profile, []protocol.PlannedToolCall{{Name: "generate_blueprint"}}, nil)
	if policy == nil {
		t.Fatal("expected approval policy")
	}
	if policy.ApprovalRequired {
		t.Fatalf("expected low-risk action to be auto-allowed, got %+v", policy)
	}
	if policy.ApprovalMode != "auto_allowed" {
		t.Fatalf("expected auto_allowed mode, got %+v", policy)
	}
}

func TestBuildApprovalPolicy_RequiresApprovalForHighRiskCapability(t *testing.T) {
	profile := userGovernanceProfile{
		Role:                 "owner",
		CostSensitivity:      "balanced",
		ReviewStrictness:     "standard",
		AutomationTolerance:  "balanced",
		EscalationPreference: "ask",
	}

	policy := buildApprovalPolicy(profile, []protocol.PlannedToolCall{{Name: "publish_signal", Arguments: map[string]any{"subject": "swarm.global.broadcast"}}}, nil)
	if policy == nil {
		t.Fatal("expected approval policy")
	}
	if !policy.ApprovalRequired {
		t.Fatalf("expected approval to be required, got %+v", policy)
	}
	if policy.ApprovalReason != "capability_risk" {
		t.Fatalf("expected capability_risk reason, got %+v", policy)
	}
}

func TestBuildApprovalPolicy_RequiresApprovalForCompanyKnowledgePromotion(t *testing.T) {
	profile := userGovernanceProfile{
		Role:                 "owner",
		CostSensitivity:      "balanced",
		ReviewStrictness:     "standard",
		AutomationTolerance:  "balanced",
		EscalationPreference: "ask",
	}

	policy := buildApprovalPolicy(profile, []protocol.PlannedToolCall{{Name: "promote_deployment_context", Arguments: map[string]any{"source_artifact_id": "ctx-1"}}}, nil)
	if policy == nil {
		t.Fatal("expected approval policy")
	}
	if !policy.ApprovalRequired {
		t.Fatalf("expected approval to be required, got %+v", policy)
	}
	if policy.CapabilityRisk != "high" {
		t.Fatalf("expected high capability risk, got %+v", policy)
	}
}

func TestHandleListAuditLog(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	contextJSON := `{"actor":"Soma","user":"erik","action":"proposal_generated","result_status":"pending","capability_used":"file_output","approval_status":"approval_required","approval_reason":"capability_risk","intent_proof_id":"proof-123"}`
	mock.ExpectQuery("SELECT (.+) FROM log_entries").
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "intent", "source", "message", "timestamp", "context"}).
			AddRow("audit-1", "chat-to-proposal", "admin", "Chat mutation detected", time.Now(), []byte(contextJSON)))

	rr := doRequest(t, http.HandlerFunc(s.handleListAuditLog), "GET", "/api/v1/audit", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatalf("expected ok response, got %+v", resp)
	}
	items, ok := resp.Data.([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected one audit item, got %+v", resp.Data)
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object audit item, got %T", items[0])
	}
	if first["action"] != "proposal_generated" {
		t.Fatalf("expected proposal_generated action, got %+v", first)
	}
	if first["approval_status"] != "approval_required" {
		t.Fatalf("expected approval status in audit item, got %+v", first)
	}
}

func TestBuildAuditRecord(t *testing.T) {
	record := buildAuditRecord(
		"audit-1",
		"chat-to-proposal",
		"admin",
		"Chat mutation detected",
		time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
		[]byte(`{"actor":"Soma","user":"local-user","action":"capability_usage","result_status":"completed","capability_used":"planning","details":{"channel":"workspace"}}`),
	)

	if record.Action != "capability_usage" {
		t.Fatalf("expected capability_usage action, got %+v", record)
	}
	if record.CapabilityUsed != "planning" {
		t.Fatalf("expected planning capability, got %+v", record)
	}
	if record.Details["channel"] != "workspace" {
		t.Fatalf("expected details to include channel, got %+v", record.Details)
	}
}

func TestHandleCancelAction(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectExec("UPDATE intent_proofs SET status = 'cancelled'").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO log_entries").
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doRequest(t, http.HandlerFunc(s.HandleCancelAction), "POST", "/api/v1/intent/cancel-action", `{"intent_proof_id":"11111111-1111-1111-1111-111111111111"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatalf("expected ok response, got %+v", resp)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
}
