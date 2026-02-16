package swarm

import (
	"testing"
	"time"

	"github.com/mycelis/core/internal/governance"
	server "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func TestSoma_Integration(t *testing.T) {
	// 1. Start Embedded NATS
	opts := server.DefaultTestOptions
	opts.Port = -1
	s := server.RunServer(&opts)
	defer s.Shutdown()

	// 2. Connect Client
	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("NATS connect failed: %v", err)
	}
	defer nc.Close()

	// 3. Setup Mocks
	guard := &governance.Guard{Engine: nil} // We need to mock ValidateIngress or it will crash if nil
	// Hack: We can't easily mock Guard struct methods in Go without an interface.
	// However, my Guard implementation is concrete.
	// For this test, valid inputs pass if Engine is nil? No, Guard.ValidateIngress doesn't use Engine.
	// It checks size/prefix. So &Guard{} is fine.

	// Registry
	reg := NewRegistry(".") // Empty path, no manifests load

	// 4. Init Soma
	soma := NewSoma(nc, guard, reg, nil, nil, nil, nil) // brain, stream, mcpExec, internalTools are nil for this test
	if err := soma.Start(); err != nil {
		t.Fatalf("Soma start failed: %v", err)
	}
	defer soma.Shutdown()

	// 5. Test Global Input -> Axon Routing
	done := make(chan bool)

	// Subscribe to where Axon should route (genesis default)
	nc.Subscribe("swarm.team.genesis.internal.command", func(msg *nats.Msg) {
		if string(msg.Data) == "hello swarm" {
			done <- true
		}
	})

	// Allow subscriptions to propagate
	nc.Flush()
	time.Sleep(100 * time.Millisecond)

	// Publish via Global Bus
	err = nc.Publish("swarm.global.input.cli.command", []byte("hello swarm"))
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	nc.Flush()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Errorf("Timeout waiting for Axon to route message")
	}
}
