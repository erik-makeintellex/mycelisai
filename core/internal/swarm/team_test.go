package swarm

import (
	"context"
	"encoding/json"
	"strings"
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

func TestTeam_StartIsReadyForImmediateCommand(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	resultCh := make(chan *nats.Msg, 1)
	if _, err := nc.Subscribe("swarm.team.immediate-team.signal.result", func(msg *nats.Msg) {
		resultCh <- msg
	}); err != nil {
		t.Fatalf("subscribe result: %v", err)
	}

	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{Providers: map[string]cognitive.ProviderConfig{
			"stub": {Enabled: true, ModelID: "stub", Location: "local"},
		}},
		Adapters: map[string]cognitive.LLMProvider{"stub": teamProviderStub{}},
	}
	manifest := &TeamManifest{
		ID:         "immediate-team",
		Name:       "Immediate Team",
		Type:       TeamTypeAction,
		Provider:   "stub",
		Inputs:     []string{"swarm.team.immediate-team.internal.command"},
		Deliveries: []string{"swarm.team.immediate-team.signal.result"},
		Members: []protocol.AgentManifest{{
			ID: "immediate-worker", Role: "worker", Provider: "stub",
		}},
	}

	team := NewTeam(manifest, nc, router, nil)
	if err := team.Start(); err != nil {
		t.Fatalf("team start: %v", err)
	}
	defer team.Stop()

	if err := nc.Publish("swarm.team.immediate-team.internal.command", []byte("deliver now")); err != nil {
		t.Fatalf("publish immediate command: %v", err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush immediate command: %v", err)
	}

	select {
	case msg := <-resultCh:
		if !strings.Contains(string(msg.Data), "ok") {
			t.Fatalf("result = %s, want provider response", string(msg.Data))
		}
	case <-time.After(time.Second):
		t.Fatal("team dropped the command published immediately after Start")
	}
}

func TestTeam_StopUnsubscribesBeforeSameIDIsRecreated(t *testing.T) {
	s, nc := startTestNATS(t)
	defer s.Shutdown()
	defer nc.Close()

	resultCh := make(chan *nats.Msg, 4)
	if _, err := nc.Subscribe("swarm.team.recreated-team.signal.result", func(msg *nats.Msg) {
		resultCh <- msg
	}); err != nil {
		t.Fatalf("subscribe result: %v", err)
	}
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{Providers: map[string]cognitive.ProviderConfig{
			"stub": {Enabled: true, ModelID: "stub", Location: "local"},
		}},
		Adapters: map[string]cognitive.LLMProvider{"stub": teamProviderStub{}},
	}
	newManifest := func() *TeamManifest {
		return &TeamManifest{
			ID:         "recreated-team",
			Name:       "Recreated Team",
			Type:       TeamTypeAction,
			Provider:   "stub",
			Inputs:     []string{"swarm.team.recreated-team.internal.command"},
			Deliveries: []string{"swarm.team.recreated-team.signal.result"},
			Members: []protocol.AgentManifest{{
				ID: "recreated-worker", Role: "worker", Provider: "stub",
			}},
		}
	}

	first := NewTeam(newManifest(), nc, router, nil)
	if err := first.Start(); err != nil {
		t.Fatalf("first team start: %v", err)
	}
	first.Stop()

	second := NewTeam(newManifest(), nc, router, nil)
	if err := second.Start(); err != nil {
		t.Fatalf("second team start: %v", err)
	}
	defer second.Stop()
	if err := nc.Publish("swarm.team.recreated-team.internal.command", []byte("deliver once")); err != nil {
		t.Fatalf("publish command: %v", err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush command: %v", err)
	}

	select {
	case <-resultCh:
	case <-time.After(time.Second):
		t.Fatal("recreated team did not return a result")
	}
	select {
	case duplicate := <-resultCh:
		t.Fatalf("stopped team left a duplicate NATS consumer: %s", string(duplicate.Data))
	case <-time.After(250 * time.Millisecond):
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
		[]byte(`{"ask_kind":"implementation","lane_role":"implementer","goal":"inspect gate state"}`),
	)
	if err != nil {
		t.Fatalf("wrap command payload: %v", err)
	}
	if err := nc.Publish("swarm.global.event.command", payload); err != nil {
		t.Fatalf("publish wrapped command: %v", err)
	}

	select {
	case got := <-done:
		var ask protocol.TeamAsk
		if err := json.Unmarshal([]byte(got), &ask); err != nil {
			t.Fatalf("decode team ask: %v", err)
		}
		if ask.Goal != "inspect gate state" {
			t.Fatalf("goal = %q", ask.Goal)
		}
		if ask.AskKind != protocol.TeamAskKindImplementation {
			t.Fatalf("ask_kind = %q", ask.AskKind)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for internal trigger payload")
	}
}

func TestNormalizeCommandPayload_PreservesStructuredTeamAsk(t *testing.T) {
	payload, err := protocol.WrapSignalPayloadWithMeta(
		protocol.SourceKindInternalTool,
		"internal_tool.delegate_task",
		protocol.PayloadKindCommand,
		"run-9",
		"alpha",
		"soma-admin",
		[]byte(`{"ask_kind":"research","lane_role":"researcher","goal":"Find the best documented approach."}`),
	)
	if err != nil {
		t.Fatalf("wrap command payload: %v", err)
	}

	got := normalizeCommandPayload(payload)
	if !strings.Contains(string(got), `"goal":"Find the best documented approach."`) {
		t.Fatalf("normalized payload = %s", string(got))
	}

	var ask protocol.TeamAsk
	if err := json.Unmarshal(got, &ask); err != nil {
		t.Fatalf("decode normalized team ask: %v", err)
	}
	if ask.AskKind != protocol.TeamAskKindResearch {
		t.Fatalf("ask_kind = %q", ask.AskKind)
	}
}
