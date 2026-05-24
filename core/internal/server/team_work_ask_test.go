package server

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestHandleTeamWorkAsk_RecordsOutputReadyResponse(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt, withNATS(t))
	now := time.Now().UTC()
	mock.MatchExpectationsInOrder(true)
	expectTeamWorkAskInsert(mock, "qa-team", protocol.TeamWorkStateQueued, false, "", now)
	expectTeamWorkAskStatus(mock, "qa-team", protocol.TeamWorkStateQueued, now)
	expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateQueued, false, "")
	expectTeamWorkAskInteraction(mock, "qa-team", "ask", string(protocol.PayloadKindCommand), now)
	expectTeamWorkAskStatus(mock, "qa-team", protocol.TeamWorkStateOutputReady, now)
	expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateOutputReady, false, "")
	expectTeamWorkAskInteraction(mock, "qa-team", "response", string(protocol.PayloadKindResult), now)

	subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, "qa-team")
	if _, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
		_ = msg.Respond([]byte("validated output package"))
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	mux := setupMux(t, "POST /api/v1/teams/{id}/work/ask", s.HandleTeamWorkAsk)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/teams/qa-team/work/ask", `{
		"message":"Validate the browser package.",
		"timeout_seconds":2,
		"expected_outputs":["validation note"]
	}`)

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	work := data["work_item"].(map[string]any)
	if work["state"] != string(protocol.TeamWorkStateOutputReady) {
		t.Fatalf("state = %v", work["state"])
	}
	if data["reply"] != "validated output package" {
		t.Fatalf("reply = %v", data["reply"])
	}
	event := data["event"].(map[string]any)
	if event["details"] != "Reply: validated output package" {
		t.Fatalf("event.details = %v", event["details"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleTeamWorkAsk_RecordsDegradedForUnreadableTeamResponse(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt, withNATS(t))
	now := time.Now().UTC()
	mock.MatchExpectationsInOrder(true)
	expectTeamWorkAskInsert(mock, "qa-team", protocol.TeamWorkStateQueued, false, "", now)
	expectTeamWorkAskStatus(mock, "qa-team", protocol.TeamWorkStateQueued, now)
	expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateQueued, false, "")
	expectTeamWorkAskInteraction(mock, "qa-team", "ask", string(protocol.PayloadKindCommand), now)
	expectTeamWorkAskStatus(mock, "qa-team", protocol.TeamWorkStateDegraded, now)
	expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateDegraded, true, "team_response_unreadable")
	expectTeamWorkAskInteraction(mock, "qa-team", "degraded", string(protocol.PayloadKindError), now)

	subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, "qa-team")
	if _, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
		_ = msg.Respond([]byte(`{"tool_call":{"name":"write_file","arguments":{"path":"output.txt"}}}`))
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	mux := setupMux(t, "POST /api/v1/teams/{id}/work/ask", s.HandleTeamWorkAsk)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/teams/qa-team/work/ask", `{
		"message":"Validate the browser package.",
		"timeout_seconds":2
	}`)

	assertStatus(t, rr, http.StatusAccepted)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	work := data["work_item"].(map[string]any)
	if work["state"] != string(protocol.TeamWorkStateDegraded) {
		t.Fatalf("state = %v", work["state"])
	}
	if work["degradation_state"] != "team_response_unreadable" {
		t.Fatalf("degradation_state = %v", work["degradation_state"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleTeamWorkAsk_RecordsDegradedWhenNATSOffline(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	now := time.Now().UTC()
	mock.MatchExpectationsInOrder(true)
	expectTeamWorkAskInsert(mock, "qa-team", protocol.TeamWorkStateQueued, false, "", now)
	expectTeamWorkAskStatus(mock, "qa-team", protocol.TeamWorkStateQueued, now)
	expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateQueued, false, "")
	expectTeamWorkAskInteraction(mock, "qa-team", "ask", string(protocol.PayloadKindCommand), now)
	expectTeamWorkAskStatus(mock, "qa-team", protocol.TeamWorkStateDegraded, now)
	expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateDegraded, true, "nats_offline")
	expectTeamWorkAskInteraction(mock, "qa-team", "degraded", string(protocol.PayloadKindError), now)

	mux := setupMux(t, "POST /api/v1/teams/{id}/work/ask", s.HandleTeamWorkAsk)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/teams/qa-team/work/ask", `{"message":"Ship the report."}`)

	assertStatus(t, rr, http.StatusAccepted)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	work := data["work_item"].(map[string]any)
	if work["state"] != string(protocol.TeamWorkStateDegraded) {
		t.Fatalf("state = %v", work["state"])
	}
	if work["degradation_state"] != "nats_offline" {
		t.Fatalf("degradation_state = %v", work["degradation_state"])
	}
	if work["needs_operator"] != true {
		t.Fatalf("needs_operator = %v", work["needs_operator"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamWorkAskFollowupContextSurvivesRequestCancellation(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()

	ctx, cancel := teamWorkAskFollowupContext(parent)
	defer cancel()

	select {
	case <-ctx.Done():
		t.Fatalf("follow-up context should stay usable after request cancellation: %v", ctx.Err())
	default:
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("follow-up context should retain a bounded deadline")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > teamAskFollowupTimeout {
		t.Fatalf("deadline remaining = %v, want within %v", remaining, teamAskFollowupTimeout)
	}
}

func expectTeamWorkAskInsert(mock sqlmock.Sqlmock, teamID string, state protocol.TeamWorkState, needsOperator bool, degradation string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_work_items").
		WithArgs(
			sqlmock.AnyArg(), teamID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Soma",
			string(protocol.TeamExecutionShapeDelegatedWork), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			string(protocol.ApprovalPostureAutoAllowed), string(state), needsOperator,
			degradation, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))
}

func expectTeamWorkAskStatus(mock sqlmock.Sqlmock, teamID string, state protocol.TeamWorkState, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_status_events").
		WithArgs(
			sqlmock.AnyArg(), teamID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			string(state), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), string(protocol.SourceKindWebAPI),
			teamWorkAskSourceChannel, string(protocol.PayloadKindStatus), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}

func expectTeamWorkAskUpdate(mock sqlmock.Sqlmock, state protocol.TeamWorkState, needsOperator bool, degradation string) {
	mock.ExpectExec("UPDATE team_work_items").
		WithArgs(
			sqlmock.AnyArg(), string(state), sqlmock.AnyArg(), needsOperator, degradation,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectTeamWorkAskInteraction(mock sqlmock.Sqlmock, teamID, verb, payloadKind string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_interactions").
		WithArgs(
			sqlmock.AnyArg(), teamID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), string(protocol.SourceKindWebAPI),
			teamWorkAskSourceChannel, sqlmock.AnyArg(), verb, sqlmock.AnyArg(),
			payloadKind, "", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}
