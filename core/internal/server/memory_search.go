package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
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

	// 1. Embed the query text
	vec, err := s.Cognitive.Embed(r.Context(), query, "")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"embedding failed — no embed provider available"}`, http.StatusBadGateway)
		return
	}

	// 2. Semantic search
	results, err := s.Mem.SemanticSearch(r.Context(), vec, limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"search query failed"}`, http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{
		"query":   query,
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
// Returns the sensor library: configured zero-compute feeds + dynamic agents.
func (s *AdminServer) HandleSensors(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type SensorNode struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Status   string `json:"status"`
		LastSeen string `json:"last_seen"`
		Label    string `json:"label"`
	}

	now := time.Now().Format(time.RFC3339)

	// Base sensor library: configured zero-compute peripherals.
	// These are external data integrations, not swarm agents.
	sensors := []SensorNode{
		{ID: "sensor-gmail-inbox", Type: "email", Status: "online", LastSeen: now, Label: "Gmail Inbox"},
		{ID: "sensor-gmail-sent", Type: "email", Status: "online", LastSeen: now, Label: "Gmail Sent"},
		{ID: "sensor-weather-local", Type: "weather", Status: "online", LastSeen: now, Label: "Local Weather"},
		{ID: "sensor-weather-forecast", Type: "weather", Status: "degraded", LastSeen: now, Label: "5-Day Forecast"},
		{ID: "sensor-pg-primary", Type: "database", Status: "online", LastSeen: now, Label: "PostgreSQL Primary"},
		{ID: "sensor-nats-bus", Type: "messaging", Status: "online", LastSeen: now, Label: "NATS JetStream"},
		{ID: "sensor-ollama-health", Type: "llm", Status: "online", LastSeen: now, Label: "Ollama LLM Health"},
	}

	// Merge dynamic agents from recent log activity (if available)
	if s.Mem != nil {
		rows, err := s.Mem.ListRecent(20)
		if err == nil {
			seen := make(map[string]bool)
			for _, sn := range sensors {
				seen[sn.ID] = true
			}
			for _, entry := range rows {
				if entry.Source != "" && !seen[entry.Source] {
					seen[entry.Source] = true
					sensorType := entry.Intent
					if sensorType == "" {
						sensorType = "agent"
					}
					sensors = append(sensors, SensorNode{
						ID:       entry.Source,
						Type:     sensorType,
						Status:   "online",
						LastSeen: entry.Timestamp.Format(time.RFC3339),
						Label:    entry.Source,
					})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"sensors": sensors,
		"count":   len(sensors),
	})
}
