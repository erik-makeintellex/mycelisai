package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func explicitConfigMutationPlan(
	agentResult chatAgentResult,
	latestRequest string,
) ([]protocol.PlannedToolCall, bool) {
	tools, recognized := outcomeTemplateMutationTools(strings.ToLower(strings.TrimSpace(latestRequest)))
	if !recognized || len(tools) == 0 {
		return nil, false
	}
	content, format, _, hasInlineDocument := parseInlineConfigDocument(latestRequest)
	planned := make([]protocol.PlannedToolCall, 0, len(tools))
	for _, tool := range tools {
		call := protocol.PlannedToolCall{Name: tool, Arguments: map[string]any{}}
		for _, providerCall := range agentResult.PlannedToolCalls {
			if strings.EqualFold(strings.TrimSpace(providerCall.Name), tool) {
				call = providerCall
				break
			}
		}
		if tool == "store_config_document" && hasInlineDocument {
			call.Arguments = map[string]any{"content": content, "format": format}
		}
		planned = append(planned, normalizePlannedToolCall(call))
	}
	return planned, true
}

func deterministicConfigMutationResult(planned []protocol.PlannedToolCall, mutTools []string) (chatAgentResult, bool) {
	if len(planned) == 0 {
		return chatAgentResult{}, false
	}
	for _, call := range planned {
		switch strings.TrimSpace(call.Name) {
		case "store_config_document":
			if !configStoreCallHasSource(call) {
				return chatAgentResult{}, false
			}
		case "activate_config_document":
			continue
		default:
			return chatAgentResult{}, false
		}
	}
	return chatAgentResult{
		Text:             "Soma can update this saved configuration through a governed proposal.",
		ToolsUsed:        toolsForPlannedCalls(planned, mutTools),
		PlannedToolCalls: planned,
	}, true
}

func configStoreCallHasSource(call protocol.PlannedToolCall) bool {
	return firstNonEmptyString(call.Arguments["content"], call.Arguments["path"]) != ""
}
