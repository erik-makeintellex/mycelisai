package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// buildPlannedToolCalls marks calls the model chose (its tool_call text or
// captured planned calls) as agent origin and everything Core inferred from
// the request as core origin; scopePlannedCallsOrRespond enforces both (S7b).
func buildPlannedToolCalls(agentResult chatAgentResult, latestRequest string, mutTools []string) []protocol.PlannedToolCall {
	planned := assemblePlannedToolCalls(agentResult, latestRequest, mutTools)
	for i := range planned {
		if planned[i].Origin != protocol.PlannedCallOriginAgent {
			planned[i].Origin = protocol.PlannedCallOriginCore
		}
	}
	return planned
}

func assemblePlannedToolCalls(agentResult chatAgentResult, latestRequest string, mutTools []string) []protocol.PlannedToolCall {
	var planned []protocol.PlannedToolCall
	parsedCall, hasParsedCall := parsePlannedToolCall(agentResult.Text)
	parsedCall.Origin = protocol.PlannedCallOriginAgent
	if configCalls, ok := explicitConfigMutationPlan(agentResult, latestRequest); ok {
		return configCalls
	}
	if continuationCalls, ok := inferTeamEvocationContinuationPlanFromRequest(latestRequest); ok {
		planned = append(planned, continuationCalls...)
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
			if deliveryCalls, ok := inferInitialComplexDeliveryPlanFromRequest(latestRequest, planned[0], fileCall); ok {
				for _, deliveryCall := range deliveryCalls {
					planned = append(planned, normalizePlannedToolCall(deliveryCall))
				}
			}
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
	if len(agentResult.PlannedToolCalls) > 0 {
		for _, call := range agentResult.PlannedToolCalls {
			call.Origin = protocol.PlannedCallOriginAgent
			planned = append(planned, normalizePlannedToolCall(call))
		}
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

// deterministicGovernedMutationResult short-circuits the model only for saved
// configuration changes whose document source was supplied inline. Every other
// mutation request goes to Soma's agent, so no proposal is built from keyword
// heuristics alone.
func deterministicGovernedMutationResult(latestRequest string, mutTools []string) (chatAgentResult, bool) {
	planned := buildPlannedToolCalls(chatAgentResult{}, latestRequest, mutTools)
	return deterministicConfigMutationResult(planned, mutTools)
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
		case "store_config_document":
			if rawPath, ok := call.Arguments["path"].(string); ok && strings.TrimSpace(rawPath) != "" {
				resources = append(resources, strings.TrimSpace(rawPath))
				continue
			}
			resources = append(resources, "Outcome Template revision")
			continue
		case "activate_config_document":
			resources = append(resources, "selected Outcome Template")
			continue
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
