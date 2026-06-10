package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

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
	mock.ExpectBegin()
	expectTeamWorkAskStatus(mock, "qa-team", protocol.TeamWorkStateOutputReady, now)
	expectTeamWorkAskUpdateWithRetainedTextRefs(mock, protocol.TeamWorkStateOutputReady, false, "")
	expectTeamWorkAskInteraction(mock, "qa-team", "response", string(protocol.PayloadKindResult), now)
	mock.ExpectCommit()

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
	outputRefs, ok := work["output_refs"].([]any)
	if !ok || len(outputRefs) != 1 {
		t.Fatalf("output_refs = %#v, want one retained text output ref", work["output_refs"])
	}
	outputRef := outputRefs[0].(map[string]any)
	if outputRef["kind"] != "text_reply" || outputRef["label"] != "Team text reply" {
		t.Fatalf("output_ref = %#v, want retained text reply ref", outputRef)
	}
	if !strings.HasPrefix(fmt.Sprint(outputRef["storage_ref"]), "team_interaction:") {
		t.Fatalf("storage_ref = %v, want team interaction ref", outputRef["storage_ref"])
	}
	if !strings.HasPrefix(fmt.Sprint(outputRef["proof_ref"]), "team_status_event:") {
		t.Fatalf("proof_ref = %v, want status event proof ref", outputRef["proof_ref"])
	}
	proofRefs, ok := work["proof_refs"].([]any)
	if !ok || len(proofRefs) == 0 {
		t.Fatalf("proof_refs = %#v, want retained proof refs", work["proof_refs"])
	}
	auditRefs, ok := work["audit_refs"].([]any)
	if !ok || len(auditRefs) < 2 {
		t.Fatalf("audit_refs = %#v, want status and interaction audit refs", work["audit_refs"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRecordTeamWorkAskOutput_RollsBackWhenInteractionInsertFails(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	now := time.Now().UTC()
	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	expectTeamWorkAskStatus(mock, "qa-team", protocol.TeamWorkStateOutputReady, now)
	expectTeamWorkAskUpdateWithRetainedTextRefs(mock, protocol.TeamWorkStateOutputReady, false, "")
	expectTeamWorkAskInteractionFailure(mock, "qa-team", "response", string(protocol.PayloadKindResult), fmt.Errorf("interaction insert failed"))
	mock.ExpectRollback()

	item := protocol.NormalizeTeamWorkItem(protocol.TeamWorkItem{
		WorkItemID:        "11111111-1111-1111-1111-111111111111",
		TeamID:            "qa-team",
		Objective:         "Validate the browser package.",
		Owner:             "Soma",
		ExecutionShape:    protocol.TeamExecutionShapeDelegatedWork,
		GovernancePosture: protocol.ApprovalPostureAutoAllowed,
		State:             protocol.TeamWorkStateQueued,
	})
	_, err := s.recordTeamWorkAskOutput(t.Context(), &item, "team.subject", "validated output package")
	if err == nil {
		t.Fatal("recordTeamWorkAskOutput succeeded, want interaction failure")
	}
	if item.LastEvent != nil {
		t.Fatal("recordTeamWorkAskOutput set in-memory LastEvent after rollback")
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
	mock.ExpectBegin()
	expectTeamWorkAskStatus(mock, "qa-team", protocol.TeamWorkStateDegraded, now)
	expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateDegraded, true, "team_response_unreadable")
	expectTeamWorkAskInteraction(mock, "qa-team", "degraded", string(protocol.PayloadKindError), now)
	mock.ExpectCommit()

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
	mock.ExpectBegin()
	expectTeamWorkAskStatus(mock, "qa-team", protocol.TeamWorkStateDegraded, now)
	expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateDegraded, true, "nats_offline")
	expectTeamWorkAskInteraction(mock, "qa-team", "degraded", string(protocol.PayloadKindError), now)
	mock.ExpectCommit()

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
