package trust

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestF16bVerifyExecutionClaimFailsClosedWithoutStore(t *testing.T) {
	good := ExecutionClaim{IntentProofID: uuid.NewString(), ContractID: uuid.NewString(), RunID: uuid.NewString(), WorkItemID: "wi", TeamID: "team", IdempotencyKey: "key"}
	if err := VerifyExecutionClaim(context.Background(), nil, good); !errors.Is(err, ErrExecutionClaimRejected) {
		t.Fatalf("nil store: err=%v", err)
	}
	for name, claim := range map[string]ExecutionClaim{
		"forged-ids":   {IntentProofID: "forged-proof", ContractID: "forged-contract", RunID: "forged-run", WorkItemID: "wi", TeamID: "team"},
		"no-team":      {IntentProofID: good.IntentProofID, ContractID: good.ContractID, RunID: good.RunID, WorkItemID: "wi"},
		"no-work-item": {IntentProofID: good.IntentProofID, ContractID: good.ContractID, RunID: good.RunID, TeamID: "team"},
	} {
		if err := VerifyExecutionClaim(context.Background(), nil, claim); !errors.Is(err, ErrExecutionClaimRejected) {
			t.Fatalf("%s: err=%v", name, err)
		}
	}
}

// Real PostgreSQL proof. MYCELIS_TRUST_TEST_DSN must point at a disposable
// database with 001_current_schema.sql installed. Skip is not proof.
func TestF16bVerifyExecutionClaimRealDB(t *testing.T) {
	dsn := os.Getenv("MYCELIS_TRUST_TEST_DSN")
	if dsn == "" {
		t.Skip("requires disposable MYCELIS_TRUST_TEST_DSN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	proofID, runID, contractID, workItem := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	callKey := ConfirmedCallDeliveryKey("confirm-action:"+proofID, workItem) // TPD
	scope, _ := json.Marshal(map[string]any{"planned_tool_calls": []map[string]any{{
		"name": "delegate_task",
		"arguments": map[string]any{"team_id": "prime-development", "context": map[string]any{
			"team_id": "prime-development", "work_item_id": workItem, "intent_proof_id": proofID, "run_id": runID, "contract_id": contractID, "idempotency_key": callKey,
		}},
	}}})
	for _, stmt := range []struct {
		q    string
		args []any
	}{
		{`INSERT INTO intent_proofs (id, template_id, resolved_intent, status, scope_validation, expires_at) VALUES ($1, 'chat-to-proposal', 'f16b', 'confirmed', $2::jsonb, NOW() + interval '10 minutes')`, []any{proofID, string(scope)}},
		{`INSERT INTO mission_runs (id, mission_id, status) VALUES ($1, 'f16b', 'running')`, []any{runID}},
		{`INSERT INTO execution_contracts (id, intent_proof_id, run_id, template_id) VALUES ($1, $2, $3, 'chat-to-proposal')`, []any{contractID, proofID, runID}},
		{`INSERT INTO execution_dispatch_outbox (id, idempotency_key, dispatch_kind, status, run_id, intent_proof_id, contract_id, team_id, work_item_id, source_kind, source_channel, payload_kind)
		  VALUES ($1, $2, 'confirmed_action_team_plan', 'executing', $3, $4, $5, 'prime-development', $6, 'web_api', 'api.intent.confirm-action', 'command')`,
			[]any{uuid.NewString(), "confirm-action:" + proofID, runID, proofID, contractID, workItem}},
	} {
		if _, err := db.ExecContext(ctx, stmt.q, stmt.args...); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM execution_dispatch_outbox WHERE intent_proof_id = $1`, proofID)
		_, _ = db.Exec(`DELETE FROM execution_contracts WHERE id = $1`, contractID)
		_, _ = db.Exec(`DELETE FROM mission_runs WHERE id = $1`, runID)
		_, _ = db.Exec(`DELETE FROM intent_proofs WHERE id = $1`, proofID)
	})
	good := ExecutionClaim{IntentProofID: proofID, ContractID: contractID, RunID: runID, WorkItemID: workItem, TeamID: "prime-development", IdempotencyKey: callKey}
	if err := VerifyExecutionClaim(ctx, db, good); err != nil {
		t.Fatalf("confirmed dispatched proof rejected: %v", err)
	}
	for name, mutate := range map[string]func(*ExecutionClaim){
		"wrong-team":      func(c *ExecutionClaim) { c.TeamID = "admin-core" },
		"wrong-run":       func(c *ExecutionClaim) { c.RunID = uuid.NewString() },
		"wrong-contract":  func(c *ExecutionClaim) { c.ContractID = uuid.NewString() },
		"unknown-proof":   func(c *ExecutionClaim) { c.IntentProofID = uuid.NewString() },
		"wrong-work-item": func(c *ExecutionClaim) { c.WorkItemID = uuid.NewString() },
	} {
		claim := good
		mutate(&claim)
		if err := VerifyExecutionClaim(ctx, db, claim); !errors.Is(err, ErrExecutionClaimRejected) {
			t.Errorf("%s: err=%v, want rejection", name, err)
		}
	}
	if _, err := db.ExecContext(ctx, `UPDATE intent_proofs SET status = 'failed' WHERE id = $1`, proofID); err != nil {
		t.Fatal(err)
	}
	if err := VerifyExecutionClaim(ctx, db, good); !errors.Is(err, ErrExecutionClaimRejected) {
		t.Fatalf("failed proof still verifies: %v", err)
	}
}
