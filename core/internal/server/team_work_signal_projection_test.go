package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestTeamWorkSignalProjection_ResultWithoutRetainedOutputsDegradesDeliverable(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectProjectedStatusEvent(mock, "research-team", workID, protocol.TeamWorkStateDegraded, protocol.PayloadKindResult, now)
	expectProjectedTeamWorkUpdate(mock, workID, protocol.TeamWorkStateDegraded, true, "missing_retained_output")
	expectProjectedInteraction(mock, "research-team", workID, "degraded", protocol.PayloadKindResult, now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
			AgentID:       "builder",
		},
		Payload: json.RawMessage(`{"context":{"work_item_id":"` + workID + `"},"state":"output_ready","summary":"Draft ready","details":"Output passed local checks."}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestDeliverableResultMissingOutputs_AppliesToDelegatedWork(t *testing.T) {
	item := protocol.TeamWorkItem{
		ExecutionShape:  protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{"retained project package"},
	}
	if !deliverableResultMissingOutputs(item, protocol.PayloadKindResult, nil) {
		t.Fatal("delegated work with expected outputs must degrade when a result has no retained output refs")
	}
}

func TestTeamWorkSignalProjection_StatusUsesExplicitState(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateQueued, false, "", now)
	expectProjectedStatusEvent(mock, "research-team", workID, protocol.TeamWorkStateRunning, protocol.PayloadKindStatus, now)
	expectProjectedTeamWorkUpdate(mock, workID, protocol.TeamWorkStateRunning, false, "")
	expectProjectedInteraction(mock, "research-team", workID, "status", protocol.PayloadKindStatus, now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindStatus,
			TeamID:        "research-team",
			AgentID:       "planner",
		},
		Payload: json.RawMessage(`{"work_item_id":"` + workID + `","state":"running","headline":"Work started","message":"The team started execution."}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.status", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamWorkSignalProjection_ResultWithRetainedOutputRecordsCompletionProof(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	runID := "22222222-2222-2222-2222-222222222222"
	intentProofID := "33333333-3333-3333-3333-333333333333"
	contractID := "44444444-4444-4444-4444-444444444444"
	mock.MatchExpectationsInOrder(true)
	mockLinkedTeamWorkItem(mock, "game-team", workID, runID, intentProofID, contractID, now)
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO proof_artifacts").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("55555555-5555-5555-5555-555555555555"))
	mock.ExpectExec("UPDATE execution_contracts").
		WillReturnResult(sqlmock.NewResult(0, 1))
	expectProjectedStatusEventInsertOnly(mock, "game-team", workID, protocol.TeamWorkStateOutputReady, protocol.PayloadKindResult, now)
	mock.ExpectExec("INSERT INTO mission_events").
		WillReturnResult(sqlmock.NewResult(0, 1))
	expectProjectedTeamWorkUpdate(mock, workID, protocol.TeamWorkStateOutputReady, false, "")
	expectProjectedInteractionInsertOnly(mock, "game-team", workID, "output_ready", protocol.PayloadKindResult, now)
	mock.ExpectCommit()

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.game-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "game-team",
			RunID:         runID,
		},
		Payload: json.RawMessage(`{
			"context":{"work_item_id":"` + workID + `"},
			"summary":"Playable package ready",
			"outputs":[{
				"output_id":"game-package",
				"kind":"project_package",
				"label":"Playable game package",
				"storage_ref":"groups/game-team/generated/game",
				"entrypoint":"index.html",
				"output_class":"user_deliverable"
			}]
		}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.game-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamWorkSignalProjection_ResultHonorsExplicitDegradedState(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectProjectedStatusEventWithSource(mock, "research-team", workID, protocol.TeamWorkStateDegraded, protocol.PayloadKindResult, string(protocol.SourceKindSystem), "swarm.team.research-team.internal.response", now)
	expectProjectedTeamWorkUpdate(mock, workID, protocol.TeamWorkStateDegraded, true, "provider_timeout")
	expectProjectedInteractionWithSource(mock, "research-team", workID, "degraded", protocol.PayloadKindResult, string(protocol.SourceKindSystem), "swarm.team.research-team.internal.response", now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindSystem,
			SourceChannel: "swarm.team.research-team.internal.response",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
		},
		Payload: json.RawMessage(`{"work_item_id":"` + workID + `","state":"degraded","headline":"Team ask degraded","details":"Provider timed out.","degradation_state":"provider_timeout"}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func mockLinkedTeamWorkItem(mock sqlmock.Sqlmock, teamID, workID, runID, intentProofID, contractID string, now time.Time) {
	mock.ExpectQuery("SELECT id::text, team_id").
		WithArgs(teamID, workID).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, teamID, runID, intentProofID, contractID, "", "Build playable game", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), "", []byte(`null`), []byte(`["playable app package"]`), []byte(`["launch smoke proof"]`), []byte(`[]`),
			"approved", string(protocol.TeamWorkStateRunning), []byte(`null`), false, "",
			[]byte(`[]`), []byte(`[]`), []byte(`[]`), []byte(`[]`), now, now, "v1",
		))
}

