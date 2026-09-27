package server

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

var explicitOutputFilePathPattern = regexp.MustCompile(`(?i)(?:create|write|save|retain|store|produce)\b[^.\n]{0,140}?\b(?:at|to|as|path)\s+[` + "`" + `'\"]?([^` + "`" + `'\"\s]+)[` + "`" + `'\"]?`)

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
	}

	// Without explicit content the plan carries only the target. The proposal
	// builder drafts real content with the model (draftMissingWriteFileContent)
	// or returns an honest blocker; Soma never fills a file from a template.
	args := map[string]any{
		"path":       targetPath,
		"validation": textOutputValidationForRequest(trimmed, targetPath),
	}
	if content != "" {
		args["content"] = content
	}
	return protocol.PlannedToolCall{Name: "write_file", Arguments: args}, true
}

func textOutputValidationForRequest(request, targetPath string) string {
	if requestAsksForTextOutput(strings.ToLower(request)) {
		return "Retained text output must reopen from the workspace, match the requested structure, and separate assumptions or source claims when relevant."
	}
	if strings.EqualFold(filepathExt(targetPath), ".html") {
		return "Retained browser output must open locally and expose the requested interactive behavior."
	}
	return "Retained file output must reopen from the workspace and match the requested operator intent."
}

func extractRequestedFilePath(text string) string {
	if explicitTarget := extractExplicitOutputFilePath(text); explicitTarget != "" {
		return explicitTarget
	}
	var candidates []string
	matches := namedFilePattern.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		candidate := strings.TrimSpace(match[1])
		if looksLikeFilePath(candidate) {
			candidates = append(candidates, strings.Trim(candidate, `"'.,;`))
		}
	}
	for _, field := range strings.Fields(text) {
		candidate := strings.Trim(field, `"'.,;:()[]{}<>`)
		if looksLikeFilePath(candidate) {
			candidates = append(candidates, candidate)
		}
	}
	return preferredRequestedFilePath(candidates)
}

func extractExplicitOutputFilePath(text string) string {
	matches := explicitOutputFilePathPattern.FindAllStringSubmatch(text, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		if len(matches[i]) < 2 {
			continue
		}
		candidate := strings.Trim(matches[i][1], `"'.,;`)
		if looksLikeFilePath(candidate) {
			return candidate
		}
	}
	return ""
}

func preferredRequestedFilePath(candidates []string) string {
	for i := len(candidates) - 1; i >= 0; i-- {
		if strings.TrimSpace(filepathExt(candidates[i])) != "" {
			return candidates[i]
		}
	}
	for i := len(candidates) - 1; i >= 0; i-- {
		if strings.TrimSpace(candidates[i]) != "" {
			return candidates[i]
		}
	}
	return ""
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
	case strings.Contains(lower, "javascript") || strings.Contains(lower, "typescript") ||
		strings.Contains(lower, "html") || strings.Contains(lower, "browser"):
		return ".html"
	case hasExactWord(lower, "python") || hasExactWord(lower, "script"):
		return ".py"
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
	if strings.ContainsAny(trimmed, "<>=;") {
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
