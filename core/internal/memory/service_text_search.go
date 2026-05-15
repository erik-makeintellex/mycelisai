package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// TextSearchWithOptions provides a bounded local-source fallback when semantic
// embeddings are unavailable. It reuses the same tenant and scope contract as
// SemanticSearchWithOptions, but ranks recent retained context by text matches.
func (s *Service) TextSearchWithOptions(ctx context.Context, query string, opts SemanticSearchOptions) ([]VectorResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 5
	}
	clauses, args, nextArg := textSearchScopeClauses(opts)
	terms := textSearchTerms(query)
	if len(terms) == 0 {
		return []VectorResult{}, nil
	}
	textParts := make([]string, 0, len(terms))
	for _, term := range terms {
		textParts = append(textParts, fmt.Sprintf("(lower(content) LIKE $%d OR lower(metadata::text) LIKE $%d)", nextArg, nextArg))
		args = append(args, "%"+term+"%")
		nextArg++
	}
	clauses = append(clauses, "("+strings.Join(textParts, " OR ")+")")

	sqlQuery := `
		SELECT id, content, metadata, created_at
		FROM context_vectors
		WHERE ` + strings.Join(clauses, " AND ") + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprintf("%d", nextArg)
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("text search failed: %w", err)
	}
	defer rows.Close()

	results := []VectorResult{}
	for rows.Next() {
		var r VectorResult
		var metaJSON []byte
		if err := rows.Scan(&r.ID, &r.Content, &metaJSON, &r.CreatedAt); err != nil {
			return nil, err
		}
		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &r.Metadata)
		}
		r.Score = textSearchScore(r.Content, r.Metadata, terms)
		results = append(results, r)
	}
	return results, nil
}

func textSearchScopeClauses(opts SemanticSearchOptions) ([]string, []any, int) {
	tenantID := strings.TrimSpace(opts.TenantID)
	if tenantID == "" {
		tenantID = "default"
	}
	clauses := []string{"COALESCE(metadata->>'tenant_id', 'default') = $1"}
	args := []any{tenantID}
	nextArg := 2
	if runID := strings.TrimSpace(opts.RunID); runID != "" {
		clauses = append(clauses, fmt.Sprintf("metadata->>'run_id' = $%d", nextArg))
		args = append(args, runID)
		nextArg++
	}
	clauses, args, nextArg = appendTextSearchTypes(clauses, args, nextArg, opts.Types)
	clauses, args, nextArg = appendTextSearchScope(clauses, args, nextArg, opts)
	return clauses, args, nextArg
}

func appendTextSearchTypes(clauses []string, args []any, nextArg int, types []string) ([]string, []any, int) {
	typeParts := make([]string, 0, len(types))
	for _, raw := range types {
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

func textSearchScore(content string, metadata map[string]any, terms []string) float64 {
	if len(terms) == 0 {
		return 0
	}
	haystack := strings.ToLower(content)
	if len(metadata) > 0 {
		metaJSON, _ := json.Marshal(metadata)
		haystack += " " + strings.ToLower(string(metaJSON))
	}
	matches := 0
	for _, term := range terms {
		if strings.Contains(haystack, term) {
			matches++
		}
	}
	return float64(matches) / float64(len(terms))
}
