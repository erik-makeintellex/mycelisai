package workers

import (
	"context"
	"testing"
)

func TestCentralBackendStreamReturnsSnapshotWithoutFinalizing(t *testing.T) {
	backend := NewCentralBackend()
	req := correlatedTestRunRequest("run-1", "create a retained output")
	req.UserID = "operator-1"
	handle, err := backend.CreateRun(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if handle.Backend != BackendCentral || handle.Status != StatusAccepted {
		t.Fatalf("handle backend/status = %s/%s", handle.Backend, handle.Status)
	}
	if handle.Version != 1 || handle.Correlation != req.Correlation {
		t.Fatalf("central authority = %#v", handle)
	}
	if handle.AuditRecord == nil || handle.AuditRecord.RunID != handle.RunID {
		t.Fatalf("expected audit record tied to run")
	}

	events, err := backend.StreamRunEvents(context.Background(), handle.RunID)
	if err != nil {
		t.Fatalf("StreamRunEvents: %v", err)
	}
	event := <-events
	if event.EventID == "" || event.Kind != EventAccepted || event.Status != StatusAccepted {
		t.Fatalf("snapshot event = %#v", event)
	}
	if event.Sequence != 1 || event.Version != handle.Version || event.Correlation != handle.Correlation {
		t.Fatalf("snapshot cursor/correlation = %#v", event)
	}
	run, err := backend.GetRun(context.Background(), handle.RunID)
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if run.Status != StatusAccepted || run.Result != nil {
		t.Fatalf("run status/result = %s/%v", run.Status, run.Result)
	}
}

func TestCentralBackendCompleteRun(t *testing.T) {
	backend := NewCentralBackend()
	handle, err := backend.CreateRun(context.Background(), WorkerRunRequest{RunID: "run-complete", Intent: "create output"})
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
	if run.Version != 2 {
		t.Fatalf("completed run version = %d", run.Version)
	}
}

func TestCentralBackendFailRun(t *testing.T) {
	backend := NewCentralBackend()
	handle, err := backend.CreateRun(context.Background(), WorkerRunRequest{RunID: "run-fail", Intent: "create output"})
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
	if run.Version != 2 {
		t.Fatalf("failed run version = %d", run.Version)
	}
}

func TestCentralBackendRejectsEmptyIntent(t *testing.T) {
	_, err := NewCentralBackend().CreateRun(context.Background(), WorkerRunRequest{RunID: "run-1"})
	if err == nil {
		t.Fatal("expected empty intent error")
	}
}

func TestCentralBackendRequiresCoreRunID(t *testing.T) {
	_, err := NewCentralBackend().CreateRun(context.Background(), WorkerRunRequest{Intent: "create output"})
	if err == nil {
		t.Fatal("expected missing run_id error")
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

func TestCentralBackendDoesNotAdvertiseUnreachableApprovalFlow(t *testing.T) {
	caps, err := NewCentralBackend().GetCapabilities(context.Background())
	if err != nil || caps.SupportsApprovals {
		t.Fatalf("central capabilities = %#v, %v", caps, err)
	}
}
