//go:build postgres_integration

package journal_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mycelis/framework-runs/internal/controller"
	"github.com/mycelis/framework-runs/internal/journal"
	"github.com/mycelis/framework-runs/internal/protocol"
)

type recoveryExecutor struct {
	mu       sync.Mutex
	effects  map[string]int
	outcomes map[string]protocol.ExecutorOutcome
	starts   map[string]protocol.ExecutorOutcome
}

func (executor *recoveryExecutor) Apply(_ context.Context, command journal.Command) (protocol.ExecutorOutcome, error) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	if outcome, exists := executor.outcomes[command.CommandID]; exists {
		return protocol.Clone(outcome), nil
	}
	executor.effects[command.CommandID]++
	var outcome protocol.ExecutorOutcome
	switch command.Kind {
	case "start":
		outcome = executor.starts[command.RunID]
		if outcome.Status == "" {
			outcome = completedPostgresOutcome(command.RunID)
		}
	case "stop":
		outcome = protocol.ExecutorOutcome{Status: protocol.StatusCancelled, Message: "Cancelled."}
	case "approve":
		outcome = completedPostgresOutcome(command.RunID)
	case "deny":
		outcome = protocol.ExecutorOutcome{Status: protocol.StatusFailed,
			Error: &protocol.Error{Code: "approval_denied", Message: "Denied.", Recoverable: false}}
	}
	executor.outcomes[command.CommandID] = protocol.Clone(outcome)
	return outcome, nil
}

func TestPostgresAcceptedCommitRestartReplayConflictAndCapacity(t *testing.T) {
	ctx := context.Background()
	databaseURL, _ := newDisposableDatabase(t, ctx)
	executor := newRecoveryExecutor()
	repository, err := journal.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	service := controller.New(repository, executor)
	service.MaxRuns = 1
	request := postgresRequest("run-accepted-restart")
	accepted, replay, err := service.Create(ctx, request)
	if err != nil || replay || accepted.Version != 1 {
		t.Fatalf("accepted commit = %#v, %v, %v", accepted, replay, err)
	}
	repository.Close()
	repository, err = journal.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	service = controller.New(repository, executor)
	service.MaxRuns = 1
	if same, replay, err := service.Create(ctx, request); err != nil || !replay || same.RunID != request.RunID {
		t.Fatalf("create replay = %#v, %v, %v", same, replay, err)
	}
	changed := request
	changed.Intent = "different"
	if _, _, err := service.Create(ctx, changed); !errors.Is(err, journal.ErrRunConflict) {
		t.Fatalf("create conflict = %v", err)
	}
	if _, _, err := service.Create(ctx, postgresRequest("run-over-capacity")); !errors.Is(err, journal.ErrCapacity) {
		t.Fatalf("capacity = %v", err)
	}
	if processed, err := service.DispatchOnce(ctx, "restart-worker"); err != nil || !processed {
		t.Fatalf("restart dispatch = %v, %v", processed, err)
	}
}

func TestPostgresCrashAfterStopEffectReusesCommandWithoutDuplicate(t *testing.T) {
	ctx := context.Background()
	databaseURL, _ := newDisposableDatabase(t, ctx)
	clock := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	executor := newRecoveryExecutor()
	executor.starts["run-stop-crash"] = protocol.ExecutorOutcome{Status: protocol.StatusRunning}
	repository, err := journal.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	service := controller.New(repository, executor)
	service.Now = func() time.Time { return clock }
	service.LeaseDuration = time.Second
	request := postgresRequest("run-stop-crash")
	_, _, _ = service.Create(ctx, request)
	_, _ = service.DispatchOnce(ctx, "worker-a")
	stop := protocol.StopRequest{CommandID: "stop-crash", ExpectedVersion: 2, ActorID: "core"}
	if _, replay, err := service.Stop(ctx, request.RunID, stop, "mycelis-core"); err != nil || replay {
		t.Fatalf("stage stop = replay %v, error %v", replay, err)
	}
	service.Fault = func(controller.FaultPoint, journal.Command) error { return errors.New("crash after effect") }
	if processed, err := service.DispatchOnce(ctx, "worker-a"); !processed || err == nil {
		t.Fatalf("stop crash = %v, %v", processed, err)
	}
	repository.Close()
	clock = clock.Add(2 * time.Second)
	repository, err = journal.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	service = controller.New(repository, executor)
	service.Now = func() time.Time { return clock }
	service.LeaseDuration = time.Second
	if processed, err := service.DispatchOnce(ctx, "worker-b"); err != nil || !processed {
		t.Fatalf("stop recovery = %v, %v", processed, err)
	}
	receipt, replay, err := service.Stop(ctx, request.RunID, stop, "mycelis-core")
	if err != nil || !replay || receipt.State != journal.CommandApplied || receipt.Version != 3 {
		t.Fatalf("stop receipt replay = %#v, %v, %v", receipt, replay, err)
	}
	if executor.effects[stop.CommandID] != 1 {
		t.Fatalf("stop side effect count = %d", executor.effects[stop.CommandID])
	}
}

