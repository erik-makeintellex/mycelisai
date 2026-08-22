package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

func TestTeamWorkSignalProjection_ResultWithoutRetainedOutputsDegradesDeliverable(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectProjectedStatusEvent(mock, "research-team", workID, protocol.TeamWorkStateDegraded, protocol.PayloadKindResult, now)
	recovery := []string{"Ask Soma to have the team attach or regenerate the retained deliverable."}
	expectProjectedTeamWorkUpdateWithRecovery(mock, workID, protocol.TeamWorkStateDegraded, true, "missing_retained_output", recovery)
	expectProjectedInteraction(mock, "research-team", workID, "degraded", protocol.PayloadKindResult, now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
			AgentID:       "builder",
		},
		Payload: json.RawMessage(`{"context":{"work_item_id":"` + workID + `"},"state":"output_ready","summary":"Draft ready","details":"Output passed local checks."}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestExternalMutationOutcomeLocksProjectionUntilNewGovernedWork(t *testing.T) {
	item := protocol.NormalizeTeamWorkItem(protocol.TeamWorkItem{
		TeamID: "external-team", Objective: "Update an external account",
		ExecutionShape: protocol.TeamExecutionShapeDeliverable,
		State:          protocol.TeamWorkStateDegraded,
		WorkIntent: &protocol.WorkIntent{SideEffect: &protocol.WorkSideEffectContract{
			EffectKind: protocol.WorkEffectExternalMutation, RetrySafety: protocol.WorkRetrySafe,
			IdempotencyKey: "external-update-1", SideEffectState: protocol.WorkSideEffectUnknown,
		}},
	})
	if !externalMutationOutcomeLocksProjection(item) {
		t.Fatal("unknown external outcome must reject delayed team projections")
	}
	item.WorkIntent.SideEffect.SideEffectState = protocol.WorkSideEffectVerifiedNotCommitted
	if !externalMutationOutcomeLocksProjection(item) {
		t.Fatal("verified-not-committed outcome must reject delayed team projections")
	}
	item.WorkIntent.SideEffect.SideEffectState = protocol.WorkSideEffectNotStarted
	if externalMutationOutcomeLocksProjection(item) {
		t.Fatal("not-started work should remain available to normal projection")
	}
}

func TestDeliverableResultMissingOutputs_AppliesToDelegatedWork(t *testing.T) {
	item := protocol.TeamWorkItem{
		ExecutionShape:  protocol.TeamExecutionShapeDelegatedWork,
		ExpectedOutputs: []string{"retained project package"},
	}
	if !deliverableResultMissingOutputs(item, protocol.PayloadKindResult, nil) {
		t.Fatal("delegated work with expected outputs must degrade when a result has no retained output refs")
	}
}

func TestTeamWorkSignalProjection_StatusUsesExplicitState(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateQueued, false, "", now)
	expectProjectedStatusEvent(mock, "research-team", workID, protocol.TeamWorkStateRunning, protocol.PayloadKindStatus, now)
	expectProjectedTeamWorkUpdate(mock, workID, protocol.TeamWorkStateRunning, false, "")
	expectProjectedInteraction(mock, "research-team", workID, "status", protocol.PayloadKindStatus, now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindStatus,
			TeamID:        "research-team",
			AgentID:       "planner",
		},
		Payload: json.RawMessage(`{"work_item_id":"` + workID + `","state":"running","headline":"Work started","message":"The team started execution."}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.status", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamWorkSignalProjection_ResultWithRetainedOutputRecordsCompletionProof(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	packageDir := filepath.Join(workspace, "groups", "game-team", "generated", "game")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("create package dir: %v", err)
	}
	packageHTML := `<!doctype html><title>Playable game</title><p>Click Play to begin.</p><button onclick="document.body.dataset.played='true'">Play</button>`
	if err := os.WriteFile(filepath.Join(packageDir, "index.html"), []byte(packageHTML), 0o644); err != nil {
		t.Fatalf("write package entrypoint: %v", err)
	}

	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	runID := "22222222-2222-2222-2222-222222222222"
	intentProofID := "33333333-3333-3333-3333-333333333333"
	contractID := "44444444-4444-4444-4444-444444444444"
	mock.MatchExpectationsInOrder(true)
	mockLinkedTeamWorkItem(mock, "game-team", workID, runID, intentProofID, contractID, now)
	mock.ExpectBegin()
	expectProjectedSignalReceipt(mock, "game-team", workID, "swarm.team.game-team.signal.result")
	mock.ExpectQuery("SELECT COUNT\\(\\*\\).*FROM team_work_items").
		WithArgs(runID, workID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("INSERT INTO proof_artifacts").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("55555555-5555-5555-5555-555555555555"))
	mock.ExpectExec("UPDATE execution_contracts").
		WillReturnResult(sqlmock.NewResult(0, 1))
	expectProjectedStatusEventInsertOnly(mock, "game-team", workID, protocol.TeamWorkStateOutputReady, protocol.PayloadKindResult, now)
	mock.ExpectExec("INSERT INTO mission_events").
		WillReturnResult(sqlmock.NewResult(0, 1))
	expectProjectedTeamWorkUpdate(mock, workID, protocol.TeamWorkStateOutputReady, false, "")
	expectProjectedInteractionInsertOnly(mock, "game-team", workID, "output_ready", protocol.PayloadKindResult, now)
	mock.ExpectExec("UPDATE mission_runs SET status = \\$1, completed_at = GREATEST").
		WithArgs(runs.StatusCompleted, runID, runs.StatusFailed).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO mission_events").
		WithArgs(sqlmock.AnyArg(), runID, "default", string(protocol.EventMissionCompleted), string(protocol.SeverityInfo), "admin", "governance", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.game-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "game-team",
			RunID:         runID,
		},
		Payload: json.RawMessage(`{
			"context":{"work_item_id":"` + workID + `"},
			"summary":"Playable package ready",
			"outputs":[{
				"output_id":"game-package",
				"kind":"project_package",
				"label":"Playable game package",
				"storage_ref":"groups/game-team/generated/game",
				"entrypoint":"index.html",
				"output_class":"user_deliverable"
			}]
		}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.game-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamWorkSignalProjection_ResultHonorsExplicitDegradedState(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectProjectedStatusEventWithSource(mock, "research-team", workID, protocol.TeamWorkStateDegraded, protocol.PayloadKindResult, string(protocol.SourceKindSystem), "swarm.team.research-team.internal.response", now)
	expectProjectedTeamWorkUpdate(mock, workID, protocol.TeamWorkStateDegraded, true, "provider_timeout")
	expectProjectedInteractionWithSource(mock, "research-team", workID, "degraded", protocol.PayloadKindResult, string(protocol.SourceKindSystem), "swarm.team.research-team.internal.response", now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindSystem,
			SourceChannel: "swarm.team.research-team.internal.response",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
		},
		Payload: json.RawMessage(`{"work_item_id":"` + workID + `","state":"degraded","headline":"Team ask degraded","details":"Provider timed out.","degradation_state":"provider_timeout"}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func mockLinkedTeamWorkItem(mock sqlmock.Sqlmock, teamID, workID, runID, intentProofID, contractID string, now time.Time) {
	mock.ExpectQuery("SELECT id::text, team_id").
		WithArgs(teamID, workID).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, teamID, runID, intentProofID, contractID, "", "Build playable game", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), "", []byte(`null`), []byte(`["playable app package"]`), []byte(`["launch smoke proof"]`), []byte(`[]`),
			"approved", string(protocol.TeamWorkStateRunning), []byte(`null`), false, "",
			[]byte(`[]`), []byte(`[]`), []byte(`[]`), []byte(`[]`), now, now, "v1",
		))
}

func TestTeamWorkSignalProjection_IgnoresArchivedWork(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateArchived, false, "missing_execution_plan", now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
			AgentID:       "builder",
		},
		Payload: json.RawMessage(`{"work_item_id":"` + workID + `","summary":"Late result arrived","details":"This must not revive active review."}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamWorkSignalProjection_UncorrelatedSignalIgnored(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
			AgentID:       "builder",
		},
		Payload: json.RawMessage(`{"summary":"No explicit active-work correlation."}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
