package swarm

import (
	"context"
	"fmt"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/internal/memory"
)

func mapValue(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func recallStructuredRecipes(ctx context.Context, store *inception.Store, query, category string, limit int) []recipeResult {
	if store == nil {
		return nil
	}
	var (
		recipes []inception.Recipe
		err     error
	)
	if category != "" {
		recipes, err = store.ListRecipes(ctx, category, "", limit)
	} else {
		recipes, err = store.SearchByTitle(ctx, query, limit)
	}
	if err != nil {
		return nil
	}
	results := make([]recipeResult, 0, len(recipes))
	for _, rec := range recipes {
		results = append(results, recipeResult{
			ID: rec.ID, Category: rec.Category, Title: rec.Title, IntentPattern: rec.IntentPattern, Parameters: rec.Parameters,
			ExamplePrompt: rec.ExamplePrompt, OutcomeShape: rec.OutcomeShape, QualityScore: rec.QualityScore, UsageCount: rec.UsageCount, Source: "rdbms",
		})
		go func(id string) { _ = store.IncrementUsage(context.Background(), id) }(rec.ID)
	}
	return results
}

func recallVectorRecipes(ctx context.Context, brain *cognitive.Router, mem *memory.Service, query string, limit int, scope memoryScope) []recipeResult {
	if brain == nil || mem == nil {
		return nil
	}
	vec, err := brain.Embed(ctx, fmt.Sprintf("[inception] %s", query), "")
	if err != nil {
		return nil
	}
	vecResults, err := mem.SemanticSearchWithOptions(ctx, vec, memory.SemanticSearchOptions{
		Limit:               limit,
		TenantID:            scope.TenantID,
		TeamID:              scope.TeamID,
		AgentID:             scope.AgentID,
		RunID:               scope.RunID,
		Types:               []string{"inception_recipe"},
		AllowGlobal:         true,
		AllowLegacyUnscoped: scope.TeamID == "" && scope.AgentID == "",
	})
	if err != nil {
		return nil
	}
	results := make([]recipeResult, 0, len(vecResults))
	for _, vr := range vecResults {
		if src, ok := vr.Metadata["source"].(string); ok && src == "inception_recipe" {
			results = append(results, recipeResult{ID: fmt.Sprintf("%v", vr.Metadata["recipe_id"]), Category: fmt.Sprintf("%v", vr.Metadata["category"]), Title: vr.Content, Source: "vector"})
		}
	}
	return results
}
