package server

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestExecutePlannedToolCalls_PropagatesRunAndWorkCorrelationToDelegateTask(t *testing.T) {
	wireNATS := withNATS(t)
	s := newTestServer(wireNATS)
	const runID = "33333333-3333-3333-3333-333333333333"
	const proofID = "22222222-2222-2222-2222-222222222222"
	const contractID = "44444444-4444-4444-4444-444444444444"
	subject := fmt.Sprintf(protocol.TopicTeamInternalCommand, "qa-team")
	commandCh := make(chan []byte, 1)
	sub, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
		commandCh <- msg.Data
	})
	if err != nil {
		t.Fatalf("subscribe team command: %v", err)
	}
	defer sub.Unsubscribe()
	s.NC.Flush()

	scope := &protocol.ScopeValidation{
		PlannedToolCalls: []protocol.PlannedToolCall{{
			Name: "delegate_task",
			Arguments: map[string]any{
				"team_id": "qa-team",
				"task":    "Validate retained output.",
			},
		}},
	}
	results, err := s.executePlannedToolCalls(t.Context(), scope, "test-user", runID, proofID, contractID)
	if err != nil {
		t.Fatalf("executePlannedToolCalls: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %#v, want one", results)
	}
	workID := confirmedDelegationWorkItemID(results[0].Arguments)
	if workID == "" {
		t.Fatal("delegation work_item_id was not annotated")
	}

	var env protocol.SignalEnvelope
	select {
	case raw := <-commandCh:
		if err := json.Unmarshal(raw, &env); err != nil {
			t.Fatalf("decode signal envelope: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for delegated command")
	}
	if env.Meta.RunID != runID {
		t.Fatalf("envelope run_id = %q, want %q", env.Meta.RunID, runID)
	}
	var ask protocol.TeamAsk
	if err := json.Unmarshal(env.Payload, &ask); err != nil {
		t.Fatalf("decode ask payload: %v", err)
	}
	if ask.Context["work_item_id"] != workID {
		t.Fatalf("payload work_item_id = %v, want %s", ask.Context["work_item_id"], workID)
	}
	if ask.Context["contract_id"] != contractID || ask.Context["intent_proof_id"] != proofID {
		t.Fatalf("payload context = %#v, want proof and contract linkage", ask.Context)
	}
}

func TestPersistConfirmedActionTeamWork_DelegatedWorkReusesCorrelationID(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	const workID = "55555555-5555-5555-5555-555555555555"
	link := testConfirmedActionTeamWorkLink(&protocol.ScopeValidation{})

	mock.ExpectBegin()
	expectTeamWorkItemInsert(mock, "qa-team", protocol.TeamExecutionShapeDelegatedWork, protocol.TeamWorkStateQueued, now)
	expectTeamStatusEventInsert(mock, "qa-team", protocol.TeamWorkStateQueued, now)
	expectTeamWorkItemUpdate(mock, protocol.TeamWorkStateQueued, sqlmock.AnyArg())
	expectTeamInteractionInsert(mock, "qa-team", "delegate", now)
	mock.ExpectCommit()

	refs, err := s.persistConfirmedActionTeamWork(t.Context(), link, []plannedToolExecutionResult{{
		Name: "delegate_task",
		Arguments: map[string]any{
			"team_id":      "qa-team",
			"task":         "Validate retained output.",
			"work_item_id": workID,
			"context": map[string]any{
				"work_item_id": workID,
			},
		},
		Output: "Delegated validation to QA.",
	}})
	if err != nil {
		t.Fatalf("persistConfirmedActionTeamWork: %v", err)
	}
	assertTeamWorkRef(t, refs, "qa-team", protocol.TeamWorkStateQueued, 0)
	if refs[0].WorkItemID != workID {
		t.Fatalf("ref work_item_id = %q, want %q", refs[0].WorkItemID, workID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
