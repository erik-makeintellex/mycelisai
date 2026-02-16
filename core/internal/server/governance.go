package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mycelis/core/internal/governance"
)

// defaultPolicyPath is the disk location for persisting policy changes.
const defaultPolicyPath = "config/policy.yaml"

// pendingApprovalJSON is the simplified JSON representation of a pending approval request.
// It maps the complex proto ApprovalRequest into a cortex-friendly structure.
type pendingApprovalJSON struct {
	ID          string `json:"id"`
	Reason      string `json:"reason"`
	SourceAgent string `json:"source_agent"`
	TeamID      string `json:"team_id"`
	Intent      string `json:"intent"`
	Timestamp   string `json:"timestamp"`
	ExpiresAt   string `json:"expires_at"`
}

// handleGetPolicy returns the current governance policy configuration as JSON.
// GET /api/v1/governance/policy
func (s *AdminServer) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	if s.Guard == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Governance engine not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	cfg := s.Guard.GetPolicyConfig()
	respondJSON(w, cfg)
}

// handleUpdatePolicy replaces the entire governance policy configuration.
// It persists the new config to disk at the default policy path.
// PUT /api/v1/governance/policy
func (s *AdminServer) handleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	if s.Guard == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Governance engine not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	var cfg governance.PolicyConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// Validate: at least defaults must be set
	if cfg.Defaults.DefaultAction == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"defaults.default_action is required"}`, http.StatusBadRequest)
		return
	}

	// Update in-memory config
	s.Guard.UpdatePolicyConfig(&cfg)

	// Persist to disk
	if err := s.Guard.SavePolicyToFile(defaultPolicyPath); err != nil {
		log.Printf("Failed to persist policy to disk: %v", err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"policy updated in memory but failed to persist to disk"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("Governance policy updated and persisted to %s", defaultPolicyPath)
	respondJSON(w, map[string]string{"status": "updated"})
}

// handleGetPendingApprovals returns all pending approval requests in a simplified JSON format.
// GET /api/v1/governance/pending
func (s *AdminServer) handleGetPendingApprovals(w http.ResponseWriter, r *http.Request) {
	if s.Guard == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Governance engine not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	pending := s.Guard.ListPending()
	result := make([]pendingApprovalJSON, 0, len(pending))

	for _, req := range pending {
		item := pendingApprovalJSON{
			ID:     req.RequestId,
			Reason: req.Reason,
		}

		// Extract fields from the original message if present
		if msg := req.OriginalMessage; msg != nil {
			item.SourceAgent = msg.SourceAgentId
			item.TeamID = msg.TeamId
			if msg.GetEvent() != nil {
				item.Intent = msg.GetEvent().EventType
			}
			if msg.Timestamp != nil {
				item.Timestamp = msg.Timestamp.AsTime().Format(time.RFC3339)
			}
		}

		if req.ExpiresAt != nil {
			item.ExpiresAt = req.ExpiresAt.AsTime().Format(time.RFC3339)
		}

		result = append(result, item)
	}

	respondJSON(w, result)
}

// handleResolveApproval resolves a pending approval request by approving or rejecting it.
// POST /api/v1/governance/resolve/{id}
func (s *AdminServer) handleResolveApproval(w http.ResponseWriter, r *http.Request) {
	if s.Guard == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Governance engine not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	reqID := r.PathValue("id")
	if reqID == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"missing approval request ID"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		Action string `json:"action"` // "APPROVE" or "REJECT"
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if payload.Action != "APPROVE" && payload.Action != "REJECT" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"action must be APPROVE or REJECT"}`, http.StatusBadRequest)
		return
	}

	approved := payload.Action == "APPROVE"
	msg, err := s.Guard.Resolve(reqID, approved, "governance-api")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	// If approved and there is a message to re-inject, publish it back into the system
	if approved && msg != nil {
		if s.Router != nil {
			if err := s.Router.PublishDirect(msg); err != nil {
				log.Printf("Failed to re-publish approved message %s: %v", reqID, err)
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"resolved but failed to re-publish message"}`, http.StatusInternalServerError)
				return
			}
		}
	}

	respondJSON(w, map[string]string{
		"status":     "resolved",
		"request_id": reqID,
		"action":     payload.Action,
	})
}
