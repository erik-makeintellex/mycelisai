package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/mcp"
)

type mcpLibraryRequest struct {
	Name              string               `json:"name"`
	Env               map[string]string    `json:"env,omitempty"`
	GovernanceContext mcpGovernanceContext `json:"governance_context,omitempty"`
}

type mcpActivityEntry struct {
	ID          string `json:"id"`
	ServerID    string `json:"server_id,omitempty"`
	ServerName  string `json:"server_name"`
	ToolName    string `json:"tool_name"`
	State       string `json:"state"`
	Summary     string `json:"summary"`
	Message     string `json:"message"`
	ChannelName string `json:"channel_name"`
	RunID       string `json:"run_id,omitempty"`
	TeamID      string `json:"team_id,omitempty"`
	AgentID     string `json:"agent_id,omitempty"`
	Timestamp   string `json:"timestamp"`
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

// handleMCPActivity returns persisted MCP activity across managed exchange channels.
// GET /api/v1/mcp/activity
func (s *AdminServer) handleMCPActivity(w http.ResponseWriter, r *http.Request) {
	if s.Exchange == nil {
		http.Error(w, `{"error":"exchange service not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	limit := parsePositiveInt(r.URL.Query().Get("limit"), 20)
	r = exchangeContext(r)

	channels := []string{"browser.research.results", "media.image.output", "api.data.output"}
	entries := make([]mcpActivityEntry, 0, limit)
	seen := make(map[string]struct{})
	for _, channel := range channels {
		items, err := s.Exchange.ListItems(r.Context(), channel, nil, limit)
		if err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, item := range items {
			entry, ok := normalizeMCPActivityItem(item)
			if !ok {
				continue
			}
			if _, exists := seen[entry.ID]; exists {
				continue
			}
			seen[entry.ID] = struct{}{}
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp > entries[j].Timestamp
	})
	if len(entries) > limit {
		entries = entries[:limit]
	}

	respondJSON(w, map[string]any{"ok": true, "data": entries})
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

func normalizeMCPActivityItem(item exchange.ExchangeItem) (mcpActivityEntry, bool) {
	payload := decodeJSONMap(item.Payload)
	metadata := decodeJSONMap(item.Metadata)
	mcpMeta := nestedMap(metadata, "mcp")

	sourceRole := lookupString([]map[string]any{payload, metadata}, "source_role")
	if sourceRole == "" {
		sourceRole = item.SourceRole
	}
	isMCP := strings.HasPrefix(strings.ToLower(strings.TrimSpace(sourceRole)), "mcp:")
	if !isMCP {
		if lookupString([]map[string]any{metadata, payload}, "source_kind") == "mcp" {
			isMCP = true
		}
		if lookupString([]map[string]any{mcpMeta, payload}, "server_name") != "" || lookupString([]map[string]any{mcpMeta, payload}, "server_id") != "" {
			isMCP = true
		}
	}
	if !isMCP {
		return mcpActivityEntry{}, false
	}

	serverID := lookupString([]map[string]any{payload, mcpMeta, metadata}, "server_id")
	serverName := lookupString([]map[string]any{payload, mcpMeta, metadata}, "server_name")
	if serverName == "" && strings.HasPrefix(sourceRole, "mcp:") {
		serverName = strings.TrimPrefix(sourceRole, "mcp:")
	}
	if serverName == "" {
		serverName = serverID
	}
	if serverName == "" {
		serverName = "mcp"
	}

	toolName := lookupString([]map[string]any{payload, mcpMeta, metadata}, "tool_name")
	if toolName == "" {
		toolName = lookupString([]map[string]any{payload, metadata}, "tool")
	}
	if toolName == "" {
		toolName = "unknown_tool"
	}

	summary := strings.TrimSpace(lookupString([]map[string]any{payload, metadata}, "summary"))
	if summary == "" {
		summary = strings.TrimSpace(item.Summary)
	}
	state := lookupString([]map[string]any{payload, mcpMeta, metadata}, "state")
	if state == "" {
		state = lookupString([]map[string]any{payload, metadata}, "status")
	}
	if state == "" {
		state = "completed"
	}
	message := lookupString([]map[string]any{payload, metadata}, "result_preview")
	if message == "" {
		message = lookupString([]map[string]any{payload, metadata}, "message")
	}
	if message == "" {
		message = summary
	}
	if message == "" {
		message = "MCP activity recorded."
	}
	timestamp := item.CreatedAt.UTC().Format(time.RFC3339)
	if raw := lookupString([]map[string]any{payload, metadata}, "created_at"); raw != "" {
		timestamp = raw
	}

	return mcpActivityEntry{
		ID:          item.ID.String(),
		ServerID:    serverID,
		ServerName:  serverName,
		ToolName:    toolName,
		State:       state,
		Summary:     summary,
		Message:     message,
		ChannelName: item.ChannelName,
		RunID:       lookupString([]map[string]any{payload, mcpMeta, metadata}, "run_id", "continuity_key"),
		TeamID:      lookupString([]map[string]any{payload, mcpMeta, metadata}, "source_team"),
		AgentID:     lookupString([]map[string]any{payload, mcpMeta, metadata}, "agent_id"),
		Timestamp:   timestamp,
	}, true
}

func decodeJSONMap(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func nestedMap(source map[string]any, key string) map[string]any {
	if source == nil {
		return map[string]any{}
	}
	value, ok := source[key]
	if !ok {
		return map[string]any{}
	}
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return map[string]any{}
}

func lookupString(sources []map[string]any, keys ...string) string {
	for _, key := range keys {
		for _, source := range sources {
			if source == nil {
				continue
			}
			if value, ok := source[key].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
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
			"status":            "requires_approval",
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

	if err := s.MCPPool.Connect(ctx, *installed); err != nil {
		log.Printf("MCP library apply: connect to %s failed (best-effort): %v", installed.Name, err)
	}

	tools, err := s.MCP.ListTools(ctx, installed.ID)
	if err != nil {
		log.Printf("MCP library apply: list tools for %s failed: %v", installed.Name, err)
		tools = []mcp.ToolDef{}
	}
	if tools == nil {
		tools = []mcp.ToolDef{}
	}

	respondJSON(w, map[string]any{
		"status":            "installed",
		"requires_approval": false,
		"server":            installed,
		"tools":             tools,
		"inspection":        inspection,
		"governance":        inspection["governance"],
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
