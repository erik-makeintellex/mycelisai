package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) auditConfirmedAction(proofID, runID string, scope *protocol.ScopeValidation, auditUser string, actorIdentity map[string]any, pendingTeamWork bool) string {
	executionState := "verified"
	runResult := "completed"
	runMessage := "Execution run completed for confirmed chat proposal"
	if pendingTeamWork {
		executionState = "running"
		runResult = "started"
		runMessage = "Execution run started for confirmed chat proposal"
	}
	auditID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal, "confirm-action",
		"Chat proposal confirmed and execution record created",
		withActorIdentity(map[string]any{
			"proof_id":        proofID,
			"run_id":          runID,
			"execution_state": executionState,
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
		runMessage,
		withActorIdentity(map[string]any{
			"actor":           "Soma",
			"user":            auditUser,
			"action":          "execution_run",
			"result_status":   runResult,
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
