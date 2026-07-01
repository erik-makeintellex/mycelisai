package searchcap

import (
	"context"
	"database/sql"
	"fmt"
)

// SourceStore persists operator-configured search source metadata.
// Secret values are never stored here; only secret references are retained.
type SourceStore struct {
	DB *sql.DB
}

func NewSourceStore(db *sql.DB) *SourceStore {
	return &SourceStore{DB: db}
}

func (s *SourceStore) List(ctx context.Context) ([]Source, error) {
	if s == nil || s.DB == nil {
		return []Source{}, nil
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, name, provider, source_type, COALESCE(endpoint, ''), COALESCE(scope_kind, 'all'),
		       COALESCE(scope_ref, ''), boundary, COALESCE(auth_scheme, 'none'), COALESCE(secret_ref, ''),
		       COALESCE(mode, 'preview'), COALESCE(sensitivity_class, 'public'),
		       COALESCE(trust_class, 'bounded_external'), COALESCE(status, 'available'), COALESCE(recovery, '')
		FROM search_sources
		WHERE tenant_id = 'default'
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list search sources: %w", err)
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var source Source
		if err := rows.Scan(
			&source.ID, &source.Name, &source.Provider, &source.SourceType, &source.Endpoint,
			&source.ScopeKind, &source.ScopeRef, &source.Boundary, &source.AuthScheme,
			&source.SecretRef, &source.Mode, &source.SensitivityClass, &source.TrustClass,
			&source.Status, &source.Recovery,
		); err != nil {
			return nil, fmt.Errorf("scan search source: %w", err)
		}
		source.BaseURL = source.Endpoint
		source.Managed = true
		sources = append(sources, source)
	}
	return sources, rows.Err()
}

func (s *SourceStore) Create(ctx context.Context, source Source) (Source, error) {
	if s == nil || s.DB == nil {
		return source, nil
	}
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO search_sources (
			id, name, provider, source_type, endpoint, scope_kind, scope_ref, boundary,
			auth_scheme, secret_ref, mode, sensitivity_class, trust_class, status, recovery, tenant_id
		) VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, NULLIF($7, ''), $8, $9, NULLIF($10, ''),
		          $11, $12, $13, $14, NULLIF($15, ''), 'default')`,
		source.ID, source.Name, source.Provider, source.SourceType, source.Endpoint,
		source.ScopeKind, source.ScopeRef, source.Boundary, source.AuthScheme,
		source.SecretRef, source.Mode, source.SensitivityClass, source.TrustClass,
		source.Status, source.Recovery,
	)
	if err != nil {
		return Source{}, fmt.Errorf("create search source: %w", err)
	}
	return source, nil
}

func (s *SourceStore) Update(ctx context.Context, source Source) error {
	if s == nil || s.DB == nil {
		return nil
	}
	res, err := s.DB.ExecContext(ctx, `
		UPDATE search_sources
		SET name = $2, provider = $3, source_type = $4, endpoint = NULLIF($5, ''),
		    scope_kind = $6, scope_ref = NULLIF($7, ''), boundary = $8,
		    auth_scheme = $9, secret_ref = NULLIF($10, ''), mode = $11,
		    sensitivity_class = $12, trust_class = $13, status = $14,
		    recovery = NULLIF($15, ''), updated_at = CURRENT_TIMESTAMP
		WHERE tenant_id = 'default' AND id = $1`,
		source.ID, source.Name, source.Provider, source.SourceType, source.Endpoint,
		source.ScopeKind, source.ScopeRef, source.Boundary, source.AuthScheme,
		source.SecretRef, source.Mode, source.SensitivityClass, source.TrustClass,
		source.Status, source.Recovery,
	)
	if err != nil {
		return fmt.Errorf("update search source: %w", err)
	}
	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errSourceNotFound
	}
	return nil
}

func (s *SourceStore) Delete(ctx context.Context, id string) error {
	if s == nil || s.DB == nil {
		return nil
	}
	res, err := s.DB.ExecContext(ctx, `
		DELETE FROM search_sources
		WHERE tenant_id = 'default' AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete search source: %w", err)
	}
	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errSourceNotFound
	}
	return nil
}
