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
