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
		content = synthesizeRequestedFileContent(trimmed, targetPath)
	}

	return protocol.PlannedToolCall{
		Name: "write_file",
		Arguments: map[string]any{
			"path":    targetPath,
			"content": content,
		},
	}, true
}

func synthesizeRequestedFileContent(request, targetPath string) string {
	request = strings.TrimSpace(request)
	ext := filepathExt(targetPath)
	switch ext {
	case ".md", ".markdown":
		return "# Generated note\n\n" +
			"Soma created this file from your request.\n\n" +
			"## What you asked for\n\n" +
			request + "\n"
	default:
		if strings.Contains(strings.ToLower(request), "where") && strings.Contains(strings.ToLower(request), "output") {
			return "Welcome to Mycelis.\n\n" +
				"Ask Soma for the work you want done. When Soma creates files, images, or other outputs, they appear in the latest output area with controls to open the file or open the containing folder. The run proof remains linked so you can review what happened later.\n"
		}
		return "Soma created this file from your request:\n\n" + request + "\n"
	}
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
	for _, field := range strings.Fields(text) {
		candidate := strings.Trim(field, `"'.,;:()[]{}<>`)
		if looksLikeFilePath(candidate) {
			return candidate
		}
	}
	return ""
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

func requestHasExplicitWriteFileContent(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	if quoted := quotedContentPattern.FindStringSubmatch(trimmed); len(quoted) >= 2 {
		return true
	}
	if prints := printsPattern.FindStringSubmatch(trimmed); len(prints) >= 2 {
		return true
	}
	return false
}

func inferWriteFilePlanFromArtifacts(artifacts []protocol.ChatArtifactRef, latestRequest string) (protocol.PlannedToolCall, bool) {
	for _, artifact := range artifacts {
		content := strings.TrimSpace(artifact.Content)
		if content == "" {
			continue
		}
		path := firstNonEmptyString(artifact.Entrypoint, artifact.SavedPath)
		if path == "" && artifact.Folder != "" && len(artifact.Files) > 0 {
			path = strings.TrimRight(strings.TrimSpace(artifact.Folder), `/\`) + "/" + strings.TrimLeft(strings.TrimSpace(artifact.Files[0]), `/\`)
		}
		if path == "" {
			path = extractRequestedFilePath(latestRequest)
		}
		if path == "" {
			path = defaultWriteFilePathForArtifact(artifact, latestRequest)
		}
		if path == "" {
			continue
		}
		return protocol.PlannedToolCall{
			Name: "write_file",
			Arguments: map[string]any{
				"path":    path,
				"content": content,
			},
		}, true
	}
	return protocol.PlannedToolCall{}, false
}

func inferWriteFileContent(agentResult chatAgentResult, latestRequest, targetPath string) string {
	if content := inferWriteFileContentFromAgentOutput(agentResult); content != "" {
		return content
	}
	if targetPath == "" || strings.TrimSpace(latestRequest) == "" {
		return ""
	}
	return synthesizeRequestedFileContent(latestRequest, targetPath)
}

func inferWriteFileContentFromAgentOutput(agentResult chatAgentResult) string {
	for _, artifact := range agentResult.Artifacts {
		if content := strings.TrimSpace(artifact.Content); content != "" {
			return content
		}
	}
	text := strings.TrimSpace(agentResult.Text)
	if text != "" && !containsToolCallJSON(text) && !isWeakDirectAnswerFallback(text) && !isGovernedProposalSummaryText(text) {
		return strings.TrimRight(text, "\r\n") + "\n"
	}
	return ""
}

func isGovernedProposalSummaryText(text string) bool {
	trimmed := strings.TrimSpace(text)
	return strings.HasPrefix(trimmed, "Soma can create `") && strings.Contains(trimmed, "through a governed proposal")
}

func defaultWriteFilePathForArtifact(artifact protocol.ChatArtifactRef, latestRequest string) string {
	ext := ".md"
	switch strings.ToLower(strings.TrimSpace(artifact.Type)) {
	case "code", "file":
		ext = extensionForWriteFileRequest(latestRequest)
	case "data", "chart":
		ext = ".json"
	}
	title := firstNonEmptyString(artifact.Title, "soma-output")
	return "workspace/generated/" + slugID(title) + ext
}

func defaultWriteFilePathForRequest(latestRequest string) string {
	if strings.TrimSpace(latestRequest) == "" {
		return ""
	}
	return "workspace/generated/" + slugID(firstWords(latestRequest, 8)) + extensionForWriteFileRequest(latestRequest)
}

func extensionForWriteFileRequest(text string) string {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "python") || strings.Contains(lower, "script"):
		return ".py"
	case strings.Contains(lower, "html") || strings.Contains(lower, "browser"):
		return ".html"
	case strings.Contains(lower, "json"):
		return ".json"
	case strings.Contains(lower, "yaml") || strings.Contains(lower, " yml"):
		return ".yaml"
	case strings.Contains(lower, "markdown") || strings.Contains(lower, "readme"):
		return ".md"
	default:
		return ".md"
	}
}

func firstWords(text string, limit int) string {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return "soma-output"
	}
	if len(fields) > limit {
		fields = fields[:limit]
	}
	return strings.Join(fields, " ")
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
