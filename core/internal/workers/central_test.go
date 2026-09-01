package workers

import (
	"context"
	"testing"
)

func TestCentralBackendLifecycle(t *testing.T) {
	backend := NewCentralBackend()
	handle, err := backend.CreateRun(context.Background(), WorkerRunRequest{
		UserID: "operator-1",
		Intent: "create a retained output",
	})
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if handle.Backend != BackendCentral || handle.Status != StatusAccepted {
		t.Fatalf("handle backend/status = %s/%s", handle.Backend, handle.Status)
	}
	if handle.AuditRecord == nil || handle.AuditRecord.RunID != handle.RunID {
		t.Fatalf("expected audit record tied to run")
	}

	events, err := backend.StreamRunEvents(context.Background(), handle.RunID)
	if err != nil {
		t.Fatalf("StreamRunEvents: %v", err)
	}
	var sawCompleted bool
	for event := range events {
		if event.Kind == EventCompleted {
			sawCompleted = true
		}
	}
	if !sawCompleted {
		t.Fatal("expected completed event")
	}
	run, err := backend.GetRun(context.Background(), handle.RunID)
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if run.Status != StatusCompleted || run.Result == nil {
		t.Fatalf("run status/result = %s/%v", run.Status, run.Result)
	}
}

func TestCentralBackendApprovalDecision(t *testing.T) {
	backend := NewCentralBackend()
	handle, err := backend.CreateRun(context.Background(), WorkerRunRequest{Intent: "dangerous command"})
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	backend.mu.Lock()
	run := backend.runs[handle.RunID]
	run.Status = StatusApprovalNeeded
	run.Approval = &WorkerApprovalRequest{ID: "approval-1", Kind: "command", RiskLevel: "high"}
	backend.runs[handle.RunID] = run
	backend.mu.Unlock()

	err = backend.SubmitApproval(context.Background(), handle.RunID, WorkerApprovalDecision{
		ApprovalID: "approval-1",
		Decision:   DecisionDeny,
		ActorID:    "operator-1",
	})
	if err != nil {
		t.Fatalf("SubmitApproval: %v", err)
	}
	run, err = backend.GetRun(context.Background(), handle.RunID)
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if run.Status != StatusFailed || run.Error == nil || run.Error.Code != "approval_denied" {
		t.Fatalf("expected approval-denied failure, got %s/%v", run.Status, run.Error)
	}
}

func TestCentralBackendCompleteRun(t *testing.T) {
	backend := NewCentralBackend()
	handle, err := backend.CreateRun(context.Background(), WorkerRunRequest{Intent: "create output"})
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	err = backend.CompleteRun(context.Background(), handle.RunID, WorkerResult{
		Summary: "Output ready.",
		Outputs: []WorkerOutput{{
			ID:   "out-1",
			Kind: "file",
			Name: "report.md",
			URI:  "workspace/report.md",
		}},
	})
	if err != nil {
		t.Fatalf("CompleteRun: %v", err)
	}

	run, err := backend.GetRun(context.Background(), handle.RunID)
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if run.Status != StatusCompleted || run.Result == nil || len(run.Result.Outputs) != 1 {
		t.Fatalf("run status/result = %s/%+v", run.Status, run.Result)
	}
	if run.Result.FinishedAt.IsZero() {
		t.Fatal("expected finished timestamp")
	}
}

func TestCentralBackendFailRun(t *testing.T) {
	backend := NewCentralBackend()
	handle, err := backend.CreateRun(context.Background(), WorkerRunRequest{Intent: "create output"})
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	err = backend.FailRun(context.Background(), handle.RunID, &WorkerError{
		Code:        "tool_failed",
		Message:     "tool unavailable",
		Recoverable: true,
	})
	if err != nil {
		t.Fatalf("FailRun: %v", err)
	}

	run, err := backend.GetRun(context.Background(), handle.RunID)
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if run.Status != StatusFailed || run.Error == nil || run.Error.Code != "tool_failed" || !run.Error.Recoverable {
		t.Fatalf("run status/error = %s/%+v", run.Status, run.Error)
	}
}

func TestCentralBackendRejectsEmptyIntent(t *testing.T) {
	_, err := NewCentralBackend().CreateRun(context.Background(), WorkerRunRequest{})
	if err == nil {
		t.Fatal("expected empty intent error")
	}
}

func TestCentralBackendSameRunIDRequiresStableRequestFingerprint(t *testing.T) {
	backend := NewCentralBackend()
	req := WorkerRunRequest{RunID: "run-1", Intent: "create output", Metadata: map[string]any{"contract": "contract-1"}}
	first, err := backend.CreateRun(context.Background(), req)
	if err != nil {
		t.Fatalf("first CreateRun: %v", err)
	}
	second, err := backend.CreateRun(context.Background(), req)
	if err != nil || second.RunID != first.RunID {
		t.Fatalf("idempotent CreateRun = %#v, %v", second, err)
	}
	conflict := req
	conflict.Metadata = map[string]any{"contract": "contract-2"}
	if _, err := backend.CreateRun(context.Background(), conflict); err == nil {
		t.Fatal("same run_id with conflicting request must fail closed")
	}
}
