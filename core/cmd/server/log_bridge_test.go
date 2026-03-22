package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func TestBuildMemoryLogEntryFromMessageUsesSignalEnvelope(t *testing.T) {
	data, err := json.Marshal(protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     time.Now().UTC(),
			SourceKind:    protocol.SourceKindSystem,
			SourceChannel: "swarm.team.alpha.signal.result",
			PayloadKind:   protocol.PayloadKindResult,
			RunID:         "run-1",
			TeamID:        "alpha",
			AgentID:       "alpha-lead",
		},
		Text: "Alpha Team Lead completed a review.",
	})
	if err != nil {
		t.Fatalf("marshal signal: %v", err)
	}

	entry := buildMemoryLogEntryFromMessage("swarm.team.alpha.signal.result", data)
	if entry == nil {
		t.Fatalf("entry = nil")
	}
	if entry.Intent != "result" {
		t.Fatalf("intent = %q, want result", entry.Intent)
	}
	if entry.Source != "alpha-lead" {
		t.Fatalf("source = %q, want alpha-lead", entry.Source)
	}
}

func TestBuildMemoryLogEntryFromMessageUsesTelemetryEnvelope(t *testing.T) {
	data, err := json.Marshal(protocol.CTSEnvelope{
		Meta: protocol.CTSMeta{
			SourceNode: "alpha-sensor",
			Timestamp:  time.Now().UTC(),
			TraceID:    "trace-1",
		},
		SignalType: protocol.SignalTelemetry,
		Payload:    json.RawMessage(`{"state":"healthy"}`),
	})
	if err != nil {
		t.Fatalf("marshal telemetry: %v", err)
	}

	entry := buildMemoryLogEntryFromMessage("swarm.team.alpha.telemetry", data)
	if entry == nil {
		t.Fatalf("entry = nil")
	}
	if entry.Intent != "telemetry" {
		t.Fatalf("intent = %q, want telemetry", entry.Intent)
	}
	if entry.Source != "alpha-sensor" {
		t.Fatalf("source = %q, want alpha-sensor", entry.Source)
	}
}
