package swarm

import "strings"

func correlationRunID(correlation *teamCommandCorrelation) string {
	if correlation == nil {
		return ""
	}
	return correlation.RunID
}

func signalString(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func firstNonEmptySignalString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
