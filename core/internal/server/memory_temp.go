package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
)

// /api/v1/memory/temp (MEM-LANES): a root-admin inspection route over the
// caller's own temp-memory rows. No UI, doc or automation calls it.
// GET    ?channel=<key>&limit=10    the caller's rows plus signal checkpoints
// POST   {channel, content, owner_agent_id?, ttl_minutes?, metadata?}
// DELETE ?channel=<key>             removes only the caller's rows
//
// It requires a root admin with memory:write. Rows are owned by the caller's
// user id and read with the temp read rule, so the route never exposes
// another user's checkpoints. Core is single-tenant here: the identity
// carries no tenant, and the route always uses the "default" tenant (no
// request field can choose one).
func (s *AdminServer) HandleTempMemory(w http.ResponseWriter, r *http.Request) {
	identity, ok := requireRootAdminScope(w, r, scopeMemoryWrite)
	if !ok {
		return
	}
	owner := strings.TrimSpace(identity.UserID)
	if owner == "" {
		respondBlocker(w, r, http.StatusForbidden, codeAdminRequired, "temp memory needs a signed-in user id", nil)
		return
	}
	if s.Mem == nil {
		respondAPIError(w, "Memory service offline", http.StatusServiceUnavailable)
		return
	}
	const tenant = "default"
	reader := memory.GovernedReader{UserID: owner}

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
		entries, err := s.Mem.GetTempMemory(r.Context(), tenant, channel, limit, reader)
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
		id, err := s.Mem.PutTempMemory(r.Context(), tenant, req.Channel, req.OwnerAgentID, req.Content, req.Metadata, req.TTLMinutes, owner)
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
		deleted, err := s.Mem.ClearTempMemory(r.Context(), tenant, channel, reader)
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
