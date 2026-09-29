package server

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

// TPD: every delegate_task call in one confirmed plan gets its own work item
// and a delivery key bound to the proof and that work item
// (confirm-action:<proof>:<work_item_id>), stable across re-correlation. A
// work item id repeated by the planner is replaced, so two calls never share
// a delivery.
func TestTPDCorrelateGivesEachPlannedCallItsOwnDeliveryKey(t *testing.T) {
	const proofID = "22222222-2222-4222-8222-222222222222"
	scope := &protocol.ScopeValidation{PlannedToolCalls: []protocol.PlannedToolCall{
		{Name: "delegate_task", Arguments: map[string]any{"team_id": "fixture-dev-team", "task": "first"}},
		{Name: "write_file", Arguments: map[string]any{"path": "draft.md"}},
		{Name: "delegate_task", Arguments: map[string]any{"team_id": "fixture-dev-team", "task": "second"}},
		{Name: "delegate_task", Arguments: map[string]any{"team_id": "admin-core", "task": "third", "work_item_id": "dup-wi"}},
		{Name: "delegate_task", Arguments: map[string]any{"team_id": "admin-core", "task": "fourth", "context": map[string]any{"work_item_id": "dup-wi"}}},
	}}
	correlated := correlateConfirmedActionScope(scope, "run-1", proofID, "contract-1")
	seenWork, seenKey := map[string]bool{}, map[string]bool{}
	keys := []string{}
	for i, planned := range correlated.PlannedToolCalls {
		if !isDelegateTool(planned.Name) {
			if _, ok := planned.Arguments["context"]; ok {
				t.Errorf("TPD non-delegate call %d was annotated", i)
			}
			continue
		}
		wi := confirmedDelegationWorkItemID(planned.Arguments)
		key := correlationContextValue(planned.Arguments, "idempotency_key")
		if wi == "" || seenWork[wi] {
			t.Errorf("TPD call %d work item %q is empty or shared", i, wi)
		}
		if want := "confirm-action:" + proofID + ":" + wi; key != want {
			t.Errorf("TPD call %d key = %q, want %q", i, key, want)
		}
		if seenKey[key] {
			t.Errorf("TPD call %d reuses delivery key %q", i, key)
		}
		seenWork[wi], seenKey[key] = true, true
		keys = append(keys, key)
	}
	if len(keys) != 4 {
		t.Fatalf("TPD delegate calls = %d, want 4", len(keys))
	}
	again := correlateConfirmedActionScope(correlated, "run-1", proofID, "contract-1")
	i := 0
	for _, planned := range again.PlannedToolCalls {
		if !isDelegateTool(planned.Name) {
			continue
		}
		if key := correlationContextValue(planned.Arguments, "idempotency_key"); key != keys[i] {
			t.Errorf("TPD re-correlation changed call key %q -> %q", keys[i], key)
		}
		i++
	}
}
