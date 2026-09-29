package swarm

import (
	"context"
	"fmt"
	"log"

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

// recallStructuredRecipes returns inception_recipes rows the reader may read
// (MEM-LANES-3). The table has no owner column, so each recipe is admitted
// only through its owner-stamped vector row (memory.ReadableInceptionRecipes);
// a failed check returns nothing, never the unfiltered rows.
func recallStructuredRecipes(ctx context.Context, store *inception.Store, mem *memory.Service, query, category string, limit int, reader memory.GovernedReader) []recipeResult {
	if store == nil || mem == nil {
		return nil
	}
	fetch := limit * 4
	if fetch > 100 {
		fetch = 100
	}
	var (
		recipes []inception.Recipe
		err     error
	)
	if category != "" {
		recipes, err = store.ListRecipes(ctx, category, "", fetch)
	} else {
		recipes, err = store.SearchByTitle(ctx, query, fetch)
	}
	if err != nil {
		return nil
	}
	ids := make([]string, 0, len(recipes))
	for _, rec := range recipes {
		ids = append(ids, rec.ID)
	}
	readable, err := mem.ReadableInceptionRecipes(ctx, ids, reader)
	if err != nil {
		log.Printf("recall_inception_recipes: read check failed, no recipes returned: %v", err)
		return nil
	}
	results := make([]recipeResult, 0, len(recipes))
	for _, rec := range recipes {
		if !readable[rec.ID] || len(results) >= limit {
			continue
		}
		results = append(results, recipeResult{
			ID: rec.ID, Category: rec.Category, Title: rec.Title, IntentPattern: rec.IntentPattern, Parameters: rec.Parameters,
			ExamplePrompt: rec.ExamplePrompt, OutcomeShape: rec.OutcomeShape, QualityScore: rec.QualityScore, UsageCount: rec.UsageCount, Source: "rdbms",
		})
		go func(id string) { _ = store.IncrementUsage(context.Background(), id) }(rec.ID)
	}
	return results
}

// recallVectorRecipes recalls recipe vectors through the reader's owner-lane
// rule; semantic when an embedding engine works, keyword ranking otherwise.
func recallVectorRecipes(ctx context.Context, brain *cognitive.Router, mem *memory.Service, query string, limit int, scope memoryScope, reader *memory.GovernedReader) []recipeResult {
	if mem == nil || reader == nil {
		return nil
	}
	var embedder memory.Embedder
	if brain != nil {
		embedder = brain
	}
	vecResults, _, err := mem.RecallGoverned(ctx, embedder, query, memory.SemanticSearchOptions{
		Limit:               limit,
		TenantID:            scope.TenantID,
		TeamID:              scope.TeamID,
		AgentID:             scope.AgentID,
		RunID:               scope.RunID,
		Types:               []string{"inception_recipe"},
		AllowGlobal:         true,
		AllowLegacyUnscoped: scope.TeamID == "" && scope.AgentID == "",
		Reader:              reader,
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
