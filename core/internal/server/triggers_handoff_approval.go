package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/triggers"
	"github.com/mycelis/core/pkg/protocol"
)

type triggerHandoffApprovalRequest struct {
	Status string `json:"status"`
	Action string `json:"action"`
}

// POST /api/v1/triggers/{id}/history/{executionId}/approval
func (s *AdminServer) HandleScheduleHandoffApproval(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	executionID := r.PathValue("executionId")
	if id == "" || executionID == "" {
		respondAPIError(w, "Missing trigger or execution ID", http.StatusBadRequest)
		return
	}
	if s.Triggers == nil {
		respondAPIError(w, "Trigger engine not initialized", http.StatusServiceUnavailable)
		return
	}

	var req triggerHandoffApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	status := normalizeScheduleHandoffApprovalState(req.Status, req.Action)
	if status == "" {
		respondAPIError(w, "status must be approved, rejected, or cancelled", http.StatusBadRequest)
		return
	}

	exec, err := s.Triggers.TransitionScheduleHandoffApproval(r.Context(), id, executionID, status)
	if err != nil {
		switch {
		case errors.Is(err, triggers.ErrInvalidApprovalState):
			respondAPIError(w, "status must be approved, rejected, or cancelled", http.StatusBadRequest)
		case errors.Is(err, triggers.ErrExecutionNotFound):
			respondAPIError(w, "schedule handoff not found", http.StatusNotFound)
		case errors.Is(err, triggers.ErrApprovalTransitionConflict):
			respondAPIError(w, "schedule handoff is not awaiting approval", http.StatusConflict)
		default:
			log.Printf("HandleScheduleHandoffApproval: %v", err)
			respondAPIError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"id":              exec.ID,
		"rule_id":         exec.RuleID,
		"handoff_key":     exec.HandoffKey,
		"proposal_status": exec.ProposalStatus,
		"run_id":          exec.RunID,
		"execution":       exec,
	}))
}

func normalizeScheduleHandoffApprovalState(status, action string) string {
	raw := strings.ToLower(strings.TrimSpace(firstNonEmptyString(status, action)))
	switch raw {
	case "approve", "approved":
		return "approved"
	case "reject", "rejected", "deny", "denied":
		return "rejected"
	case "cancel", "cancelled", "canceled":
		return "cancelled"
	default:
		return ""
	}
}
