package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
	fixtureScopeID, ok := s.qaFixtureScopeFromRequest(w, r)
	if !ok {
		return
	}
	var releaseFixtureFence func()
	if fixtureScopeID != "" {
		var lockErr error
		releaseFixtureFence, lockErr = acquireQAFixturePurgeLock(r.Context(), db, fixtureScopeID)
		if lockErr != nil {
			respondAPIError(w, "failed to lock execution cleanup ownership", http.StatusConflict)
			return
		}
		defer func() {
			if releaseFixtureFence != nil {
				releaseFixtureFence()
			}
		}()
		r = r.WithContext(withQAFixtureFenceHeld(r.Context(), fixtureScopeID))
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
	if fixtureScopeID != "" {
		if err := claimQAFixtureResourceTx(
			r.Context(), tx, fixtureScopeID, qaFixtureResource{Kind: "run", Ref: runID},
		); err != nil {
			respondAPIError(w, "failed to bind execution cleanup ownership", http.StatusConflict)
			return
		}
	}

	auditUser := auditActorIDFromRequest(r)
	actorIdentity := actorIdentitySnapshotFromRequest(r)
	configAction, configPlanErr := confirmedConfigMutationPlan(scope)
	if configPlanErr != nil {
		s.respondConfirmActionFailure(w, r, tx, proofID, contractID, runID, fixtureScopeID, auditUser, actorIdentity, configPlanErr)
		return
	}
	if confirmedActionNeedsAsyncDispatch(scope) {
		s.handleAsyncConfirmedAction(w, r, tx, proofID, contractID, runID, scope, fixtureScopeID, auditUser, actorIdentity)
		return
	}
	if configAction {
		if _, err := tx.ExecContext(r.Context(), "SAVEPOINT confirmed_config_action"); err != nil {
			respondAPIError(w, "failed to prepare atomic configuration action", http.StatusInternalServerError)
			return
		}
	}
	failAction := func(actionErr error) {
		if configAction {
			if _, rollbackErr := tx.ExecContext(r.Context(), "ROLLBACK TO SAVEPOINT confirmed_config_action"); rollbackErr != nil {
				log.Printf("CE-1: config-action rollback failed: %v", rollbackErr)
				respondAPIError(w, "configuration action rollback failed", http.StatusInternalServerError)
				return
			}
		}
		s.respondConfirmActionFailure(w, r, tx, proofID, contractID, runID, fixtureScopeID, auditUser, actorIdentity, actionErr)
	}
	var executionTx *sql.Tx
	if configAction {
		executionTx = tx
	}
	results, err := s.executePlannedToolCallsTx(r.Context(), executionTx, scope, auditUser, runID, proofID, contractID, fixtureScopeID, fixtureScopeID != "")
	if err != nil {
		failAction(err)
		return
	}
	if err := s.claimConfirmedConfigDocuments(r.Context(), tx, fixtureScopeID, results); err != nil {
		failAction(err)
		return
	}
	if configAction {
		if err := s.logConfigDocumentThreadReceiptsTx(r.Context(), tx, scope, results); err != nil {
			failAction(fmt.Errorf("persist configuration receipt: %w", err))
			return
		}
	}
	pendingTeamWork := confirmedActionHasPendingTeamWork(results)
	if !pendingTeamWork && !configAction {
		s.completeConfirmedActionWorkerRun(r.Context(), runID, results)
	}
	if !pendingTeamWork {
		if err := s.markRunCompletedTx(tx, runID, proofID); err != nil {
			log.Printf("CE-1: confirm-action run completion failed: %v", err)
			failAction(fmt.Errorf("finalize execution record: %w", err))
			return
		}
	}
	if err := s.confirmChatProofTx(tx, proofID); err != nil {
		log.Printf("CE-1: confirm-action proof update failed: %v", err)
		failAction(fmt.Errorf("confirm intent proof: %w", err))
		return
	}
	if err := tx.Commit(); err != nil {
		log.Printf("CE-1: confirm-action tx commit failed: %v", err)
		respondAPIError(w, "transaction commit failed", http.StatusInternalServerError)
		return
	}
	if !pendingTeamWork && configAction {
		s.completeConfirmedActionWorkerRun(r.Context(), runID, results)
	}
	if configAction {
		for _, result := range results {
			s.auditExecutedPlannedTool(protocol.PlannedToolCall{
				Name: result.Name, ToolRef: result.ToolRef, Arguments: result.Arguments,
			}, result.Name, auditUser)
		}
	}
	auditID := s.auditConfirmedAction(proofID, runID, scope, auditUser, actorIdentity, pendingTeamWork)
	proofArtifactID := ""
	if !pendingTeamWork {
		proofArtifactID = s.persistConfirmActionSuccessProof(r.Context(), proofID, contractID, runID, auditID, scope, results)
	}
	link := confirmedActionTeamWorkLink{
		ProofID:         proofID,
		ContractID:      contractID,
		ProofArtifactID: proofArtifactID,
		RunID:           runID,
		AuditID:         auditID,
		AuditUser:       auditUser,
		Scope:           scope,
		FixtureScopeID:  fixtureScopeID,
	}
	teamWorkRefs, outcomeProject, err := s.persistConfirmedActionVisibility(r.Context(), link, results)
	if err != nil {
		log.Printf("CE-1: confirm-action visibility persistence failed: %v", err)
	}
	if !isSynchronousConfigAction(results) {
		s.broadcastConfirmActionThreadEvent(runID, proofID, contractID, teamWorkRefs)
	}
	runStatus := runs.StatusCompleted
	if pendingTeamWork {
		runStatus = runs.StatusRunning
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(confirmActionResponseDataForStatus(proofID, contractID, proofArtifactID, runID, auditID, runStatus, scope, results, teamWorkRefs, outcomeProject)))
}

func confirmedConfigMutationPlan(scope *protocol.ScopeValidation) (bool, error) {
	if scope == nil || len(scope.PlannedToolCalls) == 0 {
		return false, nil
	}
	configCount := 0
	for _, call := range scope.PlannedToolCalls {
		if isConfigDocumentMutationTool(call.Name) {
			configCount++
		}
	}
	if configCount == 0 {
		return false, nil
	}
	if configCount != len(scope.PlannedToolCalls) {
		return false, fmt.Errorf("configuration changes must be approved separately from delegated or external work")
	}
	return true, nil
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

	runID, err := s.createExecutionRunTx(r.Context(), tx, proofID, scope, auditActorIDFromRequest(r))
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

func (s *AdminServer) respondConfirmActionFailure(w http.ResponseWriter, r *http.Request, tx *sql.Tx, proofID, contractID, runID, fixtureScopeID, auditUser string, actorIdentity map[string]any, err error) {
	s.failConfirmedActionWorkerRun(r.Context(), runID, err)
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
		FixtureScopeID:  fixtureScopeID,
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

func confirmedActionHasPendingTeamWork(results []plannedToolExecutionResult) bool {
	for _, result := range results {
		if isDelegateTool(result.Name) {
			return true
		}
	}
	return false
}
