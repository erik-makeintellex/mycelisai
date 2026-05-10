package server

import (
	"encoding/json"
	"strings"
)

func artifactRuntimeMetadata(raw json.RawMessage) (runID, runClass, noRunReason, retentionClass string) {
	if len(raw) == 0 {
		return "", "", "", ""
	}
	var metadata map[string]any
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return "", "", "", ""
	}
	runID = stringMetadata(metadata, "run_id")
	runClass = stringMetadata(metadata, "run_class")
	noRunReason = stringMetadata(metadata, "no_run_reason")
	retentionClass = stringMetadata(metadata, "retention_class")
	return runID, runClass, noRunReason, retentionClass
}

func stringMetadata(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	if value, ok := metadata[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}
