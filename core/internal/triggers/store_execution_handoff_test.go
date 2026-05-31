package triggers

import (
	"context"
	"encoding/json"
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
