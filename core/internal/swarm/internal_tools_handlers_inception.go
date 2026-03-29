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

	scope := resolveMemoryScope(ctx, args)
	recipe := inception.Recipe{Category: category, Title: title, IntentPattern: intentPattern, AgentID: scope.AgentID, Parameters: mapValue(args["parameters"]), ExamplePrompt: stringValue(args["example_prompt"]), OutcomeShape: stringValue(args["outcome_shape"]), Tags: stringSlice(args["tags"])}
	id, err := r.inception.CreateRecipe(ctx, recipe)
	if err != nil {
		return "", fmt.Errorf("store inception recipe failed: %w", err)
	}
	storeInceptionVector(ctx, r.brain, r.mem, category, title, intentPattern, id, scope)
	return mustJSON(map[string]any{"message": fmt.Sprintf("Inception recipe stored: '%s' (category: %s, id: %s)", title, category, id), "recipe_id": id}), nil
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

	results := recallStructuredRecipes(ctx, r.inception, query, category, limit)
	results = append(results, recallVectorRecipes(ctx, r.brain, r.mem, query, limit, resolveMemoryScope(ctx, args))...)
	if results == nil {
		results = []recipeResult{}
	}
	return mustJSON(results), nil
}

func storeInceptionVector(ctx context.Context, brain *cognitive.Router, mem *memory.Service, category, title, intentPattern, id string, scope memoryScope) {
	if brain == nil || mem == nil {
		return
	}
	embeddingText := fmt.Sprintf("[inception:%s] %s — %s", category, title, intentPattern)
	vec, err := brain.Embed(ctx, embeddingText, "")
	if err != nil {
		log.Printf("store_inception_recipe: embedding failed (non-fatal): %v", err)
		return
	}
	meta := map[string]any{"type": "inception_recipe", "category": category, "source": "inception_recipe", "recipe_id": id, "tenant_id": scope.TenantID, "team_id": scope.TeamID, "agent_id": scope.AgentID, "run_id": scope.RunID, "visibility": scope.Visibility}
	if insertErr := mem.StoreVector(ctx, embeddingText, vec, meta); insertErr != nil {
		log.Printf("store_inception_recipe: vector insert failed (non-fatal): %v", insertErr)
	}
}
