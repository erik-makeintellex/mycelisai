package deploymentcontext

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// Entry lifecycle (M2). Archive is reversible and hides an entry from recall
// and the default list; delete permanently removes the content and chunks.
// Chunk rows carry lifecycle_state so every recall leg filters them in SQL.
const (
	LifecycleActive   = "active"
	LifecycleArchived = "archived"
)

// ErrEntryNotFound means no governed entry has this id (never saved, or deleted).
var ErrEntryNotFound = errors.New("deployment context entry not found")

// EntryRecord is what authority and audit need about an entry. It carries no
// content.
type EntryRecord struct {
	ArtifactID     string
	Title          string
	KnowledgeClass string
	Visibility     string
	TeamID         string
	LoadedBy       string
	OwnerUserID    string
	LifecycleState string
	ChunkCount     int
	ContentLength  int
}

// DeleteResult reports a permanent delete.
type DeleteResult struct {
	ChunksRemoved int `json:"chunks_removed"`
}

const governedArtifact = `id::text = $1 AND metadata->>'knowledge_store' = 'governed_context_store'`
const governedChunks = `metadata->>'artifact_id' = $1 AND metadata->>'knowledge_store' = 'governed_context_store'`

func (s *Service) store() (*sql.DB, error) {
	if s == nil || s.Artifacts == nil || s.Artifacts.DB == nil {
		return nil, fmt.Errorf("deployment context store unavailable")
	}
	return s.Artifacts.DB, nil
}

// Lookup reads one governed entry for an authority decision.
func (s *Service) Lookup(ctx context.Context, artifactID string) (*EntryRecord, error) {
	db, err := s.store()
	if err != nil {
		return nil, err
	}
	var raw []byte
	record := EntryRecord{}
	err = db.QueryRowContext(ctx, `SELECT id::text, title, metadata,
		(SELECT count(*) FROM context_vectors WHERE `+governedChunks+`)
		FROM artifacts WHERE `+governedArtifact, artifactID).Scan(&record.ArtifactID, &record.Title, &raw, &record.ChunkCount)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEntryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("look up deployment context entry: %w", err)
	}
	var meta map[string]any
	_ = json.Unmarshal(raw, &meta)
	fillRecordMeta(&record, meta)
	return &record, nil
}

// fillRecordMeta derives the authority and lifecycle fields from metadata.
func fillRecordMeta(record *EntryRecord, meta map[string]any) {
	record.KnowledgeClass = stringMeta(meta, "knowledge_class", KnowledgeClassCustomerContext)
	record.Visibility = stringMeta(meta, "visibility", "global")
	record.TeamID = stringMeta(meta, "team_id", "")
	record.LoadedBy = stringMeta(meta, "loaded_by", "")
	record.OwnerUserID = stringMeta(meta, "owner_user_id", "")
	record.LifecycleState = stringMeta(meta, "lifecycle_state", LifecycleActive)
	record.ContentLength = intMeta(meta, "content_length", 0)
}

// Archive hides an entry from recall and the default list. Reversible.
func (s *Service) Archive(ctx context.Context, artifactID, actor string) error {
	return s.setLifecycle(ctx, artifactID, LifecycleArchived, actor)
}

// Restore returns an archived entry to recall and the default list.
func (s *Service) Restore(ctx context.Context, artifactID, actor string) error {
	return s.setLifecycle(ctx, artifactID, LifecycleActive, actor)
}

func (s *Service) setLifecycle(ctx context.Context, artifactID, state, actor string) error {
	db, err := s.store()
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	status := "approved"
	if state == LifecycleArchived {
		status = LifecycleArchived
	}
	res, err := tx.ExecContext(ctx, `UPDATE artifacts SET status = $2,
		metadata = (metadata - 'archived_at' - 'archived_by') || jsonb_build_object('lifecycle_state', $3::text)
			|| CASE WHEN $3::text = 'archived' THEN jsonb_build_object('archived_at', now(), 'archived_by', $4::text) ELSE '{}'::jsonb END
		WHERE `+governedArtifact, artifactID, status, state, actor)
	if err != nil {
		return fmt.Errorf("update deployment context entry: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrEntryNotFound
	}
	if _, err := tx.ExecContext(ctx, `UPDATE context_vectors SET metadata = metadata || jsonb_build_object('lifecycle_state', $2::text)
		WHERE `+governedChunks, artifactID, state); err != nil {
		return fmt.Errorf("update deployment context chunks: %w", err)
	}
	return tx.Commit()
}

// Delete permanently removes the entry and every chunk (content and
// vectors) in one transaction. The audit record is the only tombstone.
func (s *Service) Delete(ctx context.Context, artifactID string) (DeleteResult, error) {
	db, err := s.store()
	if err != nil {
		return DeleteResult{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return DeleteResult{}, err
	}
	defer func() { _ = tx.Rollback() }()
	// Lock the artifact first, the same order as Edit and setLifecycle, so a
	// concurrent edit and delete serialize instead of deadlocking.
	if err := tx.QueryRowContext(ctx, `SELECT id::text FROM artifacts WHERE `+governedArtifact+` FOR UPDATE`, artifactID).Scan(new(string)); errors.Is(err, sql.ErrNoRows) {
		return DeleteResult{}, ErrEntryNotFound
	} else if err != nil {
		return DeleteResult{}, fmt.Errorf("lock deployment context entry: %w", err)
	}
	chunks, err := tx.ExecContext(ctx, `DELETE FROM context_vectors WHERE `+governedChunks, artifactID)
	if err != nil {
		return DeleteResult{}, fmt.Errorf("delete deployment context chunks: %w", err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM artifacts WHERE `+governedArtifact, artifactID)
	if err != nil {
		return DeleteResult{}, fmt.Errorf("delete deployment context entry: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return DeleteResult{}, ErrEntryNotFound
	}
	removed, _ := chunks.RowsAffected()
	return DeleteResult{ChunksRemoved: int(removed)}, tx.Commit()
}
