package server

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestConfirmedActionNeedsAsyncDispatchForTeamWork(t *testing.T) {
	if confirmedActionNeedsAsyncDispatch(&protocol.ScopeValidation{PlannedToolCalls: []protocol.PlannedToolCall{{Name: "write_file"}}}) {
		t.Fatal("bounded inline tool should not require team dispatch")
	}
	if !confirmedActionNeedsAsyncDispatch(&protocol.ScopeValidation{PlannedToolCalls: []protocol.PlannedToolCall{{Name: "create_team"}}}) {
		t.Fatal("create_team must dispatch after commit so profile hydration cannot block on the confirmation transaction")
	}
	if !confirmedActionNeedsAsyncDispatch(&protocol.ScopeValidation{PlannedToolCalls: []protocol.PlannedToolCall{{Name: "delegate_task"}}}) {
		t.Fatal("delegated work must use async dispatch")
	}
}

func TestCorrelateConfirmedActionScopeProducesStableDispatchIdentity(t *testing.T) {
	scope := &protocol.ScopeValidation{PlannedToolCalls: []protocol.PlannedToolCall{
		{Name: "create_team", Arguments: map[string]any{"team_id": "delivery-team"}},
		{Name: "delegate_task", Arguments: map[string]any{"team_id": "delivery-team", "task": "build the package"}},
	}}
	correlated := correlateConfirmedActionScope(scope, "run-1", "proof-1", "contract-1")
	teamID, workItemID := confirmedActionDispatchTargets(correlated)
	if teamID != "delivery-team" || workItemID == "" {
		t.Fatalf("targets = %q/%q", teamID, workItemID)
	}
	args := correlated.PlannedToolCalls[1].Arguments
	if got := correlationContextValue(args, "idempotency_key"); got != "confirm-action:proof-1" {
		t.Fatalf("idempotency key = %q", got)
	}
	if got := correlationContextValue(args, "run_id"); got != "run-1" {
		t.Fatalf("run id = %q", got)
	}
	again := correlateConfirmedActionScope(correlated, "run-1", "proof-1", "contract-1")
	if _, againWorkID := confirmedActionDispatchTargets(again); againWorkID != workItemID {
		t.Fatalf("re-correlation changed work id: %q -> %q", workItemID, againWorkID)
	}
}

func TestPlannedDispatchVisibilityIncludesOnlyDurableTeamWork(t *testing.T) {
	scope := &protocol.ScopeValidation{PlannedToolCalls: []protocol.PlannedToolCall{
		{Name: "create_team", Arguments: map[string]any{"team_id": "delivery-team"}},
		{Name: "write_file", Arguments: map[string]any{"path": "draft.md"}},
		{Name: "delegate_task", Arguments: map[string]any{"team_id": "delivery-team", "task": "build"}},
	}}
	results := plannedDispatchVisibilityResults(scope)
	if len(results) != 1 || results[0].Name != "delegate_task" {
		t.Fatalf("visibility results = %#v", results)
	}
}
