package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
)

// Inception recipe writes (MEM-LANES-3), under the M2 rules:
//
//   - create: the recipe belongs to the caller; global (org-wide) needs root
//     admin with memory:write, team needs proven membership of team_id;
//   - quality: the owner changes a private or team recipe; an org-wide one
//     needs root admin with memory:write; an unreadable id is 404 like an
//     unknown one, a readable recipe the caller does not own is 403.
//
// The audit record is written first; without it nothing changes (503).
// There is no delete route.

var recipeTeamNotMemberCopy = roleBlockerText{User: blockerText{"You can share a recipe only with a team you belong to.",
	"Pick one of your teams, or keep the recipe private. Nothing was saved."}}

func isScopedMemoryAdmin(identity *RequestIdentity) bool {
	return identity != nil && identity.Role == "admin" && hasScope(identity, scopeMemoryWrite)
}

// auditRecipeFirst writes the authorized audit record, or the 503 blocker.
func (s *AdminServer) auditRecipeFirst(w http.ResponseWriter, r *http.Request, action, recipeID, visibility string) (string, bool) {
	ctx := map[string]any{"actor": "operator", "user": auditUserLabelFromRequest(r), "action": "inception_recipe_" + action,
		"result_status": "authorized", "recipe_id": recipeID, "visibility": visibility}
	message := fmt.Sprintf("Inception recipe %s: %s (authorized)", action, recipeID)
	auditID, err := s.createAuditEvent(protocol.TemplateChatToProposal, "inception-recipes", message, attachActorIdentity(ctx, r))
	if err != nil || auditID == "" {
		respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, fmt.Sprintf("audit unavailable: %v", err), nil)
		return "", false
	}
	return auditID, true
}

// POST /api/v1/inception/recipes
// {category, title, intent_pattern, parameters?, example_prompt?,
// outcome_shape?, agent_id?, tags?, visibility? (private|team|global), team_id?}
func (s *AdminServer) HandleCreateInceptionRecipe(w http.ResponseWriter, r *http.Request) {
	identity, _, ok := s.recipeCaller(w, r)
	if !ok {
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
		Visibility    string         `json:"visibility"`
		TeamID        string         `json:"team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.Category == "" || body.Title == "" || body.IntentPattern == "" {
		respondAPIError(w, "category, title, and intent_pattern are required", http.StatusBadRequest)
		return
	}
	owner := strings.TrimSpace(identity.UserID)
	if owner == "" {
		respondBlocker(w, r, http.StatusForbidden, codeAdminRequired, "a recipe needs a signed-in user id to belong to", nil)
		return
	}
	visibility, teamID := strings.ToLower(strings.TrimSpace(body.Visibility)), strings.TrimSpace(body.TeamID)
	switch visibility {
	case "", "private":
		visibility, teamID = "private", ""
	case "global":
		if !isScopedMemoryAdmin(identity) {
			respondBlocker(w, r, http.StatusForbidden, codeAdminRequired, "Missing required scope: "+scopeMemoryWrite, map[string]string{"required_scope": scopeMemoryWrite})
			return
		}
	case "team":
		member, err := s.memoryTeamMember(r, "default", teamID)
		if err != nil {
			respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, "team membership could not be checked", nil)
			return
		}
		if !member {
			respondBlockerText(w, r, http.StatusForbidden, codeSitrepsTeamNotMember, recipeTeamNotMemberCopy, "", nil)
			return
		}
	default:
		respondAPIError(w, "visibility must be private, team or global", http.StatusBadRequest)
		return
	}
	auditID, ok := s.auditRecipeFirst(w, r, "create", body.Title, visibility)
	if !ok {
		return
	}
	id, err := s.Inception.CreateRecipe(r.Context(), inception.Recipe{Category: body.Category, Title: body.Title, IntentPattern: body.IntentPattern,
		Parameters: body.Parameters, ExamplePrompt: body.ExamplePrompt, OutcomeShape: body.OutcomeShape, AgentID: body.AgentID, Tags: body.Tags})
	if err != nil {
		log.Printf("[inception] CreateRecipe failed: %v", err)
		respondAPIError(w, "Failed to create recipe", http.StatusInternalServerError)
		return
	}
	if err := s.Mem.StoreInceptionRecipeRecord(r.Context(), memory.InceptionRecipeRecord{RecipeID: id, Category: body.Category, Title: body.Title,
		IntentPattern: body.IntentPattern, TenantID: "default", TeamID: teamID, AgentID: body.AgentID, Visibility: visibility, OwnerUserID: owner}, nil); err != nil {
		log.Printf("[inception] recipe %s owner record failed: %v", id, err)
		respondAPIError(w, "The recipe was saved but its owner could not be recorded, so nobody can read it yet", http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]string{"id": id, "visibility": visibility, "audit_event_id": auditID}))
}

// PATCH /api/v1/inception/recipes/{id}/quality {score}
func (s *AdminServer) HandleUpdateRecipeQuality(w http.ResponseWriter, r *http.Request) {
	identity, reader, ok := s.recipeCaller(w, r)
	if !ok {
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
	id := strings.TrimSpace(r.PathValue("id"))
	rec, ok := s.readableRecipeOwnership(w, r, id, reader)
	if !ok {
		return
	}
	switch {
	case rec.Visibility != "private" && rec.Visibility != "team":
		if !isScopedMemoryAdmin(identity) {
			respondBlocker(w, r, http.StatusForbidden, codeAdminRequired, "Missing required scope: "+scopeMemoryWrite, map[string]string{"required_scope": scopeMemoryWrite})
			return
		}
	case rec.OwnerUserID == "" || rec.OwnerUserID != strings.TrimSpace(identity.UserID):
		respondBlockerText(w, r, http.StatusForbidden, codeMemoryEntryNotOwned, memoryLifecycleCopy[codeMemoryEntryNotOwned], "", nil)
		return
	}
	auditID, ok := s.auditRecipeFirst(w, r, "quality", id, rec.Visibility)
	if !ok {
		return
	}
	if err := s.Inception.UpdateQuality(r.Context(), id, body.Score); err != nil {
		respondAPIError(w, "Failed to update quality", http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]string{"status": "updated", "audit_event_id": auditID}))
}
