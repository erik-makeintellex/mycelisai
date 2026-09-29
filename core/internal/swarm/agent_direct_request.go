package swarm

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/nats-io/nats.go"
)

func (a *Agent) handleDirectRequest(msg *nats.Msg) {
	select {
	case <-a.ctx.Done():
		return
	default:
	}
	input, history := a.parseConversationPayload(msg.Data)
	log.Printf("Agent [%s] direct request (%d prior turns): %s", a.Manifest.ID, len(history), truncateLog(input, 200))
	// SRU: only Core's in-process turn token carries a user's read scope; a
	// request without one (another publisher, a council consult) has no user.
	// MEM-LANES: the token counts only on the subject and reply inbox it was
	// issued for.
	result := a.processUserTurn(claimRecallTurn(msg.Header.Get(RecallTurnHeader), msg.Subject, msg.Reply), input, history)
	if msg.Reply != "" {
		if respBytes, err := json.Marshal(result); err == nil {
			msg.Respond(respBytes)
		} else {
			fallback := result.Text
			if fallback == "" && result.Availability != nil {
				fallback = result.Availability.Summary
			}
			msg.Respond([]byte(fallback))
		}
	}
	log.Printf("Agent [%s] direct request replied (tools: %v readable=%t).", a.Manifest.ID, result.ToolsUsed, strings.TrimSpace(result.Text) != "")
}

// processUserTurn runs one direct turn under the requesting user's
// saved-memory read scope (SRU). When that scope could not be verified the
// governed lane contributed nothing, and the reply says so.
func (a *Agent) processUserTurn(access RecallAccess, input string, history []cognitive.ChatMessage) ProcessResult {
	result := a.processTurn(access, input, history, true, nil)
	if access.Unavailable {
		result.Text = withRecallUnavailableNote(result.Text)
		if access.atCapacity && strings.TrimSpace(result.Text) != "" {
			result.Text += " " + recallAtCapacityNote
		}
	}
	return result
}
