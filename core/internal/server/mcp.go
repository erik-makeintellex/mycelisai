package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	mcplib "github.com/mark3labs/mcp-go/mcp"
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
		respondAPIError(w, "list servers failed: "+redactedMCPErrorText(err), http.StatusInternalServerError)
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
// MCPA D1/D9: mcp_config:write first, then resolve, audit, disconnect, delete.
func (s *AdminServer) handleMCPDelete(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireMCPConfigWrite(w, r); !ok {
		return
	}
	if s.MCP == nil || s.MCPPool == nil {
		http.Error(w, `{"error":"MCP subsystem not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	serverID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondAPIError(w, "invalid server id", http.StatusBadRequest)
		return
	}
	s.deleteMCPServer(w, r, serverID)
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
		respondAPIError(w, "invalid server id: "+idStr, http.StatusBadRequest)
		return
	}

	toolName := r.PathValue("tool")
	if toolName == "" {
		http.Error(w, `{"error":"tool name is required"}`, http.StatusBadRequest)
		return
	}

	args, err := decodeMCPToolCallArguments(r.Body)
	if err != nil {
		respondAPIError(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// MCPS: resolve, authorize and (for high risk) audit before anything runs.
	r, call, ok := s.authorizeMCPToolCall(w, r, serverID, toolName, args)
	if !ok {
		return
	}
	serverName := call.ServerName
	args = normalizeMCPToolCallArgumentsForServer(serverName, args)

	result, err := s.MCPPool.CallTool(r.Context(), serverID, toolName, args)
	if err != nil {
		// MCPA: transport/JSON-RPC errors are redacted, capped and JSON-encoded.
		respondAPIError(w, "tool call failed: "+redactedMCPErrorText(err), http.StatusBadGateway)
		return
	}
	s.finishMCPToolCall(w, r, serverID, serverName, toolName, args, result)
}

// finishMCPToolCall records and returns a completed MCP call. A result the MCP
// server flagged isError is a tool failure: it is retained as a failed
// exchange item and answered with 502, never as completed output.
func (s *AdminServer) finishMCPToolCall(w http.ResponseWriter, r *http.Request, serverID uuid.UUID, serverName, toolName string, args map[string]any, result *mcplib.CallToolResult) {
	ctx := r.Context()
	toolErr := mcp.CallToolResultError(toolName, result)
	summary := fmt.Sprintf("%s returned output.", toolName)
	if toolErr != nil {
		summary = toolErr.Error()
	} else if text := strings.TrimSpace(extractMCPResultSummary(result)); text != "" {
		summary = text
	}
	var exchangeItemID string
	if s.Exchange != nil {
		// The raw isError result would repeat the unredacted error text.
		var retained any = map[string]any{"is_error": true, "error": summary}
		if toolErr == nil {
			// MCPS D7: the Exchange keeps a redacted copy; the caller gets the raw result.
			retained = redactMCPToolResult(result)
		}
		input := attributeMCPToolCallExchange(r, mcpToolCallExchangeInput(serverID, serverName, toolName, summary, toolErr != nil, args, retained))
		item, err := s.Exchange.PublishMCPResult(ctx, input)
		if err != nil {
			log.Printf("MCP tool call %s/%s not retained: %v", serverName, toolName, err)
		}
		if item != nil {
			exchangeItemID = item.ID.String()
		}
	}
	if toolErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(mcpToolCallFailureBody(toolErr, exchangeItemID))
		return
	}
	executionSummary := buildMCPToolCallExecutionSummary(serverName, toolName, summary, exchangeItemID)
	respondJSON(w, mcpToolCallResponse(result, executionSummary, exchangeItemID))
}

func mcpToolCallExchangeInput(serverID uuid.UUID, serverName, toolName, summary string, failed bool, args map[string]any, result any) exchange.MCPNormalizationInput {
	status := "completed"
	if failed {
		status = "failed"
	}
	return exchange.MCPNormalizationInput{
		ServerID:       serverID.String(),
		ServerName:     serverName,
		ToolName:       toolName,
		Summary:        summary,
		ResultPreview:  summary,
		TargetRole:     "soma",
		Status:         status,
		Result:         map[string]any{"arguments": redactMCPToolArguments(args), "result": result},
		RunClass:       string(protocol.ExecutionRunClassNoRun),
		NoRunReason:    "Direct MCP tool call did not supply a run id.",
		RetentionClass: string(protocol.ExecutionRetentionClassRetained),
	}
}

// mcpToolCallFailureBody carries only the redacted, capped error, never the
// raw MCP result, which would repeat the unredacted server text.
func mcpToolCallFailureBody(toolErr error, exchangeItemID string) map[string]any {
	body := map[string]any{"error": "tool call failed: " + toolErr.Error(), "is_error": true}
	if strings.TrimSpace(exchangeItemID) != "" {
		body["exchange_item_id"] = exchangeItemID
	}
	return body
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
		respondAPIError(w, "list tools failed: "+redactedMCPErrorText(err), http.StatusInternalServerError)
		return
	}
	if tools == nil {
		tools = []mcp.ToolDef{}
	}

	respondJSON(w, tools)
}
