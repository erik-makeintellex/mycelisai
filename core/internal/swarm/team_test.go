package swarm

import (
	"testing"
	"time"

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

	team := NewTeam(manifest, nc, nil)
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
