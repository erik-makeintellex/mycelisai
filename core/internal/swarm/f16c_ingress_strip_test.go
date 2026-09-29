package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

const f16cKey = "confirm-action:" + f16bProof

// f16cClaim is a complete confirmed-dispatch claim, including its key.
func f16cClaim(teamID string) map[string]any {
	return map[string]any{"run_id": f16bRun, "contract_id": f16bContract, "intent_proof_id": f16bProof,
		"work_item_id": f16bWorkItem, "idempotency_key": f16cKey, "team_id": teamID}
}

// f16cTeam runs the real Team.handleTrigger on the team's canonical
// internal.command lane.
func f16cTeam(t *testing.T, nc *nats.Conn, teamID string) *Team {
	t.Helper()
	team := NewTeam(&TeamManifest{ID: teamID, Name: teamID}, nc, nil, nil)
	sub, err := nc.Subscribe(fmt.Sprintf(protocol.TopicTeamInternalCommand, teamID), team.handleTrigger)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sub.Unsubscribe() })
	_ = nc.Flush()
	return team
}

// f16cAgentInbox captures what member agents receive on internal.trigger.
func f16cAgentInbox(t *testing.T, nc *nats.Conn, teamID string) chan []byte {
	t.Helper()
	got := make(chan []byte, 8)
	sub, err := nc.Subscribe(fmt.Sprintf(protocol.TopicTeamInternalTrigger, teamID), func(m *nats.Msg) { got <- m.Data })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sub.Unsubscribe() })
	_ = nc.Flush()
	return got
}

func f16cReceive(t *testing.T, inbox chan []byte, label string) []byte {
	t.Helper()
	select {
	case data := <-inbox:
		return data
	case <-time.After(3 * time.Second):
		t.Fatalf("%s: nothing reached internal.trigger", label)
		return nil
	}
}

// f16cAuthorityKeys lists context keys that spell a posture/correlation key
// in any case or separator style (run_id, RUN_ID, runId, run-id ...).
func f16cAuthorityKeys(data []byte) []string {
	var ask protocol.TeamAsk
	if json.Unmarshal(data, &ask) != nil {
		return nil
	}
	found := []string{}
	for key := range ask.Context {
		folded := strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(key))
		switch folded {
		case "runid", "contractid", "intentproofid", "workitemid", "idempotencykey":
			found = append(found, key)
		}
	}
	return found
}

// QA port (F16b-QA MEDIUM): a model-driven delegate_task replays a confirmed
// claim with a new goal. Every argument shape Core accepts must reach the
// team's agents without run/contract/proof/work-item/idempotency keys, so the
// agent never grants posture, even when the store would confirm the ids.
func TestF16cDelegateTaskDropsModelClaim(t *testing.T) {
	_, nc := startTestNATS(t)
	variant := map[string]any{"RUN_ID": f16bRun, "Contract_Id": f16bContract, "intentProofId": f16bProof, "work-item-id": f16bWorkItem, "IdempotencyKey": f16cKey}
	shapes := map[string]func(team string) map[string]any{
		"flat-context": func(team string) map[string]any {
			return map[string]any{"team_id": team, "task": "EVIL unapproved goal", "context": f16cClaim(team)}
		},
		"ask-map": func(team string) map[string]any {
			return map[string]any{"team_id": team, "ask": map[string]any{"goal": "EVIL unapproved goal", "context": f16cClaim(team)}}
		},
		"task-map": func(team string) map[string]any {
			return map[string]any{"team_id": team, "task": map[string]any{"goal": "EVIL unapproved goal", "context": f16cClaim(team)}}
		},
		"field-goal": func(team string) map[string]any {
			return map[string]any{"team_id": team, "goal": "EVIL unapproved goal", "context": f16cClaim(team)}
		},
		"variant-keys": func(team string) map[string]any {
			return map[string]any{"team_id": team, "task": "EVIL unapproved goal", "context": variant}
		},
	}
	// The calling agent is Soma's admin agent inside the very run it replays.
	agentCtx := WithToolInvocationContext(context.Background(), ToolInvocationContext{RunID: f16bRun, TeamID: "admin-core", AgentID: "admin",
		SourceKind: protocol.SourceKindSystem, SourceChannel: fmt.Sprintf(protocol.TopicTeamInternalTrigger, "admin-core"), PayloadKind: protocol.PayloadKindCommand})
	r := &InternalToolRegistry{nc: nc}
	for name, build := range shapes {
		team := "prime-" + name
		f16cTeam(t, nc, team)
		inbox := f16cAgentInbox(t, nc, team)
		if _, err := r.handleDelegateTask(agentCtx, build(team)); err != nil {
			t.Fatalf("%s: delegate_task: %v", name, err)
		}
		data := f16cReceive(t, inbox, name)
		if keys := f16cAuthorityKeys(data); len(keys) > 0 {
			t.Errorf("%s: model-supplied %v reached the team's agents: %s", name, keys, data)
		}
		if !strings.Contains(normalizeTeamTriggerInput(data), "EVIL") {
			t.Errorf("%s: the delegated goal was lost: %s", name, data)
		}
		db, mock := f16bMockDB(t)
		f16bExpectLookup(mock, team).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		if !f16bAgent(t, team, db).triggerPlanningOnly(data) {
			t.Errorf("%s: model-replayed claim got execution posture", name)
		}
	}
}

