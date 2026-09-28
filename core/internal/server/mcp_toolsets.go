package server

import (
	"encoding/json"
	"log"
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
// MCPA D1/D7: mcp_config:write, then a fail-closed audit before the insert.
func (s *AdminServer) handleCreateToolSet(w http.ResponseWriter, r *http.Request) {
	req, ok := s.decodeToolSetWrite(w, r)
	if !ok {
		return
	}
	kind, ref := req.storedScope()
	record := mcpToolSetAuditRecord("", req.Name, kind, ref, nil, req.ToolRefs)
	auditID, ok := s.auditMCPConfigChange(w, r, "mcp_toolset_created", req.Name, record)
	if !ok {
		return
	}
	created, err := s.MCPToolSets.Create(r.Context(), req.toolSet())
	if err != nil {
		s.respondToolSetWriteError(w, r, "mcp_toolset_create_failed", req.Name, auditID, err, record)
		return
	}
	w.WriteHeader(http.StatusCreated)
	respondJSON(w, map[string]interface{}{
		"ok":             true,
		"data":           created,
		"audit_event_id": auditID,
		"governance":     buildMCPConfigGovernanceDecision(normalizeMCPGovernanceContext(r, req.GovernanceContext), "local", "low"),
	})
}

// handleUpdateToolSet updates an existing MCP tool set.
// PUT /api/v1/mcp/toolsets/{id}
func (s *AdminServer) handleUpdateToolSet(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireMCPConfigWrite(w, r); !ok {
		return
	}
	id, ok := s.toolSetPathID(w, r)
	if !ok {
		return
	}
	req, ok := s.decodeToolSetWrite(w, r)
	if !ok {
		return
	}
	existing, ok := s.resolveToolSet(w, r, id)
	if !ok {
		return
	}
	kind, ref := req.storedScope()
	record := mcpToolSetAuditRecord(id.String(), req.Name, kind, ref, existing.ToolRefs, req.ToolRefs)
	record["previous_name"] = existing.Name
	auditID, ok := s.auditMCPConfigChange(w, r, "mcp_toolset_updated", req.Name, record)
	if !ok {
		return
	}
	updated, err := s.MCPToolSets.Update(r.Context(), id, req.toolSet())
	if err != nil {
		s.respondToolSetWriteError(w, r, "mcp_toolset_update_failed", req.Name, auditID, err, record)
		return
	}
	respondJSON(w, map[string]interface{}{
		"ok":             true,
		"data":           updated,
		"audit_event_id": auditID,
		"governance":     buildMCPConfigGovernanceDecision(normalizeMCPGovernanceContext(r, req.GovernanceContext), "local", "low"),
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
	if _, ok := requireMCPConfigWrite(w, r); !ok {
		return
	}
	id, ok := s.toolSetPathID(w, r)
	if !ok {
		return
	}
	if s.MCPToolSets == nil {
		respondError(w, "MCP Tool Set service not available", http.StatusServiceUnavailable)
		return
	}
	existing, ok := s.resolveToolSet(w, r, id)
	if !ok {
		return
	}
	record := mcpToolSetAuditRecord(id.String(), existing.Name, existing.ScopeKind, existing.ScopeRef, existing.ToolRefs, []string{})
	auditID, ok := s.auditMCPConfigChange(w, r, "mcp_toolset_deleted", existing.Name, record)
	if !ok {
		return
	}
	if err := s.MCPToolSets.Delete(r.Context(), id); err != nil {
		s.respondToolSetWriteError(w, r, "mcp_toolset_delete_failed", existing.Name, auditID, err, record)
		return
	}
	respondJSON(w, map[string]interface{}{
		"ok":             true,
		"deleted":        id.String(),
		"audit_event_id": auditID,
		"governance":     buildOwnedMCPConfigDecision(r),
	})
}

// decodeToolSetWrite runs authority (idempotent for update), the service
// check, and body validation for create/update.
func (s *AdminServer) decodeToolSetWrite(w http.ResponseWriter, r *http.Request) (mcpToolSetRequest, bool) {
	var req mcpToolSetRequest
	if _, ok := requireMCPConfigWrite(w, r); !ok {
		return req, false
	}
	if s.MCPToolSets == nil {
		respondError(w, "MCP Tool Set service not available", http.StatusServiceUnavailable)
		return req, false
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return req, false
	}
	if req.Name == "" {
		respondError(w, "name is required", http.StatusBadRequest)
		return req, false
	}
	if req.ToolRefs == nil {
		req.ToolRefs = []string{}
	}
	return req, true
}

// storedScope is the scope the tool set store will keep for this request, so
// the audit records what is stored rather than the caller's raw input. It is
// the store's own mcp.NormalizeToolSetScope (MCPL); the store refuses an
// invalid scope, and for those the trimmed caller value is recorded.
func (req mcpToolSetRequest) storedScope() (string, string) {
	if kind, ref, err := mcp.NormalizeToolSetScope(req.ScopeKind, req.ScopeRef); err == nil {
		return kind, ref
	}
	return strings.ToLower(strings.TrimSpace(req.ScopeKind)), strings.TrimSpace(req.ScopeRef)
}

func (req mcpToolSetRequest) toolSet() mcp.ToolSet {
	return mcp.ToolSet{Name: req.Name, Description: req.Description, ToolRefs: req.ToolRefs, ScopeKind: req.ScopeKind, ScopeRef: req.ScopeRef}
}

func (s *AdminServer) toolSetPathID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, "Invalid tool set id", http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

// resolveToolSet reads the current set so the audit carries refs before the
// write; an unknown id is 404 with no audit and no write.
func (s *AdminServer) resolveToolSet(w http.ResponseWriter, r *http.Request, id uuid.UUID) (*mcp.ToolSet, bool) {
	existing, err := s.MCPToolSets.Get(r.Context(), id)
	if err != nil {
		log.Printf("MCPA: resolve tool set %s: %s", id, redactedMCPErrorText(err))
		respondError(w, "Failed to read tool set", http.StatusInternalServerError)
		return nil, false
	}
	if existing == nil {
		respondError(w, "tool set not found", http.StatusNotFound)
		return nil, false
	}
	return existing, true
}

func mcpToolSetAuditRecord(id, name, scopeKind, scopeRef string, before, after []string) map[string]any {
	if before == nil {
		before = []string{}
	}
	if after == nil {
		after = []string{}
	}
	return map[string]any{"toolset_id": id, "toolset_name": name, "scope_kind": scopeKind, "scope_ref": scopeRef,
		"tool_refs_before": before, "tool_refs_after": after}
}

// respondToolSetWriteError records the post-audit failure and answers with
// validation text for 400/404 and a fixed message for anything else (D8).
func (s *AdminServer) respondToolSetWriteError(w http.ResponseWriter, r *http.Request, action, name, auditID string, err error, record map[string]any) {
	s.recordMCPConfigFailure(r, action, name, auditID, "write", err, record)
	status := mcpToolSetErrorStatus(err)
	if status == http.StatusInternalServerError {
		respondError(w, "Tool set change failed. Check the Core log for the redacted reason.", status)
		return
	}
	respondError(w, redactedMCPErrorText(err), status)
}
