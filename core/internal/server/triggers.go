package server

// Trigger Rules API — declarative IF/THEN rules evaluated on event ingest.
// Default mode is "propose" (human approval required). "auto_execute" requires
// explicit policy. All rules enforce cooldown, recursion depth, and concurrency guards.
//
// Endpoints:
//   GET    /api/v1/triggers              — list all rules
//   POST   /api/v1/triggers              — create a rule
//   PUT    /api/v1/triggers/{id}         — update a rule
//   DELETE /api/v1/triggers/{id}         — delete a rule
//   POST   /api/v1/triggers/{id}/toggle  — activate/deactivate
//   GET    /api/v1/triggers/{id}/history — execution history

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/triggers"
	"github.com/mycelis/core/pkg/protocol"
)

// triggerRuleRequest is the JSON body for create/update.
type triggerRuleRequest struct {
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	EventPattern    string          `json:"event_pattern"`
	Condition       json.RawMessage `json:"condition"`
	TargetMissionID string          `json:"target_mission_id"`
	Mode            string          `json:"mode"`
	CooldownSeconds int             `json:"cooldown_seconds"`
	MaxDepth        int             `json:"max_depth"`
	MaxActiveRuns   int             `json:"max_active_runs"`
	IsActive        bool            `json:"is_active"`
}

// GET /api/v1/triggers
func (s *AdminServer) HandleListTriggers(w http.ResponseWriter, r *http.Request) {
	if s.Triggers == nil {
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess([]triggers.TriggerRule{}))
		return
	}

	rules, err := s.Triggers.ListAll(r.Context())
	if err != nil {
		log.Printf("HandleListTriggers: %v", err)
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(rules))
}

// POST /api/v1/triggers
func (s *AdminServer) HandleCreateTrigger(w http.ResponseWriter, r *http.Request) {
	if s.Triggers == nil {
		respondAPIError(w, "Trigger engine not initialized", http.StatusServiceUnavailable)
		return
	}

	var req triggerRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		respondAPIError(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.EventPattern == "" {
		respondAPIError(w, "event_pattern is required", http.StatusBadRequest)
		return
	}
	if req.TargetMissionID == "" {
		respondAPIError(w, "target_mission_id is required", http.StatusBadRequest)
		return
	}

	// Enforce propose as default — auto_execute must be explicit
	if req.Mode != "auto_execute" {
		req.Mode = "propose"
	}

	rule := &triggers.TriggerRule{
		Name:            req.Name,
		Description:     req.Description,
		EventPattern:    req.EventPattern,
		Condition:       req.Condition,
		TargetMissionID: req.TargetMissionID,
		Mode:            req.Mode,
		CooldownSeconds: req.CooldownSeconds,
		MaxDepth:        req.MaxDepth,
		MaxActiveRuns:   req.MaxActiveRuns,
		IsActive:        req.IsActive,
	}

	if err := s.Triggers.Create(r.Context(), rule); err != nil {
		log.Printf("HandleCreateTrigger: %v", err)
		respondAPIError(w, "Failed to create trigger: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Reload engine cache
	if s.TriggerEngine != nil {
		s.TriggerEngine.ReloadRules(r.Context())
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(rule))
}

// PUT /api/v1/triggers/{id}
func (s *AdminServer) HandleUpdateTrigger(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing trigger ID", http.StatusBadRequest)
		return
	}
	if s.Triggers == nil {
		respondAPIError(w, "Trigger engine not initialized", http.StatusServiceUnavailable)
		return
	}

	var req triggerRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if req.Mode != "auto_execute" {
		req.Mode = "propose"
	}

	rule := &triggers.TriggerRule{
		ID:              id,
		Name:            req.Name,
		Description:     req.Description,
		EventPattern:    req.EventPattern,
		Condition:       req.Condition,
		TargetMissionID: req.TargetMissionID,
		Mode:            req.Mode,
		CooldownSeconds: req.CooldownSeconds,
		MaxDepth:        req.MaxDepth,
		MaxActiveRuns:   req.MaxActiveRuns,
		IsActive:        req.IsActive,
	}

	if err := s.Triggers.Update(r.Context(), rule); err != nil {
		log.Printf("HandleUpdateTrigger: %v", err)
		respondAPIError(w, "Failed to update trigger: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Reload engine cache
	if s.TriggerEngine != nil {
		s.TriggerEngine.ReloadRules(r.Context())
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{"id": id, "updated": true}))
}

// DELETE /api/v1/triggers/{id}
func (s *AdminServer) HandleDeleteTrigger(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing trigger ID", http.StatusBadRequest)
		return
	}
	if s.Triggers == nil {
		respondAPIError(w, "Trigger engine not initialized", http.StatusServiceUnavailable)
		return
	}

	if err := s.Triggers.Delete(r.Context(), id); err != nil {
		log.Printf("HandleDeleteTrigger: %v", err)
		respondAPIError(w, "Failed to delete trigger: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Reload engine cache
	if s.TriggerEngine != nil {
		s.TriggerEngine.ReloadRules(r.Context())
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{"id": id, "deleted": true}))
}

// POST /api/v1/triggers/{id}/toggle
func (s *AdminServer) HandleToggleTrigger(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing trigger ID", http.StatusBadRequest)
		return
	}
	if s.Triggers == nil {
		respondAPIError(w, "Trigger engine not initialized", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if err := s.Triggers.SetActive(r.Context(), id, req.IsActive); err != nil {
		log.Printf("HandleToggleTrigger: %v", err)
		respondAPIError(w, "Failed to toggle trigger: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Reload engine cache
	if s.TriggerEngine != nil {
		s.TriggerEngine.ReloadRules(r.Context())
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{"id": id, "is_active": req.IsActive}))
}

// GET /api/v1/triggers/{id}/history
func (s *AdminServer) HandleTriggerHistory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondAPIError(w, "Missing trigger ID", http.StatusBadRequest)
		return
	}
	if s.Triggers == nil {
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess([]triggers.TriggerExecution{}))
		return
	}

	execs, err := s.Triggers.ListExecutions(r.Context(), id, 50)
	if err != nil {
		log.Printf("HandleTriggerHistory: %v", err)
		respondAPIError(w, "Database error", http.StatusInternalServerError)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(execs))
}