// Core's confirmed dispatch (server.confirmedActionToolContext, marked by
// WithConfirmedDispatchToolContext since TPD) is the one caller that keeps the
// correlation it built, and only for its own run.
func TestF16cConfirmedDispatchDelegateKeepsCoreClaim(t *testing.T) {
	_, nc := startTestNATS(t)
	const team = "prime-development"
	f16cTeam(t, nc, team)
	inbox := f16cAgentInbox(t, nc, team)
	r := &InternalToolRegistry{nc: nc}
	confirmed := ToolInvocationContext{RunID: f16bRun, AgentID: "operator", UserLabel: "operator",
		SourceKind: protocol.SourceKindWebAPI, SourceChannel: "api.intent.confirm-action", PayloadKind: protocol.PayloadKindCommand}
	args := func() map[string]any {
		return map[string]any{"team_id": team, "task": "write the approved note", "work_item_id": f16bWorkItem, "context": f16cClaim(team)}
	}

	if _, err := r.handleDelegateTask(WithConfirmedDispatchToolContext(context.Background(), confirmed), args()); err != nil {
		t.Fatal(err)
	}
	data := f16cReceive(t, inbox, "confirmed")
	var ask protocol.TeamAsk
	if err := json.Unmarshal(data, &ask); err != nil {
		t.Fatal(err)
	}
	for key, want := range f16cClaim(team) {
		if ask.Context[key] != want {
			t.Fatalf("confirmed dispatch lost %s: %#v", key, ask.Context)
		}
	}
	db, mock := f16bMockDB(t)
	f16bExpectLookup(mock, team).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	if f16bAgent(t, team, db).triggerPlanningOnly(data) {
		t.Fatal("verified confirmed dispatch lost execution posture")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

	otherRun, planning, systemKind := confirmed, confirmed, confirmed
	otherRun.RunID = "99999999-9999-4999-8999-999999999999"
	planning.PlanningOnly = true
	systemKind.SourceKind = protocol.SourceKindSystem
	for name, ctx := range map[string]context.Context{
		"other-run": WithConfirmedDispatchToolContext(context.Background(), otherRun), "planning-only": WithConfirmedDispatchToolContext(context.Background(), planning),
		"system-kind": WithConfirmedDispatchToolContext(context.Background(), systemKind), "no-invocation": context.Background(),
		"no-marker": WithToolInvocationContext(context.Background(), confirmed),
	} {
		if _, err := r.handleDelegateTask(ctx, args()); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if keys := f16cAuthorityKeys(f16cReceive(t, inbox, name)); len(keys) > 0 {
			t.Errorf("%s: kept %v outside Core's confirmed dispatch", name, keys)
		}
	}
}

// publish_signal is the other model-args path onto a team's command and
// trigger subjects; it gets the same strip, including the one JSON-string
// layer the team decodes (normalizeCommandPayload).
func TestF16cPublishSignalDropsClaim(t *testing.T) {
	_, nc := startTestNATS(t)
	const team = "prime-development"
	f16cTeam(t, nc, team)
	inbox := f16cAgentInbox(t, nc, team)
	askJSON, _ := json.Marshal(map[string]any{"goal": "EVIL unapproved goal", "context": f16cClaim(team)})
	quoted, _ := json.Marshal(string(askJSON))
	agentCtx := WithToolInvocationContext(context.Background(), ToolInvocationContext{RunID: f16bRun, TeamID: "admin-core", AgentID: "admin", SourceKind: protocol.SourceKindSystem})
	r := &InternalToolRegistry{nc: nc}
	for name, call := range map[string][2]string{
		"command-object": {fmt.Sprintf(protocol.TopicTeamInternalCommand, team), string(askJSON)},
		"command-quoted": {fmt.Sprintf(protocol.TopicTeamInternalCommand, team), string(quoted)},
		"trigger-object": {fmt.Sprintf(protocol.TopicTeamInternalTrigger, team), string(askJSON)},
	} {
		if _, err := r.handlePublishSignal(agentCtx, map[string]any{"subject": call[0], "message": call[1]}); err != nil {
			t.Fatalf("%s: publish_signal: %v", name, err)
		}
		data := f16cReceive(t, inbox, name)
		if keys := f16cAuthorityKeys(data); len(keys) > 0 {
			t.Errorf("%s: publish_signal delivered %v to the team's agents: %s", name, keys, data)
		}
		if !teamTriggerPlanningOnly(data) {
			t.Errorf("%s: publish_signal payload passes the posture pre-check: %s", name, data)
		}
	}
}

type f16cReceiptStore struct {
	mu    sync.Mutex
	teams []string
}

func (s *f16cReceiptStore) AcceptCommand(_ context.Context, c teamCommandCorrelation, _ string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.teams = append(s.teams, c.TeamID)
	return true, nil
}

// The durable command receipt is the single-use gate; it must be keyed on the
// team that received the command, not on a team_id written in the payload.
func TestF16cCommandReceiptBindsReceivingTeam(t *testing.T) {
	_, nc := startTestNATS(t)
	const team = "prime-development"
	store := &f16cReceiptStore{}
	f16cTeam(t, nc, team).commandReceipts = store
	inbox := f16cAgentInbox(t, nc, team)
	claim := f16cClaim(team)
	claim["team_id"] = "someone-else"
	raw, _ := json.Marshal(protocol.TeamAsk{Goal: "write the approved note", Context: claim})
	env, err := protocol.WrapSignalPayloadWithMeta(protocol.SourceKindInternalTool, "internal_tool.delegate_task", protocol.PayloadKindCommand, f16bRun, team, "", raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := nc.Publish(fmt.Sprintf(protocol.TopicTeamInternalCommand, team), env); err != nil {
		t.Fatal(err)
	}
	f16cReceive(t, inbox, "receipted command")
	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.teams) != 1 || store.teams[0] != team {
		t.Fatalf("receipt keyed on teams %v, want [%s]", store.teams, team)
	}
}
