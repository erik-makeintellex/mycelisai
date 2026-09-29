package swarm

import (
	"context"
	"fmt"
	"log"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/internal/memory"
)

type recipeResult struct {
	ID            string         `json:"id"`
	Category      string         `json:"category"`
	Title         string         `json:"title"`
	IntentPattern string         `json:"intent_pattern"`
	Parameters    map[string]any `json:"parameters,omitempty"`
	ExamplePrompt string         `json:"example_prompt,omitempty"`
	OutcomeShape  string         `json:"outcome_shape,omitempty"`
	QualityScore  float64        `json:"quality_score"`
	UsageCount    int            `json:"usage_count"`
	Source        string         `json:"source"`
}

func (r *InternalToolRegistry) handleStoreInceptionRecipe(ctx context.Context, args map[string]any) (string, error) {
	category := stringValue(args["category"])
	title := stringValue(args["title"])
	intentPattern := stringValue(args["intent_pattern"])
	if category == "" || title == "" || intentPattern == "" {
		return "", fmt.Errorf("store_inception_recipe requires 'category', 'title', and 'intent_pattern'")
	}
	if r.inception == nil {
		return "", fmt.Errorf("inception store not available")
	}

	// MEM-LANES-3: a recipe belongs to the turn's user like a remembered fact.
	access := recallAccessFromContext(ctx)
	owner, err := access.ownerUserID()
	if err != nil {
		return "", fmt.Errorf("store_inception_recipe refused: %w. Nothing was saved.", err)
	}
	scope := resolveMemoryScope(ctx, args)
	var note string
	scope.Visibility, note = ownerLaneVisibility(access, scope)
	recipe := inception.Recipe{Category: category, Title: title, IntentPattern: intentPattern, AgentID: scope.AgentID, Parameters: mapValue(args["parameters"]), ExamplePrompt: stringValue(args["example_prompt"]), OutcomeShape: stringValue(args["outcome_shape"]), Tags: stringSlice(args["tags"])}
	id, err := r.inception.CreateRecipe(ctx, recipe)
	if err != nil {
		return "", fmt.Errorf("store inception recipe failed: %w", err)
	}
	if err := storeInceptionVector(ctx, r.brain, r.mem, category, title, intentPattern, id, scope, owner); err != nil {
		if owner != "" {
			return "", fmt.Errorf("store_inception_recipe saved recipe %s but could not record its owner, so nobody can recall it: %w", id, err)
		}
		log.Printf("store_inception_recipe: owner record failed for a no-user recipe (non-fatal): %v", err)
	}
	message := withVisibilityNote(fmt.Sprintf("Inception recipe stored: '%s' (category: %s, id: %s)", title, category, id), note)
	return mustJSON(map[string]any{"message": message, "recipe_id": id}), nil
}

func (r *InternalToolRegistry) handleRecallInceptionRecipes(ctx context.Context, args map[string]any) (string, error) {
	query := stringValue(args["query"])
	if query == "" {
		return "", fmt.Errorf("recall_inception_recipes requires 'query'")
	}
	category := stringValue(args["category"])
	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	// MEM-LANES-3: recipes are read through the owner-lane rule.
	reader, err := recallAccessFromContext(ctx).governedReader()
	if err != nil {
		return "", fmt.Errorf("recall_inception_recipes unavailable: %w", err)
	}
	results := recallStructuredRecipes(ctx, r.inception, r.mem, query, category, limit, *reader)
	results = append(results, recallVectorRecipes(ctx, r.brain, r.mem, query, limit, resolveMemoryScope(ctx, args), reader)...)
	if results == nil {
		results = []recipeResult{}
	}
	return mustJSON(results), nil
}

// storeInceptionVector writes the recipe's owner-stamped vector row, its
// ownership record for recall (MEM-LANES-3). Without a working embedding
// engine the row is stored pending, so keyword recall and the owner check
// still work.
func storeInceptionVector(ctx context.Context, brain *cognitive.Router, mem *memory.Service, category, title, intentPattern, id string, scope memoryScope, owner string) error {
	if mem == nil {
		return fmt.Errorf("memory service offline")
	}
	embeddingText := memory.InceptionRecipeText(category, title, intentPattern)
	var vec []float64
	if brain != nil {
		if embedded, err := brain.Embed(ctx, embeddingText, ""); err != nil {
			log.Printf("store_inception_recipe: embedding failed, stored for keyword recall: %v", err)
		} else {
			vec = embedded
		}
	}
	return mem.StoreInceptionRecipeRecord(ctx, memory.InceptionRecipeRecord{RecipeID: id, Category: category, Title: title, IntentPattern: intentPattern,
		TenantID: scope.TenantID, TeamID: scope.TeamID, AgentID: scope.AgentID, RunID: scope.RunID, Visibility: scope.Visibility, OwnerUserID: owner}, vec)
}
