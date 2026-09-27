package deploymentcontext

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"unicode/utf8"
)

func (s *Service) List(ctx context.Context, limit int) ([]Entry, error) {
	return s.ListEntries(ctx, limit, false)
}

// ListEntries lists governed entries, newest first. Archived entries are
// included only when includeArchived is set; deleted entries no longer exist.
func (s *Service) ListEntries(ctx context.Context, limit int, includeArchived bool) ([]Entry, error) {
	if s == nil || s.Artifacts == nil || s.Artifacts.DB == nil {
		return nil, fmt.Errorf("deployment context store unavailable")
	}
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.Artifacts.DB.QueryContext(ctx, `
		SELECT a.id::text,
		       a.title,
		       a.content,
		       a.metadata,
		       a.created_at,
		       c.chunks,
		       c.embedded
		FROM artifacts a
		LEFT JOIN LATERAL (
			SELECT count(*) AS chunks, count(*) FILTER (WHERE v.embedding IS NOT NULL) AS embedded
			FROM context_vectors v WHERE v.metadata->>'artifact_id' = a.id::text
		) c ON true
		WHERE COALESCE(a.metadata->>'knowledge_store', '') = 'governed_context_store'
		  AND ($2 OR COALESCE(a.metadata->>'lifecycle_state', 'active') = 'active')
		ORDER BY a.created_at DESC
		LIMIT $1
	`, limit, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("list deployment context entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var (
			id        string
			title     string
			content   sql.NullString
			metaJSON  []byte
			createdAt time.Time
			chunks    int
			embedded  int
		)
		if err := rows.Scan(&id, &title, &content, &metaJSON, &createdAt, &chunks, &embedded); err != nil {
			return nil, fmt.Errorf("scan deployment context entry: %w", err)
		}

		entry := Entry{
			ArtifactID:       id,
			KnowledgeClass:   KnowledgeClassCustomerContext,
			Title:            title,
			ContentPreview:   previewContent(content.String, 220),
			ContentLength:    utf8.RuneCountInString(content.String),
			CreatedAt:        createdAt,
			Visibility:       "global",
			SensitivityClass: "role_scoped",
			TrustClass:       "user_provided",
			SourceKind:       "user_document",
			SourceLabel:      "operator provided",
			LifecycleState:   LifecycleActive,
		}

		if len(metaJSON) > 0 {
			var meta map[string]any
			if err := json.Unmarshal(metaJSON, &meta); err == nil {
				entry.SourceLabel = stringMeta(meta, "source_label", entry.SourceLabel)
				entry.KnowledgeClass = stringMeta(meta, "knowledge_class", entry.KnowledgeClass)
				entry.SourceKind = stringMeta(meta, "source_kind", entry.SourceKind)
				entry.Visibility = stringMeta(meta, "visibility", entry.Visibility)
				entry.SensitivityClass = stringMeta(meta, "sensitivity_class", entry.SensitivityClass)
				entry.TrustClass = stringMeta(meta, "trust_class", entry.TrustClass)
				entry.ChunkCount = intMeta(meta, "chunk_count", 0)
				entry.EmbeddingStatus = stringMeta(meta, "embedding_status", "")
				entry.ContentLength = intMeta(meta, "content_length", entry.ContentLength)
				entry.ContentDomain = stringMeta(meta, "content_domain", entry.ContentDomain)
				entry.TargetGoalSets = stringSliceMeta(meta, "target_goal_sets")
				entry.LifecycleState = stringMeta(meta, "lifecycle_state", LifecycleActive)
				entry.ArchivedAt = stringMeta(meta, "archived_at", "")
				entry.LoadedBy = stringMeta(meta, "loaded_by", "")
				entry.OwnerUserID = stringMeta(meta, "owner_user_id", "")
				entry.TeamID = stringMeta(meta, "team_id", "")
			}
		}

		// Counts come from the stored chunk rows, never from saved claims.
		entry.VectorCount = embedded
		entry.EmbeddingStatus = entryEmbeddingStatus(entry.EmbeddingStatus, chunks, embedded)
		entries = append(entries, entry)
	}

	if entries == nil {
		entries = []Entry{}
	}
	return entries, rows.Err()
}

// entryEmbeddingStatus derives an honest status for rows saved before M1
// (no embedding_status) and for entries without any chunk rows.
func entryEmbeddingStatus(saved string, chunks, embedded int) string {
	switch {
	case chunks == 0:
		return "not_indexed"
	case saved != "":
		return saved
	case embedded == chunks:
		return EmbeddingStatusEmbedded
	default:
		return EmbeddingStatusPending
	}
}
