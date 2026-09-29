package server

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
)

// Inception recipe routes (MEM-LANES-3). A recipe is saved memory with an
// owner, like a remembered fact: inception_recipes has no owner column, so
// its owner-stamped vector row is the ownership record
// (memory.ReadableInceptionRecipes, memory.InceptionRecipeOwnership), the
// same rule the recall_inception_recipes tool applies.
//
//   - Every recipe route needs an identity (anon 401 blocker).
//   - Reads return only recipes the caller may read: their own, team shares
//     for proven members, and org-wide ones. An unknown and an unreadable id
//     answer with the same 404.
//   - Writes (inception_writes.go) follow the M2 save and manage rules and
//     are audited before they happen.
const (
	codeRecipeSignInRequired = "inception_sign_in_required"
	codeRecipeNotFound       = "inception_recipe_not_found"
	maxRecipeLimit           = 100
)

var (
	recipeSignInCopy = roleBlockerText{User: blockerText{"Sign in to see prompt recipes.",
		"Sign in, then try again. Nothing was shown or changed."}}
	recipeNotFoundCopy = roleBlockerText{User: blockerText{"This prompt recipe does not exist.",
		"Refresh the list. It may have been removed, or it is not shared with you."}}
)

// GET /api/v1/inception/contracts
// Returns frozen P0 contract shapes for decision/runtime integration.
func (s *AdminServer) HandleInceptionContracts(w http.ResponseWriter, r *http.Request) {
	bundle := protocol.DefaultInceptionContractBundle()
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(bundle))
}

// recipeCaller is the identity and owner-lane reader of a recipe request; it
// writes the blocker and returns ok=false when there is none.
func (s *AdminServer) recipeCaller(w http.ResponseWriter, r *http.Request) (*RequestIdentity, memory.GovernedReader, bool) {
	if s.Inception == nil {
		respondAPIError(w, "Inception store not available", http.StatusServiceUnavailable)
		return nil, memory.GovernedReader{}, false
	}
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		respondBlockerText(w, r, http.StatusUnauthorized, codeRecipeSignInRequired, recipeSignInCopy, "", nil)
		return nil, memory.GovernedReader{}, false
	}
	if s.Mem == nil {
		respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, "memory service offline: recipe access cannot be checked", nil)
		return nil, memory.GovernedReader{}, false
	}
	reader, err := s.memoryReader(r)
	if err != nil {
		log.Printf("[inception] recipe access check failed: %v", err)
		respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, "team membership could not be checked", nil)
		return nil, memory.GovernedReader{}, false
	}
	return identity, reader, true
}

func recipeLimit(r *http.Request, def int) int {
	if parsed, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && parsed > 0 {
		def = parsed
	}
	if def > maxRecipeLimit {
		def = maxRecipeLimit
	}
	return def
}

// readableRecipes keeps the recipes reader may read, at most limit.
func (s *AdminServer) readableRecipes(ctx context.Context, recipes []inception.Recipe, reader memory.GovernedReader, limit int) ([]inception.Recipe, error) {
	ids := make([]string, 0, len(recipes))
	for _, rec := range recipes {
		ids = append(ids, rec.ID)
	}
	readable, err := s.Mem.ReadableInceptionRecipes(ctx, ids, reader)
	if err != nil {
		return nil, err
	}
	out := []inception.Recipe{}
	for _, rec := range recipes {
		if readable[rec.ID] && len(out) < limit {
			out = append(out, rec)
		}
	}
	return out, nil
}

// respondRecipes answers a fetched page filtered to the caller's readable
// recipes; a failed check returns nothing, never the unfiltered page.
func (s *AdminServer) respondRecipes(w http.ResponseWriter, r *http.Request, recipes []inception.Recipe, err error, reader memory.GovernedReader, limit int) {
	if err == nil {
		recipes, err = s.readableRecipes(r.Context(), recipes, reader, limit)
	}
	if err != nil {
		log.Printf("[inception] recipe read failed: %v", err)
		respondAPIError(w, "Failed to read recipes", http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.APIResponse{OK: true, Data: recipes})
}

// GET /api/v1/inception/recipes?category=X&agent=Y&limit=N
func (s *AdminServer) HandleListInceptionRecipes(w http.ResponseWriter, r *http.Request) {
	_, reader, ok := s.recipeCaller(w, r)
	if !ok {
		return
	}
	limit := recipeLimit(r, 20)
	recipes, err := s.Inception.ListRecipes(r.Context(), r.URL.Query().Get("category"), r.URL.Query().Get("agent"), maxRecipeLimit)
	s.respondRecipes(w, r, recipes, err, reader, limit)
}

// GET /api/v1/inception/recipes/search?q=...&limit=N
func (s *AdminServer) HandleSearchInceptionRecipes(w http.ResponseWriter, r *http.Request) {
	_, reader, ok := s.recipeCaller(w, r)
	if !ok {
		return
	}
	query := r.URL.Query().Get("q")
	if query == "" {
		respondAPIError(w, "Missing search query 'q'", http.StatusBadRequest)
		return
	}
	limit := recipeLimit(r, 10)
	recipes, err := s.Inception.SearchByTitle(r.Context(), query, maxRecipeLimit)
	s.respondRecipes(w, r, recipes, err, reader, limit)
}

// readableRecipeOwnership answers 404 for an unknown and an unreadable id
// alike, and 503 when the check failed.
func (s *AdminServer) readableRecipeOwnership(w http.ResponseWriter, r *http.Request, id string, reader memory.GovernedReader) (memory.RecipeOwnership, bool) {
	rec, found, err := s.Mem.InceptionRecipeOwnership(r.Context(), id, reader)
	if err != nil {
		log.Printf("[inception] recipe owner check failed: %v", err)
		respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, "recipe access could not be checked", nil)
		return rec, false
	}
	if !found {
		respondBlockerText(w, r, http.StatusNotFound, codeRecipeNotFound, recipeNotFoundCopy, "", nil)
		return rec, false
	}
	return rec, true
}

// GET /api/v1/inception/recipes/{id}
func (s *AdminServer) HandleGetInceptionRecipe(w http.ResponseWriter, r *http.Request) {
	_, reader, ok := s.recipeCaller(w, r)
	if !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if _, ok := s.readableRecipeOwnership(w, r, id, reader); !ok {
		return
	}
	recipe, err := s.Inception.GetRecipe(r.Context(), id)
	if err != nil {
		respondBlockerText(w, r, http.StatusNotFound, codeRecipeNotFound, recipeNotFoundCopy, "", nil)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.APIResponse{OK: true, Data: recipe})
}
