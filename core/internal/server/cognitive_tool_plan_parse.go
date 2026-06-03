package server

import (
	"encoding/json"
	"strings"

	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/pkg/protocol"
)

func parsePlannedToolCall(text string) (protocol.PlannedToolCall, bool) {
	var envelope struct {
		ToolCall struct {
			Name      string         `json:"name"`
			ToolRef   string         `json:"tool_ref"`
			Arguments map[string]any `json:"arguments"`
		} `json:"tool_call"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &envelope); err != nil {
		return protocol.PlannedToolCall{}, false
	}
	if strings.TrimSpace(envelope.ToolCall.Name) == "" && strings.TrimSpace(envelope.ToolCall.ToolRef) == "" {
		return protocol.PlannedToolCall{}, false
	}
	if envelope.ToolCall.Arguments == nil {
		envelope.ToolCall.Arguments = map[string]any{}
	}
	return normalizePlannedToolCall(protocol.PlannedToolCall{
		Name:      strings.TrimSpace(envelope.ToolCall.Name),
		ToolRef:   strings.TrimSpace(envelope.ToolCall.ToolRef),
		Arguments: envelope.ToolCall.Arguments,
	}), true
}

func normalizePlannedToolCall(call protocol.PlannedToolCall) protocol.PlannedToolCall {
	if call.Arguments == nil {
		call.Arguments = map[string]any{}
	}
	call.Name = strings.TrimSpace(call.Name)
	call.ToolRef = strings.TrimSpace(call.ToolRef)
	if call.ToolRef == "" && mcp.IsMCPRef(call.Name) {
		call.ToolRef = call.Name
	}
	if ref := mcp.ParseToolRef(call.ToolRef); ref != nil && ref.ToolName != "*" {
		call.Name = ref.ToolName
	}

	switch strings.TrimSpace(call.Name) {
	case "write_file":
		if firstNonEmptyString(call.Arguments["path"]) == "" {
			if path := firstNonEmptyString(
				call.Arguments["file_path"],
				call.Arguments["target_path"],
				call.Arguments["filename"],
				call.Arguments["file"],
			); path != "" {
				call.Arguments["path"] = path
			}
		}
		if firstNonEmptyString(call.Arguments["content"]) == "" {
			if content := firstNonEmptyString(
				call.Arguments["body"],
				call.Arguments["text"],
			); content != "" {
				call.Arguments["content"] = content
			}
		}
	}

	return call
}

func inferWriteFileExecutionPlan(agentResult chatAgentResult, latestRequest string) (protocol.PlannedToolCall, bool) {
	if parsed, ok := parsePlannedToolCall(agentResult.Text); ok && strings.EqualFold(strings.TrimSpace(parsed.Name), "write_file") {
		parsed = normalizePlannedToolCall(parsed)
		if writeFilePlanHasPathAndContent(parsed) {
			return parsed, true
		}
		if fallback, ok := inferWriteFilePlanFromRequest(latestRequest); ok {
			return normalizePlannedToolCall(mergeMissingPlannedToolArguments(parsed, fallback)), true
		}
		path := firstNonEmptyString(parsed.Arguments["path"], extractRequestedFilePath(latestRequest))
		if path != "" {
			parsed.Arguments["path"] = path
			parsed.Arguments["content"] = inferWriteFileContent(agentResult, latestRequest, path)
			if writeFilePlanHasPathAndContent(parsed) {
				return normalizePlannedToolCall(parsed), true
			}
		}
	}

	if artifactPlan, ok := inferWriteFilePlanFromArtifacts(agentResult.Artifacts, latestRequest); ok {
		return normalizePlannedToolCall(artifactPlan), true
	}
	if path := extractRequestedFilePath(latestRequest); path != "" && !requestHasExplicitWriteFileContent(latestRequest) {
		if content := inferWriteFileContentFromAgentOutput(agentResult); content != "" {
			return protocol.PlannedToolCall{
				Name: "write_file",
				Arguments: map[string]any{
					"path":    path,
					"content": content,
				},
			}, true
		}
	}
	if inferred, ok := inferWriteFilePlanFromRequest(latestRequest); ok {
		return normalizePlannedToolCall(inferred), true
	}

	path := extractRequestedFilePath(latestRequest)
	if path == "" {
		path = defaultWriteFilePathForRequest(latestRequest)
	}
	content := inferWriteFileContent(agentResult, latestRequest, path)
	if path == "" || content == "" {
		return protocol.PlannedToolCall{}, false
	}
	return protocol.PlannedToolCall{
		Name: "write_file",
		Arguments: map[string]any{
			"path":    path,
			"content": content,
		},
	}, true
}

func writeFilePlanHasPathAndContent(call protocol.PlannedToolCall) bool {
	return firstNonEmptyString(call.Arguments["path"]) != "" && firstNonEmptyString(call.Arguments["content"]) != ""
}
