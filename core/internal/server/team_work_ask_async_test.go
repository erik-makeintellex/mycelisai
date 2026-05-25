package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestHandleTeamWorkAsk_AsyncPublishesCommandAndReturnsQueued(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt, withNATS(t))
	now := time.Now().UTC()
	mock.MatchExpectationsInOrder(true)
	expectTeamWorkAskInsert(mock, "qa-team", protocol.TeamWorkStateQueued, false, "", now)
	expectTeamWorkAskStatus(mock, "qa-team", protocol.TeamWorkStateQueued, now)
	expectTeamWorkAskUpdate(mock, protocol.TeamWorkStateQueued, false, "")
	expectTeamWorkAskInteraction(mock, "qa-team", "ask", string(protocol.PayloadKindCommand), now)

	subject := fmt.Sprintf(protocol.TopicTeamInternalCommand, "qa-team")
	received := make(chan []byte, 1)
	if _, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
		received <- append([]byte(nil), msg.Data...)
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	mux := setupMux(t, "POST /api/v1/teams/{id}/work/ask", s.HandleTeamWorkAsk)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/teams/qa-team/work/ask", `{
		"message":"Create the next validation note.",
		"async":true,
		"expected_outputs":["validation note"]
	}`)

	assertStatus(t, rr, http.StatusAccepted)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["accepted"] != true {
		t.Fatalf("accepted = %v", data["accepted"])
	}
	if data["dispatch_state"] != "published" {
		t.Fatalf("dispatch_state = %v", data["dispatch_state"])
	}
	work := data["work_item"].(map[string]any)
	if work["state"] != string(protocol.TeamWorkStateQueued) {
		t.Fatalf("state = %v", work["state"])
	}

	select {
	case raw := <-received:
		assertAsyncTeamAskEnvelope(t, raw)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for team command")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleTeamWorkAsk_AsyncRecordsDegradedWhenNATSOffline(t *testing.T) {
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
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/teams/qa-team/work/ask", `{
		"message":"Create the next validation note.",
		"async":true
	}`)

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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func assertAsyncTeamAskEnvelope(t *testing.T, raw []byte) {
	t.Helper()
	var env protocol.SignalEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("signal envelope: %v", err)
	}
	if env.Meta.TeamID != "qa-team" {
		t.Fatalf("team_id = %q", env.Meta.TeamID)
	}
	if env.Meta.SourceKind != protocol.SourceKindWebAPI {
		t.Fatalf("source_kind = %q", env.Meta.SourceKind)
	}
	if env.Meta.PayloadKind != protocol.PayloadKindCommand {
		t.Fatalf("payload_kind = %q", env.Meta.PayloadKind)
	}
	var payload map[string]any
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		t.Fatalf("payload: %v", err)
	}
	if payload["message"] != "Create the next validation note." {
		t.Fatalf("payload message = %v", payload["message"])
	}
	context, ok := payload["context"].(map[string]any)
	if !ok {
		t.Fatalf("payload context = %#v", payload["context"])
	}
	if context["work_item_id"] == "" {
		t.Fatal("payload context work_item_id is empty")
	}
	if context["team_id"] != "qa-team" {
		t.Fatalf("payload context team_id = %v", context["team_id"])
	}
}
