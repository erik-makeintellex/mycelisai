package trust

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// A claim must name the dispatch's single-use idempotency key; without one it
// is rejected before the store is asked (F16c).
func TestF16cVerifyExecutionClaimRequiresIdempotencyKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	// The store would confirm the ids; the missing key alone must reject.
	for range 2 {
		mock.ExpectQuery(`SELECT EXISTS`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	}
	claim := ExecutionClaim{IntentProofID: uuid.NewString(), ContractID: uuid.NewString(), RunID: uuid.NewString(), WorkItemID: "wi", TeamID: "team"}
	for _, key := range []string{"", "   "} {
		claim.IdempotencyKey = key
		if err := VerifyExecutionClaim(context.Background(), db, claim); !errors.Is(err, ErrExecutionClaimRejected) {
			t.Fatalf("claim with idempotency_key %q verified: err=%v", key, err)
		}
	}
}

// f16cSeed is one confirmed, dispatched, live claim on real PostgreSQL.
type f16cSeed struct{ proof, run, contract, workItem, key string }

func (s f16cSeed) claim() ExecutionClaim {
	return ExecutionClaim{IntentProofID: s.proof, ContractID: s.contract, RunID: s.run, WorkItemID: s.workItem, TeamID: "prime-development", IdempotencyKey: s.key}
}

func f16cRealDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("MYCELIS_TRUST_TEST_DSN")
	if dsn == "" {
		t.Skip("requires disposable MYCELIS_TRUST_TEST_DSN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func f16cExec(t *testing.T, db *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := db.Exec(q, args...); err != nil {
		t.Fatalf("exec %.70s: %v", q, err)
	}
}

