package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/mcp"
)

type mcpLibraryRequest struct {
	Name              string               `json:"name"`
	Env               map[string]string    `json:"env,omitempty"`
	GovernanceContext mcpGovernanceContext `json:"governance_context,omitempty"`
}

type mcpPreparedLibraryRequest struct {
	Request    mcpLibraryRequest
	Entry      *mcp.LibraryEntry
	Inspection map[string]any
}

type mcpLibraryInstallResult struct {
	Server mcp.ServerConfig
	Tools  []mcp.ToolDef
}

// handleMCPLibrary returns the curated MCP server library organized by category.
// GET /api/v1/mcp/library
func (s *AdminServer) handleMCPLibrary(w http.ResponseWriter, r *http.Request) {
	if s.MCPLibrary == nil {
		http.Error(w, `{"error":"MCP library not loaded"}`, http.StatusServiceUnavailable)
		return
	}
	respondJSON(w, s.MCPLibrary.Categories)
}

// handleMCPLibraryInspect previews policy posture for a curated MCP library install.
// POST /api/v1/mcp/library/inspect
func (s *AdminServer) handleMCPLibraryInspect(w http.ResponseWriter, r *http.Request) {
	if s.MCPLibrary == nil {
		http.Error(w, `{"error":"MCP library not loaded"}`, http.StatusServiceUnavailable)
		return
	}

	prepared, ok := s.prepareMCPLibraryRequest(w, r)
	if !ok {
		return
	}
	respondJSON(w, prepared.Inspection)
}

// handleMCPLibraryInstall installs an MCP server from the curated library by name.
// POST /api/v1/mcp/library/install
func (s *AdminServer) handleMCPLibraryInstall(w http.ResponseWriter, r *http.Request) {
	if s.MCP == nil || s.MCPPool == nil {
		http.Error(w, `{"error":"MCP subsystem not initialized"}`, http.StatusServiceUnavailable)
		return
	}
	if s.MCPLibrary == nil {
		http.Error(w, `{"error":"MCP library not loaded"}`, http.StatusServiceUnavailable)
		return
	}

	prepared, ok := s.prepareMCPLibraryRequest(w, r)
	if !ok {
		return
	}

	if decision, _ := prepared.Inspection["decision"].(string); decision == "require_approval" {
		w.WriteHeader(http.StatusAccepted)
		respondJSON(w, map[string]any{
			"requires_approval": true,
			"inspection":        prepared.Inspection,
		})
		return
	}

	result, ok := s.installMCPLibraryEntry(w, r, prepared, "install")
	if !ok {
		return
	}

	respondJSON(w, map[string]interface{}{
		"server":     redactMCPServerConfig(result.Server),
		"tools":      result.Tools,
		"governance": prepared.Inspection["governance"],
	})
}

// handleMCPLibraryApply runs the curated MCP inspect+install flow as a single API call.
// POST /api/v1/mcp/library/apply
func (s *AdminServer) handleMCPLibraryApply(w http.ResponseWriter, r *http.Request) {
	if s.MCP == nil || s.MCPPool == nil {
		http.Error(w, `{"error":"MCP subsystem not initialized"}`, http.StatusServiceUnavailable)
		return
	}
	if s.MCPLibrary == nil {
		http.Error(w, `{"error":"MCP library not loaded"}`, http.StatusServiceUnavailable)
		return
	}

	prepared, ok := s.prepareMCPLibraryRequest(w, r)
	if !ok {
		return
	}

	if decision, _ := prepared.Inspection["decision"].(string); decision == "require_approval" {
		w.WriteHeader(http.StatusAccepted)
		respondJSON(w, map[string]any{
			"status":            "requires_approval",
			"requires_approval": true,
			"inspection":        prepared.Inspection,
		})
		return
	}

	result, ok := s.installMCPLibraryEntry(w, r, prepared, "apply")
	if !ok {
		return
	}

	respondJSON(w, map[string]any{
		"status":            "installed",
		"requires_approval": false,
		"server":            redactMCPServerConfig(result.Server),
		"tools":             result.Tools,
		"inspection":        prepared.Inspection,
		"governance":        prepared.Inspection["governance"],
	})
}

func (s *AdminServer) prepareMCPLibraryRequest(w http.ResponseWriter, r *http.Request) (mcpPreparedLibraryRequest, bool) {
	var req mcpLibraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON body: %s"}`, err.Error()), http.StatusBadRequest)
		return mcpPreparedLibraryRequest{}, false
	}
	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return mcpPreparedLibraryRequest{}, false
	}

	entry := s.MCPLibrary.FindByName(req.Name)
	if entry == nil {
		http.Error(w, fmt.Sprintf(`{"error":"server %q not found in library"}`, req.Name), http.StatusNotFound)
		return mcpPreparedLibraryRequest{}, false
	}

	inspectCtx := normalizeMCPGovernanceContext(r, req.GovernanceContext)
	return mcpPreparedLibraryRequest{
		Request:    req,
		Entry:      entry,
		Inspection: buildMCPLibraryInspectionReport(entry, inspectCtx),
	}, true
}

func (s *AdminServer) installMCPLibraryEntry(w http.ResponseWriter, r *http.Request, prepared mcpPreparedLibraryRequest, logAction string) (mcpLibraryInstallResult, bool) {
	cfg := prepared.Entry.ToServerConfig(prepared.Request.Env)
	runtimeCfg, err := mcp.ApplyRuntimeDefaults(cfg)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"install failed: prepare runtime defaults: %s"}`, err.Error()), http.StatusInternalServerError)
		return mcpLibraryInstallResult{}, false
	}

	installed, err := s.MCP.Install(r.Context(), runtimeCfg)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"install failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return mcpLibraryInstallResult{}, false
	}

	if err := s.MCPPool.Connect(r.Context(), *installed); err != nil {
		log.Printf("MCP library %s: connect to %s failed (best-effort): %v", logAction, installed.Name, err)
	}

	tools, err := s.MCP.ListTools(r.Context(), installed.ID)
	if err != nil {
		log.Printf("MCP library %s: list tools for %s failed: %v", logAction, installed.Name, err)
		tools = []mcp.ToolDef{}
	}
	if tools == nil {
		tools = []mcp.ToolDef{}
	}

	return mcpLibraryInstallResult{
		Server: *installed,
		Tools:  tools,
	}, true
}
