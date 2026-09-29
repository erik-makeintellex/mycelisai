package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
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

// PutTempMemory stores a temporary working-memory checkpoint owned by
// ownerUserID, the verified user of the writing turn (MEM-LANES).
// ownerAgentID only names the writing agent; ttlMinutes <= 0 means no expiry.
// A row with no user is refused here (MEM-LANES-3): only Core's own writers
// store one, through PutSystemTempMemory.
func (s *Service) PutTempMemory(ctx context.Context, tenantID, channelKey, ownerAgentID, content string, metadata map[string]any, ttlMinutes int, ownerUserID string) (string, error) {
	if strings.TrimSpace(ownerUserID) == "" {
		return "", fmt.Errorf("put temp memory: a row with no user can only come from a Core system writer")
	}
	return s.putTemp(ctx, tenantID, channelKey, ownerAgentID, content, OwnerLaneMetadata(metadata, ownerUserID), ttlMinutes)
}

// PutSystemTempMemory stores a no-user (system-class) row written by one of
// Core's own writers, named by writer: the bus checkpoint and AutoSummarize.
// It is the only path to a row that no-user prompts and users' signal reads
// admit (tempReadClause); model tools never call it (MEM-LANES-3).
func (s *Service) PutSystemTempMemory(ctx context.Context, tenantID, channelKey, ownerAgentID, content string, metadata map[string]any, ttlMinutes int, writer string) (string, error) {
	writer = strings.TrimSpace(writer)
	if writer == "" {
		return "", fmt.Errorf("put system temp memory: the Core writer is required")
	}
	meta := OwnerLaneMetadata(metadata, "")
	meta[SystemWriterKey] = writer
	return s.putTemp(ctx, tenantID, channelKey, ownerAgentID, content, meta, ttlMinutes)
}

func (s *Service) putTemp(ctx context.Context, tenantID, channelKey, ownerAgentID, content string, meta map[string]any, ttlMinutes int) (string, error) {
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
	metaJSON, _ := json.Marshal(meta)

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

// GetTempMemory fetches recent, non-expired entries of a channel that reader
// may read (tempReadClause).
func (s *Service) GetTempMemory(ctx context.Context, tenantID, channelKey string, limit int, reader GovernedReader) ([]TempMemoryEntry, error) {
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
	read, args, next := tempReadClause(reader, []any{tenantID, channelKey}, 3)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id::text, tenant_id, channel_key, owner_agent_id, content, metadata,
		       expires_at, created_at, updated_at
		FROM temp_memory_channels
		WHERE tenant_id = $1
		  AND channel_key = $2
		  AND (expires_at IS NULL OR expires_at > NOW())
		  AND `+read+`
		ORDER BY updated_at DESC
		LIMIT $`+fmt.Sprint(next), append(args, limit)...)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read temp memory: %w", err)
	}
	if out == nil {
		out = []TempMemoryEntry{}
	}
	return out, nil
}

// ClearTempMemory deletes the reader's own entries of a channel
// (tempClearClause); other users' and system rows stay.
func (s *Service) ClearTempMemory(ctx context.Context, tenantID, channelKey string, reader GovernedReader) (int64, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("memory service offline")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	if channelKey == "" {
		return 0, fmt.Errorf("channel_key is required")
	}
	own, args, _ := tempClearClause(reader, []any{tenantID, channelKey}, 3)
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM temp_memory_channels
		WHERE tenant_id = $1 AND channel_key = $2 AND `+own, args...)
	if err != nil {
		return 0, fmt.Errorf("clear temp memory: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}
