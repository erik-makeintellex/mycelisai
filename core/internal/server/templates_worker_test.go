package server

import (
	"context"
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
	return b.handle, nil
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

func TestCreateExecutionRunTxUsesWorkerRunIDAndMetadata(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	workerRunID := "45555555-5555-5555-5555-555555555555"
	backend := &recordingWorkerBackend{
		handle: workers.WorkerRunHandle{
			RunID:   workerRunID,
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
		WithArgs(workerRunID, proofID, "default", runs.StatusRunning, 0, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	tx, err := s.getDB().BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	runID, err := s.createExecutionRunTx(context.Background(), tx, proofID, scope, "operator")
	if err != nil {
		t.Fatalf("createExecutionRunTx: %v", err)
	}
	if runID != workerRunID {
		t.Fatalf("runID = %q, want worker run id %q", runID, workerRunID)
	}
	if backend.request.Intent != "Build a browser app" {
		t.Fatalf("worker intent = %q", backend.request.Intent)
	}
	if backend.request.Metadata["intent_proof_id"] != proofID {
		t.Fatalf("worker metadata = %#v", backend.request.Metadata)
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
