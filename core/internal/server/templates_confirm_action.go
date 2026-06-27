package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

	proofID, contractID, scope, runID, ok := s.prepareConfirmedAction(w, r, tx, req.ConfirmToken)
	if !ok {
		return
	}

	auditUser := auditUserLabelFromRequest(r)
	actorIdentity := actorIdentitySnapshotFromRequest(r)
	results, err := s.executePlannedToolCalls(r.Context(), scope, auditUser, runID, proofID, contractID)
	if err != nil {
		s.respondConfirmActionFailure(w, r, tx, proofID, contractID, runID, auditUser, actorIdentity, err)
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

	auditID := s.auditConfirmedAction(proofID, runID, scope, auditUser, actorIdentity)
	proofArtifactID := s.persistConfirmActionSuccessProof(r.Context(), proofID, contractID, runID, auditID, scope, results)
	link := confirmedActionTeamWorkLink{
		ProofID:         proofID,
		ContractID:      contractID,
		ProofArtifactID: proofArtifactID,
		RunID:           runID,
		AuditID:         auditID,
		AuditUser:       auditUser,
		Scope:           scope,
	}
	teamWorkRefs, outcomeProject, err := s.persistConfirmedActionVisibility(r.Context(), link, results)
	if err != nil {
		log.Printf("CE-1: confirm-action visibility persistence failed: %v", err)
	}
	s.broadcastConfirmActionThreadEvent(runID, proofID, contractID, teamWorkRefs)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(confirmActionResponseData(proofID, contractID, proofArtifactID, runID, auditID, scope, results, teamWorkRefs, outcomeProject)))
}

func (s *AdminServer) prepareConfirmedAction(w http.ResponseWriter, r *http.Request, tx *sql.Tx, token string) (string, string, *protocol.ScopeValidation, string, bool) {
	proofID, err := s.consumeConfirmTokenTx(tx, token)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return "", "", nil, "", false
	}

	scope, err := s.loadIntentProofScopeTx(tx, proofID)
	if err != nil {
		log.Printf("CE-1: confirm-action scope load failed: %v", err)
		respondAPIError(w, "failed to load approved execution plan", http.StatusInternalServerError)
		return "", "", nil, "", false
	}

	runID, err := s.createExecutionRunTx(tx, proofID)
	if err != nil {
		log.Printf("CE-1: confirm-action run creation failed: %v", err)
		respondAPIError(w, "failed to create execution record", http.StatusInternalServerError)
		return "", "", nil, "", false
	}

	contractID, err := s.ensureExecutionContractTx(r.Context(), tx, proofID, runID)
	if err != nil {
		log.Printf("CE-1: confirm-action contract persistence failed: %v", err)
		respondAPIError(w, "failed to create execution contract", http.StatusInternalServerError)
		return "", "", nil, "", false
	}

	return proofID, contractID, scope, runID, true
}

func (s *AdminServer) respondConfirmActionFailure(w http.ResponseWriter, r *http.Request, tx *sql.Tx, proofID, contractID, runID, auditUser string, actorIdentity map[string]any, err error) {
	if failErr := s.failChatProofTx(tx, proofID); failErr != nil {
		log.Printf("CE-1: confirm-action proof failure update failed: %v", failErr)
	}
	if runErr := s.markRunFailedTx(tx, runID, proofID, err.Error()); runErr != nil {
		log.Printf("CE-1: confirm-action failed run update failed: %v", runErr)
	}
	if commitErr := tx.Commit(); commitErr != nil {
		log.Printf("CE-1: confirm-action failure tx commit failed: %v", commitErr)
	}
	auditID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal, "confirm-action",
		"Chat proposal confirmation failed",
		withActorIdentity(map[string]any{
			"actor":           "Soma",
			"user":            auditUser,
			"action":          "proposal_confirmed",
			"result_status":   "failed",
			"run_id":          runID,
			"intent_proof_id": proofID,
			"approval_status": "failed",
			"approval_reason": err.Error(),
		}, actorIdentity),
	)
	proofArtifactID := s.persistConfirmActionFailureProof(r.Context(), proofID, contractID, runID, auditID, err)
	link := confirmedActionTeamWorkLink{
		ProofID:         proofID,
		ContractID:      contractID,
		ProofArtifactID: proofArtifactID,
		RunID:           runID,
		AuditID:         auditID,
		AuditUser:       auditUser,
	}
	if scope, scopeErr := s.loadIntentProofScopeForFailure(r.Context(), proofID); scopeErr == nil {
		link.Scope = scope
		if teamWorkErr := s.persistFailedConfirmedActionTeamWork(r.Context(), link, err); teamWorkErr != nil {
			log.Printf("CE-1: failed confirm-action team-work persistence failed: %v", teamWorkErr)
		}
	}
	message := fmt.Sprintf("approved execution failed: %v", err)
	respondAPIJSON(w, http.StatusInternalServerError, protocol.APIResponse{
		OK:    false,
		Error: message,
		Data:  confirmActionFailureResponseData(proofID, contractID, proofArtifactID, runID, auditID, err),
	})
}

func (s *AdminServer) loadIntentProofScopeForFailure(ctx context.Context, proofID string) (*protocol.ScopeValidation, error) {
	db := s.getDB()
	if db == nil {
		return nil, errDBUnavailable
	}
	var scopeJSON []byte
	if err := db.QueryRowContext(ctx, `SELECT scope_validation FROM intent_proofs WHERE id = $1`, proofID).Scan(&scopeJSON); err != nil {
		return nil, err
	}
	scope := &protocol.ScopeValidation{}
	if len(scopeJSON) > 0 {
		if err := json.Unmarshal(scopeJSON, scope); err != nil {
			return nil, err
		}
	}
	return scope, nil
}

func (s *AdminServer) auditConfirmedAction(proofID, runID string, scope *protocol.ScopeValidation, auditUser string, actorIdentity map[string]any) string {
	auditID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal, "confirm-action",
		"Chat proposal confirmed and verified execution record created",
		withActorIdentity(map[string]any{
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
		}, actorIdentity),
	)
	_, _ = s.createAuditEvent(
		protocol.TemplateChatToProposal, "confirm-action",
		"Execution run completed for confirmed chat proposal",
		withActorIdentity(map[string]any{
			"actor":           "Soma",
			"user":            auditUser,
			"action":          "execution_run",
			"result_status":   "completed",
			"run_id":          runID,
			"intent_proof_id": proofID,
			"capability_used": strings.Join(scope.CapabilityIDs, ","),
		}, actorIdentity),
	)
	return auditID
}

func withActorIdentity(ctx map[string]any, actorIdentity map[string]any) map[string]any {
	if ctx == nil {
		ctx = map[string]any{}
	}
	if len(actorIdentity) > 0 {
		ctx["actor_identity"] = actorIdentity
	}
	return ctx
}
