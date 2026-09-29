package swarm

// TPD (team plan delivery): one confirmed plan may hold several delegate_task
// calls, to one team or several. Every approved call must reach its team once,
// with execution posture, and never twice. Ported from F16c-QA's
// TestZZQAC_RealDB_TwoDelegatesSameTeam ("task 2 delivered=0").

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

type tpdCall struct{ team, workItem, key string }

// tpdPlan is one confirmed, dispatched plan on real PostgreSQL, shaped like
// the confirm-action transaction commits it: one outbox row keyed
// confirm-action:<proof>, and one annotated planned call per delegation.
type tpdPlan struct {
	proof, run, contract, outboxKey string
	calls                           []tpdCall
}

// tpdCallKey is the per-planned-call delivery key: the outbox key bound to the
// call's work item.
func tpdCallKey(outboxKey, workItem string) string { return outboxKey + ":" + workItem }

func tpdSeedPlan(t *testing.T, db *sql.DB, teams ...string) tpdPlan {
	t.Helper()
	p := tpdPlan{proof: uuid.NewString(), run: uuid.NewString(), contract: uuid.NewString()}
	p.outboxKey = "confirm-action:" + p.proof
	planned := []map[string]any{}
	for _, team := range teams {
		c := tpdCall{team: team, workItem: uuid.NewString()}
		c.key = tpdCallKey(p.outboxKey, c.workItem)
		p.calls = append(p.calls, c)
		planned = append(planned, map[string]any{"name": "delegate_task", "arguments": p.args(c, "approved task for "+team)})
	}
	scope, _ := json.Marshal(map[string]any{"planned_tool_calls": planned})
	f16cMustExec(t, db, `INSERT INTO intent_proofs (id, template_id, resolved_intent, status, scope_validation, confirmed_at, expires_at)
		VALUES ($1,'chat-to-proposal','tpd','confirmed',$2::jsonb,NOW(),NOW()+interval '10 minutes')`, p.proof, string(scope))
	f16cMustExec(t, db, `INSERT INTO mission_runs (id, mission_id, status) VALUES ($1,'tpd','running')`, p.run)
	f16cMustExec(t, db, `INSERT INTO execution_contracts (id, intent_proof_id, run_id, template_id) VALUES ($1,$2,$3,'chat-to-proposal')`, p.contract, p.proof, p.run)
	f16cMustExec(t, db, `INSERT INTO execution_dispatch_outbox (id, idempotency_key, dispatch_kind, status, run_id, intent_proof_id, contract_id, team_id, work_item_id, source_kind, source_channel, payload_kind)
		VALUES ($1,$2,'confirmed_action_team_plan','executing',$3,$4,$5,$6,$7,'web_api','api.intent.confirm-action','command')`,
		uuid.NewString(), p.outboxKey, p.run, p.proof, p.contract, p.calls[0].team, p.calls[0].workItem)
	for _, c := range p.calls {
		f16cMustExec(t, db, `INSERT INTO team_work_items (id, team_id, run_id, intent_proof_id, contract_id, objective, execution_shape, state)
			VALUES ($1,$2,$3,$4,$5,'tpd','delegated_work','queued')`, c.workItem, c.team, p.run, p.proof, p.contract)
	}
	t.Cleanup(func() {
		for _, c := range p.calls {
			_, _ = db.Exec(`DELETE FROM team_signal_receipts WHERE work_item_id=$1`, c.workItem)
			_, _ = db.Exec(`DELETE FROM team_work_items WHERE id=$1`, c.workItem)
		}
		_, _ = db.Exec(`DELETE FROM execution_dispatch_outbox WHERE intent_proof_id=$1`, p.proof)
		_, _ = db.Exec(`DELETE FROM execution_contracts WHERE id=$1`, p.contract)
		_, _ = db.Exec(`DELETE FROM mission_runs WHERE id=$1`, p.run)
		_, _ = db.Exec(`DELETE FROM intent_proofs WHERE id=$1`, p.proof)
	})
	return p
}

