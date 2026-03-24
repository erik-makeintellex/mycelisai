package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/exchange"
)

func (r *InternalToolRegistry) handleListExchangeChannels(ctx context.Context, args map[string]any) (string, error) {
	if r.exchange == nil {
		return "", fmt.Errorf("managed exchange not available")
	}
	ctx = exchange.WithActor(ctx, exchange.Actor{Role: defaultExchangeRole(args)})
	channels, err := r.exchange.ListChannels(ctx)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(channels)
	return string(data), nil
}

func (r *InternalToolRegistry) handleListExchangeThreads(ctx context.Context, args map[string]any) (string, error) {
	if r.exchange == nil {
		return "", fmt.Errorf("managed exchange not available")
	}
	ctx = exchange.WithActor(ctx, exchange.Actor{Role: defaultExchangeRole(args)})
	channel, _ := args["channel"].(string)
	status, _ := args["status"].(string)
	limit := 20
	if raw, ok := args["limit"].(float64); ok && raw > 0 {
		limit = int(raw)
	}
	threads, err := r.exchange.ListThreads(ctx, channel, status, limit)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(threads)
	return string(data), nil
}

func (r *InternalToolRegistry) handleSearchExchangeItems(ctx context.Context, args map[string]any) (string, error) {
	if r.exchange == nil {
		return "", fmt.Errorf("managed exchange not available")
	}
	ctx = exchange.WithActor(ctx, exchange.Actor{Role: defaultExchangeRole(args)})
	query, _ := args["query"].(string)
	if strings.TrimSpace(query) == "" {
		return "", fmt.Errorf("search_exchange_items requires 'query'")
	}
	limit := 5
	if raw, ok := args["limit"].(float64); ok && raw > 0 {
		limit = int(raw)
	}
	results, err := r.exchange.Search(ctx, query, limit)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(results)
	return string(data), nil
}

func (r *InternalToolRegistry) handleCreateExchangeThread(ctx context.Context, args map[string]any) (string, error) {
	if r.exchange == nil {
		return "", fmt.Errorf("managed exchange not available")
	}
	channel, _ := args["channel"].(string)
	title, _ := args["title"].(string)
	threadType, _ := args["thread_type"].(string)
	createdBy, _ := args["created_by"].(string)
	if strings.TrimSpace(channel) == "" || strings.TrimSpace(title) == "" {
		return "", fmt.Errorf("create_exchange_thread requires 'channel' and 'title'")
	}
	participants := []string{}
	if raw, ok := args["participants"].([]any); ok {
		for _, item := range raw {
			if value, ok := item.(string); ok && strings.TrimSpace(value) != "" {
				participants = append(participants, value)
			}
		}
	}
	ctx = exchange.WithActor(ctx, exchange.Actor{Role: defaultExchangeRole(map[string]any{"created_by": createdBy, "role": args["role"]})})
	thread, err := r.exchange.CreateThread(ctx, exchange.CreateThreadInput{
		ChannelName:   channel,
		ThreadType:    threadType,
		Title:         title,
		Participants:  participants,
		AllowedReviewers: stringSlice(args["allowed_reviewers"]),
		EscalationRights: stringSlice(args["escalation_rights"]),
		ContinuityKey: stringValue(args["continuity_key"]),
		CreatedBy:     createdBy,
	})
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(thread)
	return string(data), nil
}

func (r *InternalToolRegistry) handlePublishExchangeItem(ctx context.Context, args map[string]any) (string, error) {
	if r.exchange == nil {
		return "", fmt.Errorf("managed exchange not available")
	}
	channel, _ := args["channel"].(string)
	schemaID, _ := args["schema_id"].(string)
	payload, _ := args["payload"].(map[string]any)
	createdBy, _ := args["created_by"].(string)
	if strings.TrimSpace(channel) == "" || strings.TrimSpace(schemaID) == "" || payload == nil {
		return "", fmt.Errorf("publish_exchange_item requires 'channel', 'schema_id', and 'payload'")
	}
	ctx = exchange.WithActor(ctx, exchange.Actor{Role: defaultExchangeRole(map[string]any{"created_by": createdBy, "role": args["role"]})})
	var threadID *uuid.UUID
	if raw := stringValue(args["thread_id"]); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return "", fmt.Errorf("invalid thread_id")
		}
		threadID = &parsed
	}
	item, err := r.exchange.Publish(ctx, exchange.PublishInput{
		ChannelName: channel,
		SchemaID:    schemaID,
		Payload:     payload,
		CreatedBy:   createdBy,
		AddressedTo: stringValue(args["addressed_to"]),
		ThreadID:    threadID,
		Visibility:  stringValue(args["visibility"]),
		SensitivityClass: stringValue(args["sensitivity_class"]),
		SourceRole:  stringValue(args["source_role"]),
		SourceTeam:  stringValue(args["source_team"]),
		TargetRole:  stringValue(args["target_role"]),
		TargetTeam:  stringValue(args["target_team"]),
		AllowedConsumers: stringSlice(args["allowed_consumers"]),
		CapabilityID: stringValue(args["capability_id"]),
		TrustClass:  stringValue(args["trust_class"]),
		ReviewRequired: boolValue(args["review_required"]),
		Summary:     stringValue(args["summary"]),
	})
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(item)
	return string(data), nil
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func stringSlice(v any) []string {
	switch raw := v.(type) {
	case []string:
		return raw
	case []any:
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func boolValue(v any) bool {
	b, _ := v.(bool)
	return b
}

func defaultExchangeRole(args map[string]any) string {
	for _, key := range []string{"role", "created_by"} {
		if value := stringValue(args[key]); strings.TrimSpace(value) != "" {
			return value
		}
	}
	return "soma"
}
