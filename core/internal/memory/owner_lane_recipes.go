package memory

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// ReadableInceptionRecipes reports which recipe ids the reader may read
// (MEM-LANES-3). inception_recipes has no owner column, so a recipe's
// owner-stamped context_vectors row (type inception_recipe, recipe_id) is its
// ownership record and the owner-lane rule decides: the reader's own recipes,
// team shares for proven members, and org-wide ones. A recipe with no
// readable row is not returned.
func (s *Service) ReadableInceptionRecipes(ctx context.Context, ids []string, reader GovernedReader) (map[string]bool, error) {
	readable := map[string]bool{}
	ids = nonEmpty(ids)
	if len(ids) == 0 {
		return readable, nil
	}
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("memory service offline")
	}
	lanes, args, _ := ownerLaneClause("metadata", reader, []any{textArrayLiteral(ids)}, 2)
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT metadata->>'recipe_id' FROM context_vectors
		WHERE metadata->>'type' = 'inception_recipe' AND metadata->>'recipe_id' = ANY($1::text[])
		AND COALESCE(metadata->>'lifecycle_state', 'active') = 'active' AND `+lanes, args...)
	if err != nil {
		return nil, fmt.Errorf("readable inception recipes: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		readable[strings.TrimSpace(id)] = true
	}
	return readable, rows.Err()
}

// RecipeOwnership is a recipe's owner record: its owner-stamped vector row.
type RecipeOwnership struct {
	OwnerUserID string
	Visibility  string
	TeamID      string
	TenantID    string
}

// InceptionRecipeOwnership returns the owner record of a recipe the reader
// may read (MEM-LANES-3). found is false for an unknown recipe and for one
// the reader cannot read alike, so callers answer both the same way.
func (s *Service) InceptionRecipeOwnership(ctx context.Context, id string, reader GovernedReader) (RecipeOwnership, bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return RecipeOwnership{}, false, nil
	}
	if s == nil || s.db == nil {
		return RecipeOwnership{}, false, fmt.Errorf("memory service offline")
	}
	lanes, args, _ := ownerLaneClause("metadata", reader, []any{id}, 2)
	var rec RecipeOwnership
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(metadata->>'owner_user_id', ''), COALESCE(metadata->>'visibility', ''),
		COALESCE(metadata->>'team_id', ''), COALESCE(NULLIF(metadata->>'tenant_id', ''), 'default')
		FROM context_vectors WHERE metadata->>'type' = 'inception_recipe' AND metadata->>'recipe_id' = $1
		AND COALESCE(metadata->>'lifecycle_state', 'active') = 'active' AND `+lanes+` ORDER BY created_at LIMIT 1`, args...).
		Scan(&rec.OwnerUserID, &rec.Visibility, &rec.TeamID, &rec.TenantID)
	if err == sql.ErrNoRows {
		return RecipeOwnership{}, false, nil
	}
	if err != nil {
		return RecipeOwnership{}, false, fmt.Errorf("inception recipe owner: %w", err)
	}
	return rec, true, nil
}

// InceptionRecipeRecord is the owner record written beside a recipe row.
type InceptionRecipeRecord struct {
	RecipeID, Category, Title, IntentPattern     string
	TenantID, TeamID, AgentID, RunID, Visibility string
	OwnerUserID                                  string
}

// StoreInceptionRecipeRecord writes a recipe's owner-stamped vector row, its
// ownership record for reads (MEM-LANES-3). A nil embedding stores the row
// pending, so keyword recall and the owner check still work. The tool path
// and the HTTP create route both write through it.
func (s *Service) StoreInceptionRecipeRecord(ctx context.Context, rec InceptionRecipeRecord, embedding []float64) error {
	text := InceptionRecipeText(rec.Category, rec.Title, rec.IntentPattern)
	meta := OwnerLaneMetadata(map[string]any{"type": "inception_recipe", "category": rec.Category, "source": "inception_recipe",
		"recipe_id": rec.RecipeID, "tenant_id": rec.TenantID, "team_id": rec.TeamID, "agent_id": rec.AgentID, "run_id": rec.RunID,
		"visibility": rec.Visibility}, rec.OwnerUserID)
	return s.StoreVector(ctx, text, embedding, meta)
}

// InceptionRecipeText is the indexed text of a recipe.
func InceptionRecipeText(category, title, intentPattern string) string {
	return fmt.Sprintf("[inception:%s] %s — %s", category, title, intentPattern)
}
