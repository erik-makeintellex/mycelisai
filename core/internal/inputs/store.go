package inputs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) List(ctx context.Context) ([]Source, error) {
	if s == nil || s.db == nil {
		return nil, ErrUnavailable
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, name, source_type, adapter_kind, scope_kind, scope_ref, target_outcome_id,
       target_group_id, target_host_id, auth_scheme, secret_ref, allowed_ingress_subject,
       payload_schema_ref, buffer_mode, buffer_policy, sensitivity_class, trust_class,
       status, recovery, tenant_id, created_at, updated_at
FROM input_sources
ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Source
	for rows.Next() {
		source, err := scanSource(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, source)
	}
	return result, rows.Err()
}

func (s *Store) Get(ctx context.Context, id string) (Source, error) {
	if s == nil || s.db == nil {
		return Source{}, ErrUnavailable
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, name, source_type, adapter_kind, scope_kind, scope_ref, target_outcome_id,
       target_group_id, target_host_id, auth_scheme, secret_ref, allowed_ingress_subject,
       payload_schema_ref, buffer_mode, buffer_policy, sensitivity_class, trust_class,
       status, recovery, tenant_id, created_at, updated_at
FROM input_sources
WHERE id = $1`, id)
	source, err := scanSource(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Source{}, ErrNotFound
	}
	return source, err
}

func (s *Store) Create(ctx context.Context, source Source) (Source, error) {
	if s == nil || s.db == nil {
		return Source{}, ErrUnavailable
	}
	if len(source.BufferPolicy) == 0 {
		source.BufferPolicy = json.RawMessage(`{}`)
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO input_sources (
    id, name, source_type, adapter_kind, scope_kind, scope_ref, target_outcome_id,
    target_group_id, target_host_id, auth_scheme, secret_ref, allowed_ingress_subject,
    payload_schema_ref, buffer_mode, buffer_policy, sensitivity_class, trust_class,
    status, recovery, tenant_id
) VALUES ($1, $2, $3, $4, $5, NULLIF($6,''), NULLIF($7,''), NULLIF($8,''),
          NULLIF($9,''), $10, NULLIF($11,''), $12, NULLIF($13,''), $14, $15::jsonb,
          $16, $17, $18, NULLIF($19,''), $20)
ON CONFLICT (id) DO NOTHING`,
		source.ID, source.Name, source.SourceType, source.AdapterKind, source.ScopeKind,
		source.ScopeRef, source.TargetOutcomeID, source.TargetGroupID, source.TargetHostID,
		source.AuthScheme, source.SecretRef, source.AllowedIngressSubject, source.PayloadSchemaRef,
		source.BufferMode, string(source.BufferPolicy), source.SensitivityClass, source.TrustClass,
		source.Status, source.Recovery, source.TenantID,
	)
	if err != nil {
		return Source{}, err
	}
	return s.Get(ctx, source.ID)
}

func (s *Store) Update(ctx context.Context, source Source) (Source, error) {
	if s == nil || s.db == nil {
		return Source{}, ErrUnavailable
	}
	if len(source.BufferPolicy) == 0 {
		source.BufferPolicy = json.RawMessage(`{}`)
	}
	result, err := s.db.ExecContext(ctx, `
UPDATE input_sources
SET name = $2, source_type = $3, adapter_kind = $4, scope_kind = $5, scope_ref = NULLIF($6,''),
    target_outcome_id = NULLIF($7,''), target_group_id = NULLIF($8,''), target_host_id = NULLIF($9,''),
    auth_scheme = $10, secret_ref = NULLIF($11,''), allowed_ingress_subject = $12,
    payload_schema_ref = NULLIF($13,''), buffer_mode = $14, buffer_policy = $15::jsonb,
    sensitivity_class = $16, trust_class = $17, status = $18, recovery = NULLIF($19,''),
    tenant_id = $20, updated_at = NOW()
WHERE id = $1`,
		source.ID, source.Name, source.SourceType, source.AdapterKind, source.ScopeKind,
		source.ScopeRef, source.TargetOutcomeID, source.TargetGroupID, source.TargetHostID,
		source.AuthScheme, source.SecretRef, source.AllowedIngressSubject, source.PayloadSchemaRef,
		source.BufferMode, string(source.BufferPolicy), source.SensitivityClass, source.TrustClass,
		source.Status, source.Recovery, source.TenantID,
	)
	if err != nil {
		return Source{}, err
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return Source{}, ErrNotFound
	}
	return s.Get(ctx, source.ID)
}

func (s *Store) Delete(ctx context.Context, id string) error {
	if s == nil || s.db == nil {
		return ErrUnavailable
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM input_sources WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) Buffer(ctx context.Context, sourceID, mode, channelKey string, limit int) (BufferView, error) {
	source, err := s.Get(ctx, sourceID)
	if err != nil {
		return BufferView{}, err
	}
	if mode == "" {
		mode = source.BufferMode
	}
	view := BufferView{Mode: mode, Source: source}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if channelKey == "" {
		channelKey = "default"
	}
	switch mode {
	case BufferLatestState, BufferAppendLatest:
		view.Latest, err = s.ListLatest(ctx, sourceID, channelKey)
	case BufferWindowedRollup:
		view.Windows, err = s.ListWindows(ctx, sourceID, channelKey, limit)
	default:
		view.Events, err = s.ListEvents(ctx, sourceID, channelKey, limit)
	}
	return view, err
}

func (s *Store) ListEvents(ctx context.Context, sourceID, channelKey string, limit int) ([]BufferEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT event_id::text, source_id, channel_key, payload, payload_hash, source_timestamp, received_at,
       run_id, team_id, agent_id, source_kind, source_channel, payload_kind, tenant_id
FROM input_source_events
WHERE source_id = $1 AND ($2 = '' OR channel_key = $2)
ORDER BY received_at DESC
LIMIT $3`, sourceID, channelKey, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []BufferEvent
	for rows.Next() {
		var event BufferEvent
		var sourceTimestamp sql.NullTime
		var payloadHash, runID, teamID, agentID sql.NullString
		if err := rows.Scan(&event.EventID, &event.SourceID, &event.ChannelKey, &event.Payload,
			&payloadHash, &sourceTimestamp, &event.ReceivedAt, &runID, &teamID, &agentID,
			&event.SourceKind, &event.SourceChannel, &event.PayloadKind, &event.TenantID); err != nil {
			return nil, err
		}
		event.PayloadHash = payloadHash.String
		event.RunID = runID.String
		event.TeamID = teamID.String
		event.AgentID = agentID.String
		if sourceTimestamp.Valid {
			event.SourceTimestamp = &sourceTimestamp.Time
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) ListLatest(ctx context.Context, sourceID, channelKey string) ([]LatestValue, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT source_id, channel_key, COALESCE(event_id::text, ''), payload, received_at, source_timestamp, tenant_id
FROM input_source_latest
WHERE source_id = $1 AND ($2 = '' OR channel_key = $2)
ORDER BY received_at DESC`, sourceID, channelKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var latest []LatestValue
	for rows.Next() {
		var item LatestValue
		var sourceTimestamp sql.NullTime
		if err := rows.Scan(&item.SourceID, &item.ChannelKey, &item.EventID, &item.Payload,
			&item.ReceivedAt, &sourceTimestamp, &item.TenantID); err != nil {
			return nil, err
		}
		if sourceTimestamp.Valid {
			item.SourceTimestamp = &sourceTimestamp.Time
		}
		latest = append(latest, item)
	}
	return latest, rows.Err()
}

func (s *Store) ListWindows(ctx context.Context, sourceID, channelKey string, limit int) ([]WindowSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT source_id, channel_key, window_key, summary, payload, count, started_at, ended_at, tenant_id
FROM input_source_windows
WHERE source_id = $1 AND ($2 = '' OR channel_key = $2)
ORDER BY ended_at DESC
LIMIT $3`, sourceID, channelKey, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var windows []WindowSummary
	for rows.Next() {
		var item WindowSummary
		if err := rows.Scan(&item.SourceID, &item.ChannelKey, &item.WindowKey, &item.Summary,
			&item.Payload, &item.Count, &item.StartedAt, &item.EndedAt, &item.TenantID); err != nil {
			return nil, err
		}
		windows = append(windows, item)
	}
	return windows, rows.Err()
}

type sourceScanner interface {
	Scan(dest ...any) error
}

func scanSource(row sourceScanner) (Source, error) {
	var source Source
	var scopeRef, targetOutcomeID, targetGroupID, targetHostID, secretRef, schemaRef, recovery sql.NullString
	err := row.Scan(
		&source.ID, &source.Name, &source.SourceType, &source.AdapterKind,
		&source.ScopeKind, &scopeRef, &targetOutcomeID, &targetGroupID, &targetHostID,
		&source.AuthScheme, &secretRef, &source.AllowedIngressSubject, &schemaRef,
		&source.BufferMode, &source.BufferPolicy, &source.SensitivityClass, &source.TrustClass,
		&source.Status, &recovery, &source.TenantID, &source.CreatedAt, &source.UpdatedAt,
	)
	source.ScopeRef = scopeRef.String
	source.TargetOutcomeID = targetOutcomeID.String
	source.TargetGroupID = targetGroupID.String
	source.TargetHostID = targetHostID.String
	source.SecretRef = secretRef.String
	source.PayloadSchemaRef = schemaRef.String
	source.Recovery = recovery.String
	return source, err
}
