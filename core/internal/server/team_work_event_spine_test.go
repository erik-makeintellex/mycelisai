package server

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestInsertTeamStatusEventDB_RunLinkedEmitsMissionEvent(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	event := protocol.TeamStatusEvent{
		EventID:           "11111111-1111-1111-1111-111111111111",
		TeamID:            "qa-team",
		WorkItemID:        "22222222-2222-2222-2222-222222222222",
		RunID:             "33333333-3333-3333-3333-333333333333",
		IntentProofID:     "44444444-4444-4444-4444-444444444444",
		ContractID:        "contract-1",
		ProofID:           "proof-1",
		State:             protocol.TeamWorkStateDegraded,
		Headline:          "Team ask degraded",
		Details:           "The team response timed out.",
		ConfidencePosture: "operator_attention",
		BlockedBy:         []string{"team_response_timeout"},
		NextAction:        "Recover or steer this work item before retrying.",
		SourceKind:        string(protocol.SourceKindWebAPI),
		SourceChannel:     "api.teams.work.ask",
		PayloadKind:       string(protocol.PayloadKindStatus),
		AuditRefs:         []string{"audit-1"},
		Version:           "v1",
	}

	expectRawTeamStatusEventInsert(mock, event, now)
	expectTeamWorkMissionEventInsert(mock, "qa-team", protocol.TeamWorkStateDegraded)

	if err := s.insertTeamStatusEventDB(t.Context(), &event); err != nil {
		t.Fatalf("insertTeamStatusEventDB: %v", err)
	}
	if event.Timestamp.IsZero() {
		t.Fatal("expected status event timestamp to be populated")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestInsertTeamStatusEventDB_WithoutRunDoesNotEmitMissionEvent(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	event := protocol.TeamStatusEvent{
		EventID:           "11111111-1111-1111-1111-111111111111",
		TeamID:            "qa-team",
		WorkItemID:        "22222222-2222-2222-2222-222222222222",
		State:             protocol.TeamWorkStateRunning,
		Headline:          "Team work running",
		Details:           "The team started the work.",
		ConfidencePosture: "verified",
		SourceKind:        string(protocol.SourceKindWorkspaceUI),
		SourceChannel:     "soma.team_work",
		PayloadKind:       string(protocol.PayloadKindStatus),
		Version:           "v1",
	}

	expectRawTeamStatusEventInsert(mock, event, now)

	if err := s.insertTeamStatusEventDB(t.Context(), &event); err != nil {
		t.Fatalf("insertTeamStatusEventDB: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamWorkMissionEventSeverity(t *testing.T) {
	cases := []struct {
		name  string
		state protocol.TeamWorkState
		want  protocol.EventSeverity
	}{
		{"normal", protocol.TeamWorkStateRunning, protocol.SeverityInfo},
		{"needs operator", protocol.TeamWorkStateNeedsOperator, protocol.SeverityWarn},
		{"paused", protocol.TeamWorkStatePaused, protocol.SeverityWarn},
		{"degraded", protocol.TeamWorkStateDegraded, protocol.SeverityError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := teamWorkMissionEventSeverity(tc.state); got != tc.want {
				t.Fatalf("severity = %q, want %q", got, tc.want)
			}
		})
	}
}

func expectRawTeamStatusEventInsert(mock sqlmock.Sqlmock, event protocol.TeamStatusEvent, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_status_events").
		WithArgs(
			event.EventID, event.TeamID, event.WorkItemID, sqlmock.AnyArg(),
			sqlmock.AnyArg(), event.ContractID, event.ProofID,
			string(event.State), event.Headline, event.Details, event.ConfidencePosture,
			sqlmock.AnyArg(), event.NextAction, event.SourceKind, event.SourceChannel,
			event.PayloadKind, sqlmock.AnyArg(), event.Version,
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}
