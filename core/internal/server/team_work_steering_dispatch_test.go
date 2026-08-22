package server

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestDispatchClaimedTeamWorkSteeringPublishesCorrelatedCommand(t *testing.T) {
	withDBOption, mock := withDB(t)
	s := newTestServer(withDBOption, withNATS(t))
	s.DispatchOutbox = dispatchoutbox.NewStore(s.getDB())
	const (
		runID  = "11111111-1111-4111-8111-111111111111"
		workID = "22222222-2222-4222-8222-222222222222"
	)
	subject := fmt.Sprintf(protocol.TopicTeamInternalCommand, "launch-team")
	commandCh := make(chan []byte, 1)
	sub, err := s.NC.Subscribe(subject, func(msg *nats.Msg) { commandCh <- msg.Data })
	if err != nil {
		t.Fatalf("subscribe steering command: %v", err)
	}
	defer sub.Unsubscribe()
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush subscription: %v", err)
	}

	payload, _ := json.Marshal(teamWorkSteeringDispatchPayload{
		Action:         protocol.TeamWorkActionSteer,
		Summary:        "Also include a concise launch note.",
		ActorRef:       "operator@example.com",
		IdempotencyKey: "team-steer:" + workID + ":33333333-3333-4333-8333-333333333333",
	})
	mock.ExpectExec("UPDATE execution_dispatch_outbox").
		WithArgs("dispatch-1", dispatchoutbox.StatusCompleted, "", int64(0), dispatchoutbox.StatusCompleted, true).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = s.dispatchClaimedTeamWorkSteering(t.Context(), &dispatchoutbox.Item{
		ID: "dispatch-1", AttemptCount: 1, Payload: payload,
		RunID: runID, TeamID: "launch-team", WorkItemID: workID,
	})
	if err != nil {
		t.Fatalf("dispatch steering: %v", err)
	}

	select {
	case raw := <-commandCh:
		var env protocol.SignalEnvelope
		if err := json.Unmarshal(raw, &env); err != nil {
			t.Fatalf("decode steering envelope: %v", err)
		}
		if env.Meta.RunID != runID || env.Meta.TeamID != "launch-team" || env.Meta.PayloadKind != protocol.PayloadKindCommand {
			t.Fatalf("steering metadata = %#v", env.Meta)
		}
		var command map[string]any
		if err := json.Unmarshal(env.Payload, &command); err != nil {
			t.Fatalf("decode steering command: %v", err)
		}
		contextValue := command["context"].(map[string]any)
		if contextValue["work_item_id"] != workID || contextValue["action"] != "steer" {
			t.Fatalf("steering context = %#v", contextValue)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for team steering command")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