func TestPostgresCrashAfterApprovalEffectReusesCommandWithoutDuplicate(t *testing.T) {
	ctx := context.Background()
	databaseURL, _ := newDisposableDatabase(t, ctx)
	clock := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	executor := newRecoveryExecutor()
	executor.starts["run-approval-crash"] = protocol.ExecutorOutcome{
		Status: protocol.StatusApprovalNeeded, Approval: &protocol.Approval{
			ID: "approval-1", Kind: "tool", Summary: "Use tool", RiskLevel: "low", RequestedAction: "invoke",
		},
	}
	repository, err := journal.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	service := controller.New(repository, executor)
	service.Now = func() time.Time { return clock }
	service.LeaseDuration = time.Second
	request := postgresRequest("run-approval-crash")
	_, _, _ = service.Create(ctx, request)
	_, _ = service.DispatchOnce(ctx, "worker-a")
	decision := protocol.ApprovalDecisionRequest{
		ApprovalID: "approval-1", Decision: "approve", CommandID: "approval-crash",
		ExpectedVersion: 2, ActorID: "operator-1",
	}
	if _, replay, err := service.Decide(ctx, request.RunID, decision.ApprovalID, decision, "mycelis-core"); err != nil || replay {
		t.Fatalf("stage approval = replay %v, error %v", replay, err)
	}
	service.Fault = func(controller.FaultPoint, journal.Command) error { return errors.New("crash after effect") }
	if processed, err := service.DispatchOnce(ctx, "worker-a"); !processed || err == nil {
		t.Fatalf("approval crash = %v, %v", processed, err)
	}
	repository.Close()
	clock = clock.Add(2 * time.Second)
	repository, err = journal.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	service = controller.New(repository, executor)
	service.Now = func() time.Time { return clock }
	service.LeaseDuration = time.Second
	if processed, err := service.DispatchOnce(ctx, "worker-b"); err != nil || !processed {
		t.Fatalf("approval recovery = %v, %v", processed, err)
	}
	receipt, replay, err := service.Decide(ctx, request.RunID, decision.ApprovalID, decision, "mycelis-core")
	if err != nil || !replay || receipt.State != journal.CommandApplied || receipt.Version != 3 {
		t.Fatalf("approval receipt replay = %#v, %v, %v", receipt, replay, err)
	}
	if executor.effects[decision.CommandID] != 1 {
		t.Fatalf("approval side effect count = %d", executor.effects[decision.CommandID])
	}
}

func newRecoveryExecutor() *recoveryExecutor {
	return &recoveryExecutor{
		effects: map[string]int{}, outcomes: map[string]protocol.ExecutorOutcome{},
		starts: map[string]protocol.ExecutorOutcome{},
	}
}

func completedPostgresOutcome(runID string) protocol.ExecutorOutcome {
	return protocol.ExecutorOutcome{Status: protocol.StatusCompleted, Result: &protocol.Result{
		Summary: "Candidate ready.", FinishedAt: time.Date(2026, 9, 2, 12, 0, 1, 0, time.UTC),
		Outputs: []protocol.Output{{
			ID: "output-1", Kind: "document", URI: "candidate://" + runID + "/output-1",
			ContentType: "application/json", SizeBytes: 2,
			SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		}},
	}}
}
