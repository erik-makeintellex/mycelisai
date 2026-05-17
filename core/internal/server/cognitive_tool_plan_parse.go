package server

import (
	"encoding/json"
	"fmt"
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

func inferWriteFilePlanFromRequest(text string) (protocol.PlannedToolCall, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return protocol.PlannedToolCall{}, false
	}

	targetPath := extractRequestedFilePath(trimmed)
	if targetPath == "" {
		return protocol.PlannedToolCall{}, false
	}

	content := ""
	if quoted := quotedContentPattern.FindStringSubmatch(trimmed); len(quoted) >= 2 {
		content = strings.TrimSpace(quoted[1])
	} else if prints := printsPattern.FindStringSubmatch(trimmed); len(prints) >= 2 {
		value := strings.TrimSpace(prints[1])
		switch strings.ToLower(filepathExt(targetPath)) {
		case ".py":
			content = fmt.Sprintf("print(%q)\n", value)
		default:
			content = value + "\n"
		}
	} else if strings.EqualFold(filepathExt(targetPath), ".py") {
		content = "print(\"hello world\")\n"
	}

	if content == "" {
		return protocol.PlannedToolCall{}, false
	}

	return protocol.PlannedToolCall{
		Name: "write_file",
		Arguments: map[string]any{
			"path":    targetPath,
			"content": content,
		},
	}, true
}

func extractRequestedFilePath(text string) string {
	matches := namedFilePattern.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		candidate := strings.TrimSpace(match[1])
		if looksLikeFilePath(candidate) {
			return strings.Trim(candidate, `"'.,;`)
		}
	}
	return ""
}

func looksLikeFilePath(value string) bool {
	trimmed := strings.Trim(value, `"'.,;`)
	if trimmed == "" {
		return false
	}
	return strings.ContainsAny(trimmed, `/\`) || filepathExt(trimmed) != ""
}

func filepathExt(targetPath string) string {
	lastDot := strings.LastIndex(targetPath, ".")
	if lastDot < 0 {
		return ""
	}
	return strings.ToLower(targetPath[lastDot:])
}
