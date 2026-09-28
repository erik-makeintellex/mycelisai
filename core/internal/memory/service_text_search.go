package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// keywordDocument is the full-text document for a context_vectors row: the
// source title plus the stored chunk text.
const keywordDocument = `to_tsvector('english', coalesce(metadata->>'artifact_title', '') || ' ' || content)`

// keywordQuery ORs the stemmed, stop-word-free lexemes of the query. Only
// plain alphanumeric lexemes are kept so no user text is parsed as tsquery
// syntax; the parts of hyphenated words are kept by the english parser.
func keywordQuery(arg int) string {
	return fmt.Sprintf(`to_tsquery('english', coalesce((SELECT string_agg(lexeme, ' | ')
		FROM unnest(tsvector_to_array(to_tsvector('english', $%d))) AS lexeme
		WHERE lexeme ~ '^[[:alnum:]]+$'), ''))`, arg)
}

// TextSearchWithOptions is the keyword (PostgreSQL full-text) recall leg. It
// shares the tenant and scope contract of SemanticSearchWithOptions and ranks
// by ts_rank_cd. Stop-word-only queries return no rows.
func (s *Service) TextSearchWithOptions(ctx context.Context, query string, opts SemanticSearchOptions) ([]VectorResult, error) {
	return s.keywordSearch(ctx, query, opts, false)
}

func (s *Service) keywordSearch(ctx context.Context, query string, opts SemanticSearchOptions, pendingOnly bool) ([]VectorResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 5
	}
	if len(textSearchTerms(query)) == 0 {
		return []VectorResult{}, nil
	}
	clauses, args, nextArg := recallScopeClauses(opts, nil, 1)
	if pendingOnly {
		clauses = append(clauses, "embedding IS NULL")
	}
	args = append(args, query, limit)
	sqlQuery := `
		SELECT id, content, metadata, ts_rank_cd(` + keywordDocument + `, kq.q) AS score, created_at
		FROM context_vectors, (SELECT ` + keywordQuery(nextArg) + ` AS q) kq
		WHERE ` + strings.Join(append(clauses, keywordDocument+" @@ kq.q"), " AND ") + `
		ORDER BY score DESC, created_at DESC
		LIMIT $` + fmt.Sprint(nextArg+1)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("keyword search failed: %w", err)
	}
	defer rows.Close()

	results := []VectorResult{}
	for rows.Next() {
		var r VectorResult
		var metaJSON []byte
		if err := rows.Scan(&r.ID, &r.Content, &metaJSON, &r.Score, &r.CreatedAt); err != nil {
			return nil, err
		}
		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &r.Metadata)
		}
		r.RetrievalMode = RecallModeKeyword
		results = append(results, r)
	}
	return results, rows.Err()
}

// recallScopeClauses builds the tenant, run, type, visibility, class,
// sensitivity and goal-set clauses shared by semantic and keyword recall.
func recallScopeClauses(opts SemanticSearchOptions, args []any, nextArg int) ([]string, []any, int) {
	tenantID := strings.TrimSpace(opts.TenantID)
	if tenantID == "" {
		tenantID = "default"
	}
	// Archived governed rows never reach recall, citations or prompts;
	// deleted rows no longer exist.
	clauses := []string{fmt.Sprintf("COALESCE(metadata->>'tenant_id', 'default') = $%d", nextArg),
		"COALESCE(metadata->>'lifecycle_state', 'active') = 'active'"}
	args = append(args, tenantID)
	nextArg++
	if runID := strings.TrimSpace(opts.RunID); runID != "" {
		clauses = append(clauses, fmt.Sprintf("metadata->>'run_id' = $%d", nextArg))
		args = append(args, runID)
		nextArg++
	}
	clauses, args, nextArg = appendTextSearchTypes(clauses, args, nextArg, opts.Types)
	clauses, args, nextArg = appendTextSearchScope(clauses, args, nextArg, opts)
	if opts.Reader != nil {
		var read string
		read, args, nextArg = governedRecallClause(opts.Reader, args, nextArg)
		clauses = append(clauses, read)
	}
	return appendGovernedScope(clauses, args, nextArg, opts)
}

func appendGovernedScope(clauses []string, args []any, nextArg int, opts SemanticSearchOptions) ([]string, []any, int) {
	if classes := nonEmpty(opts.ExcludeClasses); len(classes) > 0 {
		clauses = append(clauses, fmt.Sprintf("NOT (COALESCE(metadata->>'knowledge_class', '') = ANY($%d::text[]))", nextArg))
		args = append(args, textArrayLiteral(classes))
		nextArg++
	}
	if levels := nonEmpty(opts.ExcludeSensitivity); len(levels) > 0 {
		clause := fmt.Sprintf("NOT (COALESCE(metadata->>'sensitivity_class', '') = ANY($%d::text[]))", nextArg)
		args = append(args, textArrayLiteral(levels))
		nextArg++
		if teamID := strings.TrimSpace(opts.TeamID); teamID != "" {
			clause = fmt.Sprintf("(%s OR metadata->>'team_id' = $%d)", clause, nextArg)
			args = append(args, teamID)
			nextArg++
		}
		clauses = append(clauses, clause)
	}
	goalScoped := "COALESCE(jsonb_array_length(CASE WHEN jsonb_typeof(metadata->'target_goal_sets') = 'array' THEN metadata->'target_goal_sets' END), 0) = 0"
	if goals := nonEmpty(opts.GoalSets); len(goals) > 0 {
		goalScoped = fmt.Sprintf("(%s OR metadata->'target_goal_sets' ?| $%d::text[])", goalScoped, nextArg)
		args = append(args, textArrayLiteral(goals))
		nextArg++
	}
	return append(clauses, goalScoped), args, nextArg
}

// textArrayLiteral encodes values as a PostgreSQL text[] literal so every
// driver passes it as one text parameter.
func textArrayLiteral(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		value = strings.ReplaceAll(value, `\`, `\\`)
		quoted[i] = `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return "{" + strings.Join(quoted, ",") + "}"
}

func nonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func appendTextSearchTypes(clauses []string, args []any, nextArg int, types []string) ([]string, []any, int) {
	typeParts := make([]string, 0, len(types))
	for _, value := range nonEmpty(types) {
		typeParts = append(typeParts, fmt.Sprintf("metadata->>'type' = $%d", nextArg))
		args = append(args, value)
		nextArg++
	}
	if len(typeParts) > 0 {
		clauses = append(clauses, "("+strings.Join(typeParts, " OR ")+")")
	}
	return clauses, args, nextArg
}

func appendTextSearchScope(clauses []string, args []any, nextArg int, opts SemanticSearchOptions) ([]string, []any, int) {
	teamID := strings.TrimSpace(opts.TeamID)
	agentID := strings.TrimSpace(opts.AgentID)
	visibility := strings.ToLower(strings.TrimSpace(opts.Visibility))
	if teamID != "" || agentID != "" {
		scopeParts := []string{}
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
	} else if visibility != "" {
		clauses = append(clauses, fmt.Sprintf("COALESCE(metadata->>'visibility', '') = $%d", nextArg))
		args = append(args, visibility)
		nextArg++
	}
	return clauses, args, nextArg
}

// textSearchTerms reports whether a query has any word worth ranking.
func textSearchTerms(query string) []string {
	seen := map[string]bool{}
	terms := []string{}
	for _, term := range strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	}) {
		if len(term) < 2 || seen[term] {
			continue
		}
		seen[term] = true
		terms = append(terms, term)
	}
	return terms
}
