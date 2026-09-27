package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/swarm"
)

// GET /api/v1/memory/search?q=<text>&limit=5
// Embeds the query text, then performs cosine similarity search against context_vectors.
func (s *AdminServer) HandleMemorySearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Cognitive == nil {
		http.Error(w, `{"error":"Cognitive engine offline — cannot embed query"}`, http.StatusServiceUnavailable)
		return
	}

	if s.Mem == nil {
		http.Error(w, `{"error":"Memory service offline"}`, http.StatusServiceUnavailable)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error":"query parameter 'q' is required"}`, http.StatusBadRequest)
		return
	}

	limit := 5
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}
	teamID := strings.TrimSpace(r.URL.Query().Get("team_id"))
	agentID := strings.TrimSpace(r.URL.Query().Get("agent_id"))
	runID := strings.TrimSpace(r.URL.Query().Get("run_id"))
	visibility := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("visibility")))
	rawTypes := strings.Split(strings.TrimSpace(r.URL.Query().Get("types")), ",")
	searchTypes := make([]string, 0, len(rawTypes))
	for _, raw := range rawTypes {
		value := strings.TrimSpace(raw)
		if value != "" {
			searchTypes = append(searchTypes, value)
		}
	}

	// 1. Embed the query text
	vec, err := s.Cognitive.Embed(r.Context(), query, "")
	if err != nil {
		respondJSON(w, map[string]any{
			"query": query,
			"scope": map[string]any{
				"tenant_id":  "default",
				"team_id":    teamID,
				"agent_id":   agentID,
				"run_id":     runID,
				"visibility": visibility,
				"types":      searchTypes,
			},
			"results": []memory.VectorResult{},
			"count":   0,
			"degraded": map[string]any{
				"code":               "embedding_unavailable",
				"summary":            "Semantic memory search is unavailable because no embedding provider is available.",
				"recommended_action": "Configure an embedding-capable AI engine before relying on vector memory recall.",
			},
		})
		return
	}

	// 2. Semantic search
	results, err := s.Mem.SemanticSearchWithOptions(r.Context(), vec, memory.SemanticSearchOptions{
		Limit:               limit,
		TenantID:            "default",
		TeamID:              teamID,
		AgentID:             agentID,
		RunID:               runID,
		Visibility:          visibility,
		Types:               searchTypes,
		AllowGlobal:         true,
		AllowLegacyUnscoped: teamID == "" && agentID == "",
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"search query failed"}`, http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{
		"query": query,
		"scope": map[string]any{
			"tenant_id":  "default",
			"team_id":    teamID,
			"agent_id":   agentID,
			"run_id":     runID,
			"visibility": visibility,
			"types":      searchTypes,
		},
		"results": results,
		"count":   len(results),
	})
}

// GET /api/v1/memory/sitreps?team_id=<uuid>&limit=10
// Returns recent SitReps for a team.
func (s *AdminServer) HandleListSitReps(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Mem == nil {
		http.Error(w, `{"error":"Memory service offline"}`, http.StatusServiceUnavailable)
		return
	}

	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		teamID = "22222222-2222-2222-2222-222222222222" // Default team
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	sitreps, err := s.Mem.ListSitReps(r.Context(), teamID, limit)
	if err != nil {
		http.Error(w, `{"error":"failed to retrieve sitreps"}`, http.StatusInternalServerError)
		return
	}

	if sitreps == nil {
		sitreps = []map[string]any{}
	}

	respondJSON(w, map[string]any{
		"team_id": teamID,
		"sitreps": sitreps,
		"count":   len(sitreps),
	})
}

// GET /api/v1/sensors
// Lists only running endpoint-backed SensorAgents with their real probe state.
// No runtime or no configured sensors yields an empty list and an honest status.
func (s *AdminServer) HandleSensors(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	nodes, status := []SensorNode{}, sensorsStatusRuntimeUnavailable
	if s.Soma != nil {
		nodes = sensorNodesFromSnapshots(s.Soma.ListSensors())
		status = sensorsStatusOK
		if len(nodes) == 0 {
			status = sensorsStatusNoneConfigured
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"sensors": nodes, "count": len(nodes), "status": status})
}

// Top-level /api/v1/sensors status (PH-D F10). Sensors are listed only from
// running endpoint-backed SensorAgents with their real probe state.
const (
	sensorsStatusOK                 = "ok"
	sensorsStatusNoneConfigured     = "none_configured"
	sensorsStatusRuntimeUnavailable = "runtime_unavailable"
)

// SensorNode is one row of GET /api/v1/sensors.
type SensorNode struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	LastSeen string `json:"last_seen"`
	Label    string `json:"label"`
	TeamID   string `json:"team_id,omitempty"`
}

// sensorNodesFromSnapshots maps probed state to rows. last_seen is the last
// successful probe, or empty when the sensor has never answered.
func sensorNodesFromSnapshots(snaps []swarm.SensorSnapshot) []SensorNode {
	nodes := make([]SensorNode, 0, len(snaps))
	for _, snap := range snaps {
		node := SensorNode{ID: snap.ID, Type: snap.Role, Status: snap.Status, Label: snap.ID, TeamID: snap.TeamID}
		if !snap.LastSuccessAt.IsZero() {
			node.LastSeen = snap.LastSuccessAt.UTC().Format(time.RFC3339)
		}
		nodes = append(nodes, node)
	}
	return nodes
}
