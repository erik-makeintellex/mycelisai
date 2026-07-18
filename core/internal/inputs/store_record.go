package inputs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

func (s *Store) RecordEvent(ctx context.Context, source Source, event IngestEvent) (BufferEvent, error) {
	if s == nil || s.db == nil {
		return BufferEvent{}, ErrUnavailable
	}
	event = normalizeIngestEvent(source, event)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BufferEvent{}, err
	}
	defer tx.Rollback()

	var recorded BufferEvent
	var sourceTimestamp sql.NullTime
	if event.SourceTimestamp != nil {
		sourceTimestamp = sql.NullTime{Time: *event.SourceTimestamp, Valid: true}
	}
	err = tx.QueryRowContext(ctx, `
INSERT INTO input_source_events (
    source_id, channel_key, payload, payload_hash, source_timestamp, run_id,
    team_id, agent_id, source_kind, source_channel, payload_kind, tenant_id
) VALUES ($1, $2, $3::jsonb, NULLIF($4,''), $5, NULLIF($6,''), NULLIF($7,''),
          NULLIF($8,''), $9, $10, $11, $12)
RETURNING event_id::text, source_id, channel_key, payload, COALESCE(payload_hash, ''),
          source_timestamp, received_at, COALESCE(run_id, ''), COALESCE(team_id, ''),
          COALESCE(agent_id, ''), source_kind, source_channel, payload_kind, tenant_id`,
		event.SourceID, event.ChannelKey, string(event.Payload), event.PayloadHash,
		sourceTimestamp, event.RunID, event.TeamID, event.AgentID, event.SourceKind,
		event.SourceChannel, event.PayloadKind, event.TenantID,
	).Scan(
		&recorded.EventID, &recorded.SourceID, &recorded.ChannelKey, &recorded.Payload,
		&recorded.PayloadHash, &sourceTimestamp, &recorded.ReceivedAt, &recorded.RunID,
		&recorded.TeamID, &recorded.AgentID, &recorded.SourceKind, &recorded.SourceChannel,
		&recorded.PayloadKind, &recorded.TenantID,
	)
	if err != nil {
		return BufferEvent{}, err
	}
	if sourceTimestamp.Valid {
		recorded.SourceTimestamp = &sourceTimestamp.Time
	}

	switch source.BufferMode {
	case BufferLatestState, BufferAppendLatest:
		if err := upsertLatest(ctx, tx, recorded); err != nil {
			return BufferEvent{}, err
		}
	case BufferWindowedRollup:
		if err := upsertWindow(ctx, tx, source, recorded); err != nil {
			return BufferEvent{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return BufferEvent{}, err
	}
	return recorded, nil
}

func normalizeIngestEvent(source Source, event IngestEvent) IngestEvent {
	event.SourceID = source.ID
	if event.ChannelKey == "" {
		event.ChannelKey = "default"
	}
	if len(event.Payload) == 0 {
		event.Payload = json.RawMessage(`{}`)
	}
	if event.PayloadHash == "" {
		event.PayloadHash = hashPayload(event.Payload)
	}
	if event.SourceKind == "" {
		event.SourceKind = defaultSourceKind(source.AdapterKind)
	}
	if event.SourceChannel == "" {
		event.SourceChannel = source.AllowedIngressSubject
	}
	if event.PayloadKind == "" {
		event.PayloadKind = defaultPayloadKind(source.AdapterKind)
	}
	if event.TenantID == "" {
		event.TenantID = source.TenantID
	}
	if event.TenantID == "" {
		event.TenantID = "default"
	}
	return event
}

func upsertLatest(ctx context.Context, tx *sql.Tx, event BufferEvent) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO input_source_latest (
    source_id, channel_key, event_id, payload, received_at, source_timestamp, tenant_id
) VALUES ($1, $2, $3::uuid, $4::jsonb, $5, $6, $7)
ON CONFLICT (source_id, channel_key) DO UPDATE
SET event_id = EXCLUDED.event_id,
    payload = EXCLUDED.payload,
    received_at = EXCLUDED.received_at,
    source_timestamp = EXCLUDED.source_timestamp,
    tenant_id = EXCLUDED.tenant_id`,
		event.SourceID, event.ChannelKey, event.EventID, string(event.Payload),
		event.ReceivedAt, nullableTime(event.SourceTimestamp), event.TenantID,
	)
	return err
}

func upsertWindow(ctx context.Context, tx *sql.Tx, source Source, event BufferEvent) error {
	windowStart := event.ReceivedAt.UTC().Truncate(time.Minute)
	windowKey := windowStart.Format("20060102T1504Z")
	summary := fmt.Sprintf("Buffered %s input events for %s", event.ChannelKey, source.Name)
	_, err := tx.ExecContext(ctx, `
INSERT INTO input_source_windows (
    source_id, channel_key, window_key, summary, payload, count, started_at, ended_at, tenant_id
) VALUES ($1, $2, $3, $4, $5::jsonb, 1, $6, $7, $8)
ON CONFLICT (source_id, channel_key, window_key) DO UPDATE
SET summary = EXCLUDED.summary,
    payload = EXCLUDED.payload,
    count = input_source_windows.count + 1,
    ended_at = EXCLUDED.ended_at,
    tenant_id = EXCLUDED.tenant_id`,
		event.SourceID, event.ChannelKey, windowKey, summary, string(event.Payload),
		windowStart, event.ReceivedAt, event.TenantID,
	)
	return err
}

func nullableTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *value, Valid: true}
}
