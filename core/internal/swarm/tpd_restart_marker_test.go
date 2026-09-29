package swarm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// tpdStatusInbox captures the team's signal.status payloads.
func tpdStatusInbox(t *testing.T, nc *nats.Conn, teamID string) chan map[string]any {
	t.Helper()
	got := make(chan map[string]any, 16)
	sub, err := nc.Subscribe(fmt.Sprintf(protocol.TopicTeamSignalStatus, teamID), func(m *nats.Msg) {
		var env protocol.SignalEnvelope
		payload := map[string]any{}
		if json.Unmarshal(m.Data, &env) == nil {
			_ = json.Unmarshal(env.Payload, &payload)
		}
		got <- payload
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sub.Unsubscribe() })
	_ = nc.Flush()
	return got
}

// tpdRecoveryStatuses returns the degraded statuses for workItem seen within d.
func tpdRecoveryStatuses(statuses chan map[string]any, workItem string, d time.Duration) []map[string]any {
	var out []map[string]any
	deadline := time.After(d)
	for {
		select {
		case p := <-statuses:
			if signalString(p["work_item_id"]) == workItem && signalString(p["state"]) == string(protocol.TeamWorkStateDegraded) {
				out = append(out, p)
			}
		case <-deadline:
			return out
		}
	}
}

// tpdLogBuffer captures the standard logger for the rest of the test.
type tpdLogBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *tpdLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *tpdLogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func tpdCaptureLog(t *testing.T) *tpdLogBuffer {
	t.Helper()
	buf := &tpdLogBuffer{}
	prev := log.Writer()
	log.SetOutput(buf)
	t.Cleanup(func() { log.SetOutput(prev) })
	return buf
}

// Restart: Core accepted a planned call, then restarted mid-run, and the
// outbox redispatches the same call. The durable receipt recognizes it as the
// same delivery. It is never blindly re-executed (the first attempt's outcome
// is unknown), and it is never silent: the team reports the work item as
// needing recovery. A redelivery inside the same process (dispatcher retry)
// is a plain duplicate of live work and is not reported as a failure.
func TestTPDRealDBRestartRedispatchIsReportedNotDropped(t *testing.T) {
	db := f16cOpenDB(t)
	_, nc := startTestNATS(t)
	const team = "fixture-dev-team"
	p := tpdSeedPlan(t, db, team)
	c := p.calls[0]
	_, lane := tpdTeam(t, nc, db, team)
	inbox := f16cAgentInbox(t, nc, team)
	statuses := tpdStatusInbox(t, nc, team)
	reg := &InternalToolRegistry{nc: nc}
	delegate := func(goal string) {
		t.Helper()
		if _, err := reg.handleDelegateTask(tpdConfirmedCtx(p.run), p.args(c, goal)); err != nil {
			t.Fatal(err)
		}
	}

	delegate("approved task")
	if f16bAgent(t, team, db).triggerPlanningOnly(f16cReceive(t, inbox, "first delivery")) {
		t.Fatal("TPD first delivery lost execution posture")
	}
	delegate("dispatcher retry in the same process")
	if n := len(tpdDrain(inbox, 700*time.Millisecond)); n != 0 {
		t.Errorf("TPD same-process redelivery reached the agents %d times", n)
	}
	if got := tpdRecoveryStatuses(statuses, c.workItem, 300*time.Millisecond); len(got) != 0 {
		t.Errorf("TPD same-process duplicate was reported as needing recovery: %v", got)
	}

	// Core restart: the old team's lane is gone; a new Team instance has an
	// empty memory and the same durable receipt store.
	_ = lane.Unsubscribe()
	tpdTeam(t, nc, db, team)
	delegate("redispatch after restart")
	if n := len(tpdDrain(inbox, 700*time.Millisecond)); n != 0 {
		t.Errorf("TPD restart redispatch blindly re-executed the planned call (%d deliveries)", n)
	}
	got := tpdRecoveryStatuses(statuses, c.workItem, 2*time.Second)
	if len(got) != 1 {
		t.Fatalf("TPD restart redispatch left %d recovery statuses, want 1 (silent drop)", len(got))
	}
	status := got[0]
	if status["needs_operator"] != true || signalString(status["degradation_state"]) != "team_command_outcome_unknown" {
		t.Errorf("TPD recovery status is not an honest needs-recovery report: %v", status)
	}
	// The projection dedupes status signals on payload kind plus
	// idempotency_key; the report must not reuse the acceptance status key.
	if signalString(status["idempotency_key"]) == c.key {
		t.Errorf("TPD recovery status reuses the acceptance key %q and would be deduped", c.key)
	}
	if n := tpdReceipts(t, db, c); n != 1 {
		t.Errorf("TPD durable receipts = %d, want 1", n)
	}
	var runStatus string
	if err := db.QueryRow(`SELECT status FROM mission_runs WHERE id=$1`, p.run).Scan(&runStatus); err != nil || runStatus == "completed" {
		t.Errorf("TPD run status = %q (err %v); undelivered work must not complete the run", runStatus, err)
	}
}

