package server

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) recordTeamWorkAskDegraded(ctx context.Context, item *protocol.TeamWorkItem, subject, degradation, details string) (protocol.TeamStatusEvent, error) {
	item.State = protocol.TeamWorkStateDegraded
	item.NeedsOperator = true
	item.DegradationState = degradation
	item.RecoveryOptions = []string{"Recover the work item after NATS/team availability is restored.", "Add steering guidance before retrying.", "Archive if the work is no longer needed."}
	event := teamWorkAskStatusEvent(*item, protocol.TeamWorkStateDegraded, "Team ask degraded", details, "operator_attention", "Recover or steer this work item before retrying.", []string{degradation})
	if err := s.insertTeamStatusEventDB(ctx, &event); err != nil {
		return event, err
	}
	if err := s.updateTeamWorkItemLastEventDB(ctx, item, event); err != nil {
		return event, err
	}
	interaction := protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(),
		TeamID:        item.TeamID,
		WorkItemID:    item.WorkItemID,
		SourceKind:    string(protocol.SourceKindWebAPI),
		SourceChannel: teamWorkAskSourceChannel,
		ActorRef:      "Soma",
		Verb:          "degraded",
		Summary:       details,
		PayloadKind:   string(protocol.PayloadKindError),
		Payload:       map[string]any{"subject": subject, "degradation_state": degradation},
		Version:       "v1",
	})
	if err := s.insertTeamInteractionDB(ctx, &interaction); err != nil {
		return event, err
	}
	return event, nil
}

func (s *AdminServer) recordTeamWorkAskOutput(ctx context.Context, item *protocol.TeamWorkItem, subject, reply string) (protocol.TeamStatusEvent, error) {
	item.State = protocol.TeamWorkStateOutputReady
	item.NeedsOperator = false
	item.DegradationState = ""
	event := teamWorkAskStatusEvent(*item, protocol.TeamWorkStateOutputReady, "Team response ready", teamAskReplyDetails(reply), "team_response", "Review the response and decide whether to retain, steer, or ask for follow-up.", nil)
	if err := s.insertTeamStatusEventDB(ctx, &event); err != nil {
		return event, err
	}
	if err := s.updateTeamWorkItemLastEventDB(ctx, item, event); err != nil {
		return event, err
	}
	interaction := protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(),
		TeamID:        item.TeamID,
		WorkItemID:    item.WorkItemID,
		SourceKind:    string(protocol.SourceKindWebAPI),
		SourceChannel: teamWorkAskSourceChannel,
		ActorRef:      item.TeamID,
		Verb:          "response",
		Summary:       firstNonEmptyString(reply, "Team returned an empty response."),
		PayloadKind:   string(protocol.PayloadKindResult),
		Payload:       map[string]any{"subject": subject, "reply": reply},
		Version:       "v1",
	})
	if err := s.insertTeamInteractionDB(ctx, &interaction); err != nil {
		return event, err
	}
	return event, nil
}

func teamAskReplyDetails(reply string) string {
	trimmed := strings.Join(strings.Fields(reply), " ")
	if trimmed == "" {
		return "The team returned an empty bounded response for this ask."
	}
	const maxDetailRunes = 180
	runes := []rune(trimmed)
	if len(runes) > maxDetailRunes {
		trimmed = string(runes[:maxDetailRunes]) + "..."
	}
	return "Reply: " + trimmed
}

func teamWorkAskReplyReadable(reply string) bool {
	trimmed := strings.TrimSpace(reply)
	if trimmed == "" {
		return false
	}
	if strings.Contains(trimmed, "No response — LLM may be unavailable") ||
		strings.Contains(trimmed, "No response - LLM may be unavailable") {
		return false
	}
	normalized := strings.Trim(trimmed, "` \t\r\n")
	if strings.HasPrefix(normalized, "json") {
		normalized = strings.TrimSpace(strings.TrimPrefix(normalized, "json"))
	}
	return !(strings.HasPrefix(normalized, "{") && strings.Contains(normalized, `"tool_call"`))
}
