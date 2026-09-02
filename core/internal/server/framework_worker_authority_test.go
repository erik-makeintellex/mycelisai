package server

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/workerauthority"
	"github.com/mycelis/core/internal/workers"
	"github.com/mycelis/core/pkg/protocol"
)

func TestStageFrameworkRunCreateIsTransactionOnlyAndNonDispatchable(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	s.WorkerAuthority = workerauthority.NewStore(s.getDB())
	s.DispatchOutbox = dispatchoutbox.NewStore(s.getDB())
	s.WorkerBackend = workers.NewUnavailableBackend(errors.New("worker HTTP must not be called"))
	correlation := projectionCorrelation()
	request := buildConfirmedActionWorkerRunRequest(&protocol.ScopeValidation{
		WorkIntent: &protocol.WorkIntent{Objective: "Build the approved package"},
	}, "operator", correlation)

	mock.ExpectBegin()
	tx, err := s.getDB().BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO worker_run_bindings")).
		WithArgs(correlation.RunID, correlation.IntentProofID, correlation.ExecutionContractID,
			correlation.TeamID, correlation.WorkItemID, correlation.OutcomeID,
			correlation.IdempotencyKey, correlation.GraphRevision, correlation.SourceKind,
			correlation.SourceChannel, correlation.PayloadKind, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"run_id"}).AddRow(correlation.RunID))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO execution_dispatch_outbox")).
		WithArgs(sqlmock.AnyArg(), "framework-run-create:"+correlation.IdempotencyKey,
			dispatchoutbox.DispatchKindFrameworkRunCreate, dispatchoutbox.StatusAwaitingHandler,
			correlation.RunID, correlation.IntentProofID, correlation.ExecutionContractID,
			correlation.TeamID, correlation.WorkItemID, correlation.SourceKind,
			correlation.SourceChannel, correlation.PayloadKind, sqlmock.AnyArg(),
			sqlmock.AnyArg(), `{"action":"activate_framework_run_create_after_slice_c","operator_required":false}`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("dispatch-1"))
	created, err := s.stageFrameworkRunCreateTx(t.Context(), tx, request)
	if err != nil || !created {
		t.Fatalf("stage external create = %v, %v", created, err)
	}
	mock.ExpectRollback()
	_ = tx.Rollback()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