// args is the planned call's arguments exactly as Core annotated them
// (server.annotateConfirmedDelegationCall).
func (p tpdPlan) args(c tpdCall, goal string) map[string]any {
	return map[string]any{"team_id": c.team, "task": goal, "work_item_id": c.workItem, "context": map[string]any{
		"work_item_id": c.workItem, "team_id": c.team, "idempotency_key": c.key, "run_id": p.run, "intent_proof_id": p.proof, "contract_id": p.contract}}
}

// tpdTeam starts a Team on the canonical internal.command lane with the real
// durable receipt store. Unsubscribing its lane simulates a Core restart.
func tpdTeam(t *testing.T, nc *nats.Conn, db *sql.DB, teamID string) (*Team, *nats.Subscription) {
	t.Helper()
	team := NewTeam(&TeamManifest{ID: teamID, Name: teamID}, nc, nil, nil)
	team.commandReceipts = NewPostgresCommandReceiptStore(db)
	sub, err := nc.Subscribe(fmt.Sprintf(protocol.TopicTeamInternalCommand, teamID), team.handleTrigger)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sub.Unsubscribe() })
	_ = nc.Flush()
	return team, sub
}

// tpdConfirmedCtx is Core's confirmed-dispatch tool context for run.
func tpdConfirmedCtx(run string) context.Context {
	return WithConfirmedDispatchToolContext(context.Background(), ToolInvocationContext{RunID: run, AgentID: "operator", UserLabel: "operator",
		SourceKind: protocol.SourceKindWebAPI, SourceChannel: "api.intent.confirm-action", PayloadKind: protocol.PayloadKindCommand})
}

// tpdCollect gathers deliveries by work item until want arrive or d passes.
func tpdCollect(inbox chan []byte, want int, d time.Duration) map[string][][]byte {
	got := map[string][][]byte{}
	deadline := time.After(d)
	for n := 0; ; {
		select {
		case data := <-inbox:
			var ask protocol.TeamAsk
			_ = json.Unmarshal(data, &ask)
			wi := signalString(ask.Context["work_item_id"])
			got[wi] = append(got[wi], data)
			if n++; n >= want {
				// keep draining briefly so a duplicate would be seen
				deadline = time.After(300 * time.Millisecond)
				want = 1 << 30
			}
		case <-deadline:
			return got
		}
	}
}

