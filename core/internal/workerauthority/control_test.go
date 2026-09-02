package workerauthority

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestStageApprovalExactReplayAndConflict(t *testing.T) {
	request := ApprovalRequest{
		ID: "31111111-1111-1111-1111-111111111111", RunID: testBinding().RunID,
		ApprovalID: "approval-1", RequestDigest: testDigest, Kind: "tool",
		Summary: "Use approved tool", RiskLevel: "low", Action: "invoke",
	}
	for _, conflict := range []bool{false, true} {
		db, mock, _ := sqlmock.New()
		mock.ExpectBegin()
		tx, _ := db.BeginTx(t.Context(), nil)
		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO worker_approval_requests")).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		summary := request.Summary
		if conflict {
			summary = "different"
		}
		mock.ExpectQuery("SELECT id::text").WillReturnRows(sqlmock.NewRows([]string{
			"id", "run_id", "approval_id", "request_digest", "kind", "summary", "risk_level", "requested_action",
		}).AddRow(request.ID, request.RunID, request.ApprovalID, request.RequestDigest,
			request.Kind, summary, request.RiskLevel, request.Action))
		created, err := NewStore(db).StageApprovalTx(t.Context(), tx, request)
		if conflict && (err != ErrConflict || created) {
			t.Fatalf("conflicting approval replay = %v, %v", created, err)
		}
		if !conflict && (err != nil || created) {
			t.Fatalf("exact approval replay = %v, %v", created, err)
		}
		mock.ExpectRollback()
		_ = tx.Rollback()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		db.Close()
	}
}

func TestStageCommandReplayUsesSemanticJSONAndRejectsConflict(t *testing.T) {
	command := ControlCommand{
		CommandID: "41111111-1111-1111-1111-111111111111", RunID: testBinding().RunID,
		ApprovalRequestID: "31111111-1111-1111-1111-111111111111",
		IdempotencyKey:    "approval-command-1", Kind: "approve",
		ExpectedServiceVersion: 2, PayloadDigest: testDigest,
		Payload: []byte(`{"decision":"approve","reason":"ok"}`),
	}
	for _, conflict := range []bool{false, true} {
		db, mock, _ := sqlmock.New()
		mock.ExpectBegin()
		tx, _ := db.BeginTx(t.Context(), nil)
		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO worker_control_commands")).
			WillReturnRows(sqlmock.NewRows([]string{"command_id"}))
		kind := command.Kind
		if conflict {
			kind = "deny"
		}
		mock.ExpectQuery("SELECT command_id::text").WillReturnRows(sqlmock.NewRows([]string{
			"command_id", "run_id", "approval_id", "key", "kind", "expected", "digest", "payload",
		}).AddRow(command.CommandID, command.RunID, command.ApprovalRequestID,
			command.IdempotencyKey, kind, command.ExpectedServiceVersion,
			command.PayloadDigest, []byte(`{"reason":"ok","decision":"approve"}`)))
		created, err := NewStore(db).StageCommandTx(t.Context(), tx, command)
		if conflict && (err != ErrConflict || created) {
			t.Fatalf("conflicting command replay = %v, %v", created, err)
		}
		if !conflict && (err != nil || created) {
			t.Fatalf("semantic command replay = %v, %v", created, err)
		}
		mock.ExpectRollback()
		_ = tx.Rollback()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		db.Close()
	}
}
