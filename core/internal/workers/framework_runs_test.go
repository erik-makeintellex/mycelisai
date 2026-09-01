package workers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mapSecretResolver map[string]string

func (m mapSecretResolver) ResolveSecret(_ context.Context, ref string) (string, error) {
	return m[ref], nil
}

func TestFrameworkRunsBackendExposesNormalizedLifecycle(t *testing.T) {
	var authHeaders []string
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeaders = append(authHeaders, r.Header.Get("Authorization"))
		paths = append(paths, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/health":
			writeJSON(t, w, map[string]any{"status": "ok"})
		case "/v1/capabilities":
			writeJSON(t, w, map[string]any{
				"healthy": true, "protocols": []string{"runs_api"}, "events": true,
				"cancellation": true, "approvals": true, "usage": true,
			})
		case "/v1/runs":
			writeJSON(t, w, map[string]any{"id": "external-run-1", "status": "queued"})
		case "/v1/runs/external-run-1":
			writeJSON(t, w, map[string]any{"id": "external-run-1", "status": "in_progress"})
		case "/v1/runs/external-run-1/events":
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("event: run.progress\ndata: {\"type\":\"requires_action\",\"approval\":{\"approval_id\":\"approval-1\",\"summary\":\"Run command\"}}\n\n"))
			_, _ = w.Write([]byte("data: {\"status\":\"succeeded\",\"result\":{\"summary\":\"done\",\"output_refs\":[\"proof://output-1\"]}}\n\n"))
		case "/v1/runs/external-run-1/stop", "/v1/runs/external-run-1/approvals/approval-1":
			writeJSON(t, w, map[string]any{"ok": true})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	backend := newTestFrameworkRunsBackend(t, server.URL)
	health, err := backend.HealthCheck(context.Background())
	if err != nil || !health.Healthy {
		t.Fatalf("HealthCheck = %#v, %v", health, err)
	}
	caps, err := backend.GetCapabilities(context.Background())
	if err != nil || !caps.SupportsEvents || !caps.SupportsCancellation || !caps.SupportsApprovals || !caps.SupportsUsage {
		t.Fatalf("GetCapabilities = %#v, %v", caps, err)
	}
	handle, err := backend.CreateRun(context.Background(), WorkerRunRequest{Intent: "build output"})
	if err != nil || handle.RunID != "external-run-1" || handle.Status != StatusAccepted || handle.Backend != BackendFrameworkRuns {
		t.Fatalf("CreateRun = %#v, %v", handle, err)
	}
	handle, err = backend.GetRun(context.Background(), handle.RunID)
	if err != nil || handle.Status != StatusRunning {
		t.Fatalf("GetRun = %#v, %v", handle, err)
	}
	events, err := backend.StreamRunEvents(context.Background(), handle.RunID)
	if err != nil {
		t.Fatalf("StreamRunEvents: %v", err)
	}
	var approvalSeen, completedSeen bool
	for event := range events {
		if event.Kind == EventApprovalNeeded && event.Approval != nil && event.Approval.ID == "approval-1" {
			approvalSeen = true
		}
		if event.Kind == EventCompleted && event.Result != nil && len(event.Result.Outputs) == 1 {
			completedSeen = true
		}
	}
	if !approvalSeen || !completedSeen {
		t.Fatalf("approvalSeen=%v completedSeen=%v", approvalSeen, completedSeen)
	}
	if err := backend.SubmitApproval(context.Background(), handle.RunID, WorkerApprovalDecision{ApprovalID: "approval-1", Decision: DecisionApprove}); err != nil {
		t.Fatalf("SubmitApproval: %v", err)
	}
	if err := backend.StopRun(context.Background(), handle.RunID); err != nil {
		t.Fatalf("StopRun: %v", err)
	}
	for _, header := range authHeaders {
		if header != "Bearer test-secret" {
			t.Fatalf("authorization header = %q", header)
		}
	}
	if len(paths) < 8 {
		t.Fatalf("normalized lifecycle paths = %#v", paths)
	}
	if _, ok := any(backend).(RunFinalizer); ok {
		t.Fatal("external backend must not directly finalize Mycelis runs or Outcomes")
	}
}

func TestFrameworkRunsBackendRejectsInvalidResponsesAndApprovals(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/capabilities" {
			writeJSON(t, w, map[string]any{"healthy": true, "protocols": []string{"runs_api"}})
			return
		}
		writeJSON(t, w, map[string]any{"status": "accepted"})
	}))
	defer server.Close()
	backend := newTestFrameworkRunsBackend(t, server.URL)
	if _, err := backend.CreateRun(context.Background(), WorkerRunRequest{Intent: "build"}); err == nil {
		t.Fatal("expected missing run_id to fail closed")
	}
	if err := backend.SubmitApproval(context.Background(), "run-1", WorkerApprovalDecision{ApprovalID: "approval-1", Decision: "later"}); err == nil {
		t.Fatal("expected invalid approval decision to fail closed")
	}
}

func TestFrameworkRunsEventStreamFailsClosedOnInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: not-json\n\n"))
	}))
	defer server.Close()
	backend := newTestFrameworkRunsBackend(t, server.URL)
	events, err := backend.StreamRunEvents(context.Background(), "run-1")
	if err != nil {
		t.Fatalf("StreamRunEvents: %v", err)
	}
	event := <-events
	if event.Kind != EventFailed || event.Status != StatusFailed || event.Error == nil || event.Error.Code != "invalid_event_stream" {
		t.Fatalf("event = %#v", event)
	}
}

func newTestFrameworkRunsBackend(t *testing.T, baseURL string) *FrameworkRunsBackend {
	t.Helper()
	backend, err := NewFrameworkRunsBackend(WorkerConfig{
		Backend: BackendFrameworkRuns, BaseURL: baseURL, APIKeySecretRef: "secret://framework/api",
	}, mapSecretResolver{"secret://framework/api": "test-secret"})
	if err != nil {
		t.Fatalf("NewFrameworkRunsBackend: %v", err)
	}
	return backend
}

func writeJSON(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Fatalf("write json: %v", err)
	}
}
