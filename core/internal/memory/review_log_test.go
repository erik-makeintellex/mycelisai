package memory

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func TestNormalizeLogEntryForReviewAppliesCanonicalContext(t *testing.T) {
	entry := &LogEntry{
		TraceId: "trace-1",
		Level:   "warn",
		Message: "Team checkpoint available",
		Context: map[string]any{
			"team_id":        "alpha",
			"source_channel": "swarm.team.alpha.signal.status",
			"payload_kind":   "status",
		},
	}

	got := NormalizeLogEntryForReview(entry)

	if got.Level != "WARN" {
		t.Fatalf("level = %q, want WARN", got.Level)
	}
	if got.Intent != "status" {
		t.Fatalf("intent = %q, want status", got.Intent)
	}
	if got.Source != "alpha" {
		t.Fatalf("source = %q, want alpha", got.Source)
	}
	if got.Context["schema_version"] != protocol.OperationalLogSchemaVersion {
		t.Fatalf("schema_version = %v", got.Context["schema_version"])
	}
	if got.Context["centralized_review"] != true {
		t.Fatalf("centralized_review = %v, want true", got.Context["centralized_review"])
	}
}

func TestNewSignalReviewLogEntryMirrorsSignalIntoCentralReview(t *testing.T) {
	payload := json.RawMessage(`{"summary":"Creative Team Lead completed a review."}`)
	entry := NewSignalReviewLogEntry(
		"swarm.team.creative.signal.result",
		protocol.SignalEnvelope{
			Meta: protocol.SignalMeta{
				Timestamp:     time.Now().UTC(),
				SourceKind:    protocol.SourceKindSystem,
				SourceChannel: "swarm.team.creative.signal.result",
				PayloadKind:   protocol.PayloadKindResult,
				RunID:         "run-42",
				TeamID:        "creative",
				AgentID:       "creative-lead",
			},
			Payload: payload,
		},
	)

	if entry == nil {
		t.Fatalf("entry = nil")
	}
	if entry.Intent != "result" {
		t.Fatalf("intent = %q, want result", entry.Intent)
	}
	if entry.Source != "creative-lead" {
		t.Fatalf("source = %q, want creative-lead", entry.Source)
	}
	if entry.Message != "Creative Team Lead completed a review." {
		t.Fatalf("message = %q", entry.Message)
	}
}

func TestNewTelemetryReviewLogEntryUsesCentralizedTelemetryContract(t *testing.T) {
	entry := NewTelemetryReviewLogEntry(
		"swarm.team.alpha.telemetry",
		&protocol.CTSEnvelope{
			Meta: protocol.CTSMeta{
				SourceNode: "alpha-sensor",
				Timestamp:  time.Now().UTC(),
				TraceID:    "trace-telemetry",
			},
			SignalType: protocol.SignalTelemetry,
			Payload:    json.RawMessage(`{"state":"healthy"}`),
		},
	)

	if entry == nil {
		t.Fatalf("entry = nil")
	}
	if entry.Intent != "telemetry" {
		t.Fatalf("intent = %q, want telemetry", entry.Intent)
	}
	if entry.Source != "alpha-sensor" {
		t.Fatalf("source = %q, want alpha-sensor", entry.Source)
	}
	if entry.Context["review_scope"] != string(protocol.LogReviewScopeCentralReview) {
		t.Fatalf("review_scope = %v", entry.Context["review_scope"])
	}
}