// Without a durable store the in-memory dedupe is the only gate. It must
// dedupe per planned call (work item plus key), never drop a different call
// that shares a key, and log every drop.
func TestTPDInMemoryDedupeIsPerPlannedCall(t *testing.T) {
	_, nc := startTestNATS(t)
	const team = "tpd-memory"
	f16cTeam(t, nc, team)
	inbox := f16cAgentInbox(t, nc, team)
	logs := tpdCaptureLog(t)
	const shared = "confirm-action:" + f16bProof
	send := func(workItem string) {
		t.Helper()
		raw, _ := json.Marshal(protocol.TeamAsk{Goal: "task " + workItem, Context: map[string]any{"work_item_id": workItem, "idempotency_key": shared}})
		env, _ := protocol.WrapSignalPayloadWithMeta(protocol.SourceKindInternalTool, "internal_tool.delegate_task", protocol.PayloadKindCommand, "", team, "", raw)
		if err := nc.Publish(fmt.Sprintf(protocol.TopicTeamInternalCommand, team), env); err != nil {
			t.Fatal(err)
		}
	}
	const wi1, wi2 = "11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"
	send(wi1)
	send(wi2)
	send(wi1)
	got := tpdCollect(inbox, 2, 2*time.Second)
	if len(got[wi1]) != 1 || len(got[wi2]) != 1 {
		t.Fatalf("TPD in-memory deliveries wi1=%d wi2=%d, want 1 and 1", len(got[wi1]), len(got[wi2]))
	}
	if out := logs.String(); !strings.Contains(out, "work_item_id="+wi1) || !strings.Contains(out, "action=drop_duplicate") {
		t.Errorf("TPD in-memory duplicate drop was not logged with its work item:\n%s", out)
	}
}

// Hardening: the confirmed-dispatch exemption needs Core's marker, not the
// source-channel string. A context that only carries web_api and
// api.intent.confirm-action is stripped like any model call.
func TestTPDConfirmedDispatchNeedsCoreOnlyMarker(t *testing.T) {
	_, nc := startTestNATS(t)
	const team = "tpd-marker"
	f16cTeam(t, nc, team)
	inbox := f16cAgentInbox(t, nc, team)
	r := &InternalToolRegistry{nc: nc}
	channelOnly := ToolInvocationContext{RunID: f16bRun, AgentID: "operator", UserLabel: "operator",
		SourceKind: protocol.SourceKindWebAPI, SourceChannel: "api.intent.confirm-action", PayloadKind: protocol.PayloadKindCommand}
	args := func() map[string]any {
		return map[string]any{"team_id": team, "task": "write the approved note", "work_item_id": f16bWorkItem, "context": f16cClaim(team)}
	}
	// Reading a marked context back and re-wrapping it through the ordinary
	// constructor drops the marker, so a copied context cannot carry it.
	copied, _ := ToolInvocationContextFromContext(WithConfirmedDispatchToolContext(context.Background(), channelOnly))
	for name, ctx := range map[string]context.Context{
		"channel-string-only":   WithToolInvocationContext(context.Background(), channelOnly),
		"rewrapped-marked-copy": WithToolInvocationContext(context.Background(), copied),
	} {
		if _, err := r.handleDelegateTask(ctx, args()); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if keys := f16cAuthorityKeys(f16cReceive(t, inbox, name)); len(keys) > 0 {
			t.Errorf("TPD %s kept %v without Core's confirmed-dispatch marker", name, keys)
		}
	}
	// Core's marker keeps the claim, for its own run only.
	if _, err := r.handleDelegateTask(WithConfirmedDispatchToolContext(context.Background(), channelOnly), args()); err != nil {
		t.Fatal(err)
	}
	if keys := f16cAuthorityKeys(f16cReceive(t, inbox, "marked")); len(keys) != 5 {
		t.Errorf("TPD Core's confirmed dispatch kept %v, want all 5 correlation keys", keys)
	}
	otherRun := channelOnly
	otherRun.RunID = "99999999-9999-4999-8999-999999999999"
	if _, err := r.handleDelegateTask(WithConfirmedDispatchToolContext(context.Background(), otherRun), args()); err != nil {
		t.Fatal(err)
	}
	if keys := f16cAuthorityKeys(f16cReceive(t, inbox, "marked-other-run")); len(keys) > 0 {
		t.Errorf("TPD marked context for another run kept %v", keys)
	}
}
