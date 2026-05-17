package capabilities

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) List(ctx context.Context) (Snapshot, error) {
	if s == nil || s.db == nil {
		return Snapshot{}, fmt.Errorf("capability manifests: database not available")
	}
	rows, err := s.db.QueryContext(ctx, selectCapabilityManifestsSQL()+`
		ORDER BY kind, id`)
	if err != nil {
		return Snapshot{}, fmt.Errorf("capability manifests: list: %w", err)
	}
	defer rows.Close()
	return scanManifestRows(rows)
}

func (s *Store) ReplaceSnapshot(ctx context.Context, snap Snapshot) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("capability manifests: database not available")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("capability manifests: begin refresh: %w", err)
	}
	if err := replaceSnapshotTx(ctx, tx, snap); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("capability manifests: commit refresh: %w", err)
	}
	return nil
}

func replaceSnapshotTx(ctx context.Context, tx *sql.Tx, snap Snapshot) error {
	ids := make([]string, 0, len(snap.Manifests))
	for _, manifest := range snap.Manifests {
		m := cloneManifest(manifest)
		completeManifestState(&m)
		if strings.TrimSpace(m.ID) == "" {
			continue
		}
		if err := upsertManifest(ctx, tx, m); err != nil {
			return err
		}
		ids = append(ids, m.ID)
	}
	if err := deleteStaleManifests(ctx, tx, ids); err != nil {
		return err
	}
	return nil
}

func scanManifestRows(rows *sql.Rows) (Snapshot, error) {
	manifests := []Manifest{}
	var generatedAt time.Time
	for rows.Next() {
		manifest, err := scanManifest(rows)
		if err != nil {
			return Snapshot{}, err
		}
		if manifest.UpdatedAt.After(generatedAt) {
			generatedAt = manifest.UpdatedAt
		}
		manifests = append(manifests, manifest)
	}
	if err := rows.Err(); err != nil {
		return Snapshot{}, fmt.Errorf("capability manifests: scan rows: %w", err)
	}
	return Snapshot{GeneratedAt: generatedAt, Count: len(manifests), Manifests: manifests}, nil
}

func scanManifest(rows *sql.Rows) (Manifest, error) {
	var m Manifest
	var toolRefsJSON, defaultRolesJSON, allowedRolesJSON, metadataJSON []byte
	var lastProbeAt sql.NullTime
	err := rows.Scan(
		&m.ID, &m.CapabilityID, &m.Version, &m.ManifestVersion, &m.DisplayName,
		&m.Kind, &m.Source, &m.Status, &m.Health, &m.RiskClass, &m.Description,
		&m.Purpose, &toolRefsJSON, &defaultRolesJSON, &allowedRolesJSON,
		&m.AuditRequired, &m.ApprovalRequired, &m.ApprovalPosture,
		&m.InputSchemaRef, &m.OutputSchemaRef, &m.LastProbeStatus,
		&lastProbeAt, &m.FailurePosture, &m.RecoveryPosture, &m.AuditPolicy,
		&m.SecretRefPolicy, &m.Owner, &metadataJSON, &m.DerivedAt, &m.UpdatedAt,
	)
	if err != nil {
		return Manifest{}, fmt.Errorf("capability manifests: scan manifest: %w", err)
	}
	_ = json.Unmarshal(toolRefsJSON, &m.ToolRefs)
	_ = json.Unmarshal(defaultRolesJSON, &m.DefaultAllowedRoles)
	_ = json.Unmarshal(allowedRolesJSON, &m.AllowedRoles)
	_ = json.Unmarshal(metadataJSON, &m.Metadata)
	if lastProbeAt.Valid {
		m.LastProbeAt = lastProbeAt.Time
	}
	completeManifestState(&m)
	return m, nil
}

func marshalJSON(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}
