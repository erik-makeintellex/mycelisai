package triggers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestLogExecution_WithHandoffRefs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectExec("INSERT INTO trigger_executions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	exec := &TriggerExecution{
		RuleID:         "r-1",
		EventID:        "schedule:r-1:2026-05-30T10:00:00Z",
		Status:         "proposed",
		HandoffKey:     "schedule:r-1:2026-05-30T10:00:00Z",
		IntentProofID:  "11111111-1111-1111-1111-111111111111",
		ContractID:     "22222222-2222-2222-2222-222222222222",
		ProposalStatus: "awaiting_approval",
		HandoffPayload: json.RawMessage(`{"autonomous_execution":false}`),
	}

	if err := s.LogExecution(context.Background(), exec); err != nil {
		t.Fatalf("LogExecution error: %v", err)
	}
}

func TestGetExecutionByHandoffKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(execColumns).
			AddRow("exec-1", "r-1", "schedule:r-1:due", "", "proposed", "", "schedule:r-1:due",
				"11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222",
				"awaiting_approval", []byte(`{"autonomous_execution":false}`), now))

	exec, err := s.GetExecutionByHandoffKey(context.Background(), "r-1", "schedule:r-1:due")
	if err != nil {
		t.Fatalf("GetExecutionByHandoffKey error: %v", err)
	}
	if exec == nil || exec.IntentProofID == "" || exec.ContractID == "" {
		t.Fatalf("expected handoff refs, got %#v", exec)
	}
}

func TestTransitionScheduleHandoffApproval_ApprovesNoRunHandoff(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	mock.ExpectQuery("UPDATE trigger_executions AS e").
		WillReturnRows(sqlmock.NewRows(execColumns).
			AddRow("exec-1", "r-1", "schedule:r-1:due", "", "proposed", "",
				"schedule:r-1:due", "11111111-1111-1111-1111-111111111111",
				"22222222-2222-2222-2222-222222222222", "approved",
				[]byte(`{"approval_state":"approved","autonomous_execution":false}`), now))

	exec, err := s.TransitionScheduleHandoffApproval(context.Background(), "r-1", "exec-1", "approved")
	if err != nil {
		t.Fatalf("TransitionScheduleHandoffApproval error: %v", err)
	}
	if exec.ProposalStatus != "approved" {
		t.Fatalf("proposal_status = %q, want approved", exec.ProposalStatus)
	}
	if exec.RunID != "" {
		t.Fatalf("run_id = %q, want empty", exec.RunID)
	}
	if string(exec.HandoffPayload) == "" {
		t.Fatal("expected updated handoff payload")
	}
	var payload map[string]any
	if err := json.Unmarshal(exec.HandoffPayload, &payload); err != nil {
		t.Fatalf("handoff payload JSON: %v", err)
	}
	if payload["autonomous_execution"] != false {
		t.Fatalf("autonomous_execution = %#v, want false", payload["autonomous_execution"])
	}
}

func TestTransitionScheduleHandoffApproval_InvalidStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	exec, err := s.TransitionScheduleHandoffApproval(context.Background(), "r-1", "exec-1", "pending")
	if exec != nil {
		t.Fatalf("exec = %#v, want nil", exec)
	}
	if !errors.Is(err, ErrInvalidApprovalState) {
		t.Fatalf("error = %v, want ErrInvalidApprovalState", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected DB call: %v", err)
	}
}

func TestTransitionScheduleHandoffApproval_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectQuery("UPDATE trigger_executions AS e").
		WillReturnRows(sqlmock.NewRows(execColumns))
	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(execColumns))

	exec, err := s.TransitionScheduleHandoffApproval(context.Background(), "r-1", "missing-exec", "approved")
	if exec != nil {
		t.Fatalf("exec = %#v, want nil", exec)
	}
	if !errors.Is(err, ErrExecutionNotFound) {
		t.Fatalf("error = %v, want ErrExecutionNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestTransitionScheduleHandoffApproval_AlreadyTransitionedConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	mock.ExpectQuery("UPDATE trigger_executions AS e").
		WillReturnRows(sqlmock.NewRows(execColumns))
	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(execColumns).
			AddRow("exec-1", "r-1", "schedule:r-1:due", "", "proposed", "",
				"schedule:r-1:due", "11111111-1111-1111-1111-111111111111",
				"22222222-2222-2222-2222-222222222222", "approved",
				[]byte(`{"approval_state":"approved","autonomous_execution":false}`), now))

	exec, err := s.TransitionScheduleHandoffApproval(context.Background(), "r-1", "exec-1", "rejected")
	if exec != nil {
		t.Fatalf("exec = %#v, want nil", exec)
	}
	if !errors.Is(err, ErrApprovalTransitionConflict) {
		t.Fatalf("error = %v, want ErrApprovalTransitionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestTransitionScheduleHandoffApproval_AttachedRunConflicts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := NewStore(db)

	now := time.Now()
	mock.ExpectQuery("UPDATE trigger_executions AS e").
		WillReturnRows(sqlmock.NewRows(execColumns))
	mock.ExpectQuery("SELECT .+ FROM trigger_executions").
		WillReturnRows(sqlmock.NewRows(execColumns).
			AddRow("exec-1", "r-1", "schedule:r-1:due", "run-1", "proposed", "",
				"schedule:r-1:due", "11111111-1111-1111-1111-111111111111",
				"22222222-2222-2222-2222-222222222222", "awaiting_approval",
				[]byte(`{"autonomous_execution":false}`), now))

	exec, err := s.TransitionScheduleHandoffApproval(context.Background(), "r-1", "exec-1", "approved")
	if exec != nil {
		t.Fatalf("exec = %#v, want nil", exec)
	}
	if !errors.Is(err, ErrApprovalTransitionConflict) {
		t.Fatalf("error = %v, want ErrApprovalTransitionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}
