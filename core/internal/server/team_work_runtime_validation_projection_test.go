package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/outputvalidation"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

func TestInteractiveTeamResultStagesValidationWithoutProofOrRunCompletion(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	packageDir := filepath.Join(workspace, "groups", "app-team", "generated", "app")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `<!doctype html><button data-mycelis-primary-action>Run</button><main data-mycelis-validation-surface>Ready</main><script>document.querySelector('button').onclick=()=>document.querySelector('main').textContent='Changed'</script>`
	if err := os.WriteFile(filepath.Join(packageDir, "index.html"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	opt, mock := withDB(t)
	s := newTestServer(opt)
	s.DispatchOutbox = dispatchoutbox.NewStore(s.getDB())
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	runID := "22222222-2222-2222-2222-222222222222"
	proofID := "33333333-3333-3333-3333-333333333333"
	contractID := "44444444-4444-4444-4444-444444444444"
	workIntent := []byte(`{"kind":"project","output_contract":{"shape":"app_package","output_validation":{"kind":"interactive_browser","required":true,"checks":["load","no_page_errors","no_failed_local_assets"],"probe":{"action":{"kind":"click","target":"[data-mycelis-primary-action]"},"observe":{"kind":"text_change","target":"[data-mycelis-validation-surface]"}}}}}`)

	mock.MatchExpectationsInOrder(true)
	mock.ExpectQuery("SELECT id::text, team_id").WithArgs("app-team", workID).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, "app-team", runID, proofID, contractID, "", "Build browser application", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), "team_async", workIntent, []byte(`["application package"]`), []byte(`["runtime proof"]`), []byte(`[]`),
			"approved", string(protocol.TeamWorkStateRunning), []byte(`null`), false, "",
			[]byte(`[]`), []byte(`[]`), []byte(`[]`), []byte(`[]`), now, now, "v1",
		))
	mock.ExpectBegin()
	expectProjectedSignalReceipt(mock, "app-team", workID, "swarm.team.app-team.signal.result")
	expectProjectedStatusEventInsertOnly(mock, "app-team", workID, protocol.TeamWorkStateReviewing, protocol.PayloadKindResult, now)
	mock.ExpectExec("INSERT INTO mission_events").WillReturnResult(sqlmock.NewResult(0, 1))
	expectProjectedTeamWorkUpdateWithOutputs(mock, workID, protocol.TeamWorkStateReviewing, false, "", outputRefsMatch{
		TeamID: "app-team", WorkItemID: workID, Kind: "project_package",
		StorageRef: "groups/app-team/generated/app", Entrypoint: "index.html",
	})
	expectProjectedInteractionInsertOnly(mock, "app-team", workID, "reviewing", protocol.PayloadKindResult, now)
	mock.ExpectExec("INSERT INTO execution_dispatch_outbox").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec("UPDATE execution_dispatch_outbox").
		WithArgs(sqlmock.AnyArg(), dispatchoutbox.StatusPending, dispatchoutbox.StatusStaged).
		WillReturnResult(sqlmock.NewResult(0, 1))

	raw := mustSignalEnvelope(t, protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{Timestamp: now, SourceKind: protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.app-team.internal.trigger", PayloadKind: protocol.PayloadKindResult,
			TeamID: "app-team", RunID: runID},
		Payload: json.RawMessage(`{"context":{"work_item_id":"` + workID + `"},"summary":"Application ready","outputs":[{"output_id":"app","kind":"project_package","label":"Application","storage_ref":"groups/app-team/generated/app","entrypoint":"index.html","output_class":"user_deliverable"}]}`),
	})
	if err := (&teamWorkSignalProjection{server: s}).project(t.Context(), "swarm.team.app-team.signal.result", raw); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRuntimeValidationPassCreatesProofAndCompletesFinalRun(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	packageDir := filepath.Join(workspace, "groups", "app-team", "generated", "app")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "index.html"), []byte("<button>Run</button>"), 0o644); err != nil {
		t.Fatal(err)
	}
	digest, err := teamWorkOutputDigest([]protocol.TeamOutputRef{{StorageRef: "groups/app-team/generated/app"}})
	if err != nil {
		t.Fatal(err)
	}

	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	runID := "22222222-2222-2222-2222-222222222222"
	intentProofID := "33333333-3333-3333-3333-333333333333"
	contractID := "44444444-4444-4444-4444-444444444444"
	completionProofID := "55555555-5555-5555-5555-555555555555"
	outputRefs := []byte(`[{"output_id":"app","team_id":"app-team","work_item_id":"` + workID + `","kind":"project_package","label":"Application","storage_ref":"groups/app-team/generated/app","entrypoint":"index.html"}]`)
	workIntent := []byte(`{"kind":"project","output_contract":{"shape":"app_package","output_validation":{"kind":"interactive_browser","required":true,"checks":["load"],"probe":{"action":{"kind":"click","target":"button"},"observe":{"kind":"text_change","target":"button"}}}}}`)

	mock.MatchExpectationsInOrder(true)
	expectReviewingValidationItem(mock, "app-team", workID, runID, intentProofID, contractID, workIntent, outputRefs, now)
	mock.ExpectBegin()
	expectReviewingValidationItem(mock, "app-team", workID, runID, intentProofID, contractID, workIntent, outputRefs, now)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\).*FROM team_work_items").WithArgs(runID, workID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("INSERT INTO proof_artifacts").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(completionProofID))
	mock.ExpectExec("UPDATE execution_contracts").WillReturnResult(sqlmock.NewResult(0, 1))
	expectProjectedStatusEventInsertWithSource(mock, "app-team", workID, protocol.TeamWorkStateOutputReady,
		protocol.PayloadKindResult, string(protocol.SourceKindSystem), "team-work.runtime-validation", now)
	mock.ExpectExec("INSERT INTO mission_events").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE team_work_items").WithArgs(
		workID, string(protocol.TeamWorkStateOutputReady), sqlmock.AnyArg(), false, "",
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO team_interactions").WithArgs(
		sqlmock.AnyArg(), "app-team", workID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		string(protocol.SourceKindSystem), "team-work.runtime-validation", "soma", "output_ready", sqlmock.AnyArg(),
		string(protocol.PayloadKindResult), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "v1",
	).WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
	mock.ExpectExec("UPDATE mission_runs SET status = \\$1, completed_at = GREATEST").
		WithArgs(runs.StatusCompleted, runID, runs.StatusFailed).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO mission_events").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	payload := teamWorkValidationDispatchPayload{
		ContentDigest: digest, LaunchURL: "http://127.0.0.1:8081/api/v1/workspace/files/view",
		EvidenceRef: "groups/app-team/proof/runtime-validation/" + workID + "/digest",
	}
	report := outputvalidation.Report{
		Status: outputvalidation.StatusPassed, ContentDigest: digest, LaunchURL: payload.LaunchURL,
		StartedAt: now, FinishedAt: now,
	}
	outboxItem := &dispatchoutbox.Item{TeamID: "app-team", WorkItemID: workID, RunID: runID, IntentProofID: intentProofID, ContractID: contractID}
	if err := s.finalizeTeamWorkValidation(t.Context(), outboxItem, payload, report, true, ""); err != nil {
		t.Fatalf("finalize: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRuntimeValidationFailureDegradesWithoutProofOrRunCompletion(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	packageDir := filepath.Join(workspace, "groups", "app-team", "generated", "app")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "index.html"), []byte("<button>Run</button>"), 0o644); err != nil {
		t.Fatal(err)
	}
	digest, err := teamWorkOutputDigest([]protocol.TeamOutputRef{{StorageRef: "groups/app-team/generated/app"}})
	if err != nil {
		t.Fatal(err)
	}
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	runID := "22222222-2222-2222-2222-222222222222"
	intentProofID := "33333333-3333-3333-3333-333333333333"
	contractID := "44444444-4444-4444-4444-444444444444"
	outputRefs := []byte(`[{"output_id":"app","team_id":"app-team","work_item_id":"` + workID + `","kind":"project_package","label":"Application","storage_ref":"groups/app-team/generated/app","entrypoint":"index.html"}]`)
	workIntent := []byte(`{"kind":"project","output_contract":{"shape":"app_package","output_validation":{"kind":"interactive_browser","required":true,"checks":["load"],"probe":{"action":{"kind":"click","target":"button"},"observe":{"kind":"text_change","target":"button"}}}}}`)
	mock.MatchExpectationsInOrder(true)
	expectReviewingValidationItem(mock, "app-team", workID, runID, intentProofID, contractID, workIntent, outputRefs, now)
	mock.ExpectBegin()
	expectReviewingValidationItem(mock, "app-team", workID, runID, intentProofID, contractID, workIntent, outputRefs, now)
	expectProjectedStatusEventInsertWithSource(mock, "app-team", workID, protocol.TeamWorkStateDegraded,
		protocol.PayloadKindResult, string(protocol.SourceKindSystem), "team-work.runtime-validation", now)
	mock.ExpectExec("INSERT INTO mission_events").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE team_work_items").WithArgs(
		workID, string(protocol.TeamWorkStateDegraded), sqlmock.AnyArg(), true, "runtime_validation_failed",
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO team_interactions").WithArgs(
		sqlmock.AnyArg(), "app-team", workID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		string(protocol.SourceKindSystem), "team-work.runtime-validation", "soma", "degraded", sqlmock.AnyArg(),
		string(protocol.PayloadKindResult), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "v1",
	).WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
	mock.ExpectCommit()
	payload := teamWorkValidationDispatchPayload{ContentDigest: digest, LaunchURL: "http://127.0.0.1:8081/output", EvidenceRef: "groups/app-team/proof/runtime-validation/failed"}
	report := outputvalidation.Report{
		Status: outputvalidation.StatusFailed, ContentDigest: digest, LaunchURL: payload.LaunchURL,
		StartedAt: now, FinishedAt: now,
		Diagnostics: []outputvalidation.Diagnostic{{Code: "probe_observation_unchanged", Message: "The primary workflow did not change the output.", Severity: "error"}},
	}
	outboxItem := &dispatchoutbox.Item{TeamID: "app-team", WorkItemID: workID, RunID: runID, IntentProofID: intentProofID, ContractID: contractID}
	if err := s.finalizeTeamWorkValidation(t.Context(), outboxItem, payload, report, false, "runtime_validation_failed"); err != nil {
		t.Fatalf("finalize: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func expectReviewingValidationItem(mock sqlmock.Sqlmock, teamID, workID, runID, intentProofID, contractID string, workIntent, outputRefs []byte, now time.Time) {
	mock.ExpectQuery("SELECT id::text, team_id").WithArgs(teamID, workID).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, teamID, runID, intentProofID, contractID, "", "Build browser application", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), "team_async", workIntent, []byte(`["application package"]`), []byte(`["runtime proof"]`), []byte(`[]`),
			"approved", string(protocol.TeamWorkStateReviewing), []byte(`null`), false, "",
			[]byte(`[]`), outputRefs, []byte(`[]`), []byte(`[]`), now, now, "v1",
		))
}
