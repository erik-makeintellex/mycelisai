package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestProjectedRecoveryOptionsForValidationDegradation(t *testing.T) {
	tests := []struct {
		reason string
		want   string
	}{
		{"missing_retained_output", "Ask Soma to have the team attach or regenerate the retained deliverable."},
		{"invalid_deliverable_shape", "Ask Soma to have the same team package, validate, and return the deliverable with a direct entrypoint."},
		{"incomplete_deliverable_files", "Ask Soma to have the same team restore the missing package files, validate the primary interaction, and return the repaired deliverable."},
		{"unverified_primary_interaction", "Ask Soma to have the same team expose a clear primary control, verify that it changes the application, and return the repaired package."},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			item := protocol.TeamWorkItem{
				State:            protocol.TeamWorkStateDegraded,
				DegradationState: tt.reason,
				RecoveryOptions:  []string{"Retry without addressing validation."},
			}
			got := projectedRecoveryOptionsForItem(item, nil)
			if len(got) != 1 || got[0] != tt.want {
				t.Fatalf("recovery options = %#v, want [%q]", got, tt.want)
			}
		})
	}
}

func TestProjectedRecoveryOptionsPreservesNonValidationRecovery(t *testing.T) {
	item := protocol.TeamWorkItem{
		State:            protocol.TeamWorkStateDegraded,
		DegradationState: "provider_timeout",
		RecoveryOptions:  []string{"Retry after provider recovery."},
	}
	got := projectedRecoveryOptionsForItem(item, nil)
	if len(got) != 1 || got[0] != item.RecoveryOptions[0] {
		t.Fatalf("recovery options = %#v, want existing provider recovery", got)
	}
}

func TestProjectedRecoveryOptionsUsesNormalizedWorkerRecovery(t *testing.T) {
	want := "Use Soma to retry the approved work after ensuring the team can write and read retained outputs."
	item := protocol.TeamWorkItem{
		State:            protocol.TeamWorkStateDegraded,
		DegradationState: "result_contract_unsatisfied",
		RecoveryOptions:  []string{"Retry without addressing the result contract."},
	}
	payload := map[string]any{
		"recovery_options": []any{"  " + want + "  ", "", want},
	}
	got := projectedRecoveryOptionsForItem(item, payload)
	if len(got) != 1 || got[0] != want {
		t.Fatalf("recovery options = %#v, want exact normalized worker option %q", got, want)
	}
}

func TestTeamWorkSignalProjection_ResultContractRecoveryReplacesStaleRecovery(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	want := "Use Soma to retry the approved work after ensuring the team can write and read retained outputs."
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectProjectedStatusEvent(mock, "research-team", workID, protocol.TeamWorkStateDegraded, protocol.PayloadKindResult, now)
	expectProjectedTeamWorkUpdateWithRecovery(mock, workID, protocol.TeamWorkStateDegraded, true, "result_contract_unsatisfied", []string{want})
	expectProjectedInteraction(mock, "research-team", workID, "degraded", protocol.PayloadKindResult, now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
		},
		Payload: json.RawMessage(`{
			"context":{"work_item_id":"` + workID + `"},
			"state":"degraded",
			"degradation_state":"result_contract_unsatisfied",
			"recovery_options":["` + want + `"]
		}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamWorkSignalProjection_OutputReadyClearsStaleRecovery(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.MatchExpectationsInOrder(true)
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateDegraded, true, "result_contract_unsatisfied", now)
	expectProjectedStatusEvent(mock, "research-team", workID, protocol.TeamWorkStateOutputReady, protocol.PayloadKindResult, now)
	expectProjectedTeamWorkUpdateWithRecovery(mock, workID, protocol.TeamWorkStateOutputReady, false, "", nil)
	expectProjectedInteraction(mock, "research-team", workID, "output_ready", protocol.PayloadKindResult, now)

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.research-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			TeamID:        "research-team",
		},
		Payload: json.RawMessage(`{
			"context":{"work_item_id":"` + workID + `"},
			"state":"output_ready",
			"outputs":[{
				"output_id":"release-proof",
				"kind":"document",
				"storage_ref":"groups/research-team/generated/release-proof.md"
			}]
		}`),
	})

	projection := &teamWorkSignalProjection{server: s}
	if err := projection.project(t.Context(), "swarm.team.research-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTeamWorkSignalProjection_UnverifiedInteractionPersistsRecoveryWithoutCompletionProof(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	packageDir := filepath.Join(workspace, "groups", "game-team", "generated", "game")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("create package dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "index.html"), []byte(`<!doctype html><title>Unverified game</title><p>Ready.</p>`), 0o644); err != nil {
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
	expectProjectedStatusEventInsertOnly(mock, "game-team", workID, protocol.TeamWorkStateDegraded, protocol.PayloadKindResult, now)
	mock.ExpectExec("INSERT INTO mission_events").WillReturnResult(sqlmock.NewResult(0, 1))
	expectProjectedTeamWorkUpdateWithRecovery(
		mock,
		workID,
		protocol.TeamWorkStateDegraded,
		true,
		"unverified_primary_interaction",
		[]string{"Ask Soma to have the same team expose a clear primary control, verify that it changes the application, and return the repaired package."},
	)
	expectProjectedInteractionInsertOnly(mock, "game-team", workID, "degraded", protocol.PayloadKindResult, now)
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
			"summary":"Package needs interaction proof",
			"outputs":[{
				"output_id":"game-package",
				"kind":"project_package",
				"label":"Unverified game package",
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

func expectProjectedTeamWorkUpdateWithRecovery(mock sqlmock.Sqlmock, workID string, state protocol.TeamWorkState, needsOperator bool, degradation string, recovery []string) {
	mock.ExpectExec("UPDATE team_work_items").
		WithArgs(
			workID, string(state), sqlmock.AnyArg(), needsOperator, degradation,
			jsonArray(recovery), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
}
