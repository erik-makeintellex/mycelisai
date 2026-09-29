package trust

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// tpdPlan is one confirmed plan with two delegate_task calls to the same team
// on real PostgreSQL: one outbox row keyed confirm-action:<proof>, and one
// planned call per work item carrying its own delivery key.
type tpdPlan struct{ proof, run, contract, outboxKey, wi1, wi2 string }

func (p tpdPlan) claim(workItem, key string) ExecutionClaim {
	return ExecutionClaim{IntentProofID: p.proof, ContractID: p.contract, RunID: p.run, WorkItemID: workItem, TeamID: "prime-development", IdempotencyKey: key}
}

// tpdKey is the expected per-planned-call key format (TPD).
func tpdKey(outboxKey, workItem string) string { return outboxKey + ":" + workItem }

func tpdSeedTwoCalls(t *testing.T, plannedKey func(p tpdPlan, workItem string) string) tpdPlan {
	t.Helper()
	db := f16cRealDB(t)
	p := tpdPlan{proof: uuid.NewString(), run: uuid.NewString(), contract: uuid.NewString(), wi1: uuid.NewString(), wi2: uuid.NewString()}
	p.outboxKey = "confirm-action:" + p.proof
	planned := []map[string]any{}
	for _, wi := range []string{p.wi1, p.wi2} {
		ctx := map[string]any{"team_id": "prime-development", "work_item_id": wi, "intent_proof_id": p.proof, "run_id": p.run, "contract_id": p.contract}
		if key := plannedKey(p, wi); key != "" {
			ctx["idempotency_key"] = key
		}
		planned = append(planned, map[string]any{"name": "delegate_task", "arguments": map[string]any{"team_id": "prime-development", "context": ctx}})
	}
	scope, _ := json.Marshal(map[string]any{"planned_tool_calls": planned})
	f16cExec(t, db, `INSERT INTO intent_proofs (id, template_id, resolved_intent, status, scope_validation, confirmed_at, expires_at)
		VALUES ($1,'chat-to-proposal','tpd','confirmed',$2::jsonb,NOW(),NOW()+interval '10 minutes')`, p.proof, string(scope))
	f16cExec(t, db, `INSERT INTO mission_runs (id, mission_id, status) VALUES ($1,'tpd','running')`, p.run)
	f16cExec(t, db, `INSERT INTO execution_contracts (id, intent_proof_id, run_id, template_id) VALUES ($1,$2,$3,'chat-to-proposal')`, p.contract, p.proof, p.run)
	f16cExec(t, db, `INSERT INTO execution_dispatch_outbox (id, idempotency_key, dispatch_kind, status, run_id, intent_proof_id, contract_id, team_id, work_item_id, source_kind, source_channel, payload_kind)
		VALUES ($1,$2,'confirmed_action_team_plan','executing',$3,$4,$5,'prime-development',$6,'web_api','api.intent.confirm-action','command')`,
		uuid.NewString(), p.outboxKey, p.run, p.proof, p.contract, p.wi1)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM execution_dispatch_outbox WHERE intent_proof_id=$1`, p.proof)
		_, _ = db.Exec(`DELETE FROM execution_contracts WHERE id=$1`, p.contract)
		_, _ = db.Exec(`DELETE FROM mission_runs WHERE id=$1`, p.run)
		_, _ = db.Exec(`DELETE FROM intent_proofs WHERE id=$1`, p.proof)
	})
	return p
}

// The claim key must be exactly the key derived for THIS planned call (the
// outbox key bound to its work item), not any key with the outbox prefix and
// not the key of another call in the same plan.
func TestTPDVerifyExecutionClaimBindsKeyToPlannedCallRealDB(t *testing.T) {
	db := f16cRealDB(t)
	ctx := context.Background()
	p := tpdSeedTwoCalls(t, func(p tpdPlan, wi string) string { return tpdKey(p.outboxKey, wi) })
	for _, wi := range []string{p.wi1, p.wi2} {
		if err := VerifyExecutionClaim(ctx, db, p.claim(wi, tpdKey(p.outboxKey, wi))); err != nil {
			t.Errorf("TPD planned call %s with its own key rejected: %v", wi, err)
		}
	}
	for name, claim := range map[string]ExecutionClaim{
		"plan-wide-outbox-key": p.claim(p.wi1, p.outboxKey),
		"other-calls-key":      p.claim(p.wi1, tpdKey(p.outboxKey, p.wi2)),
		"other-calls-key-rev":  p.claim(p.wi2, tpdKey(p.outboxKey, p.wi1)),
		"key-with-suffix":      p.claim(p.wi1, tpdKey(p.outboxKey, p.wi1)+":x"),
		"key-for-unplanned-wi": p.claim(p.wi1, tpdKey(p.outboxKey, uuid.NewString())),
		"unplanned-wi-own-key": func() ExecutionClaim { wi := uuid.NewString(); return p.claim(wi, tpdKey(p.outboxKey, wi)) }(),
		"other-proof-prefix":   p.claim(p.wi1, tpdKey("confirm-action:"+uuid.NewString(), p.wi1)),
	} {
		if err := VerifyExecutionClaim(ctx, db, claim); !errors.Is(err, ErrExecutionClaimRejected) {
			t.Errorf("TPD %s: err=%v, want rejection", name, err)
		}
	}
}

// The planned call must itself carry the derived key: a stored plan whose
// calls hold the old plan-wide key (or none) grants no posture.
func TestTPDVerifyExecutionClaimRequiresPlannedCallKeyRealDB(t *testing.T) {
	db := f16cRealDB(t)
	ctx := context.Background()
	for name, plannedKey := range map[string]func(p tpdPlan, wi string) string{
		"planned-plan-wide-key": func(p tpdPlan, _ string) string { return p.outboxKey },
		"planned-no-key":        func(tpdPlan, string) string { return "" },
	} {
		p := tpdSeedTwoCalls(t, plannedKey)
		if err := VerifyExecutionClaim(ctx, db, p.claim(p.wi1, tpdKey(p.outboxKey, p.wi1))); !errors.Is(err, ErrExecutionClaimRejected) {
			t.Errorf("TPD %s: err=%v, want rejection", name, err)
		}
	}
}
