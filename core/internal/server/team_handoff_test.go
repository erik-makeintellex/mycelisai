package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

const (
	handoffRunID       = "11111111-1111-4111-8111-111111111111"
	handoffProofID     = "22222222-2222-4222-8222-222222222222"
	handoffArtifactID  = "33333333-3333-4333-8333-333333333333"
	handoffArtifactID2 = "44444444-4444-4444-8444-444444444444"
)

func handoffLead(agentID, teamID, runID string) context.Context {
	return swarm.WithToolInvocationContext(context.Background(), swarm.ToolInvocationContext{AgentID: agentID, TeamID: teamID, AgentRole: "lead", RunID: runID})
}

func attributedArtifact(id, team, run string, extra map[string]any) teamHandoffArtifact {
	meta := map[string]any{"provenance": "attributed", "provenance_agent_id": team + "-lead", "provenance_team_id": team, "provenance_run_id": run, "provenance_sensitivity": "team_scoped"}
	for k, v := range extra {
		meta[k] = v
	}
	return teamHandoffArtifact{ID: id, AgentID: team + "-lead", Title: "Juniper & Rye fact sheet", ContentType: "text/plain", Content: "Open Fri-Sun; rye sourdough.", Metadata: meta}
}

// handoffFixture: research (team A) and marketing (team B) both hold work in
// the approved run; sales is a known team outside the run.
func handoffFixture(t *testing.T, opts ...func(*AdminServer)) (*AdminServer, *memTeamHandoffStore, *swarm.InternalToolRegistry) {
	t.Helper()
	s := newTestServer(opts...)
	store := newMemTeamHandoffStore()
	store.artifacts[handoffArtifactID] = attributedArtifact(handoffArtifactID, "research", handoffRunID, nil)
	store.workItems = []protocol.TeamWorkItem{
		{WorkItemID: "55555555-5555-4555-8555-555555555551", TeamID: "research", RunID: handoffRunID, IntentProofID: handoffProofID, State: protocol.TeamWorkStateRunning},
		{WorkItemID: "55555555-5555-4555-8555-555555555552", TeamID: "marketing", RunID: handoffRunID, IntentProofID: handoffProofID, State: protocol.TeamWorkStateRunning},
	}
	known := map[string]bool{"research": true, "marketing": true, "sales": true}
	recorder := newTeamHandoffRecorder(s, store, func(_ context.Context, id string) (bool, error) { return known[id], nil })
	tools := swarm.NewInternalToolRegistry(swarm.InternalToolDeps{})
	tools.SetHandoffRecorder(recorder)
	return s, store, tools
}

func callTool(t *testing.T, tools *swarm.InternalToolRegistry, name string, ctx context.Context, args map[string]any) (string, error) {
	t.Helper()
	tool := tools.Get(name)
	if tool == nil {
		t.Fatalf("tool %s not registered", name)
	}
	return tool.Handler(ctx, args)
}

func handOffToMarketing(ids ...string) map[string]any {
	if len(ids) == 0 {
		ids = []string{handoffArtifactID}
	}
	artifactIDs := make([]any, 0, len(ids))
	for _, id := range ids {
		artifactIDs = append(artifactIDs, id)
	}
	return map[string]any{"target_team_id": "marketing", "artifact_ids": artifactIDs, "note": "Draft a weekend promo from this fact sheet.", "expected_action": "use"}
}

func requireBlocker(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil || swarm.HandoffBlockCode(err) != code {
		t.Fatalf("error = %v, want %s", err, code)
	}
}

