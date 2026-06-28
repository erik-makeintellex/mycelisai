package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func buildPlannedToolCalls(agentResult chatAgentResult, latestRequest string, mutTools []string) []protocol.PlannedToolCall {
	var planned []protocol.PlannedToolCall
	parsedCall, hasParsedCall := parsePlannedToolCall(agentResult.Text)
	if crossTeamCalls, ok := inferContentMarketingCrossTeamPlanFromRequest(latestRequest); ok {
		planned = append(planned, crossTeamCalls...)
		return ensureWriteFileExecutionPlan(planned, agentResult, latestRequest, mutTools)
	}
	if inferredTeamCall, ok := inferCreateTeamPlanFromRequest(latestRequest); ok {
		if hasParsedCall && strings.TrimSpace(parsedCall.Name) == "create_team" {
			planned = append(planned, normalizePlannedToolCall(mergeMissingPlannedToolArguments(parsedCall, inferredTeamCall)))
		} else {
			planned = append(planned, normalizePlannedToolCall(inferredTeamCall))
		}
		if fileCall, ok := inferWriteFilePlanFromRequest(latestRequest); ok && containsToolName(mutTools, "write_file") && shouldUseRequestedWriteFilePlan(latestRequest, fileCall) {
			planned = append(planned, normalizePlannedToolCall(fileCall))
		} else if fileCall, ok := inferTeamPreparationBriefPlanFromRequest(latestRequest, planned[0]); ok {
			planned = append(planned, normalizePlannedToolCall(fileCall))
		} else if fileCall, ok := inferWriteFileExecutionPlan(agentResult, latestRequest); ok && containsToolName(mutTools, "write_file") {
			planned = append(planned, normalizePlannedToolCall(fileCall))
		}
		if imageCall, saveCall, ok := inferTeamMediaDeliverablePlanFromRequest(latestRequest, planned[0]); ok && containsToolName(mutTools, "generate_image") {
			planned = append(planned, normalizePlannedToolCall(imageCall))
			if containsToolName(mutTools, "save_cached_image") {
				planned = append(planned, normalizePlannedToolCall(saveCall))
			}
		}
		return ensureWriteFileExecutionPlan(planned, agentResult, latestRequest, mutTools)
	}
	if continuationCalls, ok := inferTeamEvocationContinuationPlanFromRequest(latestRequest); ok {
		planned = append(planned, continuationCalls...)
		return ensureWriteFileExecutionPlan(planned, agentResult, latestRequest, mutTools)
	}
	if hasParsedCall {
		planned = append(planned, normalizePlannedToolCall(parsedCall))
	}
	if len(planned) == 0 {
		for _, tool := range mutTools {
			if tool == "write_file" {
				if call, ok := inferWriteFileExecutionPlan(agentResult, latestRequest); ok {
					planned = append(planned, normalizePlannedToolCall(call))
				}
			}
			if tool == "generate_image" {
				if imageCall, saveCall, ok := inferStandaloneMediaDeliverablePlanFromRequest(latestRequest); ok {
					planned = append(planned, normalizePlannedToolCall(imageCall))
					if containsToolName(mutTools, "save_cached_image") {
						planned = append(planned, normalizePlannedToolCall(saveCall))
					}
				}
			}
		}
	}
	return ensureWriteFileExecutionPlan(planned, agentResult, latestRequest, mutTools)
}

func ensureWriteFileExecutionPlan(planned []protocol.PlannedToolCall, agentResult chatAgentResult, latestRequest string, mutTools []string) []protocol.PlannedToolCall {
	if !containsToolName(mutTools, "write_file") {
		return planned
	}

	fallback, hasFallback := inferWriteFileExecutionPlan(agentResult, latestRequest)
	for i, call := range planned {
		call = normalizePlannedToolCall(call)
		if !strings.EqualFold(strings.TrimSpace(call.Name), "write_file") {
			planned[i] = call
			continue
		}
		if hasFallback {
			call = mergeMissingPlannedToolArguments(call, fallback)
		}
		planned[i] = normalizePlannedToolCall(call)
		return planned
	}

	if hasFallback {
		return append(planned, normalizePlannedToolCall(fallback))
	}
	return planned
}

