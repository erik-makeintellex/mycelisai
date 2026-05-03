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
	if s == nil || s.Artifacts == nil || s.Artifacts.DB == nil {
		return nil, fmt.Errorf("deployment context store unavailable")
	}
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.Artifacts.DB.QueryContext(ctx, `
		SELECT id::text,
		       title,
		       content,
		       metadata,
		       created_at
		FROM artifacts
		WHERE COALESCE(metadata->>'knowledge_store', '') = 'governed_context_store'
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
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
		)
		if err := rows.Scan(&id, &title, &content, &metaJSON, &createdAt); err != nil {
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
				entry.VectorCount = intMeta(meta, "vector_count", entry.ChunkCount)
				entry.ContentLength = intMeta(meta, "content_length", entry.ContentLength)
				entry.ContentDomain = stringMeta(meta, "content_domain", entry.ContentDomain)
				entry.TargetGoalSets = stringSliceMeta(meta, "target_goal_sets")
			}
		}

		entries = append(entries, entry)
	}

	if entries == nil {
		entries = []Entry{}
	}
	return entries, rows.Err()
}