// Acceptance chain: team A lead creates and hands off, team B is notified on
// its command topic constant, reads within scope, acknowledges, and the chain
// is visible through the Exchange "Team handoffs" API.
func TestTeamHandoffAcceptanceChain(t *testing.T) {
	dbOption, mock := withDB(t)
	s, store, tools := handoffFixture(t, dbOption, withNATS(t))
	s.DispatchOutbox = dispatchoutbox.NewStore(s.getDB())

	out, err := callTool(t, tools, "hand_off", handoffLead("research-lead", "research", handoffRunID), handOffToMarketing())
	if err != nil {
		t.Fatalf("hand_off: %v", err)
	}
	var receipt swarm.HandoffReceipt
	if err := json.Unmarshal([]byte(out), &receipt); err != nil || receipt.Status != "queued" || receipt.HandoffID == "" {
		t.Fatalf("receipt = %s (%v)", out, err)
	}
	if store.commits != 1 || len(store.outbox) != 1 {
		t.Fatalf("commits=%d outbox=%d", store.commits, len(store.outbox))
	}
	handedItem, err := store.WorkItem(context.Background(), "marketing", receipt.WorkItemID)
	if err != nil || handedItem.State != protocol.TeamWorkStateQueued || handedItem.RunID != handoffRunID || handedItem.IntentProofID != handoffProofID {
		t.Fatalf("target work item = %#v (%v)", handedItem, err)
	}
	staged := store.outbox[0]
	if staged.DispatchKind != teamHandoffDispatchKind || staged.TeamID != "marketing" || staged.WorkItemID != receipt.WorkItemID {
		t.Fatalf("outbox row = %#v", staged)
	}
	if verbs := store.interactionVerbs("marketing", receipt.WorkItemID); !slices.Equal(verbs, []string{"handoff_queued"}) {
		t.Fatalf("interactions = %v", verbs)
	}

	subject := fmt.Sprintf(protocol.TopicTeamInternalCommand, "marketing")
	commands := make(chan []byte, 1)
	sub, err := s.NC.Subscribe(subject, func(msg *nats.Msg) { commands <- msg.Data })
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Unsubscribe()
	_ = s.NC.Flush()
	mock.ExpectExec("UPDATE execution_dispatch_outbox").
		WithArgs("dispatch-1", dispatchoutbox.StatusCompleted, "", int64(0), dispatchoutbox.StatusCompleted, true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	staged.ID, staged.AttemptCount = "dispatch-1", 1
	if err := s.dispatchTeamHandoff(t.Context(), &staged, store); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	select {
	case raw := <-commands:
		var env protocol.SignalEnvelope
		if err := json.Unmarshal(raw, &env); err != nil || env.Meta.TeamID != "marketing" || env.Meta.RunID != handoffRunID {
			t.Fatalf("command envelope meta = %#v (%v)", env.Meta, err)
		}
		var ask protocol.TeamAsk
		if err := json.Unmarshal(env.Payload, &ask); err != nil || ask.Context["work_item_id"] != receipt.WorkItemID || ask.Context["handoff_id"] != receipt.HandoffID {
			t.Fatalf("command ask = %#v (%v)", ask, err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("team B was not notified on its command topic")
	}
	if verbs := store.interactionVerbs("marketing", receipt.WorkItemID); !slices.Equal(verbs, []string{"handoff_queued", "handoff_sent"}) {
		t.Fatalf("interactions after dispatch = %v", verbs)
	}

	readArgs := map[string]any{"handoff_id": receipt.HandoffID, "artifact_id": handoffArtifactID}
	content, err := callTool(t, tools, "read_handoff_input", handoffLead("marketing-lead", "marketing", handoffRunID), readArgs)
	if err != nil || !json.Valid([]byte(content)) || !strings.Contains(content, "rye sourdough") {
		t.Fatalf("read_handoff_input = %s (%v)", content, err)
	}
	store.projectAcceptance("marketing", receipt.WorkItemID)

	rr := httptest.NewRecorder()
	listTeamHandoffs(store)(rr, httptest.NewRequest(http.MethodGet, "/api/v1/exchange/handoffs", nil))
	assertStatus(t, rr, http.StatusOK)
	var body struct {
		OK   bool              `json:"ok"`
		Data []teamHandoffView `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil || len(body.Data) != 1 {
		t.Fatalf("handoffs body = %s (%v)", rr.Body.String(), err)
	}
	view := body.Data[0]
	steps := []string{}
	for _, step := range view.Chain {
		steps = append(steps, step.Step)
	}
	if view.SourceTeamID != "research" || view.TargetTeamID != "marketing" || view.Status != "read" ||
		!slices.Equal(steps, []string{"created", "queued", "handed", "acknowledged"}) || len(view.Inputs) != 1 {
		t.Fatalf("handoff view = %#v", view)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
