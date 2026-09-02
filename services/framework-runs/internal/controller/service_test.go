package controller

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mycelis/framework-runs/internal/journal"
	"github.com/mycelis/framework-runs/internal/protocol"
)

type idempotentExecutor struct {
	mu       sync.Mutex
	effects  map[string]int
	outcomes map[string]protocol.ExecutorOutcome
}

type failingExecutor struct{}

func (failingExecutor) Apply(context.Context, journal.Command) (protocol.ExecutorOutcome, error) {
	return protocol.ExecutorOutcome{}, errors.New("executor unavailable")
}

func (fake *idempotentExecutor) Apply(_ context.Context, command journal.Command) (protocol.ExecutorOutcome, error) {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if outcome, ok := fake.outcomes[command.CommandID]; ok {
		return protocol.Clone(outcome), nil
	}
	fake.effects[command.CommandID]++
	var outcome protocol.ExecutorOutcome
	switch command.Kind {
	case "start":
		outcome = completedOutcome(command.RunID)
	case "stop":
		outcome = protocol.ExecutorOutcome{Status: protocol.StatusCancelled, Message: "Cancelled."}
	case "approve":
		outcome = completedOutcome(command.RunID)
	case "deny":
		outcome = protocol.ExecutorOutcome{Status: protocol.StatusFailed, Error: &protocol.Error{Code: "denied", Message: "Denied.", Recoverable: false}}
	}
	fake.outcomes[command.CommandID] = protocol.Clone(outcome)
	return outcome, nil
}

func TestCreateRequiresExecutorBeforeDurableAcceptance(t *testing.T) {
	repository := journal.NewMemoryRepository()
	service := New(repository, nil)
	_, _, err := service.Create(context.Background(), validRequest("run-no-executor"))
	if !errors.Is(err, ErrExecutorUnavailable) {
		t.Fatalf("got %v", err)
	}
	if _, err := repository.Get(context.Background(), "run-no-executor"); !errors.Is(err, journal.ErrNotFound) {
		t.Fatal("unready create durably accepted a run")
	}
}

func TestControlsRequireExecutorBeforeDurableAcceptance(t *testing.T) {
	repository := journal.NewMemoryRepository()
	service := New(repository, &idempotentExecutor{effects: map[string]int{}, outcomes: map[string]protocol.ExecutorOutcome{}})
	request := validRequest("run-retained")
	_, _, _ = service.Create(context.Background(), request)
	withoutExecutor := New(repository, nil)
	stop := protocol.StopRequest{CommandID: "stop-unready", ExpectedVersion: 1, ActorID: "core"}
	if _, _, err := withoutExecutor.Stop(context.Background(), request.RunID, stop, "core"); !errors.Is(err, ErrExecutorUnavailable) {
		t.Fatalf("stop without executor = %v", err)
	}
	if len(repository.Receipts()) != 0 {
		t.Fatal("unready control was persisted")
	}
}

func TestCreateReplayConflictCursorAndCAS(t *testing.T) {
	clock := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	repository := journal.NewMemoryRepository()
	service := New(repository, &idempotentExecutor{effects: map[string]int{}, outcomes: map[string]protocol.ExecutorOutcome{}})
	service.Now = func() time.Time { return clock }
	request := validRequest("run-replay")
	first, replay, err := service.Create(context.Background(), request)
	if err != nil || replay || first.Version != 1 {
		t.Fatalf("first create: %#v %v %v", first, replay, err)
	}
	second, replay, err := service.Create(context.Background(), request)
	if err != nil || !replay || second.RunID != first.RunID {
		t.Fatalf("replay: %#v %v %v", second, replay, err)
	}
	changed := request
	changed.Intent = "Different semantic request"
	if _, _, err := service.Create(context.Background(), changed); !errors.Is(err, journal.ErrConflict) {
		t.Fatalf("conflicting replay: %v", err)
	}
	if _, err := service.Events(context.Background(), request.RunID, 2); !errors.Is(err, journal.ErrCursorGap) {
		t.Fatalf("ahead cursor: %v", err)
	}
	stop := protocol.StopRequest{CommandID: "stop-1", ExpectedVersion: 2, ActorID: "core"}
	if _, _, err := service.Stop(context.Background(), request.RunID, stop, "core"); !errors.Is(err, journal.ErrConflict) {
		t.Fatalf("stale version accepted: %v", err)
	}
}

