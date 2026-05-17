package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleConfirmAction_PersistsFailureProofArtifact(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mock.MatchExpectationsInOrder(false)

	token := "41111111-1111-1111-1111-111111111111"
	proofID := "42222222-2222-2222-2222-222222222222"
	tokenUUID := uuid.MustParse(token)
	proofUUID := uuid.MustParse(proofID)
	expiresAt := time.Now().Add(time.Hour)
	scope := protocol.ScopeValidation{
		Tools:             []string{"write_file"},
		PlannedToolCalls:  []protocol.PlannedToolCall{},
		CapabilityIDs:     []string{"file_output"},
		AffectedResources: []string{"workspace"},
		RiskLevel:         "medium",
	}
	scopeJSON, err := json.Marshal(scope)
	if err != nil {
		t.Fatalf("marshal scope: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intent_proof_id, consumed, expires_at FROM confirm_tokens WHERE token = \\$1").
		WithArgs(tokenUUID).
		WillReturnRows(sqlmock.NewRows([]string{"intent_proof_id", "consumed", "expires_at"}).
			AddRow(proofID, false, expiresAt))
	mock.ExpectExec("UPDATE confirm_tokens SET consumed = TRUE, consumed_at = \\$1 WHERE token = \\$2 AND consumed = FALSE").
		WithArgs(sqlmock.AnyArg(), tokenUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT scope_validation FROM intent_proofs WHERE id = \\$1").
		WithArgs(proofUUID).
		WillReturnRows(sqlmock.NewRows([]string{"scope_validation"}).AddRow(scopeJSON))
	mock.ExpectExec("INSERT INTO mission_runs \\(id, mission_id, tenant_id, status, run_depth, started_at\\)").
		WithArgs(sqlmock.AnyArg(), proofID, "default", runs.StatusRunning, 0, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT template_id, resolved_intent, COALESCE\\(audit_event_id::text, ''\\) FROM intent_proofs WHERE id = \\$1").
		WithArgs(proofUUID).
		WillReturnRows(sqlmock.NewRows([]string{"template_id", "resolved_intent", "audit_event_id"}).
			AddRow(string(protocol.TemplateChatToProposal), "chat-action", ""))
	mock.ExpectQuery("INSERT INTO execution_contracts").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("43333333-3333-3333-3333-333333333333"))
	mock.ExpectExec("UPDATE intent_proofs SET status = 'failed' WHERE id = \\$1").
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE mission_runs SET status = \\$1, completed_at = NOW\\(\\) WHERE id = \\$2").
		WithArgs(runs.StatusFailed, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO mission_events").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "default", string(protocol.EventMissionFailed), string(protocol.SeverityError), "admin", "governance", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec("INSERT INTO log_entries \\(id, trace_id, timestamp, level, source, intent, message, context\\)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", "confirm-action", string(protocol.TemplateChatToProposal), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO proof_artifacts").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("44444444-4444-4444-4444-444444444444"))
	mock.ExpectExec("UPDATE execution_contracts").
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doRequest(t, http.HandlerFunc(s.HandleConfirmAction), http.MethodPost, "/api/v1/intent/confirm-action", `{"confirm_token":"`+token+`"}`)
	assertStatus(t, rr, http.StatusInternalServerError)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if resp.OK {
		t.Fatalf("expected ok=false")
	}
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data map, got %T", resp.Data)
	}
	if data["contract_id"] != "43333333-3333-3333-3333-333333333333" {
		t.Fatalf("contract_id = %v", data["contract_id"])
	}
	if data["proof_artifact_id"] != "44444444-4444-4444-4444-444444444444" {
		t.Fatalf("proof_artifact_id = %v", data["proof_artifact_id"])
	}
	summary := data["execution_summary"].(map[string]any)
	proof := summary["proof"].(map[string]any)
	if proof["proof_id"] != data["proof_artifact_id"] || proof["contract_id"] != data["contract_id"] {
		t.Fatalf("failure proof linkage = %+v", proof)
	}
	auditRecovery := summary["audit_recovery"].(map[string]any)
	if auditRecovery["recovery_state"] != "failed" {
		t.Fatalf("audit recovery = %+v", auditRecovery)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet db expectations: %v", err)
	}
}