// f16cSeedLive writes what confirm-action commits for a live delegated run:
// a confirmed unexpired proof, a running run, the contract, and the claimed
// confirmed_action_team_plan outbox row keyed confirm-action:<proof>.
func f16cSeedLive(t *testing.T, db *sql.DB) f16cSeed {
	t.Helper()
	s := f16cSeed{uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), ""}
	s.key = "confirm-action:" + s.proof
	scope, _ := json.Marshal(map[string]any{"planned_tool_calls": []map[string]any{{"name": "delegate_task", "arguments": map[string]any{
		"team_id": "prime-development", "task": "write the approved note", "context": map[string]any{
			"team_id": "prime-development", "work_item_id": s.workItem, "intent_proof_id": s.proof, "run_id": s.run, "contract_id": s.contract, "idempotency_key": s.key,
		}}}}})
	f16cExec(t, db, `INSERT INTO intent_proofs (id, template_id, resolved_intent, status, scope_validation, confirmed_at, expires_at)
		VALUES ($1,'chat-to-proposal','f16c','confirmed',$2::jsonb,NOW(),NOW()+interval '10 minutes')`, s.proof, string(scope))
	f16cExec(t, db, `INSERT INTO mission_runs (id, mission_id, status) VALUES ($1,'f16c','running')`, s.run)
	f16cExec(t, db, `INSERT INTO execution_contracts (id, intent_proof_id, run_id, template_id) VALUES ($1,$2,$3,'chat-to-proposal')`, s.contract, s.proof, s.run)
	f16cExec(t, db, `INSERT INTO execution_dispatch_outbox (id, idempotency_key, dispatch_kind, status, run_id, intent_proof_id, contract_id, team_id, work_item_id, source_kind, source_channel, payload_kind)
		VALUES ($1,$2,'confirmed_action_team_plan','executing',$3,$4,$5,'prime-development',$6,'web_api','api.intent.confirm-action','command')`,
		uuid.NewString(), s.key, s.run, s.proof, s.contract, s.workItem)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM execution_dispatch_outbox WHERE intent_proof_id=$1`, s.proof)
		_, _ = db.Exec(`DELETE FROM execution_contracts WHERE id=$1`, s.contract)
		_, _ = db.Exec(`DELETE FROM mission_runs WHERE id=$1`, s.run)
		_, _ = db.Exec(`DELETE FROM intent_proofs WHERE id=$1`, s.proof)
	})
	return s
}

// Real PostgreSQL: a live confirmed dispatch verifies, even past the proof's
// confirm-token expiry; a terminal run, a failed or undispatched outbox row,
// or a key other than the outbox key does not. Skip is not proof.
func TestF16cVerifyExecutionClaimFreshnessRealDB(t *testing.T) {
	db := f16cRealDB(t)
	ctx := context.Background()

	live := f16cSeedLive(t, db)
	if err := VerifyExecutionClaim(ctx, db, live.claim()); err != nil {
		t.Fatalf("live confirmed dispatch rejected: %v", err)
	}
	done := f16cSeedLive(t, db)
	f16cExec(t, db, `UPDATE execution_dispatch_outbox SET status='completed', completed_at=NOW() WHERE intent_proof_id=$1`, done.proof)
	if err := VerifyExecutionClaim(ctx, db, done.claim()); err != nil {
		t.Fatalf("completed dispatch with the run still live rejected: %v", err)
	}
	// expires_at is the confirm-token TTL, not an execution window: a confirmed
	// proof past it (or with none) still verifies while its run is live.
	for name, mutation := range map[string]string{
		"proof-expired-run-live":   `UPDATE intent_proofs SET expires_at=NOW()-interval '2 hours' WHERE id=$1`,
		"proof-no-expiry-run-live": `UPDATE intent_proofs SET expires_at=NULL WHERE id=$1`,
	} {
		s := f16cSeedLive(t, db)
		f16cExec(t, db, mutation, s.proof)
		if err := VerifyExecutionClaim(ctx, db, s.claim()); err != nil {
			t.Errorf("%s: live confirmed dispatch rejected: %v", name, err)
		}
	}

	cases := map[string]string{
		"run-completed":      `UPDATE mission_runs SET status='completed', completed_at=NOW() WHERE id=(SELECT run_id FROM execution_contracts WHERE intent_proof_id=$1)`,
		"run-failed":         `UPDATE mission_runs SET status='failed', completed_at=NOW() WHERE id=(SELECT run_id FROM execution_contracts WHERE intent_proof_id=$1)`,
		"run-degraded":       `UPDATE mission_runs SET status='degraded', completed_at=NOW() WHERE id=(SELECT run_id FROM execution_contracts WHERE intent_proof_id=$1)`,
		"run-cancelled":      `UPDATE mission_runs SET status='cancelled', completed_at=NOW() WHERE id=(SELECT run_id FROM execution_contracts WHERE intent_proof_id=$1)`,
		"outbox-failed":      `UPDATE execution_dispatch_outbox SET status='failed' WHERE intent_proof_id=$1`,
		"outbox-staged":      `UPDATE execution_dispatch_outbox SET status='staged' WHERE intent_proof_id=$1`,
		"outbox-awaiting":    `UPDATE execution_dispatch_outbox SET status='awaiting_handler' WHERE intent_proof_id=$1`,
		"qa-replay-30d-late": `UPDATE intent_proofs SET expires_at=NOW()-interval '29 days', confirmed_at=NOW()-interval '30 days' WHERE id=$1`,
	}
	for name, mutation := range cases {
		s := f16cSeedLive(t, db)
		f16cExec(t, db, mutation, s.proof)
		if name == "qa-replay-30d-late" {
			f16cExec(t, db, `UPDATE mission_runs SET status='completed', completed_at=NOW()-interval '30 days' WHERE id=$1`, s.run)
			f16cExec(t, db, `UPDATE execution_dispatch_outbox SET status='completed', completed_at=NOW()-interval '30 days' WHERE intent_proof_id=$1`, s.proof)
		}
		if err := VerifyExecutionClaim(ctx, db, s.claim()); !errors.Is(err, ErrExecutionClaimRejected) {
			t.Errorf("%s: err=%v, want rejection", name, err)
		}
	}
	for name, key := range map[string]string{
		"fresh-model-key": "replay-" + uuid.NewString(),
		"other-proof-key": "confirm-action:" + uuid.NewString(),
		"key-case":        "CONFIRM-ACTION:" + live.proof,
	} {
		c := live.claim()
		c.IdempotencyKey = key
		if err := VerifyExecutionClaim(ctx, db, c); !errors.Is(err, ErrExecutionClaimRejected) {
			t.Errorf("%s: err=%v, want rejection", name, err)
		}
	}
}
