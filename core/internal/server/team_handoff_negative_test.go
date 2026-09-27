package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/swarm"
)

func TestTeamHandoffOutsideRunNeedsApproval(t *testing.T) {
	_, store, tools := handoffFixture(t)
	args := handOffToMarketing()
	args["target_team_id"] = "sales"
	_, err := callTool(t, tools, "hand_off", handoffLead("research-lead", "research", handoffRunID), args)
	requireBlocker(t, err, "handoff_needs_approval")
	if store.commits != 0 || len(store.outbox) != 0 {
		t.Fatalf("out-of-run handoff wrote rows: commits=%d outbox=%d", store.commits, len(store.outbox))
	}
}

func TestTeamHandoffUnknownTargetIsRejected(t *testing.T) {
	_, store, tools := handoffFixture(t)
	args := handOffToMarketing()
	args["target_team_id"] = "ghost-team"
	_, err := callTool(t, tools, "hand_off", handoffLead("research-lead", "research", handoffRunID), args)
	requireBlocker(t, err, "handoff_target_unknown")
	if store.commits != 0 {
		t.Fatal("unknown target wrote rows")
	}
}

// Sensitivity is Core's stored classification (provenance_sensitivity), never
// the model's own sensitivity_class label; unknown counts as restricted.
func TestTeamHandoffRestrictedContentNeedsApproval(t *testing.T) {
	for name, extra := range map[string]map[string]any{
		"core restricted, model says public": {"provenance_sensitivity": "restricted", "sensitivity_class": "public"},
		"no core class, model says public":   {"provenance_sensitivity": nil, "sensitivity_class": "public"},
		"unknown core class":                 {"provenance_sensitivity": "top-secret-ish"},
	} {
		_, store, tools := handoffFixture(t)
		artifact := attributedArtifact(handoffArtifactID2, "research", handoffRunID, extra)
		if extra["provenance_sensitivity"] == nil {
			delete(artifact.Metadata, "provenance_sensitivity")
		}
		store.artifacts[handoffArtifactID2] = artifact
		_, err := callTool(t, tools, "hand_off", handoffLead("research-lead", "research", handoffRunID), handOffToMarketing(handoffArtifactID, handoffArtifactID2))
		if swarm.HandoffBlockCode(err) != "handoff_needs_approval" || store.commits != 0 {
			t.Fatalf("%s: err=%v commits=%d", name, err, store.commits)
		}
	}
}

func TestTeamHandoffRejectsUnownedArtifacts(t *testing.T) {
	for name, artifact := range map[string]teamHandoffArtifact{
		"other team":   attributedArtifact(handoffArtifactID2, "sales", handoffRunID, nil),
		"other run":    attributedArtifact(handoffArtifactID2, "research", "99999999-9999-4999-8999-999999999999", nil),
		"unattributed": {ID: handoffArtifactID2, AgentID: "internal", Metadata: map[string]any{"provenance": "unattributed"}},
		"legacy":       {ID: handoffArtifactID2, AgentID: "internal", Metadata: map[string]any{"provenance_team_id": "research", "provenance_run_id": handoffRunID}},
	} {
		_, store, tools := handoffFixture(t)
		store.artifacts[handoffArtifactID2] = artifact
		_, err := callTool(t, tools, "hand_off", handoffLead("research-lead", "research", handoffRunID), handOffToMarketing(handoffArtifactID2))
		if swarm.HandoffBlockCode(err) != "handoff_artifact_not_attributed" || store.commits != 0 {
			t.Fatalf("%s: err=%v commits=%d", name, err, store.commits)
		}
	}
	_, store, tools := handoffFixture(t)
	_, err := callTool(t, tools, "hand_off", handoffLead("research-lead", "research", handoffRunID), handOffToMarketing("66666666-6666-4666-8666-666666666666"))
	if swarm.HandoffBlockCode(err) != "handoff_artifact_not_attributed" || store.commits != 0 {
		t.Fatalf("missing artifact: err=%v commits=%d", err, store.commits)
	}
}

func TestTeamHandoffDuplicateReturnsSameHandoff(t *testing.T) {
	_, store, tools := handoffFixture(t)
	ctx := handoffLead("research-lead", "research", handoffRunID)
	first, err := callTool(t, tools, "hand_off", ctx, handOffToMarketing())
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := callTool(t, tools, "hand_off", ctx, handOffToMarketing())
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	var a, b swarm.HandoffReceipt
	_ = json.Unmarshal([]byte(first), &a)
	_ = json.Unmarshal([]byte(second), &b)
	if a.HandoffID == "" || a.HandoffID != b.HandoffID || a.WorkItemID != b.WorkItemID || !b.Replayed || store.commits != 1 || len(store.outbox) != 1 {
		t.Fatalf("duplicate: first=%#v second=%#v commits=%d outbox=%d", a, b, store.commits, len(store.outbox))
	}
}

