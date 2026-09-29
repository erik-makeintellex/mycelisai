package swarm

import (
	"testing"
	"time"

	"github.com/mycelis/core/internal/governance"
	"github.com/nats-io/nats.go"
)

// TestPheAxonProcessSignal_NoTeamResolvedReturnsHonestError is the negative/adversarial
// case for F17: Axon must never silently publish an operator signal to a team topic
// that has no team running behind it. When "genesis" is not registered, ProcessSignal
// must return an error and must not publish anything a listener could mistake for a
// real dispatch.
func TestPheAxonProcessSignal_NoTeamResolvedReturnsHonestError(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	soma := NewSoma(nc, &governance.Guard{}, NewRegistryFromRuntimeOrganization(nil), nil, nil, nil, nil)
	// Do not start Soma or register any team: "genesis" is not running.

	axon := NewAxon(nc, soma, nil)

	routed := make(chan struct{}, 1)
	if _, err := nc.Subscribe("swarm.team.genesis.internal.command", func(*nats.Msg) {
		routed <- struct{}{}
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush subscriptions: %v", err)
	}

	err := axon.ProcessSignal(&nats.Msg{Subject: "swarm.global.input.user", Data: []byte("hello swarm")})
	if err == nil {
		t.Fatal("expected an honest error when no team is resolved, got nil")
	}

	if err := nc.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	select {
	case <-routed:
		t.Fatal("signal was published to a team topic although no team is running")
	case <-time.After(200 * time.Millisecond):
	}
}

// TestPheAxonProcessSignal_ResolvedTeamRoutes is the positive case: once "genesis" is
// actually registered as a running team, ProcessSignal routes to it and returns no error.
func TestPheAxonProcessSignal_ResolvedTeamRoutes(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	soma := NewSoma(nc, &governance.Guard{}, NewRegistryFromRuntimeOrganization(nil), nil, nil, nil, nil)
	soma.teams["genesis"] = NewTeam(&TeamManifest{ID: "genesis", Name: "Genesis", Type: TeamTypeAction}, nc, nil, nil)

	axon := NewAxon(nc, soma, nil)

	routed := make(chan struct{}, 1)
	if _, err := nc.Subscribe("swarm.team.genesis.internal.command", func(*nats.Msg) {
		routed <- struct{}{}
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush subscriptions: %v", err)
	}

	if err := axon.ProcessSignal(&nats.Msg{Subject: "swarm.global.input.user", Data: []byte("hello swarm")}); err != nil {
		t.Fatalf("expected routing to succeed once genesis is running, got %v", err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	select {
	case <-routed:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for signal to route to the resolved team")
	}
}
