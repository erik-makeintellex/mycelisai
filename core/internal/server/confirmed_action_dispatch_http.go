package server

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) handleAsyncConfirmedAction(w http.ResponseWriter, r *http.Request, tx *sql.Tx, proofID, contractID, runID string, scope *protocol.ScopeValidation, fixtureScopeID, auditUser string, actorIdentity map[string]any) {
	scope = correlateConfirmedActionScope(scope, runID, proofID, contractID)
	if err := persistCorrelatedScopeTx(r.Context(), tx, proofID, scope); err != nil {
		log.Printf("CE-1: persist correlated dispatch scope failed: %v", err)
		respondAPIError(w, "failed to preserve execution correlation", http.StatusInternalServerError)
		return
	}
	payload, idempotencyKey, err := s.stageConfirmedActionDispatchTx(r.Context(), tx, proofID, contractID, runID, scope, fixtureScopeID, auditUser, actorIdentity)
	if err != nil {
		log.Printf("CE-1: stage confirmed action dispatch failed: %v", err)
		respondAPIError(w, "failed to stage approved work", http.StatusInternalServerError)
		return
	}
	if err := s.confirmChatProofTx(tx, proofID); err != nil {
		log.Printf("CE-1: async confirm-action proof update failed: %v", err)
		respondAPIError(w, "failed to confirm intent proof", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		log.Printf("CE-1: async confirm-action tx commit failed: %v", err)
		respondAPIError(w, "transaction commit failed", http.StatusInternalServerError)
		return
	}

	payload.AuditID = s.auditConfirmedAction(proofID, runID, scope, auditUser, actorIdentity, true)
	teamID, workItemID := confirmedActionDispatchTargets(scope)
	item := &dispatchoutbox.Item{RunID: runID, IntentProofID: proofID, ContractID: contractID, TeamID: teamID, WorkItemID: workItemID}
	teamWorkRefs, outcomeProject, visibilityErr := s.ensureAsyncDispatchVisibility(r.Context(), item, payload)
	if visibilityErr != nil {
		log.Printf("CE-1: async confirm-action visibility persistence deferred to recovery: %v", visibilityErr)
	}
	if err := s.activateConfirmedActionDispatch(r.Context(), idempotencyKey, payload); err != nil {
		log.Printf("CE-1: async confirm-action activation deferred to staged recovery: %v", err)
	}
	s.broadcastConfirmActionThreadEvent(runID, proofID, contractID, teamWorkRefs)
	data := confirmActionResponseDataForStatus(proofID, contractID, "", runID, payload.AuditID, runs.StatusRunning, scope, nil, teamWorkRefs, outcomeProject)
	data["dispatch_state"] = dispatchoutbox.StatusPending
	respondAPIJSON(w, http.StatusAccepted, protocol.NewAPISuccess(data))
}
