package protocol

import (
	"encoding/json"
	"testing"
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
