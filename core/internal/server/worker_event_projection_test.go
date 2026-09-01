package server

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/workers"
)

func TestProjectWorkerCompletedEventIsIdempotentCandidateEvidence(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	correlation := workers.WorkerCorrelation{
		RunID:               "11111111-1111-1111-1111-111111111111",
		IntentProofID:       "22222222-2222-2222-2222-222222222222",
		ExecutionContractID: "33333333-3333-3333-3333-333333333333",
		TeamID:              "delivery-team",
		WorkItemID:          "work-item-1",
		OutcomeID:           "outcome-1",
		IdempotencyKey:      "confirm-action:22222222-2222-2222-2222-222222222222",
		SourceKind:          "web_api",
		SourceChannel:       "api.intent.confirm-action",
		PayloadKind:         "command",
		GraphRevision:       "mycelis.central.execution-contract.v1",
	}
	event := workers.WorkerEvent{
		EventID: "framework-event-1", RunID: "external-run-1", BackendRunID: "external-run-1", Backend: workers.BackendFrameworkRuns,
		Kind: workers.EventCompleted, Status: workers.StatusCompleted, Message: "sentinel-upstream-secret",
		Result:    &workers.WorkerResult{Outputs: []workers.WorkerOutput{{Kind: "file", URI: "workspace/output.md"}}},
		Timestamp: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
	}
	projectedPayload := workerEventPayload(correlation, "external-run-1", event)
	if projectedPayload["completion_authority"] != "candidate" || projectedPayload["requires_core_validation"] != true || projectedPayload["verified"] != false {
		t.Fatalf("completion trust posture = %#v", projectedPayload)
	}
	if projectedPayload["work_item_id"] != correlation.WorkItemID || projectedPayload["source_kind"] != correlation.SourceKind || projectedPayload["source_channel"] != correlation.SourceChannel || projectedPayload["payload_kind"] != correlation.PayloadKind || projectedPayload["graph_revision"] != correlation.GraphRevision {
		t.Fatalf("projected correlation payload = %#v", projectedPayload)
	}
	payloadJSON, err := json.Marshal(projectedPayload)
	if err != nil {
		t.Fatalf("marshal projected payload: %v", err)
	}
	if strings.Contains(string(payloadJSON), "sentinel-upstream-secret") {
		t.Fatalf("projected payload exposed raw upstream message: %s", payloadJSON)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO mission_events")).
		WithArgs(sqlmock.AnyArg(), correlation.RunID, "team_work.status", "info", correlation.TeamID, sqlmock.AnyArg(), event.Timestamp).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("projection-1"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE mission_runs SET status=$2")).
		WithArgs(correlation.RunID, runs.StatusRunning, runs.StatusCompleted, runs.StatusFailed).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	projected, err := s.projectWorkerEvent(t.Context(), correlation, "external-run-1", event)
	if err != nil || !projected {
		t.Fatalf("projectWorkerEvent = %v, %v", projected, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO mission_events")).
		WithArgs(sqlmock.AnyArg(), correlation.RunID, "team_work.status", "info", correlation.TeamID, sqlmock.AnyArg(), event.Timestamp).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()
	projected, err = s.projectWorkerEvent(t.Context(), correlation, "external-run-1", event)
	if err != nil || projected {
		t.Fatalf("duplicate projectWorkerEvent = %v, %v", projected, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestProjectWorkerEventRejectsNonFrameworkBackend(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	correlation := workers.WorkerCorrelation{
		RunID: "run-1", IntentProofID: "proof-1", ExecutionContractID: "contract-1", WorkItemID: "work-1", IdempotencyKey: "dispatch-1",
		SourceKind: "web_api", SourceChannel: "api.intent.confirm-action", PayloadKind: "command", GraphRevision: "graph-v1",
	}
	_, err := s.projectWorkerEvent(t.Context(), correlation, "run-1", workers.WorkerEvent{
		EventID: "event-1", RunID: "run-1", Backend: workers.BackendCentral, Kind: workers.EventCompleted,
	})
	if err == nil {
		t.Fatal("expected non-framework worker event to fail closed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected database call: %v", err)
	}
}

func TestProjectWorkerEventRejectsBackendRunBindingMismatch(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	correlation := workers.WorkerCorrelation{
		RunID: "run-1", IntentProofID: "proof-1", ExecutionContractID: "contract-1", WorkItemID: "work-1", IdempotencyKey: "dispatch-1",
		SourceKind: "web_api", SourceChannel: "api.intent.confirm-action", PayloadKind: "command", GraphRevision: "graph-v1",
	}
	_, err := s.projectWorkerEvent(t.Context(), correlation, "bound-run", workers.WorkerEvent{
		EventID: "event-1", RunID: "bound-run", BackendRunID: "different-run", Backend: workers.BackendFrameworkRuns, Kind: workers.EventProgress,
	})
	if err == nil || !strings.Contains(err.Error(), "backend_run_id") {
		t.Fatalf("binding mismatch error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected database call: %v", err)
	}
}

func TestProjectWorkerFailedEvidenceDoesNotRewriteMissionStatus(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	correlation := workers.WorkerCorrelation{
		RunID: "run-1", IntentProofID: "proof-1", ExecutionContractID: "contract-1", WorkItemID: "work-1", IdempotencyKey: "dispatch-1",
		SourceKind: "web_api", SourceChannel: "api.intent.confirm-action", PayloadKind: "command", GraphRevision: "graph-v1",
	}
	event := workers.WorkerEvent{
		EventID: "failed-1", RunID: "run-1", BackendRunID: "run-1", Backend: workers.BackendFrameworkRuns,
		Kind: workers.EventFailed, Status: workers.StatusFailed, Timestamp: time.Date(2026, 9, 1, 13, 0, 0, 0, time.UTC),
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO mission_events")).
		WithArgs(sqlmock.AnyArg(), correlation.RunID, "team_work.status", "warn", "", sqlmock.AnyArg(), event.Timestamp).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("projection-failed"))
	mock.ExpectCommit()
	projected, err := s.projectWorkerEvent(t.Context(), correlation, "run-1", event)
	if err != nil || !projected {
		t.Fatalf("failed evidence projection = %v, %v", projected, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("failed evidence rewrote mission status: %v", err)
	}
}
