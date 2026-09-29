package swarm

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mycelis/core/pkg/protocol"
)

// f16cLive is one confirmed, dispatched, live delegated run on real
// PostgreSQL, shaped like the confirm-action transaction commits it.
type f16cLive struct{ proof, run, contract, workItem, key, team string }

func (s f16cLive) ask(goal string, override map[string]any) []byte {
	ctx := map[string]any{"run_id": s.run, "contract_id": s.contract, "intent_proof_id": s.proof, "work_item_id": s.workItem, "idempotency_key": s.key, "team_id": s.team}
	for k, v := range override {
		ctx[k] = v
	}
	raw, _ := json.Marshal(protocol.TeamAsk{Goal: goal, Context: ctx})
	return raw
}

func f16cOpenDB(t *testing.T) *sql.DB {
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

func f16cMustExec(t *testing.T, db *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := db.Exec(q, args...); err != nil {
		t.Fatalf("exec %.70s: %v", q, err)
	}
}

func f16cSeedLiveRun(t *testing.T, db *sql.DB, team string) f16cLive {
	t.Helper()
	s := f16cLive{uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), "", team}
	outboxKey := "confirm-action:" + s.proof
	s.key = outboxKey + ":" + s.workItem // TPD: per-planned-call delivery key
	scope, _ := json.Marshal(map[string]any{"planned_tool_calls": []map[string]any{{"name": "delegate_task", "arguments": map[string]any{
		"team_id": team, "task": "write the approved note", "context": map[string]any{
			"team_id": team, "work_item_id": s.workItem, "intent_proof_id": s.proof, "run_id": s.run, "contract_id": s.contract, "idempotency_key": s.key,
		}}}}})
	f16cMustExec(t, db, `INSERT INTO intent_proofs (id, template_id, resolved_intent, status, scope_validation, confirmed_at, expires_at)
		VALUES ($1,'chat-to-proposal','f16c','confirmed',$2::jsonb,NOW(),NOW()+interval '10 minutes')`, s.proof, string(scope))
	f16cMustExec(t, db, `INSERT INTO mission_runs (id, mission_id, status) VALUES ($1,'f16c','running')`, s.run)
	f16cMustExec(t, db, `INSERT INTO execution_contracts (id, intent_proof_id, run_id, template_id) VALUES ($1,$2,$3,'chat-to-proposal')`, s.contract, s.proof, s.run)
	f16cMustExec(t, db, `INSERT INTO execution_dispatch_outbox (id, idempotency_key, dispatch_kind, status, run_id, intent_proof_id, contract_id, team_id, work_item_id, source_kind, source_channel, payload_kind)
		VALUES ($1,$2,'confirmed_action_team_plan','completed',$3,$4,$5,$6,$7,'web_api','api.intent.confirm-action','command')`,
		uuid.NewString(), outboxKey, s.run, s.proof, s.contract, team, s.workItem)
	f16cMustExec(t, db, `INSERT INTO team_work_items (id, team_id, run_id, intent_proof_id, contract_id, objective, execution_shape, state)
		VALUES ($1,$2,$3,$4,$5,'write the approved note','delegated_work','queued')`, s.workItem, team, s.run, s.proof, s.contract)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM team_signal_receipts WHERE work_item_id=$1`, s.workItem)
		_, _ = db.Exec(`DELETE FROM team_work_items WHERE id=$1`, s.workItem)
		_, _ = db.Exec(`DELETE FROM execution_dispatch_outbox WHERE intent_proof_id=$1`, s.proof)
		_, _ = db.Exec(`DELETE FROM execution_contracts WHERE id=$1`, s.contract)
		_, _ = db.Exec(`DELETE FROM mission_runs WHERE id=$1`, s.run)
		_, _ = db.Exec(`DELETE FROM intent_proofs WHERE id=$1`, s.proof)
	})
	return s
}

// QA port (TestZZQA_F16b_RealDBReplayAfterCompletion): the same confirmed
// claim that grants posture while its run is live must not grant it once the
// run finished 30 days ago and the proof expired, whatever door delivers it.
func TestF16cRealDBClaimReplayAfterCompletion(t *testing.T) {
	db := f16cOpenDB(t)
	s := f16cSeedLiveRun(t, db, "prime-development")
	agent := f16bAgent(t, "prime-development", db)
	if agent.triggerPlanningOnly(s.ask("write the approved note", nil)) {
		t.Fatal("live confirmed dispatch lost execution posture")
	}
	if !agent.triggerPlanningOnly(s.ask("EVIL: new goal", map[string]any{"idempotency_key": "replay-" + uuid.NewString()})) {
		t.Error("claim with a fresh idempotency key got execution posture")
	}
	if !f16bAgent(t, "admin-core", db).triggerPlanningOnly(s.ask("EVIL: new goal", nil)) {
		t.Error("claim presented to another team's agent got execution posture")
	}
	f16cMustExec(t, db, `UPDATE mission_runs SET status='completed', completed_at=NOW()-interval '30 days' WHERE id=$1`, s.run)
	f16cMustExec(t, db, `UPDATE execution_dispatch_outbox SET completed_at=NOW()-interval '30 days' WHERE intent_proof_id=$1`, s.proof)
	f16cMustExec(t, db, `UPDATE execution_contracts SET status='completed', execution_status='completed', completed_at=NOW()-interval '30 days' WHERE id=$1`, s.contract)
	f16cMustExec(t, db, `UPDATE intent_proofs SET expires_at=NOW()-interval '29 days', confirmed_at=NOW()-interval '30 days' WHERE id=$1`, s.proof)
	if !agent.triggerPlanningOnly(s.ask("EVIL: unrelated goal the human never approved", nil)) {
		t.Fatal("REPLAY: a completed, expired confirmed claim got execution posture")
	}
}

// Lead decision (F16c): intent_proofs.expires_at is the confirm-token TTL.
// A confirmed dispatch delivered after it (late confirm, retry, restart),
// with the run still running and its key unused, keeps execution posture,
// end to end through the team's durable receipt.
func TestF16cRealDBExpiredProofLiveRunKeepsPosture(t *testing.T) {
	db := f16cOpenDB(t)
	_, nc := startTestNATS(t)
	const team = "prime-development"
	s := f16cSeedLiveRun(t, db, team)
	f16cMustExec(t, db, `UPDATE intent_proofs SET confirmed_at=NOW()-interval '20 minutes', expires_at=NOW()-interval '5 minutes' WHERE id=$1`, s.proof)
	f16cTeam(t, nc, team).commandReceipts = NewPostgresCommandReceiptStore(db)
	inbox := f16cAgentInbox(t, nc, team)
	env, err := protocol.WrapSignalPayloadWithMeta(protocol.SourceKindInternalTool, "internal_tool.delegate_task", protocol.PayloadKindCommand, s.run, team, "", s.ask("write the approved note", nil))
	if err != nil {
		t.Fatal(err)
	}
	if err := nc.Publish(fmt.Sprintf(protocol.TopicTeamInternalCommand, team), env); err != nil {
		t.Fatal(err)
	}
	if f16bAgent(t, team, db).triggerPlanningOnly(f16cReceive(t, inbox, "late dispatch")) {
		t.Fatal("confirmed dispatch past the confirm-token TTL lost execution posture while its run is live")
	}
}

// Single use: the confirmed dispatch's key passes the team's durable command
// receipt once. A second delivery of the same key, even with the payload's
// team_id changed, never reaches the agents; a different key reaches them but
// cannot verify.
func TestF16cRealDBSecondUseDoesNotExecute(t *testing.T) {
	db := f16cOpenDB(t)
	_, nc := startTestNATS(t)
	const team = "prime-development"
	s := f16cSeedLiveRun(t, db, team)
	f16cTeam(t, nc, team).commandReceipts = NewPostgresCommandReceiptStore(db)
	inbox := f16cAgentInbox(t, nc, team)
	agent := f16bAgent(t, team, db)
	send := func(raw []byte) {
		t.Helper()
		env, err := protocol.WrapSignalPayloadWithMeta(protocol.SourceKindInternalTool, "internal_tool.delegate_task", protocol.PayloadKindCommand, s.run, team, "", raw)
		if err != nil {
			t.Fatal(err)
		}
		if err := nc.Publish(fmt.Sprintf(protocol.TopicTeamInternalCommand, team), env); err != nil {
			t.Fatal(err)
		}
		_ = nc.Flush()
	}
	quiet := func(label string) {
		t.Helper()
		select {
		case data := <-inbox:
			t.Errorf("%s reached the agents: %s", label, data)
		case <-time.After(700 * time.Millisecond):
		}
	}

	send(s.ask("write the approved note", nil))
	if agent.triggerPlanningOnly(f16cReceive(t, inbox, "first use")) {
		t.Fatal("first use of the confirmed dispatch lost execution posture")
	}
	send(s.ask("EVIL: second use", nil))
	quiet("second use of the same key")
	send(s.ask("EVIL: second use, other team_id", map[string]any{"team_id": "Prime-Development"}))
	quiet("second use with a rewritten team_id")
	send(s.ask("EVIL: fresh key", map[string]any{"idempotency_key": "replay-" + uuid.NewString()}))
	if !agent.triggerPlanningOnly(f16cReceive(t, inbox, "fresh key")) {
		t.Error("a fresh idempotency key got execution posture")
	}
}
