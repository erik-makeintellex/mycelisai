package exchange

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) bootstrapFields(ctx context.Context) error {
	for _, field := range SeedFields {
		if _, err := s.DB.ExecContext(ctx, `
			INSERT INTO exchange_field_registry (name, field_type, semantic_meaning, indexed, visibility, usage_contexts)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (name) DO UPDATE SET
				field_type = EXCLUDED.field_type,
				semantic_meaning = EXCLUDED.semantic_meaning,
				indexed = EXCLUDED.indexed,
				visibility = EXCLUDED.visibility,
				usage_contexts = EXCLUDED.usage_contexts
		`, field.Name, field.Type, field.SemanticMeaning, field.Indexed, field.Visibility, marshalSlice(field.UsageContexts)); err != nil {
			return fmt.Errorf("bootstrap exchange field %s: %w", field.Name, err)
		}
	}
	return nil
}

func (s *Service) bootstrapCapabilities(ctx context.Context) error {
	for _, capability := range SeedCapabilities {
		if _, err := s.DB.ExecContext(ctx, `
			INSERT INTO exchange_capability_registry (id, label, source, risk_class, default_allowed_roles, audit_required, approval_required, description)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE SET
				label = EXCLUDED.label,
				source = EXCLUDED.source,
				risk_class = EXCLUDED.risk_class,
				default_allowed_roles = EXCLUDED.default_allowed_roles,
				audit_required = EXCLUDED.audit_required,
				approval_required = EXCLUDED.approval_required,
				description = EXCLUDED.description
		`, capability.ID, capability.Label, capability.Source, capability.RiskClass, marshalSlice(capability.DefaultAllowedRoles), capability.AuditRequired, capability.ApprovalRequired, capability.Description); err != nil {
			return fmt.Errorf("bootstrap exchange capability %s: %w", capability.ID, err)
		}
	}
	return nil
}

func (s *Service) bootstrapSchemas(ctx context.Context) error {
	for _, schema := range SeedSchemas {
		if _, err := s.DB.ExecContext(ctx, `
			INSERT INTO exchange_schema_registry (id, label, description, required_fields, optional_fields, required_capabilities)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET
				label = EXCLUDED.label,
				description = EXCLUDED.description,
				required_fields = EXCLUDED.required_fields,
				optional_fields = EXCLUDED.optional_fields,
				required_capabilities = EXCLUDED.required_capabilities
		`, schema.ID, schema.Label, schema.Description, marshalSlice(schema.RequiredFields), marshalSlice(schema.OptionalFields), marshalSlice(schema.RequiredCapabilities)); err != nil {
			return fmt.Errorf("bootstrap exchange schema %s: %w", schema.ID, err)
		}
	}
	return nil
}

func (s *Service) bootstrapChannels(ctx context.Context) error {
	for _, channel := range SeedChannels {
		if _, err := s.DB.ExecContext(ctx, `
			INSERT INTO exchange_channels (name, channel_type, owner, participants, reviewers, schema_id, retention_policy, visibility, sensitivity_class, description, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (name) DO UPDATE SET
				channel_type = EXCLUDED.channel_type,
				owner = EXCLUDED.owner,
				participants = EXCLUDED.participants,
				reviewers = EXCLUDED.reviewers,
				schema_id = EXCLUDED.schema_id,
				retention_policy = EXCLUDED.retention_policy,
				visibility = EXCLUDED.visibility,
				sensitivity_class = EXCLUDED.sensitivity_class,
				description = EXCLUDED.description,
				metadata = EXCLUDED.metadata
		`, channel.Name, channel.Type, channel.Owner, marshalParticipants(channel.Participants), marshalSlice(channel.Reviewers), channel.SchemaID, channel.RetentionPolicy, channel.Visibility, channel.SensitivityClass, channel.Description, marshalJSON(channel.Metadata)); err != nil {
			return fmt.Errorf("bootstrap exchange channel %s: %w", channel.Name, err)
		}
	}
	return nil
}

func (s *Service) ensureChannel(ctx context.Context, name string) (*Channel, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("exchange channel name is required")
	}
	channel, err := s.getChannelByName(ctx, name)
	if err == nil {
		return channel, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("load exchange channel %s: %w", name, err)
	}
	return nil, fmt.Errorf("exchange channel %s is not registered", name)
}

