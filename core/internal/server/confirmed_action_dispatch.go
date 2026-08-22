package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	confirmedActionDispatchKind = "confirmed_action_team_plan"
	confirmedActionMaxAttempts  = 3
)

type confirmedActionDispatchPayload struct {
	AuditUser      string                    `json:"audit_user"`
	ActorIdentity  map[string]any            `json:"actor_identity,omitempty"`
	AuditID        string                    `json:"audit_id,omitempty"`
	Scope          *protocol.ScopeValidation `json:"scope"`
	FixtureScopeID string                    `json:"fixture_scope_id,omitempty"`
}

func confirmedActionNeedsAsyncDispatch(scope *protocol.ScopeValidation) bool {
	if scope == nil {
		return false
	}
	for _, planned := range scope.PlannedToolCalls {
		if strings.TrimSpace(planned.Name) == "create_team" || isDelegateTool(planned.Name) {
			return true
		}
	}
	return false
}

func correlateConfirmedActionScope(scope *protocol.ScopeValidation, runID, proofID, contractID string) *protocol.ScopeValidation {
	if scope == nil {
		return nil
	}
	copyScope := *scope
	copyScope.PlannedToolCalls = make([]protocol.PlannedToolCall, 0, len(scope.PlannedToolCalls))
	for _, planned := range scope.PlannedToolCalls {
		planned = normalizePlannedToolCall(planned)
		copyScope.PlannedToolCalls = append(copyScope.PlannedToolCalls,
			annotateConfirmedDelegationCall(planned, runID, proofID, contractID, &copyScope))
	}
	return &copyScope
}

func (s *AdminServer) stageConfirmedActionDispatchTx(ctx context.Context, tx *sql.Tx, proofID, contractID, runID string, scope *protocol.ScopeValidation, fixtureScopeID, auditUser string, actorIdentity map[string]any) (confirmedActionDispatchPayload, string, error) {
	if s.DispatchOutbox == nil {
		return confirmedActionDispatchPayload{}, "", dispatchoutbox.ErrUnavailable
	}
	payload := confirmedActionDispatchPayload{
		AuditUser: auditUser, ActorIdentity: actorIdentity, Scope: scope, FixtureScopeID: fixtureScopeID,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return payload, "", err
	}
	idempotencyKey := "confirm-action:" + proofID
	teamID, workItemID := confirmedActionDispatchTargets(scope)
	_, err = s.DispatchOutbox.EnqueueTx(ctx, tx, dispatchoutbox.Item{
		ID: uuid.NewString(), IdempotencyKey: idempotencyKey,
		DispatchKind: confirmedActionDispatchKind, RunID: runID,
		IntentProofID: proofID, ContractID: contractID, TeamID: teamID, WorkItemID: workItemID,
		SourceKind: string(protocol.SourceKindWebAPI), SourceChannel: "api.intent.confirm-action",
		PayloadKind: string(protocol.PayloadKindCommand), Payload: raw,
		Recovery: json.RawMessage(`{"action":"retry_committed_dispatch","operator_required":false}`),
	})
	return payload, idempotencyKey, err
}

func confirmedActionDispatchTargets(scope *protocol.ScopeValidation) (string, string) {
	if scope == nil {
		return "", ""
	}
	for _, planned := range scope.PlannedToolCalls {
		if isDelegateTool(planned.Name) {
			return firstNonEmptyString(confirmedActionTeamID(planned.Arguments), confirmedActionCreatedTeamIDFromScope(scope)), confirmedDelegationWorkItemID(planned.Arguments)
		}
	}
	return "", ""
}

func (s *AdminServer) activateConfirmedActionDispatch(ctx context.Context, key string, payload confirmedActionDispatchPayload) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.DispatchOutbox.UpdatePayloadAndActivate(ctx, key, raw)
}

func StartConfirmedActionDispatch(ctx context.Context, s *AdminServer) error {
	if s == nil || s.getDB() == nil || s.DispatchOutbox == nil {
		return fmt.Errorf("confirmed action dispatch requires database outbox")
	}
	go s.runConfirmedActionDispatch(ctx)
	return nil
}

