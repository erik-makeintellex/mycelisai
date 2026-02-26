package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/mcp"
)

// handleListToolSets returns all MCP tool sets.
// GET /api/v1/mcp/toolsets
func (s *AdminServer) handleListToolSets(w http.ResponseWriter, r *http.Request) {
	if s.MCPToolSets == nil {
		respondError(w, "MCP Tool Set service not available", http.StatusServiceUnavailable)
		return
	}

	sets, err := s.MCPToolSets.List(r.Context())
	if err != nil {
		respondError(w, "Failed to list tool sets: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if sets == nil {
		sets = []mcp.ToolSet{}
	}
	respondJSON(w, map[string]interface{}{
		"ok":   true,
		"data": sets,
	})
}

// handleCreateToolSet creates a new MCP tool set.
// POST /api/v1/mcp/toolsets
func (s *AdminServer) handleCreateToolSet(w http.ResponseWriter, r *http.Request) {
	if s.MCPToolSets == nil {
		respondError(w, "MCP Tool Set service not available", http.StatusServiceUnavailable)
		return
	}

	var req mcp.ToolSet
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		respondError(w, "name is required", http.StatusBadRequest)
		return
	}

	created, err := s.MCPToolSets.Create(r.Context(), req)
	if err != nil {
		respondError(w, "Failed to create tool set: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	respondJSON(w, map[string]interface{}{
		"ok":   true,
		"data": created,
	})
}

// handleUpdateToolSet updates an existing MCP tool set.
// PUT /api/v1/mcp/toolsets/{id}
func (s *AdminServer) handleUpdateToolSet(w http.ResponseWriter, r *http.Request) {
	if s.MCPToolSets == nil {
		respondError(w, "MCP Tool Set service not available", http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, "Invalid UUID: "+idStr, http.StatusBadRequest)
		return
	}

	var req mcp.ToolSet
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		respondError(w, "name is required", http.StatusBadRequest)
		return
	}

	updated, err := s.MCPToolSets.Update(r.Context(), id, req)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			status = http.StatusNotFound
		}
		respondError(w, "Failed to update tool set: "+err.Error(), status)
		return
	}
	respondJSON(w, map[string]interface{}{
		"ok":   true,
		"data": updated,
	})
}

// handleDeleteToolSet deletes an MCP tool set.
// DELETE /api/v1/mcp/toolsets/{id}
func (s *AdminServer) handleDeleteToolSet(w http.ResponseWriter, r *http.Request) {
	if s.MCPToolSets == nil {
		respondError(w, "MCP Tool Set service not available", http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, "Invalid UUID: "+idStr, http.StatusBadRequest)
		return
	}

	if err := s.MCPToolSets.Delete(r.Context(), id); err != nil {
		respondError(w, "Failed to delete tool set: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, map[string]interface{}{
		"ok":      true,
		"deleted": id.String(),
	})
}