func TestCrashAfterExecutorEffectIsReconciledWithoutDuplicate(t *testing.T) {
	clock := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	repository := journal.NewMemoryRepository()
	fake := &idempotentExecutor{effects: map[string]int{}, outcomes: map[string]protocol.ExecutorOutcome{}}
	service := New(repository, fake)
	service.Now = func() time.Time { return clock }
	service.LeaseDuration = time.Second
	run, _, err := service.Create(context.Background(), validRequest("run-crash"))
	if err != nil {
		t.Fatal(err)
	}
	service.Fault = func(point FaultPoint, _ journal.Command) error { return errors.New(string(point)) }
	if processed, err := service.DispatchOnce(context.Background(), "worker-a"); !processed || err == nil {
		t.Fatalf("fault was not observed: %v %v", processed, err)
	}
	clock = clock.Add(2 * time.Second)
	restarted := New(repository, fake)
	restarted.Now = func() time.Time { return clock }
	restarted.LeaseDuration = time.Second
	if processed, err := restarted.DispatchOnce(context.Background(), "worker-b"); !processed || err != nil {
		t.Fatalf("restart dispatch: %v %v", processed, err)
	}
	snapshot, _ := restarted.Get(context.Background(), run.RunID)
	if snapshot.Status != protocol.StatusCompleted || snapshot.Version != 2 {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	startID := startCommandID(run.RunID)
	if fake.effects[startID] != 1 {
		t.Fatalf("executor effect count = %d", fake.effects[startID])
	}
	events, _ := restarted.Events(context.Background(), run.RunID, 0)
	if len(events) != 2 || events[0].Sequence != 1 || events[1].Sequence != 2 || events[1].Version != 2 {
		t.Fatalf("unexpected journal: %#v", events)
	}
}

func TestControlResponseLossReplayUsesSameReceipt(t *testing.T) {
	clock := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	repository := journal.NewMemoryRepository()
	fake := &idempotentExecutor{effects: map[string]int{}, outcomes: map[string]protocol.ExecutorOutcome{}}
	service := New(repository, fake)
	service.Now = func() time.Time { return clock }
	request := validRequest("run-stop")
	_, _, _ = service.Create(context.Background(), request)
	_, _ = service.DispatchOnce(context.Background(), "worker")
	// Use a nonterminal running run for stop semantics.
	repository = journal.NewMemoryRepository()
	fake.outcomes = map[string]protocol.ExecutorOutcome{}
	fake.effects = map[string]int{}
	fake.outcomes[startCommandID(request.RunID)] = protocol.ExecutorOutcome{Status: protocol.StatusRunning}
	service = New(repository, fake)
	service.Now = func() time.Time { return clock }
	_, _, _ = service.Create(context.Background(), request)
	_, _ = service.DispatchOnce(context.Background(), "worker")
	control := protocol.StopRequest{CommandID: "stop-response-loss", ExpectedVersion: 2, ActorID: "core"}
	first, replay, err := service.Stop(context.Background(), request.RunID, control, "core")
	if err != nil || replay {
		t.Fatalf("new stop: %#v %v %v", first, replay, err)
	}
	_, _ = service.DispatchOnce(context.Background(), "worker")
	second, replay, err := service.Stop(context.Background(), request.RunID, control, "core")
	if err != nil || !replay || second.State != journal.CommandApplied || second.Version != 3 {
		t.Fatalf("replayed stop: %#v %v %v", second, replay, err)
	}
	if fake.effects[control.CommandID] != 1 {
		t.Fatalf("stop effect count = %d", fake.effects[control.CommandID])
	}
}

func TestExhaustedStopFailsCommandWithoutWedgingRun(t *testing.T) {
	clock := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	repository := journal.NewMemoryRepository()
	start := &idempotentExecutor{effects: map[string]int{}, outcomes: map[string]protocol.ExecutorOutcome{}}
	request := validRequest("run-stop-failure")
	start.outcomes[startCommandID(request.RunID)] = protocol.ExecutorOutcome{Status: protocol.StatusRunning}
	service := New(repository, start)
	service.Now = func() time.Time { return clock }
	_, _, _ = service.Create(context.Background(), request)
	_, _ = service.DispatchOnce(context.Background(), "worker")
	stop := protocol.StopRequest{CommandID: "stop-fails", ExpectedVersion: 2, ActorID: "core"}
	_, _, _ = service.Stop(context.Background(), request.RunID, stop, "core")
	service.Executor, service.MaxAttempts = failingExecutor{}, 1
	processed, err := service.DispatchOnce(context.Background(), "worker")
	if err != nil || !processed {
		t.Fatalf("failed command persistence = %v, %v", processed, err)
	}
	receipt, replay, err := service.Stop(context.Background(), request.RunID, stop, "core")
	if err != nil || !replay || receipt.State != journal.CommandFailed || receipt.Error == nil {
		t.Fatalf("failed command replay = %#v, %v, %v", receipt, replay, err)
	}
	followup := stop
	followup.CommandID = "stop-retry-owned-by-core"
	if _, replay, err := service.Stop(context.Background(), request.RunID, followup, "core"); err != nil || replay {
		t.Fatalf("run remained wedged after failed command: replay=%v err=%v", replay, err)
	}
}

func validRequest(runID string) protocol.CreateRequest {
	return protocol.CreateRequest{RunID: runID, Intent: "Produce a candidate", Correlation: protocol.Correlation{
		RunID: runID, IntentProofID: "proof-1", ExecutionContractID: "contract-1",
		WorkItemID: "work-1", IdempotencyKey: "idem:" + runID, SourceKind: "web_api",
		SourceChannel: "api.intent", PayloadKind: "command", GraphRevision: "graph-1",
	}}
}

func completedOutcome(runID string) protocol.ExecutorOutcome {
	return protocol.ExecutorOutcome{Status: protocol.StatusCompleted, Result: &protocol.Result{
		Summary: "Candidate ready.", FinishedAt: time.Date(2026, 9, 2, 12, 0, 1, 0, time.UTC),
		Outputs: []protocol.Output{{ID: "output-1", Kind: "document", URI: "candidate://" + runID + "/output-1",
			ContentType: "application/json", SizeBytes: 2,
			SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}},
	}}
}
