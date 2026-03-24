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
		SELECT id, name, channel_type, owner, participants, schema_id, retention_policy, visibility, description, metadata, created_at
		FROM exchange_channels
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list exchange channels: %w", err)
	}
	defer rows.Close()
	return scanChannels(rows)
}

func (s *Service) ListThreads(ctx context.Context, channelName, status string, limit int) ([]Thread, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT t.id, t.channel_id, c.name, t.thread_type, t.title, t.status, t.participants, t.continuity_key, t.created_by, t.metadata, t.created_at, t.updated_at
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
	return scanThreads(rows)
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
		SELECT i.id, i.channel_id, c.name, i.schema_id, i.payload, i.created_by, COALESCE(i.addressed_to, ''), i.thread_id, i.visibility, i.metadata, i.summary, i.created_at
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
	return scanItems(rows)
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
	thread := &Thread{ChannelID: channel.ID, ChannelName: channel.Name}
	row := s.DB.QueryRowContext(ctx, `
		INSERT INTO exchange_threads (channel_id, thread_type, title, status, participants, continuity_key, created_by, metadata)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8)
		RETURNING id, created_at, updated_at
	`, channel.ID, input.ThreadType, input.Title, input.Status, marshalSlice(input.Participants), input.ContinuityKey, input.CreatedBy, marshalMap(input.Metadata))
	if err := row.Scan(&thread.ID, &thread.CreatedAt, &thread.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create exchange thread: %w", err)
	}
	thread.ThreadType = input.ThreadType
	thread.Title = input.Title
	thread.Status = input.Status
	thread.Participants = append([]string{}, input.Participants...)
	thread.ContinuityKey = input.ContinuityKey
	thread.CreatedBy = input.CreatedBy
	thread.Metadata = marshalMap(input.Metadata)
	return thread, nil
}

func (s *Service) Publish(ctx context.Context, input PublishInput) (*ExchangeItem, error) {
	if err := validatePayload(input.SchemaID, input.Payload); err != nil {
		return nil, err
	}
	channel, err := s.ensureChannel(ctx, input.ChannelName)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Visibility) == "" {
		input.Visibility = channel.Visibility
	}
	if strings.TrimSpace(input.CreatedBy) == "" {
		input.CreatedBy = "system"
	}
	if strings.TrimSpace(input.Summary) == "" {
		if text, _ := input.Payload["summary"].(string); strings.TrimSpace(text) != "" {
			input.Summary = text
		}
	}
	item := &ExchangeItem{ChannelID: channel.ID, ChannelName: channel.Name}
	row := s.DB.QueryRowContext(ctx, `
		INSERT INTO exchange_items (channel_id, schema_id, payload, created_by, addressed_to, thread_id, visibility, metadata, summary)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9)
		RETURNING id, created_at
	`, channel.ID, input.SchemaID, marshalMap(input.Payload), input.CreatedBy, input.AddressedTo, input.ThreadID, input.Visibility, marshalMap(input.Metadata), input.Summary)
	if err := row.Scan(&item.ID, &item.CreatedAt); err != nil {
		return nil, fmt.Errorf("publish exchange item: %w", err)
	}
	item.SchemaID = input.SchemaID
	item.Payload = marshalMap(input.Payload)
	item.CreatedBy = input.CreatedBy
	item.AddressedTo = input.AddressedTo
	item.ThreadID = input.ThreadID
	item.Visibility = input.Visibility
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
	}
	if item.ThreadID != nil {
		meta["thread_id"] = item.ThreadID.String()
	}
	if continuity, ok := payload["continuity_key"]; ok {
		meta["continuity_key"] = continuity
	}
	_ = s.VectorBank.StoreVector(ctx, summary, vec, meta)
}