func (s *AdminServer) runConfirmedActionDispatch(ctx context.Context) {
	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := s.dispatchOneConfirmedAction(ctx); err != nil && ctx.Err() == nil {
			log.Printf("confirmed action dispatch: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *AdminServer) dispatchOneConfirmedAction(ctx context.Context) error {
	item, err := s.DispatchOutbox.ClaimNext(ctx, 45*time.Second)
	if err != nil || item == nil {
		return err
	}
	switch item.DispatchKind {
	case confirmedActionDispatchKind:
		return s.dispatchClaimedConfirmedAction(ctx, item)
	case teamWorkValidationDispatchKind:
		return s.dispatchClaimedTeamWorkValidation(ctx, item)
	case teamWorkSteeringDispatchKind:
		return s.dispatchClaimedTeamWorkSteering(ctx, item)
	default:
		err := fmt.Errorf("unsupported dispatch kind %q", item.DispatchKind)
		_ = s.DispatchOutbox.MarkFailed(ctx, item.ID, err)
		return err
	}
}

func (s *AdminServer) dispatchClaimedConfirmedAction(ctx context.Context, item *dispatchoutbox.Item) error {
	var payload confirmedActionDispatchPayload
	if err := json.Unmarshal(item.Payload, &payload); err != nil {
		_ = s.DispatchOutbox.MarkFailed(ctx, item.ID, err)
		return fmt.Errorf("decode dispatch %s: %w", item.ID, err)
	}
	if payload.Scope == nil {
		err := fmt.Errorf("dispatch %s has no approved scope", item.ID)
		_ = s.DispatchOutbox.MarkFailed(ctx, item.ID, err)
		return err
	}
	return s.withQAFixtureScopeLock(ctx, payload.FixtureScopeID, func() error {
		lockedCtx := withQAFixtureFenceHeld(ctx, payload.FixtureScopeID)
		return s.dispatchClaimedConfirmedActionLocked(lockedCtx, item, payload)
	})
}

func (s *AdminServer) dispatchClaimedConfirmedActionLocked(ctx context.Context, item *dispatchoutbox.Item, payload confirmedActionDispatchPayload) error {
	if payload.AuditID == "" {
		payload.AuditID = s.auditConfirmedAction(item.IntentProofID, item.RunID, payload.Scope, payload.AuditUser, payload.ActorIdentity, true)
	}
	if _, _, err := s.ensureAsyncDispatchVisibility(ctx, item, payload); err != nil {
		return s.retryOrFailConfirmedDispatch(ctx, item, payload, err)
	}
	results, err := s.executePlannedToolCalls(ctx, payload.Scope, payload.AuditUser, item.RunID, item.IntentProofID, item.ContractID, payload.FixtureScopeID, payload.FixtureScopeID != "")
	if err != nil {
		return s.retryOrFailConfirmedDispatch(ctx, item, payload, err)
	}
	link := confirmedActionTeamWorkLink{ProofID: item.IntentProofID, ContractID: item.ContractID, RunID: item.RunID, AuditID: payload.AuditID, AuditUser: payload.AuditUser, Scope: payload.Scope, FixtureScopeID: payload.FixtureScopeID}
	if err := s.persistAsyncConfirmedActionResults(ctx, link, results); err != nil {
		return s.retryOrFailConfirmedDispatch(ctx, item, payload, err)
	}
	if !confirmedActionHasPendingTeamWork(results) {
		s.completeConfirmedActionWorkerRun(ctx, item.RunID, results)
		if err := s.completeAsyncDispatchRun(ctx, item); err != nil {
			return s.retryOrFailConfirmedDispatch(ctx, item, payload, err)
		}
	}
	return s.DispatchOutbox.MarkCompleted(ctx, item.ID)
}

func (s *AdminServer) persistAsyncConfirmedActionResults(ctx context.Context, link confirmedActionTeamWorkLink, results []plannedToolExecutionResult) error {
	if err := s.ensureGroupsForCreatedTeams(ctx, link.AuditID, link.AuditUser, link.Scope, link.FixtureScopeID, results); err != nil {
		return err
	}
	if err := s.persistConfirmedActionOutputArtifacts(ctx, link.RunID, link.FixtureScopeID, results); err != nil {
		return err
	}
	_, err := s.persistConfirmedCreateTeamItems(ctx, link, results)
	return err
}

func (s *AdminServer) ensureAsyncDispatchVisibility(ctx context.Context, item *dispatchoutbox.Item, payload confirmedActionDispatchPayload) ([]confirmActionTeamWorkRef, *protocol.OutcomeProject, error) {
	teamID, workItemID := confirmedActionDispatchTargets(payload.Scope)
	if teamID != "" && workItemID != "" {
		if existing, err := s.getTeamWorkItemDB(ctx, teamID, workItemID); err == nil {
			ref := confirmActionTeamWorkRefForItem(existing)
			link := confirmedActionTeamWorkLink{ProofID: item.IntentProofID, ContractID: item.ContractID, RunID: item.RunID, AuditID: payload.AuditID, AuditUser: payload.AuditUser, Scope: payload.Scope, FixtureScopeID: payload.FixtureScopeID}
			project, projectErr := s.ensureOutcomeOwnershipForConfirmedAction(ctx, link, []confirmActionTeamWorkRef{ref})
			return []confirmActionTeamWorkRef{ref}, project, projectErr
		}
	}
	link := confirmedActionTeamWorkLink{ProofID: item.IntentProofID, ContractID: item.ContractID, RunID: item.RunID, AuditID: payload.AuditID, AuditUser: payload.AuditUser, Scope: payload.Scope, FixtureScopeID: payload.FixtureScopeID}
	return s.persistConfirmedActionVisibility(ctx, link, plannedDispatchVisibilityResults(payload.Scope))
}

func plannedDispatchVisibilityResults(scope *protocol.ScopeValidation) []plannedToolExecutionResult {
	if scope == nil {
		return nil
	}
	results := []plannedToolExecutionResult{}
	for _, planned := range scope.PlannedToolCalls {
		if isDelegateTool(planned.Name) {
			results = append(results, plannedToolExecutionResult{Name: planned.Name, ToolRef: planned.ToolRef, Arguments: planned.Arguments})
		}
	}
	return results
}

func (s *AdminServer) completeAsyncDispatchRun(ctx context.Context, item *dispatchoutbox.Item) error {
	tx, err := s.getDB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := s.markRunCompletedTx(tx, item.RunID, item.IntentProofID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *AdminServer) retryOrFailConfirmedDispatch(ctx context.Context, item *dispatchoutbox.Item, payload confirmedActionDispatchPayload, cause error) error {
	if item.AttemptCount < confirmedActionMaxAttempts {
		delay := time.Duration(item.AttemptCount) * time.Second
		return s.DispatchOutbox.MarkRetry(ctx, item.ID, cause, delay)
	}
	if err := s.failAsyncDispatchRun(ctx, item, payload, cause); err != nil {
		return err
	}
	return s.DispatchOutbox.MarkFailed(ctx, item.ID, cause)
}

func (s *AdminServer) failAsyncDispatchRun(ctx context.Context, item *dispatchoutbox.Item, payload confirmedActionDispatchPayload, cause error) error {
	tx, err := s.getDB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := s.failChatProofTx(tx, item.IntentProofID); err != nil {
		return err
	}
	if err := s.markRunFailedTx(tx, item.RunID, item.IntentProofID, cause.Error()); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.failConfirmedActionWorkerRun(ctx, item.RunID, cause)
	auditID := payload.AuditID
	if auditID == "" {
		auditID, _ = s.createAuditEvent(protocol.TemplateChatToProposal, "confirm-action-dispatch", "Committed execution dispatch failed", map[string]any{"run_id": item.RunID, "intent_proof_id": item.IntentProofID, "reason": cause.Error()})
	}
	proofID := s.persistConfirmActionFailureProof(ctx, item.IntentProofID, item.ContractID, item.RunID, auditID, cause)
	link := confirmedActionTeamWorkLink{ProofID: item.IntentProofID, ContractID: item.ContractID, ProofArtifactID: proofID, RunID: item.RunID, AuditID: auditID, AuditUser: payload.AuditUser, Scope: payload.Scope, FixtureScopeID: payload.FixtureScopeID}
	return s.persistFailedConfirmedActionTeamWork(ctx, link, cause)
}

func persistCorrelatedScopeTx(ctx context.Context, tx *sql.Tx, proofID string, scope *protocol.ScopeValidation) error {
	raw, err := json.Marshal(scope)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE intent_proofs SET scope_validation=$2::jsonb WHERE id=$1`, strings.TrimSpace(proofID), string(raw))
	return err
}
