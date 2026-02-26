package swarm

import (
	"sync"
	"testing"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
	server "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func testBlueprint() *protocol.MissionBlueprint {
	return &protocol.MissionBlueprint{
		MissionID: "mission-parallel",
		Teams: []protocol.BlueprintTeam{
			{
				Name: "Research Team",
				Role: "research",
				Agents: []protocol.AgentManifest{
					{ID: "agent-research", Role: "assistant"},
				},
			},
			{
				Name: "Writer Team",
				Role: "writer",
				Agents: []protocol.AgentManifest{
					{ID: "agent-writer", Role: "assistant"},
				},
			},
		},
	}
}

func newTestSomaForActivation(t *testing.T) *Soma {
	t.Helper()

	opts := server.DefaultTestOptions
	opts.Port = -1
	ns := server.RunServer(&opts)
	t.Cleanup(ns.Shutdown)

	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatalf("nats connect failed: %v", err)
	}
	t.Cleanup(nc.Close)

	guard := &governance.Guard{}
	reg := NewRegistry(".")
	return NewSoma(nc, guard, reg, nil, nil, nil, nil)
}

func TestActivateBlueprint_IdempotentAcrossCalls(t *testing.T) {
	soma := newTestSomaForActivation(t)
	bp := testBlueprint()

	first := soma.ActivateBlueprint(bp, nil)
	if first.TeamsSpawned != 2 || first.TeamsSkipped != 0 {
		t.Fatalf("first activation unexpected result: spawned=%d skipped=%d", first.TeamsSpawned, first.TeamsSkipped)
	}

	second := soma.ActivateBlueprint(bp, nil)
	if second.TeamsSpawned != 0 || second.TeamsSkipped != 2 {
		t.Fatalf("second activation unexpected result: spawned=%d skipped=%d", second.TeamsSpawned, second.TeamsSkipped)
	}

	if got := len(soma.ListTeams()); got != 2 {
		t.Fatalf("expected 2 active teams after idempotent activation, got %d", got)
	}
}

func TestActivateBlueprint_ConcurrentCallsDoNotDuplicateTeams(t *testing.T) {
	soma := newTestSomaForActivation(t)
	bp := testBlueprint()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = soma.ActivateBlueprint(bp, nil)
	}()
	go func() {
		defer wg.Done()
		_ = soma.ActivateBlueprint(bp, nil)
	}()
	wg.Wait()

	if got := len(soma.ListTeams()); got != 2 {
		t.Fatalf("expected 2 active teams after concurrent activation, got %d", got)
	}
}

