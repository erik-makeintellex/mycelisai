package server

import (
	"encoding/json"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

func decodeChatAgentResult(data []byte) chatAgentResult {
	var result chatAgentResult
	if err := json.Unmarshal(data, &result); err == nil {
		if strings.TrimSpace(result.Text) == "" {
			if readable, artifacts, ok := extractReadableStructuredReply(string(data)); ok {
				result.Text = readable
				if len(result.Artifacts) == 0 && len(artifacts) > 0 {
					result.Artifacts = artifacts
				}
			}
		}
		if result.hasStructuredState() {
			return result
		}
	}
	if readable, artifacts, ok := extractReadableStructuredReply(string(data)); ok {
		if strings.TrimSpace(readable) == "" && len(artifacts) == 0 {
			return chatAgentResult{
				Text: string(data),
			}
		}
		return chatAgentResult{
			Text:      readable,
			Artifacts: artifacts,
		}
	}
	return chatAgentResult{
		Text: string(data),
	}
}

func (r chatAgentResult) hasStructuredState() bool {
	return strings.TrimSpace(r.Text) != "" ||
		len(r.ToolsUsed) > 0 ||
		len(r.Artifacts) > 0 ||
		r.Availability != nil ||
		strings.TrimSpace(r.ProviderID) != "" ||
		strings.TrimSpace(r.ModelUsed) != "" ||
		len(r.Consultations) > 0
}

func buildChatBlocker(agentResult chatAgentResult, fallbackSummary string) cognitive.ExecutionAvailability {
	if agentResult.Availability != nil {
		blocker := *agentResult.Availability
		if blocker.Summary == "" {
			blocker.Summary = fallbackSummary
		}
		if blocker.Code == "" {
			blocker.Code = emptyProviderOutputCode
		}
		if blocker.RecommendedAction == "" {
			blocker.RecommendedAction = "Retry the request. If the issue persists, inspect the configured provider output or switch to another engine."
		}
		if blocker.ProviderID == "" {
			blocker.ProviderID = agentResult.ProviderID
		}
		if blocker.ModelID == "" {
			blocker.ModelID = agentResult.ModelUsed
		}
		blocker.Available = false
		return blocker
	}
	return cognitive.ExecutionAvailability{
		Available:         false,
		Code:              emptyProviderOutputCode,
		Summary:           fallbackSummary,
		RecommendedAction: "Retry the request. If the issue persists, inspect the configured provider output or switch to another engine.",
		ProviderID:        agentResult.ProviderID,
		ModelID:           agentResult.ModelUsed,
	}
}

func readableChatText(agentResult chatAgentResult, isMutation bool) string {
	if isMutation {
		if _, ok := parsePlannedToolCall(agentResult.Text); ok {
			return "Soma captured a governed mutation intent. Review the proposal details below."
		}
	}
	if !isMutation {
		if readable, _, ok := extractReadableStructuredReply(agentResult.Text); ok && strings.TrimSpace(readable) != "" {
			return readable
		}
	}
	if !isMutation && (containsToolCallJSON(agentResult.Text) || isUnreadableStructuredReply(agentResult.Text)) {
		return ""
	}
	if strings.TrimSpace(agentResult.Text) != "" {
		return agentResult.Text
	}
	if isMutation {
		return "Soma captured a governed mutation intent. Review the proposal details below."
	}
	if len(agentResult.Artifacts) > 0 {
		return "Soma returned artifacts for this request."
	}
	return ""
}

func mergeMutationTools(agentTools, requestTools []string) (bool, []string) {
	if len(requestTools) == 0 {
		return false, nil
	}
	combined := uniqueOrderedTools(append(append([]string{}, requestTools...), agentTools...))
	isMutation, mutTools := hasMutationTools(combined)
	return isMutation, mutTools
}

func containsToolCallJSON(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	if _, ok := parsePlannedToolCall(trimmed); ok {
		return true
	}
	return strings.Contains(trimmed, `"tool_call"`)
}

func extractReadableStructuredReply(text string) (string, []protocol.ChatArtifactRef, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" || !json.Valid([]byte(trimmed)) {
		return "", nil, false
	}
	if _, ok := parsePlannedToolCall(trimmed); ok || strings.Contains(trimmed, `"tool_call"`) || strings.Contains(trimmed, `"operation"`) {
		return "", nil, false
	}

	var payload struct {
		Text      string                     `json:"text"`
		Message   string                     `json:"message"`
		Summary   string                     `json:"summary"`
		Artifact  *protocol.ChatArtifactRef  `json:"artifact,omitempty"`
		Artifacts []protocol.ChatArtifactRef `json:"artifacts,omitempty"`
	}
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return "", nil, false
	}

	artifacts := make([]protocol.ChatArtifactRef, 0, len(payload.Artifacts)+1)
	if payload.Artifact != nil {
		artifacts = append(artifacts, *payload.Artifact)
	}
	if len(payload.Artifacts) > 0 {
		artifacts = append(artifacts, payload.Artifacts...)
	}

	return firstNonEmptyString(payload.Text, payload.Message, payload.Summary), artifacts, true
}

func isUnreadableStructuredReply(text string) bool {
	readable, artifacts, ok := extractReadableStructuredReply(text)
	return ok && strings.TrimSpace(readable) == "" && len(artifacts) == 0
}

func isWeakDirectAnswerFallback(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" {
		return false
	}

	weakCues := []string{
		"i'm sorry, but i can't assist with that right now",
		"i am sorry, but i can't assist with that right now",
		"i'm sorry, but i cannot assist with that right now",
		"i am sorry, but i cannot assist with that right now",
		"can't assist with that right now",
		"cannot assist with that right now",
		"i'm unable to help",
		"i am unable to help",
		"please try again later",
		"let me know if there's anything else i can help with",
		"let me know if there is anything else i can help with",
	}

	for _, cue := range weakCues {
		if strings.Contains(normalized, cue) {
			return true
		}
	}
	return false
}

func shouldRetryDirectAnswer(agentResult chatAgentResult, requestMutationTools []string) bool {
	if len(requestMutationTools) != 0 {
		return false
	}
	if isMutation, _ := hasMutationTools(agentResult.ToolsUsed); isMutation {
		return true
	}
	return containsToolCallJSON(agentResult.Text) || isUnreadableStructuredReply(agentResult.Text) || isWeakDirectAnswerFallback(agentResult.Text)
}

func directAnswerDriftBlocker(agentResult chatAgentResult) chatAgentResult {
	return chatAgentResult{
		Availability: &cognitive.ExecutionAvailability{
			Available:         false,
			Code:              emptyProviderOutputCode,
			Summary:           "Soma drifted into a governed action while answering a read-only request. Retry the request or restate it more directly.",
			RecommendedAction: "Retry the request. If this repeats, simplify the question or inspect the active cognitive provider output.",
			ProviderID:        agentResult.ProviderID,
			ModelID:           agentResult.ModelUsed,
		},
		ProviderID: agentResult.ProviderID,
		ModelUsed:  agentResult.ModelUsed,
	}
}
