package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// VectorResult is a single semantic search hit from context_vectors.
type VectorResult struct {
	ID        string         `json:"id"`
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata"`
	Score     float64        `json:"score"` // cosine similarity (1.0 = identical)
	CreatedAt time.Time      `json:"created_at"`
}

// SemanticSearchOptions constrains pgvector recall so durable memory can be
// queried safely for a tenant, team, agent, or memory class.
type SemanticSearchOptions struct {
	Limit               int
	TenantID            string
	TeamID              string
	AgentID             string
	RunID               string
	Visibility          string
	Types               []string
	AllowGlobal         bool
	AllowLegacyUnscoped bool
}

// StoreVector persists an embedding into context_vectors for future RAG retrieval.
func (s *Service) StoreVector(ctx context.Context, content string, embedding []float64, metadata map[string]any) error {
	metaJSON, _ := json.Marshal(metadata)
	vecStr := formatVector(embedding)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO context_vectors (content, embedding, metadata)
		VALUES ($1, $2::vector, $3)
	`, content, vecStr, metaJSON)

	if err != nil {
		return fmt.Errorf("store vector failed: %w", err)
	}
	return nil
}

// SemanticSearch finds the top-K nearest vectors by cosine similarity.
func (s *Service) SemanticSearch(ctx context.Context, queryVec []float64, limit int) ([]VectorResult, error) {
	return s.SemanticSearchWithOptions(ctx, queryVec, SemanticSearchOptions{
		Limit:               limit,
		TenantID:            "default",
		AllowGlobal:         true,
		AllowLegacyUnscoped: true,
	})
}

// SemanticSearchWithOptions finds the top-K nearest vectors by cosine
// similarity while respecting optional scope and visibility boundaries.
func (s *Service) SemanticSearchWithOptions(ctx context.Context, queryVec []float64, opts SemanticSearchOptions) ([]VectorResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 5
	}

	tenantID := strings.TrimSpace(opts.TenantID)
	if tenantID == "" {
		tenantID = "default"
	}

	vecStr := formatVector(queryVec)
	clauses := []string{"embedding IS NOT NULL"}
	args := []any{vecStr}
	nextArg := 2

	clauses = append(clauses, fmt.Sprintf("COALESCE(metadata->>'tenant_id', 'default') = $%d", nextArg))
	args = append(args, tenantID)
	nextArg++

	teamID := strings.TrimSpace(opts.TeamID)
	agentID := strings.TrimSpace(opts.AgentID)
	runID := strings.TrimSpace(opts.RunID)
	visibility := strings.ToLower(strings.TrimSpace(opts.Visibility))

	if runID != "" {
		clauses = append(clauses, fmt.Sprintf("metadata->>'run_id' = $%d", nextArg))
		args = append(args, runID)
		nextArg++
	}

	if len(opts.Types) > 0 {
		typeParts := make([]string, 0, len(opts.Types))
		for _, raw := range opts.Types {
			value := strings.TrimSpace(raw)
			if value == "" {
				continue
			}
			typeParts = append(typeParts, fmt.Sprintf("metadata->>'type' = $%d", nextArg))
			args = append(args, value)
			nextArg++
		}
		if len(typeParts) > 0 {
			clauses = append(clauses, "("+strings.Join(typeParts, " OR ")+")")
		}
	}

	switch {
	case teamID != "" || agentID != "":
		scopeParts := make([]string, 0, 4)
		if opts.AllowGlobal {
			scopeParts = append(scopeParts, "COALESCE(metadata->>'visibility', '') = 'global'")
		}
		if teamID != "" {
			scopeParts = append(scopeParts, fmt.Sprintf("(metadata->>'team_id' = $%d AND COALESCE(NULLIF(metadata->>'visibility', ''), 'team') IN ('team', 'global'))", nextArg))
			args = append(args, teamID)
			nextArg++
		}
		if agentID != "" {
			scopeParts = append(scopeParts, fmt.Sprintf("(metadata->>'agent_id' = $%d AND COALESCE(NULLIF(metadata->>'visibility', ''), 'private') = 'private')", nextArg))
			args = append(args, agentID)
			nextArg++
		}
		if opts.AllowLegacyUnscoped {
			scopeParts = append(scopeParts, "(NOT (metadata ? 'visibility') AND NOT (metadata ? 'team_id') AND NOT (metadata ? 'agent_id'))")
		}
		if len(scopeParts) > 0 {
			clauses = append(clauses, "("+strings.Join(scopeParts, " OR ")+")")
		}
	case visibility != "":
		clauses = append(clauses, fmt.Sprintf("COALESCE(metadata->>'visibility', '') = $%d", nextArg))
		args = append(args, visibility)
		nextArg++
	}

	query := `
		SELECT id, content, metadata, 1 - (embedding <=> $1::vector) AS score, created_at
		FROM context_vectors
		WHERE ` + strings.Join(clauses, " AND ") + `
		ORDER BY embedding <=> $1::vector
		LIMIT $` + fmt.Sprintf("%d", nextArg)
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}
	defer rows.Close()

	var results []VectorResult
	for rows.Next() {
		var r VectorResult
		var metaJSON []byte
		if err := rows.Scan(&r.ID, &r.Content, &metaJSON, &r.Score, &r.CreatedAt); err != nil {
			return nil, err
		}
		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &r.Metadata)
		}
		results = append(results, r)
	}
	return results, nil
}

// ListSitReps retrieves recent SitReps for a team.
func (s *Service) ListSitReps(ctx context.Context, teamID string, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, team_id, timestamp, time_window_start, time_window_end, summary, key_events, strategies, status
		FROM sitreps
		WHERE team_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`, teamID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var id, tID, summary, status string
		var strategies sql.NullString
		var ts, winStart, winEnd time.Time
		var keyEventsJSON []byte

		if err := rows.Scan(&id, &tID, &ts, &winStart, &winEnd, &summary, &keyEventsJSON, &strategies, &status); err != nil {
			return nil, err
		}

		entry := map[string]any{
			"id":                id,
			"team_id":           tID,
			"timestamp":         ts,
			"time_window_start": winStart,
			"time_window_end":   winEnd,
			"summary":           summary,
			"strategies":        strategies.String,
			"status":            status,
		}

		var keyEvents []string
		if json.Unmarshal(keyEventsJSON, &keyEvents) == nil {
			entry["key_events"] = keyEvents
		}

		results = append(results, entry)
	}
	return results, nil
}
