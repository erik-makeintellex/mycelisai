package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/pkg/protocol"
)

// GET /api/v1/inception/contracts
// Returns frozen P0 contract shapes for decision/runtime integration.
func (s *AdminServer) HandleInceptionContracts(w http.ResponseWriter, r *http.Request) {
	bundle := protocol.DefaultInceptionContractBundle()
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(bundle))
}

// GET /api/v1/inception/recipes
// Lists inception recipes. Optional query params: ?category=X&agent=Y&limit=N
func (s *AdminServer) HandleListInceptionRecipes(w http.ResponseWriter, r *http.Request) {
	if s.Inception == nil {
		respondAPIError(w, "Inception store not available", http.StatusServiceUnavailable)
		return
	}

	category := r.URL.Query().Get("category")
	agentID := r.URL.Query().Get("agent")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	recipes, err := s.Inception.ListRecipes(r.Context(), category, agentID, limit)
	if err != nil {
		log.Printf("[inception] ListRecipes failed: %v", err)
		respondAPIError(w, "Failed to list recipes", http.StatusInternalServerError)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.APIResponse{OK: true, Data: recipes})
}

// GET /api/v1/inception/recipes/search?q=...&limit=N
// Searches recipes by title/intent pattern (trigram text search).
func (s *AdminServer) HandleSearchInceptionRecipes(w http.ResponseWriter, r *http.Request) {
	if s.Inception == nil {
		respondAPIError(w, "Inception store not available", http.StatusServiceUnavailable)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		respondAPIError(w, "Missing search query 'q'", http.StatusBadRequest)
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	recipes, err := s.Inception.SearchByTitle(r.Context(), query, limit)
	if err != nil {
		log.Printf("[inception] SearchByTitle failed: %v", err)
		respondAPIError(w, "Search failed", http.StatusInternalServerError)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.APIResponse{OK: true, Data: recipes})
}

// GET /api/v1/inception/recipes/{id}
// Returns a single recipe by ID.
func (s *AdminServer) HandleGetInceptionRecipe(w http.ResponseWriter, r *http.Request) {
	if s.Inception == nil {
		respondAPIError(w, "Inception store not available", http.StatusServiceUnavailable)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing recipe ID", http.StatusBadRequest)
		return
	}

	recipe, err := s.Inception.GetRecipe(r.Context(), id)
	if err != nil {
		respondAPIError(w, "Recipe not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.APIResponse{OK: true, Data: recipe})
}

// POST /api/v1/inception/recipes
// Creates a new inception recipe via the API (for manual recipe creation).
func (s *AdminServer) HandleCreateInceptionRecipe(w http.ResponseWriter, r *http.Request) {
	if s.Inception == nil {
		respondAPIError(w, "Inception store not available", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Category      string         `json:"category"`
		Title         string         `json:"title"`
		IntentPattern string         `json:"intent_pattern"`
		Parameters    map[string]any `json:"parameters"`
		ExamplePrompt string         `json:"example_prompt"`
		OutcomeShape  string         `json:"outcome_shape"`
		AgentID       string         `json:"agent_id"`
		Tags          []string       `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if body.Category == "" || body.Title == "" || body.IntentPattern == "" {
		respondAPIError(w, "category, title, and intent_pattern are required", http.StatusBadRequest)
		return
	}

	recipe := inception.Recipe{
		Category:      body.Category,
		Title:         body.Title,
		IntentPattern: body.IntentPattern,
		Parameters:    body.Parameters,
		ExamplePrompt: body.ExamplePrompt,
		OutcomeShape:  body.OutcomeShape,
		AgentID:       body.AgentID,
		Tags:          body.Tags,
	}

	id, err := s.Inception.CreateRecipe(r.Context(), recipe)
	if err != nil {
		log.Printf("[inception] CreateRecipe failed: %v", err)
		respondAPIError(w, "Failed to create recipe", http.StatusInternalServerError)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.APIResponse{OK: true, Data: map[string]string{"id": id}})
}

// PATCH /api/v1/inception/recipes/{id}/quality
// Updates the quality score for a recipe (feedback loop).
func (s *AdminServer) HandleUpdateRecipeQuality(w http.ResponseWriter, r *http.Request) {
	if s.Inception == nil {
		respondAPIError(w, "Inception store not available", http.StatusServiceUnavailable)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing recipe ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Score float64 `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if body.Score < 0 || body.Score > 1 {
		respondAPIError(w, "Score must be between 0.0 and 1.0", http.StatusBadRequest)
		return
	}

	if err := s.Inception.UpdateQuality(r.Context(), id, body.Score); err != nil {
		respondAPIError(w, "Failed to update quality", http.StatusInternalServerError)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.APIResponse{OK: true, Data: map[string]string{"status": "updated"}})
}
