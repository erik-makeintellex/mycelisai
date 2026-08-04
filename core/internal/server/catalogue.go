package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/catalogue"
)

// handleListCatalogue returns all reusable worker profiles.
// GET /api/v1/catalogue/agents
func (s *AdminServer) handleListCatalogue(w http.ResponseWriter, r *http.Request) {
	if s.Catalogue == nil {
		http.Error(w, `{"error":"catalogue not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	agents, err := s.Catalogue.List(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"list failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if agents == nil {
		agents = []catalogue.AgentTemplate{}
	}

	respondJSON(w, agents)
}

// handleCreateCatalogue creates a new worker profile.
// POST /api/v1/catalogue/agents
func (s *AdminServer) handleCreateCatalogue(w http.ResponseWriter, r *http.Request) {
	if s.Catalogue == nil {
		http.Error(w, `{"error":"catalogue not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	var input catalogue.AgentTemplate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	if input.Role == "" {
		http.Error(w, `{"error":"role is required"}`, http.StatusBadRequest)
		return
	}
	input.ProfileKey = ""
	input.Source = "user"
	input.Locked = false

	created, err := s.Catalogue.Create(r.Context(), input)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"create failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, created)
}

// handleUpdateCatalogue updates an existing worker profile.
// PUT /api/v1/catalogue/agents/{id}
func (s *AdminServer) handleUpdateCatalogue(w http.ResponseWriter, r *http.Request) {
	if s.Catalogue == nil {
		http.Error(w, `{"error":"catalogue not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid id: %s"}`, idStr), http.StatusBadRequest)
		return
	}

	var input catalogue.AgentTemplate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	if input.Role == "" {
		http.Error(w, `{"error":"role is required"}`, http.StatusBadRequest)
		return
	}
	input.ProfileKey = ""
	input.Source = "user"
	input.Locked = false

	updated, err := s.Catalogue.Update(r.Context(), id, input)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"update failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	respondJSON(w, updated)
}

// handleDeleteCatalogue removes a worker profile.
// DELETE /api/v1/catalogue/agents/{id}
func (s *AdminServer) handleDeleteCatalogue(w http.ResponseWriter, r *http.Request) {
	if s.Catalogue == nil {
		http.Error(w, `{"error":"catalogue not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid id: %s"}`, idStr), http.StatusBadRequest)
		return
	}

	if err := s.Catalogue.Delete(r.Context(), id); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"delete failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{"status": "deleted"})
}
