package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/mcp"
)

type mcpLibraryRequest struct {
	Name              string               `json:"name"`
	Env               map[string]string    `json:"env,omitempty"`
	GovernanceContext mcpGovernanceContext `json:"governance_context,omitempty"`
}

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
	if s.Exchange != nil {
		summary := fmt.Sprintf("%s returned output.", toolName)
		if text := strings.TrimSpace(extractMCPResultSummary(result)); text != "" {
			summary = text
		}
		_, _ = s.Exchange.PublishMCPResult(ctx, exchange.MCPNormalizationInput{
			ServerName: serversNameOrFallback(s.MCP, ctx, serverID),
			ToolName:   toolName,
			Summary:    summary,
			TargetRole: "soma",
			Status:     "completed",
			Result:     map[string]any{"arguments": body.Arguments, "result": result},
		})
	}

	respondJSON(w, result)
}

func serversNameOrFallback(svc *mcp.Service, ctx context.Context, serverID uuid.UUID) string {
	if svc == nil {
		return serverID.String()
	}
	server, err := svc.Get(ctx, serverID)
	if err != nil || server == nil || strings.TrimSpace(server.Name) == "" {
		return serverID.String()
	}
	return server.Name
}

func extractMCPResultSummary(result any) string {
	switch typed := result.(type) {
	case map[string]any:
		for _, key := range []string{"summary", "message", "text"} {
			if value, ok := typed[key].(string); ok {
				return value
			}
		}
	case []any:
		return fmt.Sprintf("MCP tool returned %d result items.", len(typed))
	}
	return ""
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

	var req mcpLibraryRequest
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

	inspectCtx := normalizeMCPGovernanceContext(r, req.GovernanceContext)
	respondJSON(w, buildMCPLibraryInspectionReport(entry, inspectCtx))
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

	var req mcpLibraryRequest
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

	inspectCtx := normalizeMCPGovernanceContext(r, req.GovernanceContext)
	inspection := buildMCPLibraryInspectionReport(entry, inspectCtx)
	if decision, _ := inspection["decision"].(string); decision == "require_approval" {
		w.WriteHeader(http.StatusAccepted)
		respondJSON(w, map[string]any{
			"requires_approval": true,
			"inspection":        inspection,
		})
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
		"server":     installed,
		"tools":      tools,
		"governance": inspection["governance"],
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
