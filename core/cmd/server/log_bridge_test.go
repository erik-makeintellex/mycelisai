package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	pb "github.com/mycelis/core/pkg/pb/swarm"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	if entry.Context["team_id"] != "alpha" {
		t.Fatalf("team_id = %v, want alpha", entry.Context["team_id"])
	}
}

func TestBuildMemoryLogEntryFromMessageUsesLegacyEnvelopeReviewContract(t *testing.T) {
	data, err := proto.Marshal(&pb.MsgEnvelope{
		SourceAgentId: "creative-lead",
		TraceId:       "trace-legacy",
		TeamId:        "creative",
		Timestamp:     timestamppb.New(time.Now().UTC()),
		Payload: &pb.MsgEnvelope_ToolResult{
			ToolResult: &pb.ToolResultPayload{
				CallId:  "call-7",
				IsError: false,
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal legacy envelope: %v", err)
	}

	entry := buildMemoryLogEntryFromMessage("swarm.team.creative.signal.result", data)
	if entry == nil {
		t.Fatalf("entry = nil")
	}
	if entry.Intent != "tool_result" {
		t.Fatalf("intent = %q, want tool_result", entry.Intent)
	}
	if entry.Source != "creative-lead" {
		t.Fatalf("source = %q, want creative-lead", entry.Source)
	}
	if entry.Context["team_id"] != "creative" {
		t.Fatalf("team_id = %v, want creative", entry.Context["team_id"])
	}
	if entry.Context["payload_kind"] != string(protocol.PayloadKindResult) {
		t.Fatalf("payload_kind = %v, want result", entry.Context["payload_kind"])
	}
	if entry.Context["source_kind"] != string(protocol.SourceKindSystem) {
		t.Fatalf("source_kind = %v, want system", entry.Context["source_kind"])
	}
}
