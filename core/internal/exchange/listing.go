package exchange

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) ListThreads(ctx context.Context, channelName, status string, limit int) ([]Thread, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT t.id, t.channel_id, c.name, t.thread_type, t.title, t.status, t.participants, t.allowed_reviewers, t.escalation_rights, t.continuity_key, t.created_by, t.metadata, t.created_at, t.updated_at
		FROM exchange_threads t
		JOIN exchange_channels c ON c.id = t.channel_id
		WHERE ($1 = '' OR c.name = $1)
		  AND ($2 = '' OR t.status = $2)
		ORDER BY t.updated_at DESC
		LIMIT $3
	`, channelName, status, limit)
	if err != nil {
		return nil, fmt.Errorf("list exchange threads: %w", err)
	}
	defer rows.Close()
	threads, err := scanThreads(rows)
	if err != nil {
		return nil, err
	}
	actor := ActorFromContext(ctx)
	if actor.IsAdmin() {
		return threads, nil
	}
	channels, err := s.ListChannels(WithActor(ctx, actor))
	if err != nil {
		return nil, err
	}
	channelByID := map[uuid.UUID]Channel{}
	for _, channel := range channels {
		channelByID[channel.ID] = channel
	}
	filtered := make([]Thread, 0, len(threads))
	for _, thread := range threads {
		threadCopy := thread
		channel, ok := channelByID[thread.ChannelID]
		if ok && canAccessThread(actor, &channel, &threadCopy) {
			filtered = append(filtered, threadCopy)
		}
	}
	return filtered, nil
}

func (s *Service) ListItems(ctx context.Context, channelName string, threadID *uuid.UUID, limit int) ([]ExchangeItem, error) {
	if limit <= 0 {
		limit = 50
	}
	var threadArg any
	if threadID != nil {
		threadArg = *threadID
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT i.id, i.channel_id, c.name, i.schema_id, i.payload, i.created_by, COALESCE(i.addressed_to, ''), i.thread_id, i.visibility, i.sensitivity_class, i.source_role, COALESCE(i.source_team, ''), COALESCE(i.target_role, ''), COALESCE(i.target_team, ''), i.allowed_consumers, COALESCE(i.capability_id, ''), i.trust_class, i.review_required, i.metadata, i.summary, i.created_at
		FROM exchange_items i
		JOIN exchange_channels c ON c.id = i.channel_id
		WHERE ($1 = '' OR c.name = $1)
		  AND ($2::uuid IS NULL OR i.thread_id = $2::uuid)
		ORDER BY i.created_at DESC
		LIMIT $3
	`, channelName, threadArg, limit)
	if err != nil {
		return nil, fmt.Errorf("list exchange items: %w", err)
	}
	defer rows.Close()
	items, err := scanItems(rows)
	if err != nil {
		return nil, err
	}
	actor := ActorFromContext(ctx)
	if actor.IsAdmin() {
		return items, nil
	}
	channels, err := s.ListChannels(WithActor(ctx, actor))
	if err != nil {
		return nil, err
	}
	channelByID := map[uuid.UUID]Channel{}
	for _, channel := range channels {
		channelByID[channel.ID] = channel
	}
	filtered := make([]ExchangeItem, 0, len(items))
	for _, item := range items {
		itemCopy := item
		channel, ok := channelByID[item.ChannelID]
		if ok && canReadItem(actor, &channel, &itemCopy) {
			filtered = append(filtered, itemCopy)
		}
	}
	return filtered, nil
}
