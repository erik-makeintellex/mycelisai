package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

// POST /api/v1/intent/confirm-action - confirm a chat-based proposal action.
func (s *AdminServer) HandleConfirmAction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConfirmToken string `json:"confirm_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if req.ConfirmToken == "" {
		respondAPIError(w, "Missing confirm_token", http.StatusBadRequest)
		return
	}

	db := s.getDB()
	if db == nil {
		respondAPIError(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	tx, err := db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("CE-1: confirm-action tx begin failed: %v", err)
		respondAPIError(w, "database transaction failed", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	proofID, scope, runID, ok := s.prepareConfirmedAction(w, r, tx, req.ConfirmToken)
	if !ok {
		return
	}

	auditUser := auditUserLabelFromRequest(r)
	if err := s.executePlannedToolCalls(r.Context(), scope, auditUser); err != nil {
		s.respondConfirmActionFailure(w, tx, proofID, runID, auditUser, err)
		return
	}

	if err := s.markRunCompletedTx(tx, runID, proofID); err != nil {
		log.Printf("CE-1: confirm-action run completion failed: %v", err)
		respondAPIError(w, "failed to finalize execution record", http.StatusInternalServerError)
		return
	}
	if err := s.confirmChatProofTx(tx, proofID); err != nil {
		log.Printf("CE-1: confirm-action proof update failed: %v", err)
		respondAPIError(w, "failed to confirm intent proof", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		log.Printf("CE-1: confirm-action tx commit failed: %v", err)
		respondAPIError(w, "transaction commit failed", http.StatusInternalServerError)
		return
	}

	auditID := s.auditConfirmedAction(proofID, runID, scope, auditUser)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"confirmed":       true,
		"verified":        true,
		"execution_state": "verified",
		"proof_id":        proofID,
		"audit_event_id":  auditID,
		"run_id":          runID,
		"run_status":      runs.StatusCompleted,
	}))
}

func (s *AdminServer) prepareConfirmedAction(w http.ResponseWriter, r *http.Request, tx *sql.Tx, token string) (string, *protocol.ScopeValidation, string, bool) {
	proofID, err := s.consumeConfirmTokenTx(tx, token)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return "", nil, "", false
	}

	scope, err := s.loadIntentProofScopeTx(tx, proofID)
	if err != nil {
		log.Printf("CE-1: confirm-action scope load failed: %v", err)
		respondAPIError(w, "failed to load approved execution plan", http.StatusInternalServerError)
		return "", nil, "", false
	}

	runID, err := s.createExecutionRunTx(tx, proofID)
	if err != nil {
		log.Printf("CE-1: confirm-action run creation failed: %v", err)
		respondAPIError(w, "failed to create execution record", http.StatusInternalServerError)
		return "", nil, "", false
	}

	return proofID, scope, runID, true
}

func (s *AdminServer) respondConfirmActionFailure(w http.ResponseWriter, tx *sql.Tx, proofID, runID, auditUser string, err error) {
	if failErr := s.failChatProofTx(tx, proofID); failErr != nil {
		log.Printf("CE-1: confirm-action proof failure update failed: %v", failErr)
	}
	if runErr := s.markRunFailedTx(tx, runID, proofID, err.Error()); runErr != nil {
		log.Printf("CE-1: confirm-action failed run update failed: %v", runErr)
	}
	if commitErr := tx.Commit(); commitErr != nil {
		log.Printf("CE-1: confirm-action failure tx commit failed: %v", commitErr)
	}
	_, _ = s.createAuditEvent(
		protocol.TemplateChatToProposal, "confirm-action",
		"Chat proposal confirmation failed",
		map[string]any{
			"actor":           "Soma",
			"user":            auditUser,
			"action":          "proposal_confirmed",
			"result_status":   "failed",
			"run_id":          runID,
			"intent_proof_id": proofID,
			"approval_status": "failed",
			"approval_reason": err.Error(),
		},
	)
	respondAPIError(w, fmt.Sprintf("approved execution failed: %v", err), http.StatusInternalServerError)
}

func (s *AdminServer) auditConfirmedAction(proofID, runID string, scope *protocol.ScopeValidation, auditUser string) string {
	auditID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal, "confirm-action",
		"Chat proposal confirmed and verified execution record created",
		map[string]any{
			"proof_id":        proofID,
			"run_id":          runID,
			"execution_state": "verified",
			"actor":           "Soma",
			"user":            auditUser,
			"action":          "proposal_confirmed",
			"result_status":   "confirmed",
			"approval_status": "confirmed",
			"intent_proof_id": proofID,
			"capability_used": strings.Join(scope.CapabilityIDs, ","),
		},
	)
	_, _ = s.createAuditEvent(
		protocol.TemplateChatToProposal, "confirm-action",
		"Execution run completed for confirmed chat proposal",
		map[string]any{
			"actor":           "Soma",
			"user":            auditUser,
			"action":          "execution_run",
			"result_status":   "completed",
			"run_id":          runID,
			"intent_proof_id": proofID,
			"capability_used": strings.Join(scope.CapabilityIDs, ","),
		},
	)
	return auditID
}
