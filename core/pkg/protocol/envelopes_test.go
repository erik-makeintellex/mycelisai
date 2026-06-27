package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestWrapSignalPayloadWithMeta_IncludesRunTeamAgent(t *testing.T) {
	raw, err := WrapSignalPayloadWithMeta(
		SourceKindInternalTool,
		"internal_tool.delegate_task",
		PayloadKindCommand,
		"run-42",
		"alpha",
		"soma-admin",
		[]byte("inspect gate state"),
	)
	if err != nil {
		t.Fatalf("wrap signal payload: %v", err)
	}

	var env SignalEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if env.Meta.RunID != "run-42" {
		t.Fatalf("run_id = %q, want run-42", env.Meta.RunID)
	}
	if env.Meta.TeamID != "alpha" {
		t.Fatalf("team_id = %q, want alpha", env.Meta.TeamID)
	}
	if env.Meta.AgentID != "soma-admin" {
		t.Fatalf("agent_id = %q, want soma-admin", env.Meta.AgentID)
	}
	if env.Text != "inspect gate state" {
		t.Fatalf("text = %q, want inspect gate state", env.Text)
	}
}

func TestWrapSignalPayload_BackwardCompatibleDefaults(t *testing.T) {
	raw, err := WrapSignalPayload(
		SourceKindSystem,
		"swarm.team.alpha.internal.response",
		PayloadKindStatus,
		"alpha",
		[]byte(`{"summary":"ok"}`),
	)
	if err != nil {
		t.Fatalf("wrap signal payload: %v", err)
	}

	var env SignalEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if env.Meta.TeamID != "alpha" {
		t.Fatalf("team_id = %q, want alpha", env.Meta.TeamID)
	}
	if env.Meta.RunID != "" {
		t.Fatalf("run_id = %q, want empty", env.Meta.RunID)
	}
	if env.Meta.AgentID != "" {
		t.Fatalf("agent_id = %q, want empty", env.Meta.AgentID)
	}
	if string(env.Payload) != `{"summary":"ok"}` {
		t.Fatalf("payload = %s", string(env.Payload))
	}
}

func TestThreadEventEnvelope_IncludesOperatorSafeMetadata(t *testing.T) {
	event := ThreadEventEnvelope{
		Type:       "thread_event",
		EventType:  EventTeamWorkStatus,
		ThreadID:   "work-1",
		ThreadKind: "team_work",
		EventID:    "event-1",
		Version:    "v1",
		Meta: SignalMeta{
			Timestamp:     time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC),
			SourceKind:    SourceKindWebAPI,
			SourceChannel: "api.intent.confirm-action",
			PayloadKind:   PayloadKindThreadEvent,
			RunID:         "run-1",
			TeamID:        "team-1",
		},
		Payload: ThreadEventPayload{
			Kind:            ThreadEventExecutionStarted,
			Label:           "Execution started",
			Detail:          "Soma accepted the approved work.",
			Tone:            "info",
			Status:          "running",
			Href:            "/runs/run-1",
			HrefLabel:       "Open run receipt",
			TargetReference: "run:run-1",
			WorkItemID:      "work-1",
			IntentProofID:   "proof-1",
			ContractID:      "contract-1",
			ProofID:         "proof-1",
		},
	}
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal thread event: %v", err)
	}
	var decoded ThreadEventEnvelope
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode thread event: %v", err)
	}
	if decoded.Meta.PayloadKind != PayloadKindThreadEvent {
		t.Fatalf("payload_kind = %q, want %q", decoded.Meta.PayloadKind, PayloadKindThreadEvent)
	}
	if decoded.Meta.SourceKind != SourceKindWebAPI {
		t.Fatalf("source_kind = %q, want %q", decoded.Meta.SourceKind, SourceKindWebAPI)
	}
	if decoded.Payload.Href != "/runs/run-1" || decoded.Payload.TargetReference != "run:run-1" {
		t.Fatalf("thread event proof target missing: %+v", decoded.Payload)
	}
}
