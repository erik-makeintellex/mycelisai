package exchange

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/memory"
)

type VectorStore interface {
	StoreVector(ctx context.Context, content string, embedding []float64, metadata map[string]any) error
	SemanticSearch(ctx context.Context, embedding []float64, limit int) ([]memory.VectorResult, error)
}

type Service struct {
	DB         *sql.DB
	Embed      EmbedFunc
	VectorBank VectorStore
}

func NewService(db *sql.DB, embed EmbedFunc, vectors VectorStore) *Service {
	return &Service{DB: db, Embed: embed, VectorBank: vectors}
}

func (s *Service) Bootstrap(ctx context.Context) error {
	if s.DB == nil {
		return nil
	}
	if err := s.bootstrapFields(ctx); err != nil {
		return err
	}
	if err := s.bootstrapCapabilities(ctx); err != nil {
		return err
	}
	if err := s.bootstrapSchemas(ctx); err != nil {
		return err
	}
	return s.bootstrapChannels(ctx)
}

func (s *Service) ListFields(ctx context.Context) ([]FieldDefinition, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT name, field_type, semantic_meaning, indexed, visibility, usage_contexts, created_at
		FROM exchange_field_registry
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list exchange fields: %w", err)
	}
	defer rows.Close()

	out := []FieldDefinition{}
	for rows.Next() {
		var def FieldDefinition
		var usageJSON []byte
		if err := rows.Scan(&def.Name, &def.Type, &def.SemanticMeaning, &def.Indexed, &def.Visibility, &usageJSON, &def.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan exchange field: %w", err)
		}
		_ = json.Unmarshal(usageJSON, &def.UsageContexts)
		out = append(out, def)
	}
	return out, rows.Err()
}

func (s *Service) ListSchemas(ctx context.Context) ([]SchemaDefinition, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, label, description, required_fields, optional_fields, required_capabilities, created_at
		FROM exchange_schema_registry
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list exchange schemas: %w", err)
	}
	defer rows.Close()

	out := []SchemaDefinition{}
	for rows.Next() {
		var def SchemaDefinition
		var reqJSON, optJSON, capsJSON []byte
		if err := rows.Scan(&def.ID, &def.Label, &def.Description, &reqJSON, &optJSON, &capsJSON, &def.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan exchange schema: %w", err)
		}
		_ = json.Unmarshal(reqJSON, &def.RequiredFields)
		_ = json.Unmarshal(optJSON, &def.OptionalFields)
		_ = json.Unmarshal(capsJSON, &def.RequiredCapabilities)
		out = append(out, def)
	}
	return out, rows.Err()
}

func (s *Service) ListChannels(ctx context.Context) ([]Channel, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, name, channel_type, owner, participants, reviewers, schema_id, retention_policy, visibility, sensitivity_class, description, metadata, created_at
		FROM exchange_channels
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list exchange channels: %w", err)
	}
	defer rows.Close()
	channels, err := scanChannels(rows)
	if err != nil {
		return nil, err
	}
	actor := ActorFromContext(ctx)
	if actor.IsAdmin() {
		return channels, nil
	}
	filtered := make([]Channel, 0, len(channels))
	for _, channel := range channels {
		channelCopy := channel
		if canReadChannel(actor, &channelCopy) {
			filtered = append(filtered, channelCopy)
		}
	}
	return filtered, nil
}

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

func (s *Service) CreateThread(ctx context.Context, input CreateThreadInput) (*Thread, error) {
	channel, err := s.ensureChannel(ctx, input.ChannelName)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.ThreadType) == "" {
		input.ThreadType = "work_thread"
	}
	if strings.TrimSpace(input.Status) == "" {
		input.Status = "active"
	}
	if strings.TrimSpace(input.CreatedBy) == "" {
		input.CreatedBy = "soma"
	}
	actor := ActorFromContext(ctx)
	if actor.Role == "" {
		actor = defaultActorFromCreatedBy(input.CreatedBy)
	}
	if !canWriteChannel(actor, channel) {
		return nil, fmt.Errorf("role %s cannot create threads in %s", actor.Role, channel.Name)
	}
	if len(input.AllowedReviewers) == 0 {
		input.AllowedReviewers = append([]string{}, channel.Reviewers...)
	}
	if len(input.EscalationRights) == 0 {
		input.EscalationRights = []string{"soma", "team_lead", "review"}
	}
	thread := &Thread{ChannelID: channel.ID, ChannelName: channel.Name}
	row := s.DB.QueryRowContext(ctx, `
		INSERT INTO exchange_threads (channel_id, thread_type, title, status, participants, allowed_reviewers, escalation_rights, continuity_key, created_by, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9, $10)
		RETURNING id, created_at, updated_at
	`, channel.ID, input.ThreadType, input.Title, input.Status, marshalSlice(input.Participants), marshalSlice(input.AllowedReviewers), marshalSlice(input.EscalationRights), input.ContinuityKey, input.CreatedBy, marshalMap(input.Metadata))
	if err := row.Scan(&thread.ID, &thread.CreatedAt, &thread.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create exchange thread: %w", err)
	}
	thread.ThreadType = input.ThreadType
	thread.Title = input.Title
	thread.Status = input.Status
	thread.Participants = append([]string{}, input.Participants...)
	thread.AllowedReviewers = append([]string{}, input.AllowedReviewers...)
	thread.EscalationRights = append([]string{}, input.EscalationRights...)
	thread.ContinuityKey = input.ContinuityKey
	thread.CreatedBy = input.CreatedBy
	thread.Metadata = marshalMap(input.Metadata)
	return thread, nil
}

