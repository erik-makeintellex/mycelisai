package server

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mycelis/core/internal/exchange"
)

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