func deterministicGovernedMutationResult(latestRequest string, mutTools []string) (chatAgentResult, bool) {
	planned := buildPlannedToolCalls(chatAgentResult{}, latestRequest, mutTools)
	if len(planned) == 0 || !plannedCallsHaveWritableOutput(planned) || !plannedCallsAreDeterministicProposalSafe(planned) {
		return chatAgentResult{}, false
	}
	tools := toolsForPlannedCalls(planned, mutTools)
	target := firstPlannedOutputTarget(planned)
	if target == "" {
		return chatAgentResult{}, false
	}
	return chatAgentResult{
		Text:      fmt.Sprintf("Soma can create `%s` through a governed proposal. Review the plan before execution.", target),
		ToolsUsed: tools,
	}, true
}

func plannedCallsHaveWritableOutput(planned []protocol.PlannedToolCall) bool {
	for _, call := range planned {
		if strings.EqualFold(strings.TrimSpace(call.Name), "write_file") {
			return true
		}
		if strings.EqualFold(strings.TrimSpace(call.Name), "generate_image") {
			return true
		}
		if strings.EqualFold(strings.TrimSpace(call.Name), "save_cached_image") {
			return true
		}
	}
	return false
}

func plannedCallsAreDeterministicProposalSafe(planned []protocol.PlannedToolCall) bool {
	for _, call := range planned {
		switch strings.TrimSpace(call.Name) {
		case "create_team", "delegate_task", "generate_image", "save_cached_image", "write_file":
			continue
		default:
			return false
		}
	}
	return true
}

func firstPlannedOutputTarget(planned []protocol.PlannedToolCall) string {
	for _, call := range planned {
		switch strings.TrimSpace(call.Name) {
		case "write_file":
			if target := firstNonEmptyString(call.Arguments["path"], call.Arguments["package_entrypoint"], call.Arguments["package_folder"]); target != "" {
				return target
			}
		case "save_cached_image":
			if target := workspaceMediaTarget(call.Arguments); target != "" {
				return target
			}
		}
	}
	for _, call := range planned {
		if strings.TrimSpace(call.Name) == "generate_image" {
			if target := firstNonEmptyString(call.Arguments["goal"], call.Arguments["prompt"]); target != "" {
				return target
			}
		}
	}
	return ""
}

func workspaceMediaTarget(arguments map[string]any) string {
	folder := strings.Trim(strings.TrimSpace(fmt.Sprint(arguments["folder"])), "/\\")
	filename := strings.Trim(strings.TrimSpace(fmt.Sprint(arguments["filename"])), "/\\")
	if filename == "" {
		return folder
	}
	if folder == "" {
		return filename
	}
	return folder + "/" + filename
}

func countUserChatMessages(messages []chatRequestMessage) int {
	count := 0
	for _, message := range messages {
		if strings.EqualFold(strings.TrimSpace(message.Role), "user") && strings.TrimSpace(message.Content) != "" {
			count++
		}
	}
	return count
}

func mergeMissingPlannedToolArguments(primary, fallback protocol.PlannedToolCall) protocol.PlannedToolCall {
	if primary.Arguments == nil {
		primary.Arguments = map[string]any{}
	}
	for key, value := range fallback.Arguments {
		if plannedArgumentIsEmpty(primary.Arguments[key]) {
			primary.Arguments[key] = value
		}
	}
	return primary
}

func plannedArgumentIsEmpty(value any) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	return strings.TrimSpace(fmt.Sprint(value)) == ""
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
		if strings.TrimSpace(call.ToolRef) != "" {
			if rawPath, ok := call.Arguments["path"].(string); ok && strings.TrimSpace(rawPath) != "" {
				resources = append(resources, strings.TrimSpace(rawPath))
				continue
			}
		}
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
		case "delegate_task":
			if teamID := firstNonEmptyString(call.Arguments["team_id"], call.Arguments["target_team"]); teamID != "" {
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
	case strings.HasPrefix(t, "mcp:"), strings.HasPrefix(t, "mcp_"), strings.Contains(t, "mcp"):
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
	teamID = resolveFocusedSomaTeamID(teamID)
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
