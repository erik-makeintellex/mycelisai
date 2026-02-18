package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
)

// handleListArtifacts returns artifacts filtered by query params.
// GET /api/v1/artifacts?mission_id=&team_id=&agent_id=&limit=
func (s *AdminServer) handleListArtifacts(w http.ResponseWriter, r *http.Request) {
	if s.Artifacts == nil {
		http.Error(w, `{"error":"artifacts not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	var result []artifacts.Artifact
	var err error

	missionIDStr := r.URL.Query().Get("mission_id")
	teamIDStr := r.URL.Query().Get("team_id")
	agentIDStr := r.URL.Query().Get("agent_id")

	switch {
	case missionIDStr != "":
		id, parseErr := uuid.Parse(missionIDStr)
		if parseErr != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid mission_id: %s"}`, missionIDStr), http.StatusBadRequest)
			return
		}
		result, err = s.Artifacts.ListByMission(ctx, id, limit)
	case teamIDStr != "":
		id, parseErr := uuid.Parse(teamIDStr)
		if parseErr != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid team_id: %s"}`, teamIDStr), http.StatusBadRequest)
			return
		}
		result, err = s.Artifacts.ListByTeam(ctx, id, limit)
	case agentIDStr != "":
		result, err = s.Artifacts.ListByAgent(ctx, agentIDStr, limit)
	default:
		result, err = s.Artifacts.ListRecent(ctx, limit)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"list failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if result == nil {
		result = []artifacts.Artifact{}
	}

	respondJSON(w, result)
}

// handleGetArtifact returns a single artifact by ID.
// GET /api/v1/artifacts/{id}
func (s *AdminServer) handleGetArtifact(w http.ResponseWriter, r *http.Request) {
	if s.Artifacts == nil {
		http.Error(w, `{"error":"artifacts not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid id: %s"}`, idStr), http.StatusBadRequest)
		return
	}

	artifact, err := s.Artifacts.Get(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	respondJSON(w, artifact)
}

// handleStoreArtifact persists a new artifact.
// POST /api/v1/artifacts
func (s *AdminServer) handleStoreArtifact(w http.ResponseWriter, r *http.Request) {
	if s.Artifacts == nil {
		http.Error(w, `{"error":"artifacts not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	var input artifacts.Artifact
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if input.AgentID == "" {
		http.Error(w, `{"error":"agent_id is required"}`, http.StatusBadRequest)
		return
	}
	if input.ArtifactType == "" {
		http.Error(w, `{"error":"artifact_type is required"}`, http.StatusBadRequest)
		return
	}
	if input.Title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}
	if input.Status == "" {
		input.Status = "pending"
	}

	stored, err := s.Artifacts.Store(r.Context(), input)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"store failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, stored)
}

// handleUpdateArtifactStatus updates the governance status of an artifact.
// PUT /api/v1/artifacts/{id}/status
func (s *AdminServer) handleUpdateArtifactStatus(w http.ResponseWriter, r *http.Request) {
	if s.Artifacts == nil {
		http.Error(w, `{"error":"artifacts not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid id: %s"}`, idStr), http.StatusBadRequest)
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if body.Status == "" {
		http.Error(w, `{"error":"status is required"}`, http.StatusBadRequest)
		return
	}

	if err := s.Artifacts.UpdateStatus(r.Context(), id, body.Status); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{"status": body.Status})
}
