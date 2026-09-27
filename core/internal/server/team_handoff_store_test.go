package server

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

func sqlHandoffCommit() teamHandoffCommit {
	record := teamHandoffRecord{
		HandoffID: "88888888-8888-4888-8888-888888888888", IdempotencyKey: "team-handoff:k", RunID: handoffRunID,
		SourceTeamID: "research", SourceAgentID: "research-lead", TargetTeamID: "marketing", Note: "Use it.",
		ExpectedAction: "use", WorkItemID: "99999999-9999-4999-8999-999999999999",
		Inputs: []swarm.HandoffInputRef{{ArtifactID: handoffArtifactID, Title: "Fact sheet"}},
	}
	item, event, interaction := teamHandoffWorkRows(record, handoffProofID, "")
	return teamHandoffCommit{Record: record, WorkItem: item, Event: event, Interaction: interaction, Outbox: teamHandoffOutboxItem(record, item)}
}

// One transaction holds the exchange note, the target work item and its
// queued event, the handoff_sent interaction, and the outbox row.
func TestSQLTeamHandoffStoreCommitsInOneTransaction(t *testing.T) {
	dbOption, mock := withDB(t)
	s := newTestServer(dbOption)
	s.DispatchOutbox = dispatchoutbox.NewStore(s.getDB())
	store := &sqlTeamHandoffStore{server: s}
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO exchange_items").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO team_work_items").WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))
	mock.ExpectQuery("INSERT INTO team_status_events").WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
	mock.ExpectExec("INSERT INTO mission_events").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE team_work_items").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO team_interactions").WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
	mock.ExpectExec("INSERT INTO execution_dispatch_outbox").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := store.CommitHandoff(context.Background(), sqlHandoffCommit()); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLTeamHandoffStoreFailsClosedWithoutChannel(t *testing.T) {
	dbOption, mock := withDB(t)
	s := newTestServer(dbOption)
	s.DispatchOutbox = dispatchoutbox.NewStore(s.getDB())
	store := &sqlTeamHandoffStore{server: s}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO exchange_items").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := store.CommitHandoff(context.Background(), sqlHandoffCommit())
	if err == nil || !strings.Contains(err.Error(), teamHandoffChannel) {
		t.Fatalf("commit without seeded channel = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamHandoffWorkRowsAreQueuedAndLinked(t *testing.T) {
	commit := sqlHandoffCommit()
	if commit.WorkItem.State != protocol.TeamWorkStateQueued || commit.WorkItem.TeamID != "marketing" || commit.WorkItem.RunID != handoffRunID {
		t.Fatalf("work item = %#v", commit.WorkItem)
	}
	if commit.Interaction.Verb != "handoff_queued" || commit.Interaction.PayloadRef != commit.Record.HandoffID {
		t.Fatalf("interaction = %#v", commit.Interaction)
	}
	if commit.Outbox.DispatchKind != teamHandoffDispatchKind || commit.Outbox.IntentProofID != handoffProofID || commit.Outbox.IdempotencyKey != "team-handoff:"+commit.Record.HandoffID {
		t.Fatalf("outbox = %#v", commit.Outbox)
	}
	if commit.Event.State != protocol.TeamWorkStateQueued || !strings.Contains(commit.Event.Headline, "research") {
		t.Fatalf("event = %#v", commit.Event)
	}
}

func TestTeamHandoffStatusDerivation(t *testing.T) {
	read := []protocol.TeamInteraction{{Verb: "handoff_queued"}, {Verb: "handoff_sent"}, {Verb: "handoff_input_read"}}
	for _, tc := range []struct {
		state        protocol.TeamWorkState
		interactions []protocol.TeamInteraction
		want         string
	}{
		{protocol.TeamWorkStateQueued, nil, "queued"},
		{protocol.TeamWorkStateRunning, nil, "accepted"},
		{protocol.TeamWorkStateRunning, read, "read"},
		{protocol.TeamWorkStateOutputReady, read, "completed"},
		{protocol.TeamWorkStateDegraded, read, "needs_attention"},
		{protocol.TeamWorkStateNeedsOperator, nil, "needs_attention"},
	} {
		if got := teamHandoffStatus(protocol.TeamWorkItem{State: tc.state}, tc.interactions); got != tc.want {
			t.Fatalf("%s/%d = %s, want %s", tc.state, len(tc.interactions), got, tc.want)
		}
	}
}

func TestListTeamHandoffsRequiresRootAdminScope(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/exchange/handoffs", s.handleListTeamHandoffs)
	if rr := doRequest(t, mux, "GET", "/api/v1/exchange/handoffs", ""); rr.Code != 401 {
		t.Fatalf("anonymous = %d, want 401", rr.Code)
	}
	standard := &RequestIdentity{UserID: "user-1", Role: "user", Scopes: []string{"groups:read"}}
	if rr := doAuthenticatedRequestAs(t, mux, "GET", "/api/v1/exchange/handoffs", "", standard); rr.Code != 403 {
		t.Fatalf("standard user = %d, want 403", rr.Code)
	}
	noScope := &RequestIdentity{UserID: "admin-2", Role: "admin", Scopes: []string{"teams:read"}}
	if rr := doAuthenticatedRequestAs(t, mux, "GET", "/api/v1/exchange/handoffs", "", noScope); rr.Code != 403 {
		t.Fatalf("admin without groups:read = %d, want 403", rr.Code)
	}
}
