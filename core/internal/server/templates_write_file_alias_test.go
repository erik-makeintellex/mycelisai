package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleConfirmAction_NormalizesWriteFileAliasesInStoredPlan(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mock.MatchExpectationsInOrder(false)

	token := "31111111-1111-1111-1111-111111111111"
	proofID := "32222222-2222-2222-2222-222222222222"
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		t.Fatalf("parse token uuid: %v", err)
	}
	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		t.Fatalf("parse proof uuid: %v", err)
	}
	expiresAt := time.Now().Add(time.Hour)
	scope := protocol.ScopeValidation{
		Tools: []string{"write_file"},
		PlannedToolCalls: []protocol.PlannedToolCall{
			normalizePlannedToolCall(protocol.PlannedToolCall{
				Name: "write_file",
				Arguments: map[string]any{
					"file_path": "output/confirmed_alias.txt",
					"body":      "hello alias",
				},
			}),
		},
		CapabilityIDs: []string{"write_file"},
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
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("33333333-3333-3333-3333-333333333333"))
	mock.ExpectExec("INSERT INTO log_entries \\(id, trace_id, timestamp, level, source, intent, message, context\\)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", "confirm-action", string(protocol.TemplateChatToProposal), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO proof_artifacts").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("44444444-4444-4444-4444-444444444444"))
	mock.ExpectExec("UPDATE execution_contracts").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE mission_runs SET status = \\$1, completed_at = NOW\\(\\) WHERE id = \\$2").
		WithArgs(runs.StatusCompleted, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO mission_events").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "default", string(protocol.EventMissionCompleted), string(protocol.SeverityInfo), "admin", "governance", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE intent_proofs SET status = 'confirmed', confirmed_at = \\$1 WHERE id = \\$2").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec("INSERT INTO log_entries \\(id, trace_id, timestamp, level, source, intent, message, context\\)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", "confirm-action", string(protocol.TemplateChatToProposal), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO log_entries \\(id, trace_id, timestamp, level, source, intent, message, context\\)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", "confirm-action", string(protocol.TemplateChatToProposal), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doRequest(t, http.HandlerFunc(s.HandleConfirmAction), http.MethodPost, "/api/v1/intent/confirm-action", `{"confirm_token":"`+token+`"}`)
	assertStatus(t, rr, http.StatusOK)

	if _, err := os.Stat(filepath.Join(workspace, "output", "confirmed_alias.txt")); err != nil {
		t.Fatalf("expected normalized write_file target to exist: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}
