package swarm

import (
	"testing"
	"time"

	"github.com/mycelis/core/internal/governance"
	"github.com/nats-io/nats.go"
)

func TestSoma_Integration(t *testing.T) {
	// 1. Start Embedded NATS
	s, nc := startTestNATS(t)
	defer s.Shutdown()
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
	err := nc.Publish("swarm.global.input.user", []byte("hello swarm"))
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

func TestSoma_DoesNotRouteRegisteredServiceInputAsOperatorIntent(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	soma := NewSoma(nc, &governance.Guard{}, NewRegistry("."), nil, nil, nil, nil)
	if err := soma.Start(); err != nil {
		t.Fatalf("Soma start failed: %v", err)
	}
	defer soma.Shutdown()

	routed := make(chan struct{}, 1)
	if _, err := nc.Subscribe("swarm.team.genesis.internal.command", func(*nats.Msg) {
		routed <- struct{}{}
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush subscriptions: %v", err)
	}
	if err := nc.Publish("swarm.global.input.shared-service-proof", []byte(`{"value":42}`)); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush publish: %v", err)
	}

	select {
	case <-routed:
		t.Fatal("registered service input was routed as operator intent")
	case <-time.After(250 * time.Millisecond):
	}
}