func (s *Service) getChannelByName(ctx context.Context, name string) (*Channel, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, name, channel_type, owner, participants, reviewers, schema_id, retention_policy, visibility, sensitivity_class, description, metadata, created_at
		FROM exchange_channels
		WHERE name = $1
	`, name)
	return scanChannelRow(row)
}

func (s *Service) getThread(ctx context.Context, id uuid.UUID) (*Thread, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT t.id, t.channel_id, c.name, t.thread_type, t.title, t.status, t.participants, t.allowed_reviewers, t.escalation_rights, t.continuity_key, t.created_by, t.metadata, t.created_at, t.updated_at
		FROM exchange_threads t
		JOIN exchange_channels c ON c.id = t.channel_id
		WHERE t.id = $1
	`, id)
	var participantsJSON, reviewersJSON, escalationJSON []byte
	var thread Thread
	if err := row.Scan(&thread.ID, &thread.ChannelID, &thread.ChannelName, &thread.ThreadType, &thread.Title, &thread.Status, &participantsJSON, &reviewersJSON, &escalationJSON, &thread.ContinuityKey, &thread.CreatedBy, &thread.Metadata, &thread.CreatedAt, &thread.UpdatedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(participantsJSON, &thread.Participants)
	_ = json.Unmarshal(reviewersJSON, &thread.AllowedReviewers)
	_ = json.Unmarshal(escalationJSON, &thread.EscalationRights)
	return &thread, nil
}

func (s *Service) getItem(ctx context.Context, id uuid.UUID) (*ExchangeItem, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT i.id, i.channel_id, c.name, i.schema_id, i.payload, i.created_by, COALESCE(i.addressed_to, ''), i.thread_id, i.visibility, i.sensitivity_class, i.source_role, COALESCE(i.source_team, ''), COALESCE(i.target_role, ''), COALESCE(i.target_team, ''), i.allowed_consumers, COALESCE(i.capability_id, ''), i.trust_class, i.review_required, i.metadata, i.summary, i.created_at
		FROM exchange_items i
		JOIN exchange_channels c ON c.id = i.channel_id
		WHERE i.id = $1
	`, id)
	return scanItemRow(row)
}

func validatePayload(schemaID string, payload map[string]any) error {
	schema, ok := SchemaByID(schemaID)
	if !ok {
		return fmt.Errorf("exchange schema %s is not registered", schemaID)
	}
	for _, fieldName := range schema.RequiredFields {
		value, exists := payload[fieldName]
		if !exists {
			return fmt.Errorf("exchange schema %s requires field %s", schemaID, fieldName)
		}
		field, ok := FieldByName(fieldName)
		if ok {
			if err := validateFieldType(field.Type, value); err != nil {
				return fmt.Errorf("field %s: %w", fieldName, err)
			}
		}
	}
	for fieldName, value := range payload {
		field, ok := FieldByName(fieldName)
		if !ok {
			continue
		}
		if err := validateFieldType(field.Type, value); err != nil {
			return fmt.Errorf("field %s: %w", fieldName, err)
		}
	}
	return nil
}

func validateFieldType(fieldType string, value any) error {
	switch fieldType {
	case "string", "enum", "reference":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string-compatible value")
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean value")
		}
	case "number":
		switch value.(type) {
		case float64, float32, int, int64, int32:
		default:
			return fmt.Errorf("expected numeric value")
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("expected object value")
		}
	case "array":
		if _, ok := value.([]any); ok {
			return nil
		}
		if _, ok := value.([]string); !ok {
			return fmt.Errorf("expected array value")
		}
	}
	return nil
}

func marshalSlice[T any](in []T) []byte {
	raw, _ := json.Marshal(in)
	return raw
}

func marshalMap(in map[string]any) []byte {
	if in == nil {
		in = map[string]any{}
	}
	raw, _ := json.Marshal(in)
	return raw
}

func marshalParticipants(in []ChannelParticipant) []byte {
	raw, _ := json.Marshal(in)
	return raw
}

func marshalJSON(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte(`{}`)
	}
	return raw
}
