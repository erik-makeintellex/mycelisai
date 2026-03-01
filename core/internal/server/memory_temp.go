package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mycelis/core/pkg/protocol"
)

// /api/v1/memory/temp
// GET    ?channel=<key>&limit=10
// POST   {channel, content, owner_agent_id?, ttl_minutes?, metadata?}
// DELETE ?channel=<key>
func (s *AdminServer) HandleTempMemory(w http.ResponseWriter, r *http.Request) {
	if s.Mem == nil {
		respondAPIError(w, "Memory service offline", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		channel := r.URL.Query().Get("channel")
		if channel == "" {
			respondAPIError(w, "channel query parameter is required", http.StatusBadRequest)
			return
		}
		limit := 10
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}
		entries, err := s.Mem.GetTempMemory(r.Context(), "default", channel, limit)
		if err != nil {
			respondAPIError(w, "Failed to read temp memory: "+err.Error(), http.StatusInternalServerError)
			return
		}
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
			"channel": channel,
			"entries": entries,
			"count":   len(entries),
		}))
		return

	case http.MethodPost:
		var req struct {
			Channel      string         `json:"channel"`
			Content      string         `json:"content"`
			OwnerAgentID string         `json:"owner_agent_id"`
			TTLMinutes   int            `json:"ttl_minutes"`
			Metadata     map[string]any `json:"metadata"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.Channel == "" || req.Content == "" {
			respondAPIError(w, "channel and content are required", http.StatusBadRequest)
			return
		}
		id, err := s.Mem.PutTempMemory(r.Context(), "default", req.Channel, req.OwnerAgentID, req.Content, req.Metadata, req.TTLMinutes)
		if err != nil {
			respondAPIError(w, "Failed to write temp memory: "+err.Error(), http.StatusInternalServerError)
			return
		}
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
			"id":      id,
			"channel": req.Channel,
			"status":  "stored",
		}))
		return

	case http.MethodDelete:
		channel := r.URL.Query().Get("channel")
		if channel == "" {
			respondAPIError(w, "channel query parameter is required", http.StatusBadRequest)
			return
		}
		deleted, err := s.Mem.ClearTempMemory(r.Context(), "default", channel)
		if err != nil {
			respondAPIError(w, "Failed to clear temp memory: "+err.Error(), http.StatusInternalServerError)
			return
		}
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
			"channel": channel,
			"deleted": deleted,
		}))
		return
	}

	respondAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

