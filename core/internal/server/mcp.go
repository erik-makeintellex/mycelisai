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
			ServerConfig: redactMCPServerConfig(srv),
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
			ServerID:      serverID.String(),
			ServerName:    serversNameOrFallback(s.MCP, ctx, serverID),
			ToolName:      toolName,
			Summary:       summary,
			ResultPreview: summary,
			TargetRole:    "soma",
			Status:        "completed",
			Result:        map[string]any{"arguments": body.Arguments, "result": result},
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
