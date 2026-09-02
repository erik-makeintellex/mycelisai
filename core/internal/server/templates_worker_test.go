package server

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/workers"
	"github.com/mycelis/core/pkg/protocol"
)

type recordingWorkerBackend struct {
	request workers.WorkerRunRequest
	handle  workers.WorkerRunHandle
}

func (b *recordingWorkerBackend) CreateRun(_ context.Context, req workers.WorkerRunRequest) (workers.WorkerRunHandle, error) {
	b.request = req
	handle := b.handle
	if handle.RunID == "" {
		handle.RunID = req.RunID
	}
	if handle.BackendRunID == "" {
		handle.BackendRunID = handle.RunID
	}
	return handle, nil
}

func (b *recordingWorkerBackend) StreamRunEvents(context.Context, string) (<-chan workers.WorkerEvent, error) {
	panic("not used")
}

func (b *recordingWorkerBackend) GetRun(context.Context, string) (workers.WorkerRunHandle, error) {
	panic("not used")
}

func (b *recordingWorkerBackend) StopRun(context.Context, string) error {
	panic("not used")
}

func (b *recordingWorkerBackend) SubmitApproval(context.Context, string, workers.WorkerApprovalDecision) error {
	panic("not used")
}

func (b *recordingWorkerBackend) GetCapabilities(context.Context) (workers.WorkerCapabilities, error) {
	panic("not used")
}

func (b *recordingWorkerBackend) HealthCheck(context.Context) (workers.WorkerHealth, error) {
	panic("not used")
}

func (b *recordingWorkerBackend) CompleteRun(context.Context, string, workers.WorkerResult) error {
	return nil
}

func (b *recordingWorkerBackend) FailRun(context.Context, string, *workers.WorkerError) error {
	return nil
}

func TestCreateExecutionRunTxMintsBeforeWorkerInitializationAndCarriesCorrelation(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	backend := &recordingWorkerBackend{
		handle: workers.WorkerRunHandle{
			Backend: workers.BackendCentral,
			Status:  workers.StatusAccepted,
		},
	}
	s.WorkerBackend = backend
	proofID := "46666666-6666-6666-6666-666666666666"
	scope := &protocol.ScopeValidation{
		AffectedResources: []string{"workspace"},
		RiskLevel:         "medium",
		PlannedToolCalls:  []protocol.PlannedToolCall{{Name: "delegate_task"}},
		WorkIntent:        &protocol.WorkIntent{Objective: "Build a browser app"},
		ExecutionMode:     "team_async",
		CapabilityIDs:     []string{"team.coordinate"},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO mission_runs").
		WithArgs(sqlmock.AnyArg(), proofID, "default", runs.StatusRunning, 0, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	tx, err := s.getDB().BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	runID, err := s.createExecutionRunTx(context.Background(), tx, proofID)
	if err != nil {
		t.Fatalf("createExecutionRunTx: %v", err)
	}
	if runID == "" {
		t.Fatal("Core did not mint run identity")
	}
	if backend.request.RunID != "" {
		t.Fatal("worker backend was called before explicit initialization")
	}
	contractID := "47777777-7777-7777-7777-777777777777"
	scope = correlateConfirmedActionScope(scope, runID, proofID, contractID)
	correlation := confirmedActionWorkerCorrelation(runID, proofID, contractID, scope, &protocol.OutcomeProject{OutcomeID: "outcome-1"})
	if _, err := s.startConfirmedActionWorkerRun(context.Background(), scope, "operator", correlation); err != nil {
		t.Fatalf("startConfirmedActionWorkerRun: %v", err)
	}
	if backend.request.Intent != "Build a browser app" {
		t.Fatalf("worker intent = %q", backend.request.Intent)
	}
	for _, duplicate := range []string{"run_id", "intent_proof_id", "execution_contract_id", "idempotency_key", "source_kind", "source_channel", "payload_kind", "graph_revision", "team_id", "outcome_id", "work_item_id"} {
		if _, found := backend.request.Metadata[duplicate]; found {
			t.Fatalf("duplicate correlation metadata %q = %#v", duplicate, backend.request.Metadata)
		}
	}
	if backend.request.RunID != runID || backend.request.Correlation.RunID != runID {
		t.Fatalf("run correlation = %#v", backend.request)
	}
	if backend.request.Correlation.ExecutionContractID != correlation.ExecutionContractID || backend.request.Correlation.OutcomeID != "outcome-1" || backend.request.Correlation.IdempotencyKey != "confirm-action:"+proofID {
		t.Fatalf("explicit correlation = %#v", backend.request.Correlation)
	}
	if backend.request.Correlation.WorkItemID == "" || backend.request.Correlation.SourceKind != string(protocol.SourceKindWebAPI) || backend.request.Correlation.SourceChannel != confirmedActionSourceChannel || backend.request.Correlation.PayloadKind != string(protocol.PayloadKindCommand) || backend.request.Correlation.GraphRevision != centralExecutionContractGraphRevision {
		t.Fatalf("execution boundary correlation = %#v", backend.request.Correlation)
	}
	if backend.request.Metadata["planned_tool_count"] != 1 {
		t.Fatalf("planned_tool_count = %#v", backend.request.Metadata["planned_tool_count"])
	}
	if backend.request.Metadata["execution_mode"] != "team_async" {
		t.Fatalf("execution_mode = %#v", backend.request.Metadata["execution_mode"])
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet db expectations: %v", err)
	}
}

func TestStartConfirmedActionWorkerRunRejectsNonFinalizingBackendBeforeCreate(t *testing.T) {
	s := newTestServer()
	s.WorkerBackend = workers.NewUnavailableBackend(errors.New("sentinel CreateRun reached"))
	correlation := workers.WorkerCorrelation{RunID: "run-1"}
	_, err := s.startConfirmedActionWorkerRun(t.Context(), nil, "operator", correlation)
	if err == nil || !strings.Contains(err.Error(), "lacks central finalization authority") || strings.Contains(err.Error(), "sentinel") {
		t.Fatalf("start error = %v", err)
	}
}