func TestTeamWorkSignalProjection_IgnoresArchivedWork(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateArchived, false, "missing_execution_plan", now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
			AgentID:       "builder",
		},
		Payload: json.RawMessage(`{"work_item_id":"` + workID + `","summary":"Late result arrived","details":"This must not revive active review."}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamWorkSignalProjection_UncorrelatedSignalIgnored(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
			AgentID:       "builder",
		},
		Payload: json.RawMessage(`{"summary":"No explicit active-work correlation."}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func expectProjectedStatusEvent(mock sqlmock.Sqlmock, teamID, workID string, state protocol.TeamWorkState, kind protocol.SignalPayloadKind, now time.Time) {
	expectProjectedStatusEventWithSource(mock, teamID, workID, state, kind, string(protocol.SourceKindInternalTool), "swarm.team."+teamID+".internal.trigger", now)
}

func expectProjectedStatusEventWithSource(mock sqlmock.Sqlmock, teamID, workID string, state protocol.TeamWorkState, kind protocol.SignalPayloadKind, sourceKind, sourceChannel string, now time.Time) {
	mock.ExpectBegin()
	expectProjectedStatusEventInsertWithSource(mock, teamID, workID, state, kind, sourceKind, sourceChannel, now)
}

func expectProjectedStatusEventInsertOnly(mock sqlmock.Sqlmock, teamID, workID string, state protocol.TeamWorkState, kind protocol.SignalPayloadKind, now time.Time) {
	expectProjectedStatusEventInsertWithSource(mock, teamID, workID, state, kind, string(protocol.SourceKindInternalTool), "swarm.team."+teamID+".internal.trigger", now)
}

func expectProjectedStatusEventInsertWithSource(mock sqlmock.Sqlmock, teamID, workID string, state protocol.TeamWorkState, kind protocol.SignalPayloadKind, sourceKind, sourceChannel string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_status_events").
		WithArgs(
			sqlmock.AnyArg(), teamID, workID, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			string(state), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sourceKind,
			sourceChannel, string(kind), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}

func expectProjectedTeamWorkUpdate(mock sqlmock.Sqlmock, workID string, state protocol.TeamWorkState, needsOperator bool, degradation string) {
	mock.ExpectExec("UPDATE team_work_items").
		WithArgs(
			workID, string(state), sqlmock.AnyArg(), needsOperator, degradation,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectProjectedInteraction(mock sqlmock.Sqlmock, teamID, workID, verb string, kind protocol.SignalPayloadKind, now time.Time) {
	expectProjectedInteractionWithSource(mock, teamID, workID, verb, kind, string(protocol.SourceKindInternalTool), "swarm.team."+teamID+".internal.trigger", now)
}

func expectProjectedInteractionWithSource(mock sqlmock.Sqlmock, teamID, workID, verb string, kind protocol.SignalPayloadKind, sourceKind, sourceChannel string, now time.Time) {
	expectProjectedInteractionInsertWithSource(mock, teamID, workID, verb, kind, sourceKind, sourceChannel, now)
	mock.ExpectCommit()
}

func expectProjectedInteractionInsertOnly(mock sqlmock.Sqlmock, teamID, workID, verb string, kind protocol.SignalPayloadKind, now time.Time) {
	expectProjectedInteractionInsertWithSource(mock, teamID, workID, verb, kind, string(protocol.SourceKindInternalTool), "swarm.team."+teamID+".internal.trigger", now)
}

func expectProjectedInteractionInsertWithSource(mock sqlmock.Sqlmock, teamID, workID, verb string, kind protocol.SignalPayloadKind, sourceKind, sourceChannel string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_interactions").
		WithArgs(
			sqlmock.AnyArg(), teamID, workID, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sourceKind,
			sourceChannel, sqlmock.AnyArg(), verb, sqlmock.AnyArg(),
			string(kind), "", sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}

func mustSignalEnvelope(t *testing.T, env protocol.SignalEnvelope) []byte {
	t.Helper()
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal signal envelope: %v", err)
	}
	return raw
}
