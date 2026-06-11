package server

import (
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestPersistConfirmedActionTeamWork_DelegatedWorkReturnsQueuedRef(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	link := testConfirmedActionTeamWorkLink(&protocol.ScopeValidation{
		PlannedToolCalls: []protocol.PlannedToolCall{{
			Name: "delegate_task",
			Arguments: map[string]any{
				"team_id": "qa-team",
				"task":    "Validate the retained output.",
			},
		}},
	})

	mock.ExpectBegin()
	expectTeamWorkItemInsert(mock, "qa-team", protocol.TeamExecutionShapeDelegatedWork, protocol.TeamWorkStateQueued, now)
	expectTeamStatusEventInsert(mock, "qa-team", protocol.TeamWorkStateQueued, now)
	expectTeamWorkItemUpdate(mock, protocol.TeamWorkStateQueued, sqlmock.AnyArg())
	expectTeamInteractionInsert(mock, "qa-team", "delegate", now)
	mock.ExpectCommit()

	refs, err := s.persistConfirmedActionTeamWork(t.Context(), link, []plannedToolExecutionResult{{
		Name: "delegate_task",
		Arguments: map[string]any{
			"team_id": "qa-team",
			"task":    "Validate the retained output.",
		},
		Output: "Delegated validation to QA.",
	}})
	if err != nil {
		t.Fatalf("persistConfirmedActionTeamWork: %v", err)
	}
	assertTeamWorkRef(t, refs, "qa-team", protocol.TeamWorkStateQueued, 0)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func assertTeamWorkRef(t *testing.T, refs []confirmActionTeamWorkRef, teamID string, state protocol.TeamWorkState, minOutputRefs int) {
	t.Helper()
	if len(refs) != 1 {
		t.Fatalf("refs = %#v, want exactly one", refs)
	}
	ref := refs[0]
	if ref.TeamID != teamID {
		t.Fatalf("team_id = %q, want %q", ref.TeamID, teamID)
	}
	if ref.State != state {
		t.Fatalf("state = %q, want %q", ref.State, state)
	}
	if strings.TrimSpace(ref.WorkItemID) == "" {
		t.Fatal("work_item_id is empty")
	}
	if ref.RunID != "33333333-3333-3333-3333-333333333333" {
		t.Fatalf("run_id = %q, want linked run", ref.RunID)
	}
	if len(ref.OutputRefs) < minOutputRefs {
		t.Fatalf("output_refs = %#v, want at least %d", ref.OutputRefs, minOutputRefs)
	}
}
