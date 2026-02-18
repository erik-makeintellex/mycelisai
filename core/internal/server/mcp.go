package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/mcp"
)

// handleMCPInstall registers and connects a new MCP server.
// POST /api/v1/mcp/install
func (s *AdminServer) handleMCPInstall(w http.ResponseWriter, r *http.Request) {
	if s.MCP == nil || s.MCPPool == nil {
		http.Error(w, `{"error":"MCP subsystem not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	var cfg mcp.ServerConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON body: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	// Validate required fields.
	if cfg.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	if cfg.Transport == "" {
		http.Error(w, `{"error":"transport is required"}`, http.StatusBadRequest)
		return
	}
	if cfg.Transport == "stdio" && cfg.Command == "" {
		http.Error(w, `{"error":"command is required for stdio transport"}`, http.StatusBadRequest)
		return
	}
	if cfg.Transport == "sse" && cfg.URL == "" {
		http.Error(w, `{"error":"url is required for sse transport"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Persist the server config.
	installed, err := s.MCP.Install(ctx, cfg)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"install failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Connect to the MCP server and discover tools (best-effort).
	if err := s.MCPPool.Connect(ctx, *installed); err != nil {
		log.Printf("MCP install: connect to %s failed (best-effort): %v", installed.Name, err)
	}

	// Fetch tools from DB (they may have been cached by Connect).
	tools, err := s.MCP.ListTools(ctx, installed.ID)
	if err != nil {
		log.Printf("MCP install: list tools for %s failed: %v", installed.Name, err)
		tools = []mcp.ToolDef{}
	}
	if tools == nil {
		tools = []mcp.ToolDef{}
	}

	respondJSON(w, map[string]interface{}{
		"server": installed,
		"tools":  tools,
	})
}

// handleMCPList returns all registered MCP servers with their tools.
// GET /api/v1/mcp/servers
func (s *AdminServer) handleMCPList(w http.ResponseWriter, r *http.Request) {
	if s.MCP == nil {
		http.Error(w, `{"error":"MCP subsystem not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()

	servers, err := s.MCP.List(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"list servers failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if servers == nil {
		servers = []mcp.ServerConfig{}
	}

	// Build response with tools attached to each server.
	type serverWithTools struct {
		mcp.ServerConfig
		Tools []mcp.ToolDef `json:"tools"`
	}

	result := make([]serverWithTools, 0, len(servers))
	for _, srv := range servers {
		tools, err := s.MCP.ListTools(ctx, srv.ID)
		if err != nil {
			log.Printf("MCP list: failed to list tools for server %s: %v", srv.ID, err)
			tools = []mcp.ToolDef{}
		}
		if tools == nil {
			tools = []mcp.ToolDef{}
		}
		result = append(result, serverWithTools{
			ServerConfig: srv,
			Tools:        tools,
		})
	}

	respondJSON(w, result)
}

// handleMCPDelete removes an MCP server and disconnects the live client.
// DELETE /api/v1/mcp/servers/{id}
func (s *AdminServer) handleMCPDelete(w http.ResponseWriter, r *http.Request) {
	if s.MCP == nil || s.MCPPool == nil {
		http.Error(w, `{"error":"MCP subsystem not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	serverID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid server id: %s"}`, idStr), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Disconnect live client (best-effort — ignore "not found in pool" errors).
	if err := s.MCPPool.Disconnect(serverID); err != nil {
		log.Printf("MCP delete: disconnect %s (best-effort): %v", serverID, err)
	}

	// Delete from DB (CASCADE deletes tools).
	if err := s.MCP.Delete(ctx, serverID); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"delete failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{"status": "deleted"})
}

// handleMCPToolCall invokes a tool on a specific MCP server.
// POST /api/v1/mcp/servers/{id}/tools/{tool}/call
func (s *AdminServer) handleMCPToolCall(w http.ResponseWriter, r *http.Request) {
	if s.MCP == nil || s.MCPPool == nil {
		http.Error(w, `{"error":"MCP subsystem not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	idStr := r.PathValue("id")
	serverID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid server id: %s"}`, idStr), http.StatusBadRequest)
		return
	}

	toolName := r.PathValue("tool")
	if toolName == "" {
		http.Error(w, `{"error":"tool name is required"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON body: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	result, err := s.MCPPool.CallTool(ctx, serverID, toolName, body.Arguments)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"tool call failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}

	respondJSON(w, result)
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

	var req struct {
		Name string            `json:"name"`
		Env  map[string]string `json:"env,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON body: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	entry := s.MCPLibrary.FindByName(req.Name)
	if entry == nil {
		http.Error(w, fmt.Sprintf(`{"error":"server %q not found in library"}`, req.Name), http.StatusNotFound)
		return
	}

	cfg := entry.ToServerConfig(req.Env)
	ctx := r.Context()

	installed, err := s.MCP.Install(ctx, cfg)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"install failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Connect and discover tools (best-effort)
	if err := s.MCPPool.Connect(ctx, *installed); err != nil {
		log.Printf("MCP library install: connect to %s failed (best-effort): %v", installed.Name, err)
	}

	tools, err := s.MCP.ListTools(ctx, installed.ID)
	if err != nil {
		log.Printf("MCP library install: list tools for %s failed: %v", installed.Name, err)
		tools = []mcp.ToolDef{}
	}
	if tools == nil {
		tools = []mcp.ToolDef{}
	}

	respondJSON(w, map[string]interface{}{
		"server": installed,
		"tools":  tools,
	})
}

// handleMCPToolsList returns a flat list of all tools across all MCP servers.
// GET /api/v1/mcp/tools
func (s *AdminServer) handleMCPToolsList(w http.ResponseWriter, r *http.Request) {
	if s.MCP == nil {
		http.Error(w, `{"error":"MCP subsystem not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()

	tools, err := s.MCP.ListAllTools(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"list tools failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if tools == nil {
		tools = []mcp.ToolDef{}
	}

	respondJSON(w, tools)
}
