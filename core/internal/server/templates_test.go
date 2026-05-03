package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleConfirmAction_CompletesVerifiedExecutionWithPlannedToolCalls(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mock.MatchExpectationsInOrder(false)

	token := "11111111-1111-1111-1111-111111111111"
	proofID := "22222222-2222-2222-2222-222222222222"
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
			{
				Name: "write_file",
				Arguments: map[string]any{
					"path":    "output/confirmed.txt",
					"content": "hello world",
				},
			},
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
	mock.ExpectExec("INSERT INTO log_entries \\(id, trace_id, timestamp, level, source, intent, message, context\\)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", "confirm-action", string(protocol.TemplateChatToProposal), sqlmock.AnyArg(), sqlmock.AnyArg()).
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

	reqBody := `{"confirm_token":"` + token + `"}`
	rr := doRequest(t, http.HandlerFunc(s.HandleConfirmAction), http.MethodPost, "/api/v1/intent/confirm-action", reqBody)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatalf("expected ok=true, got error=%q", resp.Error)
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data map, got %T", resp.Data)
	}
	if data["confirmed"] != true {
		t.Fatalf("confirmed = %v, want true", data["confirmed"])
	}
	if data["verified"] != true {
		t.Fatalf("verified = %v, want true", data["verified"])
	}
	if data["execution_state"] != "verified" {
		t.Fatalf("execution_state = %v, want verified", data["execution_state"])
	}
	runID, _ := data["run_id"].(string)
	if strings.TrimSpace(runID) == "" {
		t.Fatal("expected non-empty run_id")
	}
	if data["run_status"] != runs.StatusCompleted {
		t.Fatalf("run_status = %v, want %s", data["run_status"], runs.StatusCompleted)
	}
	auditID, _ := data["audit_event_id"].(string)
	if strings.TrimSpace(auditID) == "" {
		t.Fatal("expected non-empty audit_event_id")
	}

	written, err := os.ReadFile(filepath.Join(workspace, "output", "confirmed.txt"))
	if err != nil {
		t.Fatalf("read confirmed file: %v", err)
	}
	if got := strings.TrimSpace(string(written)); got != "hello world" {
		t.Fatalf("file content = %q, want hello world", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet db expectations: %v", err)
	}
}

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
	mock.ExpectExec("INSERT INTO log_entries \\(id, trace_id, timestamp, level, source, intent, message, context\\)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", "confirm-action", string(protocol.TemplateChatToProposal), sqlmock.AnyArg(), sqlmock.AnyArg()).
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

	reqBody := `{"confirm_token":"` + token + `"}`
	rr := doRequest(t, http.HandlerFunc(s.HandleConfirmAction), http.MethodPost, "/api/v1/intent/confirm-action", reqBody)
	assertStatus(t, rr, http.StatusOK)

	if _, err := os.Stat(filepath.Join(workspace, "output", "confirmed_alias.txt")); err != nil {
		t.Fatalf("expected normalized write_file target to exist: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}
