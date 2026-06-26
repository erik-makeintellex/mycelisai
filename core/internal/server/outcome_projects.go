package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// HandleListOutcomeProjects returns durable user-facing outcome workspaces.
// GET /api/v1/outcome-projects
func (s *AdminServer) HandleListOutcomeProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.listOutcomeProjectsDB(r.Context(), parseLimit(r.URL.Query().Get("limit"), 20))
	if err != nil {
		respondAPIError(w, "Failed to list outcome projects: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(projects))
}

// HandleCreateOutcomeProject creates a durable outcome workspace without starting execution.
// POST /api/v1/outcome-projects
func (s *AdminServer) HandleCreateOutcomeProject(w http.ResponseWriter, r *http.Request) {
	var req protocol.OutcomeProject
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	project := protocol.NormalizeOutcomeProject(req)
	if err := s.insertOutcomeProjectDB(r.Context(), &project); err != nil {
		respondAPIError(w, "Failed to create outcome project: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(project))
}

// HandleGetOutcomeProject returns one durable outcome workspace.
// GET /api/v1/outcome-projects/{id}
func (s *AdminServer) HandleGetOutcomeProject(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimSpace(r.PathValue("id"))
	if projectID == "" {
		respondAPIError(w, "project_id is required", http.StatusBadRequest)
		return
	}
	project, err := s.getOutcomeProjectDB(r.Context(), projectID)
	if err != nil {
		respondAPIError(w, "Failed to load outcome project: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(project))
}

// HandleListTeamRegistryEntries returns assignments for one outcome workspace.
// GET /api/v1/outcome-projects/{id}/team-registry
func (s *AdminServer) HandleListTeamRegistryEntries(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimSpace(r.PathValue("id"))
	if projectID == "" {
		respondAPIError(w, "project_id is required", http.StatusBadRequest)
		return
	}
	entries, err := s.listTeamRegistryEntriesDB(r.Context(), projectID, parseLimit(r.URL.Query().Get("limit"), 50))
	if err != nil {
		respondAPIError(w, "Failed to list team registry: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(entries))
}

// HandleCreateTeamRegistryEntry attaches a team or specialist to an outcome workspace.
// POST /api/v1/outcome-projects/{id}/team-registry
func (s *AdminServer) HandleCreateTeamRegistryEntry(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimSpace(r.PathValue("id"))
	if projectID == "" {
		respondAPIError(w, "project_id is required", http.StatusBadRequest)
		return
	}
	var req protocol.TeamRegistryEntry
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	req.ProjectID = projectID
	entry := protocol.NormalizeTeamRegistryEntry(req)
	if err := s.insertTeamRegistryEntryDB(r.Context(), &entry); err != nil {
		respondAPIError(w, "Failed to create team registry entry: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(entry))
}
