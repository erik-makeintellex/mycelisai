package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func expectProjectedStatusEvent(mock sqlmock.Sqlmock, teamID, workID string, state protocol.TeamWorkState, kind protocol.SignalPayloadKind, now time.Time) {
	expectProjectedStatusEventWithSource(mock, teamID, workID, state, kind, string(protocol.SourceKindInternalTool), "swarm.team."+teamID+".internal.trigger", now)
}

func expectProjectedStatusEventWithSource(mock sqlmock.Sqlmock, teamID, workID string, state protocol.TeamWorkState, kind protocol.SignalPayloadKind, sourceKind, sourceChannel string, now time.Time) {
	mock.ExpectBegin()
	expectProjectedSignalReceipt(mock, teamID, workID, "swarm.team."+teamID+".signal."+string(kind))
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
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
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
