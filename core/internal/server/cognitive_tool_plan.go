package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func parsePlannedToolCall(text string) (protocol.PlannedToolCall, bool) {
	var envelope struct {
		ToolCall struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		} `json:"tool_call"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &envelope); err != nil {
		return protocol.PlannedToolCall{}, false
	}
	if strings.TrimSpace(envelope.ToolCall.Name) == "" {
		return protocol.PlannedToolCall{}, false
	}
	if envelope.ToolCall.Arguments == nil {
		envelope.ToolCall.Arguments = map[string]any{}
	}
	return normalizePlannedToolCall(protocol.PlannedToolCall{
		Name:      strings.TrimSpace(envelope.ToolCall.Name),
		Arguments: envelope.ToolCall.Arguments,
	}), true
}

func normalizePlannedToolCall(call protocol.PlannedToolCall) protocol.PlannedToolCall {
	if call.Arguments == nil {
		call.Arguments = map[string]any{}
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

func buildPlannedToolCalls(agentResult chatAgentResult, latestRequest string, mutTools []string) []protocol.PlannedToolCall {
	var planned []protocol.PlannedToolCall
	parsedCall, hasParsedCall := parsePlannedToolCall(agentResult.Text)
	if inferredTeamCall, ok := inferCreateTeamPlanFromRequest(latestRequest); ok {
		if hasParsedCall && strings.TrimSpace(parsedCall.Name) == "create_team" {
			planned = append(planned, normalizePlannedToolCall(mergeMissingPlannedToolArguments(parsedCall, inferredTeamCall)))
		} else {
			planned = append(planned, normalizePlannedToolCall(inferredTeamCall))
		}
		if containsToolName(mutTools, "write_file") {
			if fileCall, ok := inferWriteFilePlanFromRequest(latestRequest); ok {
				planned = append(planned, normalizePlannedToolCall(fileCall))
			}
		}
		return planned
	}
	if hasParsedCall {
		planned = append(planned, normalizePlannedToolCall(parsedCall))
	}
	if len(planned) == 0 {
		for _, tool := range mutTools {
			if tool == "write_file" {
				if call, ok := inferWriteFilePlanFromRequest(latestRequest); ok {
					planned = append(planned, normalizePlannedToolCall(call))
				}
			}
		}
	}
	return planned
}

func mergeMissingPlannedToolArguments(primary, fallback protocol.PlannedToolCall) protocol.PlannedToolCall {
	if primary.Arguments == nil {
		primary.Arguments = map[string]any{}
	}
	for key, value := range fallback.Arguments {
		if strings.TrimSpace(fmt.Sprint(primary.Arguments[key])) == "" {
			primary.Arguments[key] = value
		}
	}
	return primary
}

func containsToolName(tools []string, want string) bool {
	for _, tool := range tools {
		if strings.EqualFold(strings.TrimSpace(tool), want) {
			return true
		}
	}
	return false
}

func affectedResourcesForPlannedCalls(planned []protocol.PlannedToolCall) []string {
	var resources []string
	for _, call := range planned {
		switch strings.TrimSpace(call.Name) {
		case "write_file":
			if rawPath, ok := call.Arguments["path"].(string); ok && strings.TrimSpace(rawPath) != "" {
				resources = append(resources, strings.TrimSpace(rawPath))
				continue
			}
		case "publish_signal":
			if subject, ok := call.Arguments["subject"].(string); ok && strings.TrimSpace(subject) != "" {
				resources = append(resources, strings.TrimSpace(subject))
				continue
			}
		case "promote_deployment_context":
			if artifactID, ok := call.Arguments["source_artifact_id"].(string); ok && strings.TrimSpace(artifactID) != "" {
				resources = append(resources, fmt.Sprintf("company knowledge from %s", strings.TrimSpace(artifactID)))
				continue
			}
		case "create_team":
			if teamID := firstNonEmptyString(call.Arguments["team_id"], call.Arguments["id"], call.Arguments["team_name"]); teamID != "" {
				resources = append(resources, "team:"+teamID)
				continue
			}
		}
		resources = append(resources, "state")
	}
	if len(resources) == 0 {
		return []string{"state"}
	}
	return uniqueOrderedTools(resources)
}

func inferAdapterKindFromTool(tool string) string {
	t := strings.ToLower(strings.TrimSpace(tool))
	switch {
	case strings.HasPrefix(t, "mcp_"), strings.Contains(t, "mcp"):
		return "mcp"
	case strings.HasPrefix(t, "http_"), strings.Contains(t, "api"), strings.Contains(t, "webhook"):
		return "openapi"
	case strings.HasPrefix(t, "host_"), t == "local_command":
		return "host"
	default:
		return "internal"
	}
}

func buildTeamExpressionsFromTools(tools []string, teamID string, rolePlan []string) []protocol.ChatTeamExpression {
	deduped := uniqueOrderedTools(tools)
	if strings.TrimSpace(teamID) == "" {
		teamID = "admin-core"
	}
	if len(rolePlan) == 0 {
		rolePlan = []string{"admin"}
	}
	expressions := make([]protocol.ChatTeamExpression, 0, len(deduped))
	for i, tool := range deduped {
		idx := i + 1
		bindingID := fmt.Sprintf("binding-%d-%s", idx, strings.ReplaceAll(tool, "_", "-"))
		expressionID := fmt.Sprintf("expr-%d", idx)
		expressions = append(expressions, protocol.ChatTeamExpression{
			ExpressionID: expressionID,
			TeamID:       teamID,
			Objective:    fmt.Sprintf("Execute %s through governed module binding", tool),
			RolePlan:     rolePlan,
			ModuleBindings: []protocol.ChatModuleBinding{
				{
					BindingID:   bindingID,
					ModuleID:    tool,
					AdapterKind: inferAdapterKindFromTool(tool),
					Operation:   tool,
				},
			},
		})
	}
	return expressions
}
