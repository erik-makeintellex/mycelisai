package swarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func stripToolCallJSON(text string) string {
	keyword := `"tool_call"`
	idx := strings.Index(text, keyword)
	if idx == -1 {
		return text
	}
	start := -1
	for i := idx - 1; i >= 0; i-- {
		if text[i] == '{' {
			start = i
			break
		}
	}
	if start == -1 {
		return text
	}
	if cleaned := strings.TrimSpace(text[:start]); cleaned != "" {
		return cleaned
	}
	return text
}

func (a *Agent) publishToolBusSignal(payloadKind protocol.SignalPayloadKind, sourceKind protocol.SignalSourceKind, payload map[string]any) {
	if a.nc == nil || strings.TrimSpace(a.TeamID) == "" {
		return
	}
	subject := ""
	switch payloadKind {
	case protocol.PayloadKindStatus:
		subject = fmt.Sprintf(protocol.TopicTeamSignalStatus, a.TeamID)
	case protocol.PayloadKindResult:
		subject = fmt.Sprintf(protocol.TopicTeamSignalResult, a.TeamID)
	default:
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Agent [%s] signal payload marshal failed: %v", a.Manifest.ID, err)
		return
	}
	sourceChannel := fmt.Sprintf(protocol.TopicTeamInternalTrigger, a.TeamID)
	wrapped, err := protocol.WrapSignalPayloadWithMeta(sourceKind, sourceChannel, payloadKind, a.runID, a.TeamID, a.Manifest.ID, raw)
	if err != nil {
		log.Printf("Agent [%s] signal envelope wrap failed: %v", a.Manifest.ID, err)
		return
	}
	if err := a.nc.Publish(subject, wrapped); err != nil {
		log.Printf("Agent [%s] publish signal failed on [%s]: %v", a.Manifest.ID, subject, err)
		return
	}
	if a.internalTools != nil {
		channelKey := resolveSignalCheckpointChannelKey(subject, nil)
		metadata := map[string]any{"subject": subject, "source_kind": string(sourceKind), "payload_kind": string(payloadKind), "team_id": a.TeamID, "agent_id": a.Manifest.ID}
		if strings.TrimSpace(a.runID) != "" {
			metadata["run_id"] = strings.TrimSpace(a.runID)
		}
		if _, err := a.internalTools.upsertSignalCheckpoint(a.ctx, channelKey, a.Manifest.ID, string(wrapped), metadata); err != nil {
			log.Printf("Agent [%s] checkpoint update failed on [%s]: %v", a.Manifest.ID, channelKey, err)
		}
	}
}

func (a *Agent) handleTrigger(msg *nats.Msg) {
	select {
	case <-a.ctx.Done():
		if msg.Reply != "" {
			msg.Respond([]byte("Agent shutting down."))
		}
		return
	default:
	}
	input := normalizeTeamTriggerInput(msg.Data)
	log.Printf("Agent [%s] thinking about: %s", a.Manifest.ID, input)
	responseText := a.processMessage(input, nil)
	if responseText == "" {
		if msg.Reply != "" {
			msg.Respond([]byte(fmt.Sprintf("[%s] No response — LLM may be unavailable.", a.Manifest.ID)))
		}
		return
	}
	if msg.Reply != "" {
		msg.Respond([]byte(responseText))
	}
	a.nc.Publish(fmt.Sprintf(protocol.TopicTeamInternalRespond, a.TeamID), []byte(responseText))
	log.Printf("Agent [%s] replied.", a.Manifest.ID)
}

func normalizeTeamTriggerInput(data []byte) string {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return ""
	}

	var ask protocol.TeamAsk
	if err := json.Unmarshal(trimmed, &ask); err == nil && !ask.IsZero() {
		return renderTeamAskPrompt(ask.Normalize())
	}
	return string(data)
}

func renderTeamAskPrompt(ask protocol.TeamAsk) string {
	var sb strings.Builder
	sb.WriteString("You have received a structured team ask.\n")
	sb.WriteString("Use the ask to stay aligned on mission, scope, and proof needs.\n")
	sb.WriteString("Do not force your response into a rigid template unless the ask explicitly requires one.\n")
	sb.WriteString("Deliver the best output for the job while making sure it satisfies the ask goal, constraints, and required evidence.\n")
	sb.WriteString(fmt.Sprintf("Ask kind: %s\n", ask.AskKind))
	sb.WriteString(fmt.Sprintf("Lane role: %s\n", ask.LaneRole))
	sb.WriteString(fmt.Sprintf("Goal: %s\n", ask.Goal))
	if len(ask.OwnedScope) > 0 {
		sb.WriteString("Owned scope:\n")
		for _, item := range ask.OwnedScope {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	if len(ask.Constraints) > 0 {
		sb.WriteString("Constraints:\n")
		for _, item := range ask.Constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	if len(ask.RequiredCapabilities) > 0 {
		sb.WriteString("Required capabilities:\n")
		for _, item := range ask.RequiredCapabilities {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	if len(ask.ExitCriteria) > 0 {
		sb.WriteString("Exit criteria:\n")
		for _, item := range ask.ExitCriteria {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	if len(ask.EvidenceRequired) > 0 {
		sb.WriteString("Evidence required:\n")
		for _, item := range ask.EvidenceRequired {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	if ask.Context != nil {
		if raw, err := json.Marshal(ask.Context); err == nil {
			sb.WriteString(fmt.Sprintf("Context: %s\n", string(raw)))
		}
	}
	sb.WriteString("Complete the ask within scope, match the mission, and report the outcome clearly.")
	return sb.String()
}

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
