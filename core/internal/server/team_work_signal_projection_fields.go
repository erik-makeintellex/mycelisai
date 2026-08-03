package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func teamIDFromSignalSubject(subject string) string {
	parts := strings.Split(subject, ".")
	if len(parts) == 5 && parts[0] == "swarm" && parts[1] == "team" && parts[3] == "signal" {
		return strings.TrimSpace(parts[2])
	}
	return ""
}

func payloadKindFromSignalSubject(subject string) protocol.SignalPayloadKind {
	if strings.HasSuffix(subject, ".signal.result") {
		return protocol.PayloadKindResult
	}
	return protocol.PayloadKindStatus
}

func projectedHeadline(kind protocol.SignalPayloadKind, payload map[string]any) string {
	if headline := firstNonEmptyString(stringField(payload, "headline"), stringField(payload, "title")); headline != "" {
		return headline
	}
	if kind == protocol.PayloadKindResult {
		return "Team result ready"
	}
	return "Team status update"
}

func projectedDetails(payload map[string]any) string {
	return firstNonEmptyString(stringField(payload, "details"), stringField(payload, "message"), stringField(payload, "text"), stringField(payload, "summary"))
}

func projectedSummary(kind protocol.SignalPayloadKind, payload map[string]any) string {
	if summary := firstNonEmptyString(stringField(payload, "summary"), stringField(payload, "message"), stringField(payload, "text"), stringField(payload, "details")); summary != "" {
		return summary
	}
	if kind == protocol.PayloadKindResult {
		return "Team emitted a correlated result signal."
	}
	return "Team emitted a correlated status signal."
}

func stringField(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if value, ok := values[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func stringSliceField(values map[string]any, key string) []string {
	raw, ok := values[key].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, value := range raw {
		if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
			out = append(out, strings.TrimSpace(text))
		}
	}
	return out
}
