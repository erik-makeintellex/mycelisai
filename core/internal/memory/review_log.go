package memory

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func normalizeLevel(level string) string {
	trimmed := strings.TrimSpace(strings.ToUpper(level))
	if trimmed == "" {
		return "INFO"
	}
	return trimmed
}

func summarizePayload(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return ""
	}

	var payload map[string]any
	if json.Unmarshal(raw, &payload) == nil {
		for _, key := range []string{"summary", "message", "status", "result", "error"} {
			if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}

	if len(trimmed) > 240 {
		return trimmed[:240] + "..."
	}
	return trimmed
}

func NormalizeLogEntryForReview(entry *LogEntry) *LogEntry {
	if entry == nil {
		return nil
	}

	ctx := protocol.ParseOperationalLogContext(entry.Context)
	if ctx.TraceID == "" {
		ctx.TraceID = strings.TrimSpace(entry.TraceId)
	}
	if ctx.Summary == "" {
		ctx.Summary = strings.TrimSpace(entry.Message)
	}
	if ctx.Status == "" {
		ctx.Status = strings.ToLower(strings.TrimSpace(entry.Level))
	}
	if entry.TraceId == "" {
		entry.TraceId = ctx.TraceID
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	entry.Level = normalizeLevel(entry.Level)
	if entry.Intent == "" {
		switch {
		case ctx.PayloadKind != "":
			entry.Intent = string(ctx.PayloadKind)
		case ctx.ReviewScope != "":
			entry.Intent = string(ctx.ReviewScope)
		default:
			entry.Intent = "log"
		}
	}
	if entry.Source == "" {
		switch {
		case ctx.AgentID != "":
			entry.Source = ctx.AgentID
		case ctx.TeamID != "":
			entry.Source = ctx.TeamID
		case ctx.Component != "":
			entry.Source = ctx.Component
		case ctx.Service != "":
			entry.Source = ctx.Service
		default:
			entry.Source = "system"
		}
	}
	if entry.Message == "" {
		entry.Message = ctx.Summary
	}
	entry.Context = ctx.ToMap()
	return entry
}

func NewSignalReviewLogEntry(subject string, env protocol.SignalEnvelope) *LogEntry {
	summary := strings.TrimSpace(env.Text)
	if summary == "" {
		summary = summarizePayload(env.Payload)
	}
	if summary == "" {
		summary = fmt.Sprintf("Signal received on %s", subject)
	}

	level := "INFO"
	if env.Meta.PayloadKind == protocol.PayloadKindError {
		level = "ERROR"
	}

	entry := &LogEntry{
		TraceId:   strings.TrimSpace(env.Meta.RunID),
		Timestamp: env.Meta.Timestamp,
		Level:     level,
		Source:    firstNonEmpty(env.Meta.AgentID, env.Meta.TeamID, string(env.Meta.SourceKind), "system"),
		Intent:    firstNonEmpty(string(env.Meta.PayloadKind), "signal"),
		Message:   summary,
		Context: protocol.OperationalLogContext{
			ReviewScope:    protocol.LogReviewScopeCentralReview,
			Service:        "core",
			Component:      "signal-bridge",
			Summary:        summary,
			Detail:         summarizePayload(env.Payload),
			WhyItMatters:   "This team output is mirrored into centralized review so Soma and operational leads can inspect progress without leaving the shared channel model.",
			SourceKind:     env.Meta.SourceKind,
			SourceChannel:  firstNonEmpty(env.Meta.SourceChannel, subject),
			PayloadKind:    env.Meta.PayloadKind,
			RunID:          env.Meta.RunID,
			TeamID:         env.Meta.TeamID,
			AgentID:        env.Meta.AgentID,
			Status:         strings.ToLower(level),
			ReviewChannels: []string{subject},
			Tags:           []string{"team-output", string(env.Meta.PayloadKind)},
		}.ToMap(),
	}
	return NormalizeLogEntryForReview(entry)
}

func NewTelemetryReviewLogEntry(subject string, env *protocol.CTSEnvelope) *LogEntry {
	if env == nil {
		return nil
	}
	summary := summarizePayload(env.Payload)
	if summary == "" {
		summary = fmt.Sprintf("Telemetry received from %s", env.Meta.SourceNode)
	}
	entry := &LogEntry{
		TraceId:   env.Meta.TraceID,
		Timestamp: env.Meta.Timestamp,
		Level:     "INFO",
		Source:    firstNonEmpty(env.Meta.SourceNode, "system"),
		Intent:    "telemetry",
		Message:   summary,
		Context: protocol.OperationalLogContext{
			ReviewScope:    protocol.LogReviewScopeCentralReview,
			Service:        "core",
			Component:      "telemetry-bridge",
			Summary:        summary,
			Detail:         summarizePayload(env.Payload),
			WhyItMatters:   "Telemetry remains team-owned, but it is mirrored into centralized review so Soma, meta-agentry, and team leads can reconstruct state without polling each team separately.",
			SourceChannel:  subject,
			PayloadKind:    protocol.PayloadKindTelemetry,
			TraceID:        env.Meta.TraceID,
			Status:         "info",
			ReviewChannels: []string{subject},
			Tags:           []string{"telemetry", string(env.SignalType)},
		}.ToMap(),
	}
	return NormalizeLogEntryForReview(entry)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
