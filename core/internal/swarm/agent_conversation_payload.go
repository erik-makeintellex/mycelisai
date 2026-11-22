package swarm

import (
	"encoding/json"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
)

// parseConversationPayload detects whether the NATS payload is a JSON conversation
// array or plain text. Returns the last user message as input and any prior turns
// as ChatMessage history.
func (a *Agent) parseConversationPayload(data []byte) (string, []cognitive.ChatMessage) {
	trimmed := strings.TrimSpace(string(data))
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return string(data), nil
	}

	type chatTurn struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	var turns []chatTurn
	if err := json.Unmarshal(data, &turns); err != nil {
		return string(data), nil
	}

	if len(turns) == 0 {
		return "", nil
	}

	last := turns[len(turns)-1]
	if len(turns) == 1 {
		return last.Content, nil
	}

	history := make([]cognitive.ChatMessage, 0, len(turns)-1)
	for _, t := range turns[:len(turns)-1] {
		role := t.Role
		switch role {
		case "admin", "architect", "assistant":
			role = "assistant"
		case "user":
		default:
			role = "user"
		}
		history = append(history, cognitive.ChatMessage{Role: role, Content: t.Content})
	}

	return last.Content, history
}

func isRefusalAssistantReply(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	return strings.Contains(lower, "can't assist with that") ||
		strings.Contains(lower, "cannot assist with that") ||
		strings.Contains(lower, "i'm unable to help") ||
		strings.Contains(lower, "i am unable to help") ||
		strings.Contains(lower, "can't deliver") ||
		strings.Contains(lower, "cannot deliver") ||
		strings.Contains(lower, "unable to deliver") ||
		strings.Contains(lower, "can't fulfill") ||
		strings.Contains(lower, "cannot fulfill") ||
		strings.Contains(lower, "unable to fulfill") ||
		strings.Contains(lower, "can't complete") ||
		strings.Contains(lower, "cannot complete") ||
		strings.Contains(lower, "unable to complete") ||
		strings.Contains(lower, "can't help with") ||
		strings.Contains(lower, "cannot help with")
}

func unwrapTransportUserContent(raw string) string {
	hasWrapper := strings.Contains(raw, "[DIRECT ANSWER ROUTE]") ||
		strings.Contains(raw, "[DIRECT ANSWER RETRY]") ||
		strings.Contains(raw, "[GOVERNED MUTATION ROUTE]") ||
		strings.Contains(raw, "[USER GOVERNANCE PROFILE]")
	if !hasWrapper {
		return raw
	}
	if idx := strings.LastIndex(raw, "Original request:\n"); idx >= 0 {
		if extracted := strings.TrimSpace(raw[idx+len("Original request:\n"):]); extracted != "" {
			return extracted
		}
	}
	return raw
}

func extractDirectAnswerPayload(input string, priorHistory []cognitive.ChatMessage) (string, string, []cognitive.ChatMessage, bool) {
	isDirect := strings.Contains(input, "[DIRECT ANSWER ROUTE]") || strings.Contains(input, "[DIRECT ANSWER RETRY]")
	cleanInput := unwrapTransportUserContent(input)

	var workspaceBlocks []string
	cleanedHistory := make([]cognitive.ChatMessage, 0, len(priorHistory))
	for _, msg := range priorHistory {
		trimmed := strings.TrimSpace(msg.Content)
		if trimmed == "" {
			continue
		}
		if msg.Role == "user" && (strings.HasPrefix(trimmed, "[WORKSPACE CONTEXT]") || strings.HasPrefix(trimmed, "[REFERENTIAL REVIEW]")) {
			workspaceBlocks = append(workspaceBlocks, trimmed)
			continue
		}
		if msg.Role == "assistant" && isRefusalAssistantReply(trimmed) {
			if len(cleanedHistory) > 0 && cleanedHistory[len(cleanedHistory)-1].Role == "user" {
				cleanedHistory = cleanedHistory[:len(cleanedHistory)-1]
			}
			continue
		}
		if msg.Role == "user" {
			msg.Content = unwrapTransportUserContent(trimmed)
		}
		cleanedHistory = append(cleanedHistory, msg)
	}

	return cleanInput, strings.Join(workspaceBlocks, "\n\n"), cleanedHistory, isDirect
}

func sanitizeSystemPromptForModelSafety(sys string) string {
	replacer := strings.NewReplacer(
		"The user is your symbiote partner.", "You partner closely with the operator.",
		"you have FULL filesystem access", "you have governed workspace file access",
		"NEVER ask permission", "call the appropriate tool directly",
		"## Identity — NEVER BREAK CHARACTER", "## Identity",
		`You are NOT "an AI assistant."`, "You are Soma, the Mycelis executive orchestrator.",
		`NEVER say "As an AI" or "I don't have feelings" or "I'm just a language model."`, "Speak clearly in first person as Soma.",
	)
	return replacer.Replace(sys)
}
