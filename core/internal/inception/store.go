// Package inception provides structured prompt recipe persistence for Soma.
// Inception recipes capture "how to ask for X" patterns that agents distill
// after successfully completing complex tasks. Recipes are dual-persisted:
// RDBMS (structured query) + pgvector (semantic recall via context_vectors).
package inception

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Recipe represents a structured prompt pattern stored by an agent.
type Recipe struct {
	ID              string         `json:"id"`
	TenantID        string         `json:"tenant_id"`
	Category        string         `json:"category"`
	Title           string         `json:"title"`
	IntentPattern   string         `json:"intent_pattern"`
	Parameters      map[string]any `json:"parameters,omitempty"`
	ExamplePrompt   string         `json:"example_prompt,omitempty"`
	OutcomeShape    string         `json:"outcome_shape,omitempty"`
	SourceRunID     string         `json:"source_run_id,omitempty"`
	SourceSessionID string         `json:"source_session_id,omitempty"`
	AgentID         string         `json:"agent_id"`
	Tags            []string       `json:"tags,omitempty"`
	QualityScore    float64        `json:"quality_score"`
	UsageCount      int            `json:"usage_count"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// Store handles inception recipe persistence.
type Store struct {
	db *sql.DB
}

// NewStore creates a new inception store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// CreateRecipe inserts a new inception recipe and returns its ID.
func (s *Store) CreateRecipe(ctx context.Context, r Recipe) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("inception store: database not available")
	}

	if r.TenantID == "" {
		r.TenantID = "default"
	}
	if r.AgentID == "" {
		r.AgentID = "admin"
	}
	if r.Tags == nil {
		r.Tags = []string{}
	}
	if r.Parameters == nil {
		r.Parameters = map[string]any{}
	}

	paramsJSON, err := json.Marshal(r.Parameters)
	if err != nil {
		paramsJSON = []byte("{}")
	}

	var id string
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO inception_recipes
			(tenant_id, category, title, intent_pattern, parameters, example_prompt, outcome_shape,
			 source_run_id, source_session_id, agent_id, tags, quality_score)
		VALUES ($1, $2, $3, $4, $5,
			NULLIF($6, ''), NULLIF($7, ''),
			NULLIF($8, ''), NULLIF($9, ''),
			$10, $11, $12)
		RETURNING id
	`,
		r.TenantID, r.Category, r.Title, r.IntentPattern, string(paramsJSON),
		r.ExamplePrompt, r.OutcomeShape,
		r.SourceRunID, r.SourceSessionID,
		r.AgentID, pq.Array(r.Tags), r.QualityScore,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("inception store: insert failed: %w", err)
	}

	return id, nil
}

// GetRecipe retrieves a single recipe by ID.
func (s *Store) GetRecipe(ctx context.Context, id string) (*Recipe, error) {
	if s.db == nil {
		return nil, fmt.Errorf("inception store: database not available")
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, category, title, intent_pattern, COALESCE(parameters::text, '{}'),
			COALESCE(example_prompt, ''), COALESCE(outcome_shape, ''),
			COALESCE(source_run_id::text, ''), COALESCE(source_session_id::text, ''),
			agent_id, COALESCE(tags, '{}'), quality_score, usage_count, created_at, updated_at
		FROM inception_recipes WHERE id = $1
	`, id)

	return scanRecipe(row)
}

// ListRecipes retrieves recipes with optional category and agent filters.
func (s *Store) ListRecipes(ctx context.Context, category, agentID string, limit int) ([]Recipe, error) {
	if s.db == nil {
		return nil, fmt.Errorf("inception store: database not available")
	}
	if limit <= 0 {
		limit = 20
	}

	var conditions []string
	var args []any
	argIdx := 1

	if category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}
	if agentID != "" {
		conditions = append(conditions, fmt.Sprintf("agent_id = $%d", argIdx))
		args = append(args, agentID)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	args = append(args, limit)
	query := fmt.Sprintf(`
		SELECT id, tenant_id, category, title, intent_pattern, COALESCE(parameters::text, '{}'),
			COALESCE(example_prompt, ''), COALESCE(outcome_shape, ''),
			COALESCE(source_run_id::text, ''), COALESCE(source_session_id::text, ''),
			agent_id, COALESCE(tags, '{}'), quality_score, usage_count, created_at, updated_at
		FROM inception_recipes %s
		ORDER BY usage_count DESC, quality_score DESC, created_at DESC
		LIMIT $%d
	`, where, argIdx)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("inception store: list failed: %w", err)
	}
	defer rows.Close()

	return scanRecipes(rows)
}

// SearchByTitle performs a trigram text search on recipe titles.
func (s *Store) SearchByTitle(ctx context.Context, query string, limit int) ([]Recipe, error) {
	if s.db == nil {
		return nil, fmt.Errorf("inception store: database not available")
	}
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, category, title, intent_pattern, COALESCE(parameters::text, '{}'),
			COALESCE(example_prompt, ''), COALESCE(outcome_shape, ''),
			COALESCE(source_run_id::text, ''), COALESCE(source_session_id::text, ''),
			agent_id, COALESCE(tags, '{}'), quality_score, usage_count, created_at, updated_at
		FROM inception_recipes
		WHERE title ILIKE '%' || $1 || '%' OR intent_pattern ILIKE '%' || $1 || '%'
		ORDER BY quality_score DESC, usage_count DESC
		LIMIT $2
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("inception store: search failed: %w", err)
	}
	defer rows.Close()

	return scanRecipes(rows)
}

// IncrementUsage bumps the usage_count for a recipe (called on recall).
func (s *Store) IncrementUsage(ctx context.Context, id string) error {
	if s.db == nil {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE inception_recipes SET usage_count = usage_count + 1, updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

// UpdateQuality sets the quality_score for a recipe (feedback loop).
func (s *Store) UpdateQuality(ctx context.Context, id string, score float64) error {
	if s.db == nil {
		return fmt.Errorf("inception store: database not available")
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE inception_recipes SET quality_score = $1, updated_at = NOW()
		WHERE id = $2
	`, score, id)
	return err
}

// scanRecipe scans a single row into a Recipe.
func scanRecipe(row *sql.Row) (*Recipe, error) {
	var r Recipe
	var paramsJSON string
	err := row.Scan(
		&r.ID, &r.TenantID, &r.Category, &r.Title, &r.IntentPattern, &paramsJSON,
		&r.ExamplePrompt, &r.OutcomeShape,
		&r.SourceRunID, &r.SourceSessionID,
		&r.AgentID, pq.Array(&r.Tags), &r.QualityScore, &r.UsageCount,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(paramsJSON), &r.Parameters); err != nil {
		r.Parameters = map[string]any{}
	}
	return &r, nil
}

// scanRecipes scans multiple rows into a slice of Recipe.
func scanRecipes(rows *sql.Rows) ([]Recipe, error) {
	var recipes []Recipe
	for rows.Next() {
		var r Recipe
		var paramsJSON string
		if err := rows.Scan(
			&r.ID, &r.TenantID, &r.Category, &r.Title, &r.IntentPattern, &paramsJSON,
			&r.ExamplePrompt, &r.OutcomeShape,
			&r.SourceRunID, &r.SourceSessionID,
			&r.AgentID, pq.Array(&r.Tags), &r.QualityScore, &r.UsageCount,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			log.Printf("inception store: scan row: %v", err)
			continue
		}
		if err := json.Unmarshal([]byte(paramsJSON), &r.Parameters); err != nil {
			r.Parameters = map[string]any{}
		}
		recipes = append(recipes, r)
	}
	if recipes == nil {
		recipes = []Recipe{}
	}
	return recipes, nil
}
