package swarm

import (
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestTeamSteeringRoutesToAgentInterjectionWithoutStartingSecondJob(t *testing.T) {
	server, nc := startTestNATS(t)
	defer server.Shutdown()
	defer nc.Close()

	team := NewTeam(&TeamManifest{
		ID:     "delivery-team",
		Name:   "Delivery Team",
		Type:   TeamTypeAction,
		Inputs: []string{"swarm.team.delivery-team.internal.command"},
		Members: []protocol.AgentManifest{{
			ID:   "delivery-agent",
			Role: "worker",
		}},
	}, nc, nil, nil)
	if err := team.Start(); err != nil {
		t.Fatalf("start team: %v", err)
	}
	defer team.Stop()

	interjectionCh := make(chan string, 1)
	if _, err := nc.Subscribe("swarm.agent.delivery-agent.interjection", func(msg *nats.Msg) {
		interjectionCh <- string(msg.Data)
	}); err != nil {
		t.Fatalf("subscribe interjection: %v", err)
	}
	jobCh := make(chan struct{}, 1)
	if _, err := nc.Subscribe("swarm.team.delivery-team.internal.trigger", func(*nats.Msg) {
		jobCh <- struct{}{}
	}); err != nil {
		t.Fatalf("subscribe job trigger: %v", err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush subscriptions: %v", err)
	}

	command := []byte(`{"goal":"continue with guidance","guidance":"Keep the visible Restart control.","context":{"action":"steer","work_item_id":"work-1","team_id":"delivery-team","run_id":"run-1","idempotency_key":"steer-1"}}`)
	wrapper, err := protocol.WrapSignalPayloadWithMeta(
		protocol.SourceKindWorkspaceUI,
		"soma.active_work.steer",
		protocol.PayloadKindCommand,
		"run-1",
		"delivery-team",
		"",
		command,
	)
	if err != nil {
		t.Fatalf("wrap steering command: %v", err)
	}
	if err := nc.Publish("swarm.team.delivery-team.internal.command", wrapper); err != nil {
		t.Fatalf("publish steering command: %v", err)
	}

	select {
	case got := <-interjectionCh:
		if got != "Keep the visible Restart control." {
			t.Fatalf("interjection = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for agent interjection")
	}
	select {
	case <-jobCh:
		t.Fatal("steering command started a second team job")
	case <-time.After(100 * time.Millisecond):
	}
	if pending := team.consumeCommandCorrelation(); pending != nil {
		t.Fatalf("steering command entered result correlation queue: %+v", pending)
	}
}