func (s *Service) Publish(ctx context.Context, input PublishInput) (*ExchangeItem, error) {
	channel, err := s.ensureChannel(ctx, input.ChannelName)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.CreatedBy) == "" {
		input.CreatedBy = "system"
	}
	actor := ActorFromContext(ctx)
	if actor.Role == "" {
		fallbackRole := input.SourceRole
		if strings.TrimSpace(fallbackRole) == "" {
			fallbackRole = input.CreatedBy
		}
		actor = defaultActorFromCreatedBy(fallbackRole)
	}
	if !canWriteChannel(actor, channel) {
		return nil, fmt.Errorf("role %s cannot publish to %s", actor.Role, channel.Name)
	}
	input, capability, err := enrichPublishInput(input, channel)
	if err != nil {
		return nil, err
	}
	if !canUseCapability(actor, capability) {
		return nil, fmt.Errorf("role %s cannot use capability %s", actor.Role, input.CapabilityID)
	}
	if err := validatePayload(input.SchemaID, input.Payload); err != nil {
		return nil, err
	}
	if input.ThreadID != nil {
		thread, err := s.getThread(ctx, *input.ThreadID)
		if err != nil {
			return nil, fmt.Errorf("load exchange thread: %w", err)
		}
		if thread.ChannelID != channel.ID {
			return nil, fmt.Errorf("exchange thread does not belong to channel %s", channel.Name)
		}
		if !canPublishToThread(actor, channel, thread) {
			return nil, fmt.Errorf("role %s cannot publish into thread %s", actor.Role, thread.ID.String())
		}
	}
	if strings.TrimSpace(input.Summary) == "" {
		if text, _ := input.Payload["summary"].(string); strings.TrimSpace(text) != "" {
			input.Summary = text
		}
	}
	item := &ExchangeItem{ChannelID: channel.ID, ChannelName: channel.Name}
	row := s.DB.QueryRowContext(ctx, `
		INSERT INTO exchange_items (channel_id, schema_id, payload, created_by, addressed_to, thread_id, visibility, sensitivity_class, source_role, source_team, target_role, target_team, allowed_consumers, capability_id, trust_class, review_required, metadata, summary)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9, NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, ''), $13, NULLIF($14, ''), $15, $16, $17, $18)
		RETURNING id, created_at
	`, channel.ID, input.SchemaID, marshalMap(input.Payload), input.CreatedBy, input.AddressedTo, input.ThreadID, input.Visibility, input.SensitivityClass, input.SourceRole, input.SourceTeam, input.TargetRole, input.TargetTeam, marshalSlice(input.AllowedConsumers), input.CapabilityID, input.TrustClass, input.ReviewRequired, marshalMap(input.Metadata), input.Summary)
	if err := row.Scan(&item.ID, &item.CreatedAt); err != nil {
		return nil, fmt.Errorf("publish exchange item: %w", err)
	}
	item.SchemaID = input.SchemaID
	item.Payload = marshalMap(input.Payload)
	item.CreatedBy = input.CreatedBy
	item.AddressedTo = input.AddressedTo
	item.ThreadID = input.ThreadID
	item.Visibility = input.Visibility
	item.SensitivityClass = input.SensitivityClass
	item.SourceRole = input.SourceRole
	item.SourceTeam = input.SourceTeam
	item.TargetRole = input.TargetRole
	item.TargetTeam = input.TargetTeam
	item.AllowedConsumers = append([]string{}, input.AllowedConsumers...)
	item.CapabilityID = input.CapabilityID
	item.TrustClass = input.TrustClass
	item.ReviewRequired = input.ReviewRequired
	item.Metadata = marshalMap(input.Metadata)
	item.Summary = input.Summary
	s.indexItem(ctx, channel, item, input.Payload)
	return item, nil
}

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
		"source":           "exchange_item",
		"exchange_item_id": item.ID.String(),
		"channel_id":       channel.ID.String(),
		"channel_name":     channel.Name,
		"schema_id":        item.SchemaID,
		"created_by":       item.CreatedBy,
		"visibility":       item.Visibility,
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