func TestTeamHandoffCommitFailureIsNotSuccess(t *testing.T) {
	_, store, tools := handoffFixture(t)
	store.commitErr = errors.New("database unavailable")
	out, err := callTool(t, tools, "hand_off", handoffLead("research-lead", "research", handoffRunID), handOffToMarketing())
	if err == nil || out != "" {
		t.Fatalf("commit failure reported success: out=%q err=%v", out, err)
	}
}

func TestReadHandoffInputOutsideScopeIsDenied(t *testing.T) {
	_, store, tools := handoffFixture(t)
	out, err := callTool(t, tools, "hand_off", handoffLead("research-lead", "research", handoffRunID), handOffToMarketing())
	if err != nil {
		t.Fatalf("hand_off: %v", err)
	}
	var receipt swarm.HandoffReceipt
	_ = json.Unmarshal([]byte(out), &receipt)
	cases := map[string]struct {
		ctx  context.Context
		args map[string]any
	}{
		"other team lead":     {handoffLead("sales-lead", "sales", handoffRunID), map[string]any{"handoff_id": receipt.HandoffID, "artifact_id": handoffArtifactID}},
		"source team":         {handoffLead("research-lead", "research", handoffRunID), map[string]any{"handoff_id": receipt.HandoffID, "artifact_id": handoffArtifactID}},
		"artifact not in ref": {handoffLead("marketing-lead", "marketing", handoffRunID), map[string]any{"handoff_id": receipt.HandoffID, "artifact_id": handoffArtifactID2}},
		"unknown handoff":     {handoffLead("marketing-lead", "marketing", handoffRunID), map[string]any{"handoff_id": "77777777-7777-4777-8777-777777777777", "artifact_id": handoffArtifactID}},
	}
	for name, tc := range cases {
		content, err := callTool(t, tools, "read_handoff_input", tc.ctx, tc.args)
		if swarm.HandoffBlockCode(err) != "handoff_input_not_in_scope" || content != "" {
			t.Fatalf("%s: content=%q err=%v", name, content, err)
		}
	}
	for _, verb := range store.interactionVerbs("marketing", receipt.WorkItemID) {
		if verb == "handoff_input_read" {
			t.Fatal("a denied read was recorded as a read")
		}
	}
}

func handoffViews(t *testing.T, store *memTeamHandoffStore) []teamHandoffView {
	t.Helper()
	rr := httptest.NewRecorder()
	listTeamHandoffs(store)(rr, httptest.NewRequest(http.MethodGet, "/api/v1/exchange/handoffs", nil))
	var body struct {
		Data []teamHandoffView `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil || len(body.Data) != 1 {
		t.Fatalf("handoffs body = %s (%v)", rr.Body.String(), err)
	}
	return body.Data
}

func chainSteps(view teamHandoffView) []string {
	steps := []string{}
	for _, step := range view.Chain {
		steps = append(steps, step.Step)
	}
	return steps
}

// NATS down: "handed" is never recorded. Early attempts stay honestly queued
// for retry; the final failed attempt shows needs_attention.
func TestTeamHandoffDispatchWithoutNATSIsNeverHanded(t *testing.T) {
	dbOption, mock := withDB(t)
	s, store, tools := handoffFixture(t, dbOption)
	s.DispatchOutbox = dispatchoutbox.NewStore(s.getDB())
	if _, err := callTool(t, tools, "hand_off", handoffLead("research-lead", "research", handoffRunID), handOffToMarketing()); err != nil {
		t.Fatalf("hand_off: %v", err)
	}
	staged := store.outbox[0]
	staged.ID, staged.AttemptCount = "dispatch-2", 1
	mock.ExpectExec("UPDATE execution_dispatch_outbox").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := s.dispatchTeamHandoff(t.Context(), &staged, store); err != nil {
		t.Fatalf("dispatch retry bookkeeping: %v", err)
	}
	view := handoffViews(t, store)[0]
	if view.Status != "queued" || !slices.Equal(chainSteps(view), []string{"created", "queued"}) {
		t.Fatalf("after retryable failure: status=%s chain=%v", view.Status, chainSteps(view))
	}
	staged.AttemptCount = teamHandoffMaxAttempts
	mock.ExpectExec("UPDATE execution_dispatch_outbox").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := s.dispatchTeamHandoff(t.Context(), &staged, store); err != nil {
		t.Fatalf("dispatch failure bookkeeping: %v", err)
	}
	view = handoffViews(t, store)[0]
	if view.Status != "needs_attention" || slices.Contains(chainSteps(view), "handed") {
		t.Fatalf("after final failure: status=%s chain=%v", view.Status, chainSteps(view))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
