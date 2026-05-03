package exchange

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) Search(ctx context.Context, query string, limit int) ([]SearchHit, error) {
	if strings.TrimSpace(query) == "" || s.Embed == nil || s.VectorBank == nil {
		return []SearchHit{}, nil
	}
	if limit <= 0 {
		limit = 5
	}
	vec, err := s.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed exchange search query: %w", err)
	}
	results, err := s.VectorBank.SemanticSearch(ctx, vec, limit*3)
	if err != nil {
		return nil, fmt.Errorf("semantic search exchange items: %w", err)
	}

	hits := []SearchHit{}
	actor := ActorFromContext(ctx)
	for _, result := range results {
		if src := fmt.Sprintf("%v", result.Metadata["source"]); src != "exchange_item" {
			continue
		}
		idStr := fmt.Sprintf("%v", result.Metadata["exchange_item_id"])
		itemID, parseErr := uuid.Parse(idStr)
		if parseErr != nil {
			continue
		}
		item, getErr := s.getItem(ctx, itemID)
		if getErr != nil {
			continue
		}
		channel, channelErr := s.getChannelByName(ctx, item.ChannelName)
		if channelErr != nil {
			continue
		}
		if actor.Role != "" && !canReadItem(actor, channel, item) && !actor.IsAdmin() {
			continue
		}
		hits = append(hits, SearchHit{Item: *item, Distance: result.Score, ChannelID: fmt.Sprintf("%v", result.Metadata["channel_id"])})
		if len(hits) >= limit {
			break
		}
	}
	return hits, nil
}

func (s *Service) indexItem(ctx context.Context, channel *Channel, item *ExchangeItem, payload map[string]any) {
	if s.Embed == nil || s.VectorBank == nil || item == nil || channel == nil {
		return
	}
	summary := strings.TrimSpace(item.Summary)
	if summary == "" {
		return
	}
	vec, err := s.Embed(ctx, summary)
	if err != nil {
		return
	}
	meta := map[string]any{
		"source":            "exchange_item",
		"exchange_item_id":  item.ID.String(),
		"channel_id":        channel.ID.String(),
		"channel_name":      channel.Name,
		"schema_id":         item.SchemaID,
		"created_by":        item.CreatedBy,
		"visibility":        item.Visibility,
		"sensitivity_class": item.SensitivityClass,
		"capability_id":     item.CapabilityID,
		"trust_class":       item.TrustClass,
	}
	if item.ThreadID != nil {
		meta["thread_id"] = item.ThreadID.String()
	}
	if continuity, ok := payload["continuity_key"]; ok {
		meta["continuity_key"] = continuity
	}
	_ = s.VectorBank.StoreVector(ctx, summary, vec, meta)
}
