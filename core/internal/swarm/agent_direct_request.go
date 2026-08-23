package swarm

import (
	"encoding/json"
	"log"
	"strings"

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
	result := a.processMessageStructured(input, history)
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
