package deploymentcontext

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

// Edit in place (MEM). An operator changes a saved entry's title, content or
// source label. knowledge_class and visibility never change here: moving an
// entry between classes is promotion, a separate governed path.
//
// One transaction locks the artifact row (the same lock order as Delete and
// archive/restore), replaces the content and every chunk row, drops the
// stale chunks and their vectors, and marks the entry pending. Keyword recall
// sees the new text as soon as it commits; embeddings follow through the
// post-save embed or the backfill, exactly as for a fresh save.
var (
	ErrEntryArchived = errors.New("deployment context entry is archived; restore it first")
	ErrEntryChanged  = errors.New("deployment context entry changed since it was authorized")
	ErrEditInvalid   = errors.New("invalid deployment context edit")
)

// EditRequest is a partial update; a nil field is left unchanged.
// Authorized is the record the caller's authority decision was made on. The
// edit refuses if the entry's class, visibility, team or owner moved since.
type EditRequest struct {
	ArtifactID  string
	Title       *string
	Content     *string
	SourceLabel *string
	Actor       string
	Authorized  *EntryRecord
}

// EditResult reports what the edit did. Changed is false only when the patch
// matched the saved values and no row was written.
type EditResult struct {
	ArtifactID      string   `json:"artifact_id"`
	Changed         bool     `json:"changed"`
	Title           string   `json:"title"`
	SourceLabel     string   `json:"source_label"`
	ContentChanged  bool     `json:"content_changed"`
	ContentLength   int      `json:"content_length"`
	ChunkCount      int      `json:"chunk_count"`
	ChunksRemoved   int      `json:"chunks_removed"`
	VectorCount     int      `json:"vector_count"`
	EmbeddingStatus string   `json:"embedding_status"`
	RetrievalModes  []string `json:"retrieval_modes"`
	StatusMessage   string   `json:"status_message"`
}

// Validate rejects an empty patch and blank values before any audit or lookup.
func (r EditRequest) Validate() error {
	if r.Title == nil && r.Content == nil && r.SourceLabel == nil {
		return fmt.Errorf("%w: send at least one of title, content or source_label", ErrEditInvalid)
	}
	fields := []struct {
		name  string
		value *string
	}{{"title", r.Title}, {"content", r.Content}, {"source_label", r.SourceLabel}}
	for _, field := range fields {
		if field.value != nil && strings.TrimSpace(*field.value) == "" {
			return fmt.Errorf("%w: %s cannot be empty", ErrEditInvalid, field.name)
		}
	}
	return nil
}

func editValue(patch *string, current string) string {
	if patch == nil {
		return current
	}
	return strings.TrimSpace(*patch)
}

func sameAuthority(a, b *EntryRecord) bool {
	return a.KnowledgeClass == b.KnowledgeClass && a.Visibility == b.Visibility &&
		a.TeamID == b.TeamID && a.OwnerUserID == b.OwnerUserID && a.LoadedBy == b.LoadedBy
}

