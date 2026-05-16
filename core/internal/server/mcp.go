package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/pkg/protocol"
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

	args, err := decodeMCPToolCallArguments(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON body: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	serverName := serversNameOrFallback(s.MCP, ctx, serverID)
	args = normalizeMCPToolCallArgumentsForServer(serverName, args)

	result, err := s.MCPPool.CallTool(ctx, serverID, toolName, args)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"tool call failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	summary := fmt.Sprintf("%s returned output.", toolName)
	if text := strings.TrimSpace(extractMCPResultSummary(result)); text != "" {
		summary = text
	}
	var exchangeItemID string
	if s.Exchange != nil {
		item, _ := s.Exchange.PublishMCPResult(ctx, exchange.MCPNormalizationInput{
			ServerID:       serverID.String(),
			ServerName:     serverName,
			ToolName:       toolName,
			Summary:        summary,
			ResultPreview:  summary,
			TargetRole:     "soma",
			Status:         "completed",
			Result:         map[string]any{"arguments": args, "result": result},
			RunClass:       string(protocol.ExecutionRunClassNoRun),
			NoRunReason:    "Direct MCP tool call did not supply a run id.",
			RetentionClass: string(protocol.ExecutionRetentionClassRetained),
		})
		if item != nil {
			exchangeItemID = item.ID.String()
		}
	}

	executionSummary := buildMCPToolCallExecutionSummary(serverName, toolName, summary, exchangeItemID)
	respondJSON(w, mcpToolCallResponse(result, executionSummary, exchangeItemID))
}

func decodeMCPToolCallArguments(reader io.Reader) (map[string]any, error) {
	var body map[string]any
	if err := json.NewDecoder(reader).Decode(&body); err != nil {
		return nil, err
	}
	if body == nil {
		return map[string]any{}, nil
	}
	if raw, exists := body["arguments"]; exists {
		if raw == nil {
			return map[string]any{}, nil
		}
		args, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("arguments must be an object")
		}
		return args, nil
	}
	return body, nil
}

func normalizeMCPToolCallArgumentsForServer(serverName string, args map[string]any) map[string]any {
	if !strings.EqualFold(strings.TrimSpace(serverName), "filesystem") || len(args) == 0 {
		return args
	}
	for _, key := range []string{"path", "source", "destination"} {
		if raw, ok := args[key].(string); ok {
			args[key] = normalizeFilesystemMCPPath(raw)
		}
	}
	if rawPaths, ok := args["paths"].([]any); ok {
		paths := make([]any, 0, len(rawPaths))
		for _, raw := range rawPaths {
			if pathValue, ok := raw.(string); ok {
				paths = append(paths, normalizeFilesystemMCPPath(pathValue))
				continue
			}
			paths = append(paths, raw)
		}
		args["paths"] = paths
	}
	return args
}

func normalizeFilesystemMCPPath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw
	}
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	normalized = strings.TrimPrefix(normalized, "./")

	var rel string
	switch {
	case normalized == "workspace" || normalized == "/workspace":
		rel = ""
	case strings.HasPrefix(normalized, "workspace/"):
		rel = strings.TrimPrefix(normalized, "workspace/")
	case strings.HasPrefix(normalized, "/workspace/"):
		rel = strings.TrimPrefix(normalized, "/workspace/")
	default:
		return raw
	}

	root := strings.TrimSpace(mcp.ResolveFilesystemWorkspaceRoot())
	if root == "" {
		return raw
	}
	if rel == "" {
		return root
	}
	return filepath.Join(root, filepath.FromSlash(rel))
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

func mcpToolCallResponse(result any, summary *protocol.ExecutionSummary, exchangeItemID string) any {
	if raw, err := json.Marshal(result); err == nil {
		var object map[string]any
		if err := json.Unmarshal(raw, &object); err == nil && object != nil {
			object["execution_summary"] = summary
			if strings.TrimSpace(exchangeItemID) != "" {
				object["exchange_item_id"] = exchangeItemID
			}
			return object
		}
	}
	response := map[string]any{
		"result":            result,
		"execution_summary": summary,
	}
	if strings.TrimSpace(exchangeItemID) != "" {
		response["exchange_item_id"] = exchangeItemID
	}
	return response
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
