package swarm

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

type teamProviderStub struct{}

func (teamProviderStub) Infer(_ context.Context, _ string, _ cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "ok", Provider: "stub", ModelUsed: "stub"}, nil
}

func (teamProviderStub) Probe(_ context.Context) (bool, error) {
	return true, nil
}

func TestTeam_TriggerLogic(t *testing.T) {
	// 1. Start NATS
	s, nc := startTestNATS(t)
	defer s.Shutdown()
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
	s, nc := startTestNATS(t)
	defer s.Shutdown()
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

func TestTeam_TriggerLogic_UnwrapsCommandEnvelope(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	manifest := &TeamManifest{
		ID:     "test-core",
		Name:   "Test Core",
		Type:   TeamTypeAction,
		Inputs: []string{"swarm.global.event.command"},
	}

	team := NewTeam(manifest, nc, nil, nil)
	if err := team.Start(); err != nil {
		t.Fatalf("team start: %v", err)
	}
	defer team.Stop()

	done := make(chan string, 1)
	if _, err := nc.Subscribe("swarm.team.test-core.internal.trigger", func(msg *nats.Msg) {
		done <- string(msg.Data)
	}); err != nil {
		t.Fatalf("subscribe internal trigger: %v", err)
	}
	nc.Flush()

	payload, err := protocol.WrapSignalPayloadWithMeta(
		protocol.SourceKindInternalTool,
		"internal_tool.delegate_task",
		protocol.PayloadKindCommand,
		"run-9",
		"test-core",
		"soma-admin",
		[]byte("inspect gate state"),
	)
	if err != nil {
		t.Fatalf("wrap command payload: %v", err)
	}
	if err := nc.Publish("swarm.global.event.command", payload); err != nil {
		t.Fatalf("publish wrapped command: %v", err)
	}

	select {
	case got := <-done:
		if got != "inspect gate state" {
			t.Fatalf("internal trigger payload = %q", got)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for internal trigger payload")
	}
}

func TestTeamNormalizeRuntimeProviderRoutingFallsBackExplicitDefaults(t *testing.T) {
	team := &Team{
		Manifest: &TeamManifest{
			ID:       "admin-core",
			Name:     "Soma",
			Provider: "local-ollama-dev",
			Members: []protocol.AgentManifest{
				{ID: "admin", Role: "admin"},
				{ID: "council-coder", Role: "coder", Provider: "local-ollama-dev"},
			},
		},
		brain: &cognitive.Router{
			Config: &cognitive.BrainConfig{
				Providers: map[string]cognitive.ProviderConfig{
					"ollama":           {Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local"},
					"local-ollama-dev": {Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local"},
				},
			},
			Adapters: map[string]cognitive.LLMProvider{
				"ollama": teamProviderStub{},
			},
		},
	}

	team.normalizeRuntimeProviderRouting()

	if team.Manifest.Provider != "ollama" {
		t.Fatalf("team provider = %q, want ollama", team.Manifest.Provider)
	}
	if team.Manifest.Members[0].Provider != "ollama" {
		t.Fatalf("admin provider = %q, want ollama", team.Manifest.Members[0].Provider)
	}
	if team.Manifest.Members[1].Provider != "ollama" {
		t.Fatalf("coder provider = %q, want ollama", team.Manifest.Members[1].Provider)
	}
}
