package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// TempMemoryEntry represents a persisted short-horizon working-memory checkpoint.
type TempMemoryEntry struct {
	ID           string         `json:"id"`
	TenantID     string         `json:"tenant_id"`
	ChannelKey   string         `json:"channel_key"`
	OwnerAgentID string         `json:"owner_agent_id"`
	Content      string         `json:"content"`
	Metadata     map[string]any `json:"metadata"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// PutTempMemory stores a temporary working-memory checkpoint.
// ttlMinutes <= 0 means no expiry.
func (s *Service) PutTempMemory(ctx context.Context, tenantID, channelKey, ownerAgentID, content string, metadata map[string]any, ttlMinutes int) (string, error) {
	if s == nil || s.db == nil {
		return "", fmt.Errorf("memory service offline")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	if channelKey == "" {
		return "", fmt.Errorf("channel_key is required")
	}
	if ownerAgentID == "" {
		ownerAgentID = "admin"
	}
	if content == "" {
		return "", fmt.Errorf("content is required")
	}
	if metadata == nil {
		metadata = map[string]any{}
	}

	metaJSON, _ := json.Marshal(metadata)

	var expiresAt any = nil
	if ttlMinutes > 0 {
		exp := time.Now().Add(time.Duration(ttlMinutes) * time.Minute)
		expiresAt = exp
	}

	var id string
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO temp_memory_channels
		    (tenant_id, channel_key, owner_agent_id, content, metadata, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text
	`, tenantID, channelKey, ownerAgentID, content, metaJSON, expiresAt).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("put temp memory: %w", err)
	}
	return id, nil
}

// GetTempMemory fetches recent, non-expired entries for a channel.
func (s *Service) GetTempMemory(ctx context.Context, tenantID, channelKey string, limit int) ([]TempMemoryEntry, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("memory service offline")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	if channelKey == "" {
		return nil, fmt.Errorf("channel_key is required")
	}
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id::text, tenant_id, channel_key, owner_agent_id, content, metadata,
		       expires_at, created_at, updated_at
		FROM temp_memory_channels
		WHERE tenant_id = $1
		  AND channel_key = $2
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY updated_at DESC
		LIMIT $3
	`, tenantID, channelKey, limit)
	if err != nil {
		return nil, fmt.Errorf("get temp memory: %w", err)
	}
	defer rows.Close()

	var out []TempMemoryEntry
	for rows.Next() {
		var e TempMemoryEntry
		var metaJSON []byte
		var expires sql.NullTime
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.ChannelKey, &e.OwnerAgentID, &e.Content,
			&metaJSON, &expires, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan temp memory: %w", err)
		}
		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &e.Metadata)
		}
		if e.Metadata == nil {
			e.Metadata = map[string]any{}
		}
		if expires.Valid {
			t := expires.Time
			e.ExpiresAt = &t
		}
		out = append(out, e)
	}
	if out == nil {
		out = []TempMemoryEntry{}
	}
	return out, nil
}

// ClearTempMemory deletes entries for the given channel.
func (s *Service) ClearTempMemory(ctx context.Context, tenantID, channelKey string) (int64, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("memory service offline")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	if channelKey == "" {
		return 0, fmt.Errorf("channel_key is required")
	}

	res, err := s.db.ExecContext(ctx, `
		DELETE FROM temp_memory_channels
		WHERE tenant_id = $1 AND channel_key = $2
	`, tenantID, channelKey)
	if err != nil {
		return 0, fmt.Errorf("clear temp memory: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}
