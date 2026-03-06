package swarm

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	server "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func TestTeam_TriggerLogic(t *testing.T) {
	// 1. Start NATS
	opts := server.DefaultTestOptions
	opts.Port = -1
	s := server.RunServer(&opts)
	defer s.Shutdown()

	nc, _ := nats.Connect(s.ClientURL())
	defer nc.Close()

	// 2. Create Team
	manifest := &TeamManifest{
		ID:     "test-core",
		Name:   "Test Core",
		Type:   TeamTypeAction,
		Inputs: []string{"swarm.global.event.boom"},
	}

	team := NewTeam(manifest, nc, nil, nil)
	team.Start()
	defer team.Stop()

	// 3. Verify Internal Bus Activation
	done := make(chan bool)
	nc.Subscribe("swarm.team.test-core.internal.trigger", func(msg *nats.Msg) {
		done <- true
	})

	// 4. Trigger External
	nc.Publish("swarm.global.event.boom", []byte("data"))

	select {
	case <-done:
		// Pass
	case <-time.After(1 * time.Second):
		t.Errorf("Team did not forward external trigger to internal bus")
	}
}

func TestTeam_ResponseDeliveryWrapsStatusAndResultSignals(t *testing.T) {
	opts := server.DefaultTestOptions
	opts.Port = -1
	s := server.RunServer(&opts)
	defer s.Shutdown()

	nc, _ := nats.Connect(s.ClientURL())
	defer nc.Close()

	manifest := &TeamManifest{
		ID:   "test-core",
		Name: "Test Core",
		Type: TeamTypeAction,
		Inputs: []string{
			"swarm.global.event.boom",
		},
		Deliveries: []string{
			"swarm.team.test-core.signal.status",
			"swarm.team.test-core.signal.result",
		},
	}

	team := NewTeam(manifest, nc, nil, nil)
	if err := team.Start(); err != nil {
		t.Fatalf("team start: %v", err)
	}
	defer team.Stop()

	statusCh := make(chan *nats.Msg, 1)
	resultCh := make(chan *nats.Msg, 1)
	if _, err := nc.Subscribe("swarm.team.test-core.signal.status", func(msg *nats.Msg) { statusCh <- msg }); err != nil {
		t.Fatalf("subscribe status: %v", err)
	}
	if _, err := nc.Subscribe("swarm.team.test-core.signal.result", func(msg *nats.Msg) { resultCh <- msg }); err != nil {
		t.Fatalf("subscribe result: %v", err)
	}
	nc.Flush()

	internalResponse := "swarm.team.test-core.internal.response"
	if err := nc.Publish(internalResponse, []byte(`{"summary":"done"}`)); err != nil {
		t.Fatalf("publish response: %v", err)
	}

	assertSignal := func(ch <-chan *nats.Msg, wantKind protocol.SignalPayloadKind) {
		select {
		case msg := <-ch:
			var env protocol.SignalEnvelope
			if err := json.Unmarshal(msg.Data, &env); err != nil {
				t.Fatalf("decode signal envelope: %v", err)
			}
			if env.Meta.TeamID != "test-core" {
				t.Fatalf("team_id = %q, want test-core", env.Meta.TeamID)
			}
			if env.Meta.SourceKind != protocol.SourceKindSystem {
				t.Fatalf("source_kind = %q, want %q", env.Meta.SourceKind, protocol.SourceKindSystem)
			}
			if env.Meta.PayloadKind != wantKind {
				t.Fatalf("payload_kind = %q, want %q", env.Meta.PayloadKind, wantKind)
			}
			if string(env.Payload) != `{"summary":"done"}` {
				t.Fatalf("payload = %s", string(env.Payload))
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout waiting for %s", wantKind)
		}
	}

	assertSignal(statusCh, protocol.PayloadKindStatus)
	assertSignal(resultCh, protocol.PayloadKindResult)
}
