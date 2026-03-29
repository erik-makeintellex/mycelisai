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
