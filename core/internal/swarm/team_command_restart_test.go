package swarm

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

type retainedCommandReceiptStore struct {
	keys map[string]struct{}
}

func TestDurableCommandAcceptancePublishesCorrelatedRunningStatus(t *testing.T) {
	server, nc := startTestNATS(t)
	defer server.Shutdown()
	defer nc.Close()

	store := &retainedCommandReceiptStore{keys: map[string]struct{}{}}
	team := NewTeam(&TeamManifest{
		ID:     "acceptance-team",
		Name:   "Acceptance Team",
		Type:   TeamTypeAction,
		Inputs: []string{"swarm.team.acceptance-team.internal.command"},
	}, nc, nil, nil)
	team.commandReceipts = store
	if err := team.Start(); err != nil {
		t.Fatalf("start team: %v", err)
	}
	defer team.Stop()

	statusCh := make(chan *nats.Msg, 1)
	if _, err := nc.Subscribe("swarm.team.acceptance-team.signal.status", func(msg *nats.Msg) { statusCh <- msg }); err != nil {
		t.Fatalf("subscribe acceptance status: %v", err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush subscriptions: %v", err)
	}

	const workID = "11111111-1111-1111-1111-111111111111"
	raw, err := protocol.WrapSignalPayloadWithMeta(
		protocol.SourceKindWebAPI,
		"api.teams.work.ask",
		protocol.PayloadKindCommand,
		"",
		"acceptance-team",
		"soma-admin",
		[]byte(`{"goal":"verify receipt","context":{"work_item_id":"`+workID+`","team_id":"acceptance-team"}}`),
	)
	if err != nil {
		t.Fatalf("wrap command: %v", err)
	}
	if err := nc.Publish("swarm.team.acceptance-team.internal.command", raw); err != nil {
		t.Fatalf("publish command: %v", err)
	}

	select {
	case msg := <-statusCh:
		var env protocol.SignalEnvelope
		if err := json.Unmarshal(msg.Data, &env); err != nil {
			t.Fatalf("decode acceptance envelope: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			t.Fatalf("decode acceptance payload: %v", err)
		}
		if payload["work_item_id"] != workID || payload["state"] != string(protocol.TeamWorkStateRunning) {
			t.Fatalf("acceptance payload = %#v", payload)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for durable command acceptance")
	}
}

func (s *retainedCommandReceiptStore) AcceptCommand(_ context.Context, correlation teamCommandCorrelation, _ string) (bool, error) {
	key := correlation.TeamID + ":" + correlation.commandKey()
	if _, exists := s.keys[key]; exists {
		return false, nil
	}
	s.keys[key] = struct{}{}
	return true, nil
}

func TestTeamCommandAcceptanceSurvivesTeamReconstruction(t *testing.T) {
	correlation := teamCommandCorrelation{
		WorkItemID:     "11111111-1111-1111-1111-111111111111",
		TeamID:         "restart-safe-team",
		RunID:          "run-restart-1",
		IdempotencyKey: "confirm-action:proof-restart-1",
	}

	store := &retainedCommandReceiptStore{keys: map[string]struct{}{}}
	beforeRestart := &Team{
		Manifest:        &TeamManifest{ID: correlation.TeamID},
		seenCommandKeys: map[string]time.Time{},
		commandReceipts: store,
	}
	accepted, err := beforeRestart.acceptCommandCorrelation(t.Context(), correlation, "test.command")
	if err != nil || !accepted {
		t.Fatal("first delivery must be accepted")
	}

	afterRestart := &Team{
		Manifest:        &TeamManifest{ID: correlation.TeamID},
		seenCommandKeys: map[string]time.Time{},
		commandReceipts: store,
	}
	accepted, err = afterRestart.acceptCommandCorrelation(t.Context(), correlation, "test.command")
	if err != nil {
		t.Fatalf("reconstructed team receipt check: %v", err)
	}
	if accepted {
		t.Fatal("reconstructed team accepted a duplicate command; acceptance must be durable across restart")
	}
}
