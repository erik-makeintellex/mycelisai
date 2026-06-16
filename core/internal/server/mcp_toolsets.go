package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/mcp"
)

type mcpToolSetRequest struct {
	Name              string               `json:"name"`
	Description       string               `json:"description"`
	ToolRefs          []string             `json:"tool_refs"`
	ScopeKind         string               `json:"scope_kind"`
	ScopeRef          string               `json:"scope_ref"`
	GovernanceContext mcpGovernanceContext `json:"governance_context,omitempty"`
}

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

	var req mcpToolSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		respondError(w, "name is required", http.StatusBadRequest)
		return
	}

	created, err := s.MCPToolSets.Create(r.Context(), mcp.ToolSet{
		Name:        req.Name,
		Description: req.Description,
		ToolRefs:    req.ToolRefs,
		ScopeKind:   req.ScopeKind,
		ScopeRef:    req.ScopeRef,
	})
	if err != nil {
		respondError(w, "Failed to create tool set: "+err.Error(), mcpToolSetErrorStatus(err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	respondJSON(w, map[string]interface{}{
		"ok":         true,
		"data":       created,
		"governance": buildMCPConfigGovernanceDecision(normalizeMCPGovernanceContext(r, req.GovernanceContext), "local", "low"),
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

	var req mcpToolSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		respondError(w, "name is required", http.StatusBadRequest)
		return
	}

	updated, err := s.MCPToolSets.Update(r.Context(), id, mcp.ToolSet{
		Name:        req.Name,
		Description: req.Description,
		ToolRefs:    req.ToolRefs,
		ScopeKind:   req.ScopeKind,
		ScopeRef:    req.ScopeRef,
	})
	if err != nil {
		respondError(w, "Failed to update tool set: "+err.Error(), mcpToolSetErrorStatus(err))
		return
	}
	respondJSON(w, map[string]interface{}{
		"ok":         true,
		"data":       updated,
		"governance": buildMCPConfigGovernanceDecision(normalizeMCPGovernanceContext(r, req.GovernanceContext), "local", "low"),
	})
}

func mcpToolSetErrorStatus(err error) int {
	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "not found"):
		return http.StatusNotFound
	case strings.Contains(lower, "scope_kind") || strings.Contains(lower, "scope_ref"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
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
		"ok":         true,
		"deleted":    id.String(),
		"governance": buildOwnedMCPConfigDecision(r),
	})
}