// Edit applies a validated patch. See the file comment for the invariants.
func (s *Service) Edit(ctx context.Context, req EditRequest) (*EditResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if req.Authorized == nil {
		return nil, fmt.Errorf("%w: no authority decision", ErrEditInvalid)
	}
	db, err := s.store()
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var title, content string
	var raw []byte
	err = tx.QueryRowContext(ctx, `SELECT title, COALESCE(content, ''), metadata FROM artifacts WHERE `+governedArtifact+` FOR UPDATE`,
		req.ArtifactID).Scan(&title, &content, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEntryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock deployment context entry: %w", err)
	}
	var meta map[string]any
	_ = json.Unmarshal(raw, &meta)
	current := EntryRecord{ArtifactID: req.ArtifactID}
	fillRecordMeta(&current, meta)
	if current.LifecycleState == LifecycleArchived {
		return nil, ErrEntryArchived
	}
	if !sameAuthority(&current, req.Authorized) {
		return nil, ErrEntryChanged
	}
	oldLabel := stringMeta(meta, "source_label", "")
	result := &EditResult{ArtifactID: req.ArtifactID, Title: editValue(req.Title, title), SourceLabel: editValue(req.SourceLabel, oldLabel)}
	newContent := editValue(req.Content, content)
	result.ContentChanged = newContent != content
	result.ContentLength = utf8.RuneCountInString(newContent)
	if result.Title == title && !result.ContentChanged && result.SourceLabel == oldLabel {
		result.ChunkCount = intMeta(meta, "chunk_count", 0)
		result.VectorCount = intMeta(meta, "vector_count", 0)
		result.EmbeddingStatus = stringMeta(meta, "embedding_status", EmbeddingStatusPending)
		result.RetrievalModes, result.StatusMessage = retrievalModes(result.EmbeddingStatus)
		return result, nil
	}
	pending, removed, err := s.replaceChunks(ctx, tx, req.ArtifactID, result.Title, result.SourceLabel, newContent)
	if err != nil {
		return nil, err
	}
	res, err := tx.ExecContext(ctx, `UPDATE artifacts SET title = $2, content = $3,
		metadata = metadata || jsonb_build_object('source_label', $4::text, 'chunk_count', $5::int, 'vector_count', 0,
			'embedding_status', 'pending', 'content_length', $6::int, 'edited_at', now(), 'edited_by', $7::text)
		WHERE `+governedArtifact, req.ArtifactID, result.Title, newContent, result.SourceLabel, len(pending), result.ContentLength, req.Actor)
	if err != nil {
		return nil, fmt.Errorf("update deployment context entry: %w", err)
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return nil, ErrEntryNotFound
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit deployment context edit: %w", err)
	}
	result.Changed, result.ChunkCount, result.ChunksRemoved = true, len(pending), removed
	s.embedChunks(ctx, pending)
	result.EmbeddingStatus = EmbeddingStatusPending
	if statuses, counts, err := s.refreshArtifacts(ctx, []string{req.ArtifactID}); err != nil {
		log.Printf("deployment context: refresh embedding status after edit of %s: %v", req.ArtifactID, err)
	} else if status, ok := statuses[req.ArtifactID]; ok {
		result.EmbeddingStatus, result.VectorCount = status, counts[req.ArtifactID]
	}
	result.RetrievalModes, result.StatusMessage = retrievalModes(result.EmbeddingStatus)
	return result, nil
}

// replaceChunks deletes every chunk row (content and vector) of the entry and
// inserts the new chunks with embedding NULL and embedding_status pending.
// New rows copy the scope metadata of the old first chunk so recall scope,
// class, sensitivity and lifecycle stay exactly as saved.
func (s *Service) replaceChunks(ctx context.Context, tx *sql.Tx, artifactID, title, sourceLabel, content string) ([]pendingChunk, int, error) {
	var raw []byte
	err := tx.QueryRowContext(ctx, `SELECT metadata FROM context_vectors WHERE `+governedChunks+`
		ORDER BY COALESCE((metadata->>'chunk_index')::int, 0), created_at LIMIT 1`, artifactID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, fmt.Errorf("deployment context entry %s has no chunk rows to re-chunk from", artifactID)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("read deployment context chunk scope: %w", err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM context_vectors WHERE `+governedChunks, artifactID)
	if err != nil {
		return nil, 0, fmt.Errorf("remove stale deployment context chunks: %w", err)
	}
	removed, _ := res.RowsAffected()
	chunks := chunkText(content, defaultChunkSize, defaultChunkOverlap)
	if len(chunks) == 0 {
		return nil, 0, fmt.Errorf("%w: content cannot be empty", ErrEditInvalid)
	}
	pending := make([]pendingChunk, len(chunks))
	for idx, chunk := range chunks {
		var meta map[string]any
		if err := json.Unmarshal(raw, &meta); err != nil || meta == nil {
			return nil, 0, fmt.Errorf("read deployment context chunk scope: %v", err)
		}
		meta["artifact_id"], meta["artifact_title"], meta["source_label"] = artifactID, title, sourceLabel
		meta["chunk_index"], meta["chunk_count"], meta["embedding_status"] = idx, len(chunks), EmbeddingStatusPending
		encoded, _ := json.Marshal(meta)
		if err := tx.QueryRowContext(ctx, `INSERT INTO context_vectors (content, embedding, metadata) VALUES ($1, NULL, $2) RETURNING id::text`,
			chunk, encoded).Scan(&pending[idx].id); err != nil {
			return nil, 0, fmt.Errorf("insert deployment context chunk: %w", err)
		}
		pending[idx].text = embeddingTextFromMeta(meta, chunk)
	}
	return pending, int(removed), nil
}
