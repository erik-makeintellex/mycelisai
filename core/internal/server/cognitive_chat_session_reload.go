package server

import (
	"context"
	"log"
)

func (s *AdminServer) restoreChatSessionMessages(
	ctx context.Context,
	sessionID string,
	messages []chatRequestMessage,
) ([]chatRequestMessage, int) {
	if sessionID == "" || s.Conversations == nil {
		return messages, 0
	}
	priorTurns, err := s.Conversations.GetSessionTurns(ctx, sessionID)
	if err != nil {
		log.Printf("[chat] prior session conversation lookup failed: %v", err)
		return messages, 0
	}
	return mergePersistedSessionMessages(messages, priorTurns), len(priorTurns)
}
