package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// HandleListTeamWork returns durable work items for one runtime or persisted team.
// GET /api/v1/teams/{id}/work
func (s *AdminServer) HandleListTeamWork(w http.ResponseWriter, r *http.Request) {
	teamID := strings.TrimSpace(r.PathValue("id"))
	if teamID == "" {
		respondAPIError(w, "team_id is required", http.StatusBadRequest)
		return
	}
	items, err := s.listTeamWorkItemsDB(r.Context(), teamID, parseLimit(r.URL.Query().Get("limit"), 20))
	if err != nil {
		respondAPIError(w, "Failed to list team work: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(items))
}

// HandleCreateTeamWork records a durable team-work item without starting runtime execution.
// POST /api/v1/teams/{id}/work
func (s *AdminServer) HandleCreateTeamWork(w http.ResponseWriter, r *http.Request) {
	teamID := strings.TrimSpace(r.PathValue("id"))
	if teamID == "" {
		respondAPIError(w, "team_id is required", http.StatusBadRequest)
		return
	}
	var req protocol.TeamWorkItem
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	req.TeamID = teamID
	item := protocol.NormalizeTeamWorkItem(req)
	if err := protocol.ValidateTeamWorkItem(item); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.insertTeamWorkItemDB(r.Context(), &item); err != nil {
		respondAPIError(w, "Failed to create team work: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(item))
}

// HandleListTeamInteractions returns the durable exchange history for one work item.
// GET /api/v1/teams/{id}/work/{workItemId}/interactions
func (s *AdminServer) HandleListTeamInteractions(w http.ResponseWriter, r *http.Request) {
	teamID, workItemID := teamWorkPathIDs(r)
	if teamID == "" || workItemID == "" {
		respondAPIError(w, "team_id and work_item_id are required", http.StatusBadRequest)
		return
	}
	if err := validateOptionalUUID("work_item_id", workItemID); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	items, err := s.listTeamInteractionsDB(r.Context(), teamID, workItemID, parseLimit(r.URL.Query().Get("limit"), 50))
	if err != nil {
		respondAPIError(w, "Failed to list team interactions: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(items))
}

// HandleListTeamStatusEvents returns the durable operator-readable status timeline for one work item.
// GET /api/v1/teams/{id}/work/{workItemId}/status-events
func (s *AdminServer) HandleListTeamStatusEvents(w http.ResponseWriter, r *http.Request) {
	teamID, workItemID := teamWorkPathIDs(r)
	if teamID == "" || workItemID == "" {
		respondAPIError(w, "team_id and work_item_id are required", http.StatusBadRequest)
		return
	}
	if err := validateOptionalUUID("work_item_id", workItemID); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	items, err := s.listTeamStatusEventsDB(r.Context(), teamID, workItemID, parseLimit(r.URL.Query().Get("limit"), 50))
	if err != nil {
		respondAPIError(w, "Failed to list team status events: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(items))
}

// HandleCreateTeamInteraction appends a durable Soma/Council/operator/team-lead exchange.
// POST /api/v1/teams/{id}/work/{workItemId}/interactions
func (s *AdminServer) HandleCreateTeamInteraction(w http.ResponseWriter, r *http.Request) {
	teamID, workItemID := teamWorkPathIDs(r)
	if teamID == "" || workItemID == "" {
		respondAPIError(w, "team_id and work_item_id are required", http.StatusBadRequest)
		return
	}
	var req protocol.TeamInteraction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	req.TeamID = teamID
	req.WorkItemID = workItemID
	item := protocol.NormalizeTeamInteraction(req)
	if err := protocol.ValidateTeamInteraction(item); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.insertTeamInteractionDB(r.Context(), &item); err != nil {
		respondAPIError(w, "Failed to create team interaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(item))
}

func teamWorkPathIDs(r *http.Request) (string, string) {
	return strings.TrimSpace(r.PathValue("id")), strings.TrimSpace(r.PathValue("workItemId"))
}
