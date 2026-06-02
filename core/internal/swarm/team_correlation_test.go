package swarm

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestTeam_ResponseDeliveryCarriesPendingWorkCorrelation(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	team := NewTeam(&TeamManifest{
		ID:         "test-core",
		Name:       "Test Core",
		Type:       TeamTypeAction,
		Inputs:     []string{"swarm.team.test-core.internal.command"},
		Deliveries: []string{"swarm.team.test-core.signal.result"},
	}, nc, nil, nil)
	if err := team.Start(); err != nil {
		t.Fatalf("team start: %v", err)
	}
	defer team.Stop()

	internalTriggerCh := make(chan struct{}, 1)
	if _, err := nc.Subscribe("swarm.team.test-core.internal.trigger", func(msg *nats.Msg) {
		internalTriggerCh <- struct{}{}
	}); err != nil {
		t.Fatalf("subscribe internal trigger: %v", err)
	}
	resultCh := make(chan *nats.Msg, 1)
	if _, err := nc.Subscribe("swarm.team.test-core.signal.result", func(msg *nats.Msg) { resultCh <- msg }); err != nil {
		t.Fatalf("subscribe result: %v", err)
	}
	nc.Flush()

	const workID = "11111111-1111-1111-1111-111111111111"
	publishCorrelatedCommand(t, nc, workID, "run-9")
	select {
	case <-internalTriggerCh:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for internal trigger")
	}

	if err := nc.Publish("swarm.team.test-core.internal.response", []byte("The note is ready.")); err != nil {
		t.Fatalf("publish response: %v", err)
	}

	msg := waitForTeamSignal(t, resultCh, "correlated result")
	env, projected := decodeTeamSignalPayload(t, msg.Data)
	if env.Meta.RunID != "run-9" {
		t.Fatalf("run_id = %q, want run-9", env.Meta.RunID)
	}
	assertProjectedWorkOutput(t, projected, workID, "The note is ready.")
}

func TestTeam_StatusDeliveryCarriesOutputReadyCorrelation(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	team := NewTeam(&TeamManifest{
		ID:         "test-core",
		Name:       "Test Core",
		Type:       TeamTypeAction,
		Inputs:     []string{"swarm.team.test-core.internal.command"},
		Deliveries: []string{"swarm.team.test-core.signal.status"},
	}, nc, nil, nil)
	if err := team.Start(); err != nil {
		t.Fatalf("team start: %v", err)
	}
	defer team.Stop()

	statusCh := make(chan *nats.Msg, 1)
	if _, err := nc.Subscribe("swarm.team.test-core.signal.status", func(msg *nats.Msg) { statusCh <- msg }); err != nil {
		t.Fatalf("subscribe status: %v", err)
	}
	internalTriggerCh := make(chan struct{}, 1)
	if _, err := nc.Subscribe("swarm.team.test-core.internal.trigger", func(msg *nats.Msg) {
		internalTriggerCh <- struct{}{}
	}); err != nil {
		t.Fatalf("subscribe internal trigger: %v", err)
	}
	nc.Flush()

	const workID = "11111111-1111-1111-1111-111111111111"
	publishCorrelatedCommand(t, nc, workID, "")
	select {
	case <-internalTriggerCh:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for internal trigger")
	}
	if err := nc.Publish("swarm.team.test-core.internal.response", []byte("Status-only note ready.")); err != nil {
		t.Fatalf("publish response: %v", err)
	}

	_, projected := decodeTeamSignalPayload(t, waitForTeamSignal(t, statusCh, "correlated status").Data)
	assertProjectedWorkOutput(t, projected, workID, "Status-only note ready.")
}

func publishCorrelatedCommand(t *testing.T, nc *nats.Conn, workID, runID string) {
	t.Helper()
	payload, err := protocol.WrapSignalPayloadWithMeta(
		protocol.SourceKindWebAPI,
		"api.teams.work.ask",
		protocol.PayloadKindCommand,
		runID,
		"test-core",
		"soma-admin",
		[]byte(`{"goal":"write the note","context":{"work_item_id":"`+workID+`","team_id":"test-core"}}`),
	)
	if err != nil {
		t.Fatalf("wrap command payload: %v", err)
	}
	if err := nc.Publish("swarm.team.test-core.internal.command", payload); err != nil {
		t.Fatalf("publish command: %v", err)
	}
}

func waitForTeamSignal(t *testing.T, ch <-chan *nats.Msg, label string) *nats.Msg {
	t.Helper()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for %s", label)
	}
	return nil
}

func decodeTeamSignalPayload(t *testing.T, raw []byte) (protocol.SignalEnvelope, map[string]any) {
	t.Helper()
	var env protocol.SignalEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode signal envelope: %v", err)
	}
	var projected map[string]any
	if err := json.Unmarshal(env.Payload, &projected); err != nil {
		t.Fatalf("decode projected payload: %v", err)
	}
	return env, projected
}

func assertProjectedWorkOutput(t *testing.T, projected map[string]any, workID, text string) {
	t.Helper()
	if projected["work_item_id"] != workID {
		t.Fatalf("work_item_id = %v", projected["work_item_id"])
	}
	if projected["state"] != "output_ready" {
		t.Fatalf("state = %v", projected["state"])
	}
	if projected["text"] != text {
		t.Fatalf("text = %v", projected["text"])
	}
}
