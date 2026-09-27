package deploymentcontext

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

const opportunisticBackfillLimit = 20

// Keyword-only and keyword+semantic copy shown to the operator.
const (
	statusKeywordOnly = "Saved. Soma can recall this by keywords; semantic search needs an embedding engine."
	statusEmbedded    = "Saved. Soma can recall this by keywords and by meaning."
)

type pendingChunk struct {
	id   string
	text string
}

// insertAtomic writes the artifact and every chunk row in one transaction.
// Chunk rows start with embedding NULL and embedding_status=pending.
func (s *Service) insertAtomic(ctx context.Context, agentID, title, contentType, content string, meta []byte, chunks []string, chunkMeta []map[string]any) (string, time.Time, []string, error) {
	tx, err := s.Artifacts.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var artifactID string
	var createdAt time.Time
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO artifacts (agent_id, artifact_type, title, content_type, content, metadata, status)
		VALUES ($1, 'document', $2, $3, $4, $5, 'approved')
		RETURNING id, created_at`, agentID, title, contentType, content, meta).Scan(&artifactID, &createdAt); err != nil {
		return "", time.Time{}, nil, err
	}
	ids := make([]string, 0, len(chunks))
	for i, chunk := range chunks {
		chunkMeta[i]["artifact_id"] = artifactID
		raw, _ := json.Marshal(chunkMeta[i])
		var id string
		if err := tx.QueryRowContext(ctx, `INSERT INTO context_vectors (content, embedding, metadata) VALUES ($1, NULL, $2) RETURNING id::text`, chunk, raw).Scan(&id); err != nil {
			return "", time.Time{}, nil, err
		}
		ids = append(ids, id)
	}
	if err := tx.Commit(); err != nil {
		return "", time.Time{}, nil, err
	}
	return artifactID, createdAt, ids, nil
}

// embedChunks embeds freshly committed chunks when an engine is available and
// reports how many were embedded. Failures leave rows pending.
func (s *Service) embedChunks(ctx context.Context, chunks []pendingChunk) int {
	if !s.Cognitive.EmbeddingAvailable(ctx) {
		if s.Cognitive.EmbeddingStatus().Code == cognitive.EmbeddingWidthMismatch {
			for _, chunk := range chunks {
				s.markFailedDimension(ctx, s.Artifacts.DB, chunk.id)
			}
		}
		return 0
	}
	embedded := 0
	for _, chunk := range chunks {
		ok, err := s.embedOne(ctx, s.Artifacts.DB, chunk)
		if err != nil {
			break
		}
		if ok {
			embedded++
		}
	}
	return embedded
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// embedOne embeds a chunk and stores the vector only when it fits the store.
func (s *Service) embedOne(ctx context.Context, db execer, chunk pendingChunk) (bool, error) {
	vec, err := s.Cognitive.Embed(ctx, chunk.text, "")
	if err != nil {
		return false, err
	}
	if len(vec) != cognitive.EmbeddingDimensions {
		s.markFailedDimension(ctx, db, chunk.id)
		return false, nil
	}
	_, err = db.ExecContext(ctx, `UPDATE context_vectors SET embedding = $2::vector,
		metadata = metadata || '{"embedding_status": "embedded"}'::jsonb
		WHERE id::text = $1 AND embedding IS NULL`, chunk.id, formatVector(vec))
	return err == nil, err
}

func (s *Service) markFailedDimension(ctx context.Context, db execer, id string) {
	if _, err := db.ExecContext(ctx, `UPDATE context_vectors SET metadata = metadata || '{"embedding_status": "failed_dimension"}'::jsonb
		WHERE id::text = $1 AND embedding IS NULL`, id); err != nil {
		log.Printf("deployment context: mark failed_dimension: %v", err)
	}
}

// refreshArtifacts recomputes vector_count and embedding_status from the
// committed chunk rows. The artifact rows are locked first so concurrent
// refreshes serialize and the last one reads every committed chunk.
func (s *Service) refreshArtifacts(ctx context.Context, artifactIDs []string) (map[string]string, map[string]int, error) {
	tx, err := s.Artifacts.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()
	// Artifact ids are database-issued UUIDs, so a plain text[] literal is safe.
	idList := "{" + strings.Join(artifactIDs, ",") + "}"
	if _, err := tx.ExecContext(ctx, `SELECT id FROM artifacts WHERE id::text = ANY($1::text[]) ORDER BY id FOR UPDATE`, idList); err != nil {
		return nil, nil, err
	}
	rows, err := tx.QueryContext(ctx, `
		UPDATE artifacts a SET metadata = a.metadata || jsonb_build_object('vector_count', c.embedded,
			'embedding_status', CASE WHEN c.pending > 0 THEN 'pending' WHEN c.failed > 0 THEN 'failed_dimension' ELSE 'embedded' END)
		FROM (SELECT metadata->>'artifact_id' AS aid,
			count(*) FILTER (WHERE embedding IS NOT NULL) AS embedded,
			count(*) FILTER (WHERE embedding IS NULL AND COALESCE(metadata->>'embedding_status', 'pending') = 'pending') AS pending,
			count(*) FILTER (WHERE metadata->>'embedding_status' = 'failed_dimension') AS failed
			FROM context_vectors WHERE metadata->>'artifact_id' = ANY($1::text[]) GROUP BY 1) c
		WHERE a.id::text = c.aid
		RETURNING a.id::text, a.metadata->>'embedding_status', c.embedded`, idList)
	if err != nil {
		return nil, nil, err
	}
	statuses, counts := map[string]string{}, map[string]int{}
	for rows.Next() {
		var id, status string
		var embedded int
		if err := rows.Scan(&id, &status, &embedded); err != nil {
			rows.Close()
			return nil, nil, err
		}
		statuses[id], counts[id] = status, embedded
	}
	rows.Close()
	return statuses, counts, tx.Commit()
}

func (s *Service) pendingCount(ctx context.Context) (int, error) {
	var n int
	err := s.Artifacts.DB.QueryRowContext(ctx, `SELECT count(*) FROM context_vectors
		WHERE metadata->>'knowledge_store' = 'governed_context_store' AND embedding IS NULL
		AND COALESCE(metadata->>'embedding_status', 'pending') = 'pending'`).Scan(&n)
	return n, err
}

// BackfillEmbeddings embeds up to limit pending governed chunks. Rows are
// claimed with FOR UPDATE SKIP LOCKED so concurrent runs never double-embed.
func (s *Service) BackfillEmbeddings(ctx context.Context, limit int) (BackfillResult, error) {
	if s == nil || s.Artifacts == nil || s.Artifacts.DB == nil {
		return BackfillResult{}, fmt.Errorf("deployment context store unavailable")
	}
	if limit <= 0 || limit > 100 {
		limit = opportunisticBackfillLimit
	}
	if !s.Cognitive.EmbeddingAvailable(ctx) {
		remaining, err := s.pendingCount(ctx)
		return BackfillResult{Remaining: remaining, Status: BackfillStatusUnavailable}, err
	}
	embedded, touched, engineErr, err := s.backfillBatch(ctx, limit)
	if err != nil {
		return BackfillResult{}, err
	}
	if len(touched) > 0 {
		if _, _, err := s.refreshArtifacts(ctx, touched); err != nil {
			return BackfillResult{}, err
		}
	}
	remaining, err := s.pendingCount(ctx)
	result := BackfillResult{Embedded: embedded, Remaining: remaining, Status: BackfillStatusComplete}
	switch {
	case engineErr != nil:
		result.Status = BackfillStatusUnavailable
	case remaining > 0:
		result.Status = BackfillStatusPartial
	}
	return result, err
}

func (s *Service) backfillBatch(ctx context.Context, limit int) (int, []string, error, error) {
	tx, err := s.Artifacts.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, `SELECT id::text, content, metadata FROM context_vectors
		WHERE metadata->>'knowledge_store' = 'governed_context_store' AND embedding IS NULL
		AND COALESCE(metadata->>'embedding_status', 'pending') = 'pending'
		ORDER BY created_at, id LIMIT $1 FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return 0, nil, nil, err
	}
	var claimed []pendingChunk
	artifactSet := map[string]bool{}
	for rows.Next() {
		var id, content string
		var raw []byte
		if err := rows.Scan(&id, &content, &raw); err != nil {
			rows.Close()
			return 0, nil, nil, err
		}
		var meta map[string]any
		_ = json.Unmarshal(raw, &meta)
		claimed = append(claimed, pendingChunk{id: id, text: embeddingTextFromMeta(meta, content)})
		if aid := stringMeta(meta, "artifact_id", ""); aid != "" {
			artifactSet[aid] = true
		}
	}
	rows.Close()
	embedded := 0
	var engineErr error
	for _, chunk := range claimed {
		ok, err := s.embedOne(ctx, tx, chunk)
		if err != nil {
			engineErr = err
			break
		}
		if ok {
			embedded++
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, nil, nil, err
	}
	touched := make([]string, 0, len(artifactSet))
	for id := range artifactSet {
		touched = append(touched, id)
	}
	return embedded, touched, engineErr, nil
}

func embeddingTextFromMeta(meta map[string]any, chunk string) string {
	return buildEmbeddingText(stringMeta(meta, "knowledge_class", KnowledgeClassCustomerContext), stringMeta(meta, "artifact_title", ""),
		stringMeta(meta, "source_label", ""), stringMeta(meta, "source_kind", ""), chunk,
		intMeta(meta, "chunk_index", 0)+1, intMeta(meta, "chunk_count", 1))
}

func formatVector(v []float64) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func retrievalModes(status string) ([]string, string) {
	if status == EmbeddingStatusEmbedded {
		return []string{"keyword", "semantic"}, statusEmbedded
	}
	return []string{"keyword"}, statusKeywordOnly
}