func tpdReceipts(t *testing.T, db *sql.DB, c tpdCall) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM team_signal_receipts WHERE team_id=$1 AND work_item_id=$2 AND direction='command'`, c.team, c.workItem).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// tpdDeliverPlan runs every planned call through the real delegate_task tool,
// the bus, the team's durable receipt and the agent's claim verification.
func tpdDeliverPlan(t *testing.T, db *sql.DB, nc *nats.Conn, p tpdPlan) map[string]chan []byte {
	t.Helper()
	inboxes := map[string]chan []byte{}
	for _, c := range p.calls {
		if _, ok := inboxes[c.team]; !ok {
			tpdTeam(t, nc, db, c.team)
			inboxes[c.team] = f16cAgentInbox(t, nc, c.team)
		}
	}
	reg := &InternalToolRegistry{nc: nc}
	for i, c := range p.calls {
		if _, err := reg.handleDelegateTask(tpdConfirmedCtx(p.run), p.args(c, fmt.Sprintf("approved task %d", i+1))); err != nil {
			t.Fatalf("delegate call %d: %v", i+1, err)
		}
	}
	perTeam := map[string]int{}
	for _, c := range p.calls {
		perTeam[c.team]++
	}
	for team, want := range perTeam {
		got := tpdCollect(inboxes[team], want, 3*time.Second)
		for _, c := range p.calls {
			if c.team != team {
				continue
			}
			if n := len(got[c.workItem]); n != 1 {
				t.Errorf("TPD %s work item %s delivered %d times, want 1", team, c.workItem, n)
				continue
			}
			if f16bAgent(t, team, db).triggerPlanningOnly(got[c.workItem][0]) {
				t.Errorf("TPD %s work item %s lost execution posture", team, c.workItem)
			}
			if n := tpdReceipts(t, db, c); n != 1 {
				t.Errorf("TPD %s work item %s has %d durable command receipts, want 1", team, c.workItem, n)
			}
		}
	}
	return inboxes
}

// Two approved delegate_task calls to the same team: both delivered, both
// with execution posture, and a replay of either is refused.
func TestTPDRealDBTwoDelegatesSameTeamBothDelivered(t *testing.T) {
	db := f16cOpenDB(t)
	_, nc := startTestNATS(t)
	const team = "prime-development"
	p := tpdSeedPlan(t, db, team, team)
	inbox := tpdDeliverPlan(t, db, nc, p)[team]

	reg := &InternalToolRegistry{nc: nc}
	for i, c := range p.calls {
		_, _ = reg.handleDelegateTask(tpdConfirmedCtx(p.run), p.args(c, "EVIL replay through the confirmed dispatch"))
		raw, _ := json.Marshal(protocol.TeamAsk{Goal: "EVIL raw replay", Context: p.args(c, "")["context"].(map[string]any)})
		env, _ := protocol.WrapSignalPayloadWithMeta(protocol.SourceKindInternalTool, "internal_tool.delegate_task", protocol.PayloadKindCommand, p.run, team, "", raw)
		_ = nc.Publish(fmt.Sprintf(protocol.TopicTeamInternalCommand, team), env)
		if got := tpdDrain(inbox, 700*time.Millisecond); len(got) != 0 {
			t.Errorf("TPD replay of call %d reached the agents %d times", i+1, len(got))
		}
	}
	// A key is bound to its own planned call: call 1's key with call 2's work
	// item (and the reverse) never verifies.
	agent := f16bAgent(t, team, db)
	a, b := p.calls[0], p.calls[1]
	for label, ctx := range map[string]map[string]any{
		"key1-on-wi2":   {"run_id": p.run, "contract_id": p.contract, "intent_proof_id": p.proof, "work_item_id": b.workItem, "idempotency_key": a.key},
		"key2-on-wi1":   {"run_id": p.run, "contract_id": p.contract, "intent_proof_id": p.proof, "work_item_id": a.workItem, "idempotency_key": b.key},
		"plan-wide-key": {"run_id": p.run, "contract_id": p.contract, "intent_proof_id": p.proof, "work_item_id": a.workItem, "idempotency_key": p.outboxKey},
	} {
		raw, _ := json.Marshal(protocol.TeamAsk{Goal: "EVIL", Context: ctx})
		if !agent.triggerPlanningOnly(raw) {
			t.Errorf("TPD %s got execution posture", label)
		}
	}
}

// Three approved calls across two teams: each reaches its own team once.
func TestTPDRealDBThreeCallsAcrossTwoTeams(t *testing.T) {
	db := f16cOpenDB(t)
	_, nc := startTestNATS(t)
	p := tpdSeedPlan(t, db, "prime-development", "admin-core", "prime-development")
	tpdDeliverPlan(t, db, nc, p)
	// A call planned for admin-core never verifies at prime-development.
	cross := p.calls[1]
	raw, _ := json.Marshal(protocol.TeamAsk{Goal: "EVIL", Context: p.args(cross, "")["context"].(map[string]any)})
	if !f16bAgent(t, "prime-development", db).triggerPlanningOnly(raw) {
		t.Error("TPD admin-core's planned call verified at prime-development")
	}
}

func tpdDrain(inbox chan []byte, d time.Duration) [][]byte {
	var got [][]byte
	deadline := time.After(d)
	for {
		select {
		case m := <-inbox:
			got = append(got, m)
		case <-deadline:
			return got
		}
	}
}
