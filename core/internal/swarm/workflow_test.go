package swarm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/governance"
	server "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

// MockProvider implements common.LLMProvider for testing
type MockProvider struct {
	Response string
}

func (m *MockProvider) Infer(ctx context.Context, prompt string, opts cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{
		Text: m.Response,
	}, nil
}
func (m *MockProvider) Probe(ctx context.Context) (bool, error) { return true, nil }

func TestWorkflow_FullLoop(t *testing.T) {
	// 1. Start NATS
	opts := server.DefaultTestOptions
	opts.Port = -1
	s := server.RunServer(&opts)
	defer s.Shutdown()

	nc, _ := nats.Connect(s.ClientURL())
	defer nc.Close()

	// 2. Setup Mock Cognitive Engine
	brain := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{
				"chat": "mock",
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"mock": &MockProvider{Response: "I am ready to serve."},
		},
	}

	// 3. Setup Soma with Registry
	// We'll manually inject a team instead of loading from disk to keep test self-contained
	guard := &governance.Guard{}
	registry := NewRegistry(".") // Empty
	soma := NewSoma(nc, guard, registry, brain)

	// Manually inject a team
	manifest := &TeamManifest{
		ID:   "workflow-team",
		Name: "Workflow Team",
		Type: TeamTypeAction,
		Members: []TeamMember{
			{ID: "worker-1", Role: "assistant"},
		},
		Inputs: []string{"swarm.global.input.workflow.start"},
	}
	team := NewTeam(manifest, nc, brain)
	soma.teams["workflow-team"] = team
	team.Start()
	soma.axon = NewAxon(nc, soma) // Axon needs to know about the team map, but currently it routes by convention.
	// Actually Axon routes based on config/logic.
	// Let's just test the Team-Agent loop directly for now, as routing logic is simple string matching.

	// 4. Verification Channel
	done := make(chan string)
	nc.Subscribe("swarm.team.workflow-team.internal.response", func(msg *nats.Msg) {
		done <- string(msg.Data)
	})

	// 5. Trigger
	// Publish to the team's internal trigger directly to simulate Axon routing it there
	// Or publish to global and hope Axon routes it?
	// Axon routes "swarm.global.event.genesis.boot" -> "swarm.team.genesis.internal.command"
	// Our manifest listens to "swarm.global.input.workflow.start".
	// Team.Start() subscribes to that.
	// When Team receives it, it calls handleTrigger?
	// Wait, Team.handleTrigger just logs?
	// No, Team needs to forward to Internal Bus for Agents to see?
	// Check Team.handleTrigger implementation.

	// Publish directly to Internal Bus to test Agent logic
	agentTrigger := "swarm.team.workflow-team.internal.trigger"
	nc.Publish(agentTrigger, []byte("Do work!"))

	select {
	case resp := <-done:
		if !strings.Contains(resp, "ready to serve") {
			t.Errorf("Unexpected agent response: %s", resp)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("Timeout waiting for agent response")
	}
}
