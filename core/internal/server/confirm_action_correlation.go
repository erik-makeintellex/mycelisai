package server

import (
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func annotateConfirmedDelegationCall(planned protocol.PlannedToolCall, runID, proofID, contractID string, scope *protocol.ScopeValidation) protocol.PlannedToolCall {
	if !isDelegateTool(planned.Name) {
		return planned
	}
	args := cloneAnyMap(planned.Arguments)
	teamID := firstNonEmptyString(confirmedActionTeamID(args), confirmedActionCreatedTeamIDFromScope(scope))
	workID := firstNonEmptyString(correlationContextValue(args, "work_item_id"), args["work_item_id"], uuid.NewString())
	context := map[string]any{
		"work_item_id": workID,
		"team_id":      teamID,
	}
	addIfNotEmpty(context, "run_id", runID)
	addIfNotEmpty(context, "intent_proof_id", proofID)
	addIfNotEmpty(context, "contract_id", contractID)
	args["work_item_id"] = workID
	args["context"] = mergeAnyMap(args["context"], context)
	if ask, ok := args["ask"].(map[string]any); ok {
		ask = cloneAnyMap(ask)
		ask["context"] = mergeAnyMap(ask["context"], context)
		args["ask"] = ask
	}
	if task, ok := args["task"].(map[string]any); ok {
		task = cloneAnyMap(task)
		task["context"] = mergeAnyMap(task["context"], context)
		args["task"] = task
	}
	planned.Arguments = args
	return planned
}

func confirmedDelegationWorkItemID(args map[string]any) string {
	return firstNonEmptyString(correlationContextValue(args, "work_item_id"), args["work_item_id"])
}

func correlationContextValue(args map[string]any, key string) string {
	if args == nil || strings.TrimSpace(key) == "" {
		return ""
	}
	if value := mapStringValue(args["context"], key); value != "" {
		return value
	}
	if value := nestedMapStringValue(args["ask"], "context", key); value != "" {
		return value
	}
	return nestedMapStringValue(args["task"], "context", key)
}

func nestedMapStringValue(raw any, nestedKey, valueKey string) string {
	values, ok := raw.(map[string]any)
	if !ok {
		return ""
	}
	return mapStringValue(values[nestedKey], valueKey)
}

func mapStringValue(raw any, key string) string {
	values, ok := raw.(map[string]any)
	if !ok {
		return ""
	}
	return firstNonEmptyString(values[key])
}

func mergeAnyMap(raw any, extra map[string]any) map[string]any {
	merged := map[string]any{}
	if existing, ok := raw.(map[string]any); ok {
		for key, value := range existing {
			merged[key] = value
		}
	}
	for key, value := range extra {
		if strings.TrimSpace(firstNonEmptyString(value)) == "" {
			continue
		}
		merged[key] = value
	}
	return merged
}

func cloneAnyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func addIfNotEmpty(values map[string]any, key, value string) {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		values[key] = trimmed
	}
}
