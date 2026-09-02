package server

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/workerauthority"
	"github.com/mycelis/core/internal/workers"
)

func TestProjectWorkerCompletedEventUsesReceiptAndCursorTransaction(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	s.WorkerAuthority = workerauthority.NewStore(s.getDB())
	correlation := projectionCorrelation()
	event := workers.WorkerEvent{
		EventID: "framework-event-1", RunID: correlation.RunID,
		Backend: workers.BackendFrameworkRuns, Sequence: 1, Version: 5,
		Correlation: correlation, Kind: workers.EventCompleted, Status: workers.StatusCompleted,
		Message:   "sentinel-upstream-secret",
		Result:    &workers.WorkerResult{Outputs: []workers.WorkerOutput{{Kind: "file", URI: "workspace/output.md"}}},
		Timestamp: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
	}
	payload := workerEventPayload(correlation, event)
	if payload["completion_authority"] != "candidate" || payload["requires_core_validation"] != true || payload["verified"] != false {
		t.Fatalf("completion trust posture = %#v", payload)
	}
	if strings.Contains(string(mustJSON(payload)), event.Message) {
		t.Fatal("normalized projection exposed an untrusted upstream message")
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT state").WithArgs(correlation.RunID).
		WillReturnRows(workerAuthorityRow(correlation, 4, 0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO worker_event_receipts")).
		WithArgs(sqlmock.AnyArg(), correlation.RunID, event.EventID, event.Sequence,
			string(event.Kind), event.Version, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO mission_events")).
		WithArgs(sqlmock.AnyArg(), correlation.RunID, "team_work.status", "info",
			correlation.TeamID, sqlmock.AnyArg(), event.Timestamp).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("31111111-1111-1111-1111-111111111111"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE mission_runs SET status=$2")).
		WithArgs(correlation.RunID, runs.StatusRunning, runs.StatusCompleted, runs.StatusFailed).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE worker_event_receipts").
		WithArgs(correlation.RunID, event.Sequence, "31111111-1111-1111-1111-111111111111").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE worker_run_bindings").
		WithArgs(correlation.RunID, event.Sequence, event.Version,
			string(event.Status), int64(0), int64(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	projected, err := s.projectWorkerEvent(t.Context(), correlation, event)
	if err != nil || !projected {
		t.Fatalf("projectWorkerEvent = %v, %v", projected, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProjectWorkerEventRejectsIdentityCorrelationAndCursorBeforeDatabase(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	correlation := projectionCorrelation()
	valid := workers.WorkerEvent{
		EventID: "event-1", RunID: correlation.RunID, Backend: workers.BackendFrameworkRuns,
		Sequence: 1, Version: 1, Correlation: correlation,
		Kind: workers.EventProgress, Status: workers.StatusRunning,
	}
	for name, mutate := range map[string]func(*workers.WorkerEvent){
		"backend":     func(e *workers.WorkerEvent) { e.Backend = workers.BackendCentral },
		"run":         func(e *workers.WorkerEvent) { e.RunID = "different-run" },
		"correlation": func(e *workers.WorkerEvent) { e.Correlation.WorkItemID = "different-work" },
		"sequence":    func(e *workers.WorkerEvent) { e.Sequence = 0 },
		"version":     func(e *workers.WorkerEvent) { e.Version = -1 },
	} {
		t.Run(name, func(t *testing.T) {
			event := valid
			mutate(&event)
			if _, err := s.projectWorkerEvent(t.Context(), correlation, event); err == nil {
				t.Fatal("expected event to fail closed")
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("invalid evidence reached database: %v", err)
	}
}

func projectionCorrelation() workers.WorkerCorrelation {
	return workers.WorkerCorrelation{
		RunID:               "11111111-1111-1111-1111-111111111111",
		IntentProofID:       "11111111-1111-1111-1111-111111111112",
		ExecutionContractID: "11111111-1111-1111-1111-111111111113",
		TeamID:              "delivery-team", WorkItemID: "11111111-1111-1111-1111-111111111114",
		OutcomeID: "outcome-1", IdempotencyKey: "confirm-action:proof-1",
		SourceKind: "web_api", SourceChannel: confirmedActionSourceChannel,
		PayloadKind: "command", GraphRevision: "graph-v1",
	}
}

func workerAuthorityRow(c workers.WorkerCorrelation, serviceVersion, lastSequence, cursorVersion int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"state", "proof", "contract", "team", "work", "outcome", "key", "graph", "source_kind", "source_channel", "payload_kind", "digest", "service_version", "last_sequence", "cursor_version"}).
		AddRow(workerauthority.BindingBound, c.IntentProofID, c.ExecutionContractID,
			c.TeamID, c.WorkItemID, c.OutcomeID, c.IdempotencyKey, c.GraphRevision,
			c.SourceKind, c.SourceChannel, c.PayloadKind,
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			serviceVersion, lastSequence, cursorVersion)
}
