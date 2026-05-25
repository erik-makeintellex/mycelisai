package server

import (
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestPersistConfirmedActionTeamWork_CreateTeamStaysNew(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	link := testConfirmedActionTeamWorkLink(&protocol.ScopeValidation{
		PlannedToolCalls: []protocol.PlannedToolCall{{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id": "research-team",
				"name":    "Research Team",
			},
		}},
	})

	expectTeamWorkItemInsert(mock, "research-team", protocol.TeamExecutionShapeCreateTeam, protocol.TeamWorkStateNew, now)
	expectTeamStatusEventInsert(mock, "research-team", protocol.TeamWorkStateNew, now)
	expectTeamWorkItemUpdate(mock, protocol.TeamWorkStateNew, jsonContainsArg("research-team"))
	expectTeamInteractionInsert(mock, "research-team", "create_team", now)

	err := s.persistConfirmedActionTeamWork(t.Context(), link, []plannedToolExecutionResult{{
		Name: "create_team",
		Arguments: map[string]any{
			"team_id": "research-team",
			"name":    "Research Team",
		},
		Output: `{"status":"created","team_id":"research-team","name":"Research Team"}`,
	}})
	if err != nil {
		t.Fatalf("persistConfirmedActionTeamWork: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestPersistConfirmedActionTeamWork_DeliverableOutputReadyHasRefs(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	link := testConfirmedActionTeamWorkLink(&protocol.ScopeValidation{
		PlannedToolCalls: []protocol.PlannedToolCall{
			{
				Name:      "create_team",
				Arguments: map[string]any{"team_id": "qa-team", "name": "QA Team"},
			},
			{
				Name: "write_file",
				Arguments: map[string]any{
					"path":    "output/confirmed.txt",
					"content": "hello world",
				},
			},
		},
	})
	mock.MatchExpectationsInOrder(true)

	expectTeamWorkItemInsert(mock, "qa-team", protocol.TeamExecutionShapeCreateTeam, protocol.TeamWorkStateNew, now)
	expectTeamStatusEventInsert(mock, "qa-team", protocol.TeamWorkStateNew, now)
	expectTeamWorkItemUpdate(mock, protocol.TeamWorkStateNew, sqlmock.AnyArg())
	expectTeamInteractionInsert(mock, "qa-team", "create_team", now)

	expectTeamWorkItemInsert(mock, "qa-team", protocol.TeamExecutionShapeDeliverable, protocol.TeamWorkStateOutputReady, now)
	expectTeamStatusEventInsert(mock, "qa-team", protocol.TeamWorkStateQueued, now)
	expectTeamStatusEventInsert(mock, "qa-team", protocol.TeamWorkStateRunning, now)
	expectTeamStatusEventInsert(mock, "qa-team", protocol.TeamWorkStateOutputReady, now)
	expectTeamWorkItemUpdate(mock, protocol.TeamWorkStateOutputReady, jsonContainsArg("output/confirmed.txt"))
	expectTeamInteractionInsert(mock, "qa-team", "output_ready", now)

	err := s.persistConfirmedActionTeamWork(t.Context(), link, []plannedToolExecutionResult{
		{
			Name:      "create_team",
			Arguments: map[string]any{"team_id": "qa-team", "name": "QA Team"},
			Output:    `{"status":"created","team_id":"qa-team","name":"QA Team"}`,
		},
		{
			Name: "write_file",
			Arguments: map[string]any{
				"path":    "output/confirmed.txt",
				"content": "hello world",
			},
			Output: "Wrote output/confirmed.txt",
		},
	})
	if err != nil {
		t.Fatalf("persistConfirmedActionTeamWork: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestPersistFailedConfirmedActionTeamWork_RecordsDegradedWork(t *testing.T) {
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

	expectTeamWorkItemInsertWithPosture(mock, "qa-team", protocol.TeamExecutionShapeDelegatedWork, protocol.TeamWorkStateDegraded, true, "confirmed_action_failed", now)
	expectTeamStatusEventInsert(mock, "qa-team", protocol.TeamWorkStateDegraded, now)
	expectTeamWorkItemUpdateWithPosture(mock, protocol.TeamWorkStateDegraded, true, "confirmed_action_failed", sqlmock.AnyArg())
	expectTeamInteractionInsert(mock, "qa-team", "degraded", now)

	err := s.persistFailedConfirmedActionTeamWork(t.Context(), link, assertErr("tool failed"))
	if err != nil {
		t.Fatalf("persistFailedConfirmedActionTeamWork: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}

func testConfirmedActionTeamWorkLink(scope *protocol.ScopeValidation) confirmedActionTeamWorkLink {
	return confirmedActionTeamWorkLink{
		ProofID:         "22222222-2222-2222-2222-222222222222",
		ContractID:      "contract-1",
		ProofArtifactID: "proof-artifact-1",
		RunID:           "33333333-3333-3333-3333-333333333333",
		AuditID:         "audit-1",
		AuditUser:       "test-user",
		Scope:           scope,
	}
}

func expectTeamWorkItemInsert(mock sqlmock.Sqlmock, teamID string, shape protocol.TeamExecutionShape, state protocol.TeamWorkState, now time.Time) {
	expectTeamWorkItemInsertWithPosture(mock, teamID, shape, state, false, "", now)
}

func expectTeamWorkItemInsertWithPosture(mock sqlmock.Sqlmock, teamID string, shape protocol.TeamExecutionShape, state protocol.TeamWorkState, needsOperator bool, degradation string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_work_items").
		WithArgs(
			sqlmock.AnyArg(), teamID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Soma",
			string(shape), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			string(protocol.ApprovalPostureRequired), string(state), needsOperator,
			degradation, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))
}

func expectTeamStatusEventInsert(mock sqlmock.Sqlmock, teamID string, state protocol.TeamWorkState, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_status_events").
		WithArgs(
			sqlmock.AnyArg(), teamID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			string(state), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), string(protocol.SourceKindWebAPI),
			"api.intent.confirm-action", string(protocol.PayloadKindStatus), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
	expectTeamWorkMissionEventInsert(mock, teamID, state)
}

func expectTeamWorkMissionEventInsert(mock sqlmock.Sqlmock, teamID string, state protocol.TeamWorkState) {
	mock.ExpectExec("INSERT INTO mission_events").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), "default", string(protocol.EventTeamWorkStatus),
			string(teamWorkMissionEventSeverity(state)), "soma", teamID, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectTeamWorkItemUpdate(mock sqlmock.Sqlmock, state protocol.TeamWorkState, outputRefs driver.Value) {
	expectTeamWorkItemUpdateWithPosture(mock, state, false, "", outputRefs)
}

func expectTeamWorkItemUpdateWithPosture(mock sqlmock.Sqlmock, state protocol.TeamWorkState, needsOperator bool, degradation string, outputRefs driver.Value) {
	mock.ExpectExec("UPDATE team_work_items").
		WithArgs(
			sqlmock.AnyArg(), string(state), sqlmock.AnyArg(), needsOperator, degradation,
			sqlmock.AnyArg(), outputRefs, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectTeamInteractionInsert(mock sqlmock.Sqlmock, teamID, verb string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_interactions").
		WithArgs(
			sqlmock.AnyArg(), teamID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), string(protocol.SourceKindWebAPI),
			"api.intent.confirm-action", sqlmock.AnyArg(), verb, sqlmock.AnyArg(),
			string(protocol.PayloadKindCommand), "", sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}

type jsonContainsArg string

func (j jsonContainsArg) Match(value driver.Value) bool {
	switch typed := value.(type) {
	case []byte:
		return strings.Contains(string(typed), string(j))
	case string:
		return strings.Contains(typed, string(j))
	default:
		return false
	}
}
