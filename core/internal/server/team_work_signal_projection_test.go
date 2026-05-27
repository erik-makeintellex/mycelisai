package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestTeamWorkSignalProjection_ResultMapsToOutputReady(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectProjectedStatusEvent(mock, "research-team", workID, protocol.TeamWorkStateOutputReady, protocol.PayloadKindResult, now)
	expectProjectedTeamWorkUpdate(mock, workID, protocol.TeamWorkStateOutputReady, false, "")
	expectProjectedInteraction(mock, "research-team", workID, "output_ready", protocol.PayloadKindResult, now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
			AgentID:       "builder",
		},
		Payload: json.RawMessage(`{"context":{"work_item_id":"` + workID + `"},"summary":"Draft ready","details":"Output passed local checks."}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
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
	mock.ExpectQuery("INSERT INTO team_status_events").
		WithArgs(
			sqlmock.AnyArg(), teamID, workID, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			string(state), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), string(protocol.SourceKindInternalTool),
			"swarm.team."+teamID+".internal.trigger", string(kind), sqlmock.AnyArg(), "v1",
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
	mock.ExpectQuery("INSERT INTO team_interactions").
		WithArgs(
			sqlmock.AnyArg(), teamID, workID, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), string(protocol.SourceKindInternalTool),
			"swarm.team."+teamID+".internal.trigger", sqlmock.AnyArg(), verb, sqlmock.AnyArg(),
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
