package workers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mapSecretResolver map[string]string

func (m mapSecretResolver) ResolveSecret(_ context.Context, ref string) (string, error) {
	return m[ref], nil
}

func TestFrameworkRunsBackendExposesNormalizedLifecycle(t *testing.T) {
	var authHeaders []string
	var paths []string
	var createPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeaders = append(authHeaders, r.Header.Get("Authorization"))
		paths = append(paths, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/health":
			writeJSON(t, w, map[string]any{"healthy": true})
		case "/v1/capabilities":
			writeJSON(t, w, map[string]any{
				"healthy": true, "supported_protocols": []string{"runs_api"}, "supports_events": true,
				"supports_cancellation": true, "supports_approvals": true, "supports_usage": true,
			})
		case "/v1/runs":
			if err := json.NewDecoder(r.Body).Decode(&createPayload); err != nil {
				t.Fatalf("decode create run request: %v", err)
			}
			writeJSON(t, w, map[string]any{"run_id": "mycelis-run-1", "status": "accepted"})
		case "/v1/runs/mycelis-run-1":
			writeJSON(t, w, map[string]any{"run_id": "mycelis-run-1", "status": "running"})
		case "/v1/runs/mycelis-run-1/events":
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"event_id\":\"approval-event-1\",\"run_id\":\"mycelis-run-1\",\"kind\":\"approval_needed\",\"status\":\"approval_needed\",\"approval\":{\"id\":\"approval-1\",\"summary\":\"Run command\"}}\n\n"))
			_, _ = w.Write([]byte("data: {\"event_id\":\"completed-event-1\",\"run_id\":\"mycelis-run-1\",\"kind\":\"completed\",\"status\":\"completed\",\"result\":{\"summary\":\"done\",\"outputs\":[{\"id\":\"output-1\",\"kind\":\"reference\",\"uri\":\"proof://output-1\"}]}}\n\n"))
		case "/v1/runs/mycelis-run-1/stop", "/v1/runs/mycelis-run-1/approvals/approval-1":
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
	handle, err := backend.CreateRun(context.Background(), correlatedTestRunRequest("mycelis-run-1", "build output"))
	if err != nil || handle.RunID != "mycelis-run-1" || handle.BackendRunID != "mycelis-run-1" || handle.Status != StatusAccepted || handle.Backend != BackendFrameworkRuns {
		t.Fatalf("CreateRun = %#v, %v", handle, err)
	}
	correlation, _ := createPayload["correlation"].(map[string]any)
	if createPayload["run_id"] != "mycelis-run-1" || correlation["run_id"] != "mycelis-run-1" || correlation["intent_proof_id"] != "proof-1" || correlation["execution_contract_id"] != "contract-1" || correlation["work_item_id"] != "work-1" || correlation["idempotency_key"] != "dispatch-1" || correlation["source_kind"] != "web_api" || correlation["source_channel"] != "api.intent.confirm-action" || correlation["payload_kind"] != "command" || correlation["graph_revision"] != "graph-v1" {
		t.Fatalf("create run correlation = %#v", createPayload)
	}
	backendRunID := handle.BackendRunID
	handle, err = backend.GetRun(context.Background(), backendRunID)
	if err != nil || handle.Status != StatusRunning {
		t.Fatalf("GetRun = %#v, %v", handle, err)
	}
	events, err := backend.StreamRunEvents(context.Background(), backendRunID)
	if err != nil {
		t.Fatalf("StreamRunEvents: %v", err)
	}
	var approvalSeen, completedSeen bool
	for event := range events {
		if event.Kind == EventApprovalNeeded && event.Approval != nil && event.Approval.ID == "approval-1" {
			if event.EventID != "approval-event-1" {
				t.Fatalf("SSE event id = %q", event.EventID)
			}
			approvalSeen = true
		}
		if event.Kind == EventCompleted && event.Result != nil && len(event.Result.Outputs) == 1 {
			completedSeen = true
		}
	}
	if !approvalSeen || !completedSeen {
		t.Fatalf("approvalSeen=%v completedSeen=%v", approvalSeen, completedSeen)
	}
	if err := backend.SubmitApproval(context.Background(), backendRunID, WorkerApprovalDecision{ApprovalID: "approval-1", Decision: DecisionApprove}); err != nil {
		t.Fatalf("SubmitApproval: %v", err)
	}
	if err := backend.StopRun(context.Background(), backendRunID); err != nil {
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
			writeJSON(t, w, map[string]any{"healthy": true, "supported_protocols": []string{"runs_api"}})
			return
		}
		writeJSON(t, w, map[string]any{"status": "accepted"})
	}))
	defer server.Close()
	backend := newTestFrameworkRunsBackend(t, server.URL)
	if _, err := backend.CreateRun(context.Background(), correlatedTestRunRequest("mycelis-run-1", "build")); err == nil {
		t.Fatal("expected missing run_id to fail closed")
	}
	if err := backend.SubmitApproval(context.Background(), "run-1", WorkerApprovalDecision{ApprovalID: "approval-1", Decision: "later"}); err == nil {
		t.Fatal("expected invalid approval decision to fail closed")
	}
}

func correlatedTestRunRequest(runID, intent string) WorkerRunRequest {
	return WorkerRunRequest{
		RunID: runID, Intent: intent,
		Correlation: WorkerCorrelation{
			RunID: runID, IntentProofID: "proof-1", ExecutionContractID: "contract-1", WorkItemID: "work-1", IdempotencyKey: "dispatch-1",
			SourceKind: "web_api", SourceChannel: "api.intent.confirm-action", PayloadKind: "command", GraphRevision: "graph-v1",
		},
	}
}

func TestFrameworkRunsBackendRejectsMismatchedAuthoritativeRunID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/capabilities" {
			writeJSON(t, w, map[string]any{"healthy": true, "supported_protocols": []string{"runs_api"}})
			return
		}
		writeJSON(t, w, map[string]any{"run_id": "different-run", "status": "accepted"})
	}))
	defer server.Close()
	backend := newTestFrameworkRunsBackend(t, server.URL)
	_, err := backend.CreateRun(context.Background(), correlatedTestRunRequest("mycelis-run-1", "build"))
	workerErr, ok := err.(*WorkerError)
	if !ok || workerErr.Code != "run_identity_mismatch" || workerErr.Recoverable {
		t.Fatalf("identity mismatch error = %#v", err)
	}
}

func TestFrameworkRunsBackendGetRunRejectsMismatchedAuthoritativeRunID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"run_id": "different-run", "status": "running"})
	}))
	defer server.Close()
	backend := newTestFrameworkRunsBackend(t, server.URL)
	_, err := backend.GetRun(context.Background(), "mycelis-run-1")
	workerErr, ok := err.(*WorkerError)
	if !ok || workerErr.Code != "run_identity_mismatch" || workerErr.Recoverable {
		t.Fatalf("get identity mismatch error = %#v", err)
	}
}

func TestFrameworkRunsBackendGetRunRejectsNoncanonicalReturnedIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"run_id": " mycelis-run-1 ", "status": "running"})
	}))
	defer server.Close()
	backend := newTestFrameworkRunsBackend(t, server.URL)
	if _, err := backend.GetRun(context.Background(), "mycelis-run-1"); err == nil {
		t.Fatal("expected padded returned identity to fail closed")
	}
}

func TestFrameworkRunsBackendRequiresCompleteCorrelation(t *testing.T) {
	backend := newTestFrameworkRunsBackend(t, "https://workers.example.test")
	req := correlatedTestRunRequest("mycelis-run-1", "build")
	req.Correlation.GraphRevision = ""
	if _, err := backend.CreateRun(context.Background(), req); err == nil || !strings.Contains(err.Error(), "graph_revision") {
		t.Fatalf("missing graph revision error = %v", err)
	}
}

func TestFrameworkRunsBackendRejectsDuplicateCorrelationMetadata(t *testing.T) {
	backend := newTestFrameworkRunsBackend(t, "https://workers.example.test")
	req := correlatedTestRunRequest("mycelis-run-1", "build")
	req.Metadata = map[string]any{"intent_proof_id": "duplicate-proof"}
	if _, err := backend.CreateRun(context.Background(), req); err == nil || !strings.Contains(err.Error(), "duplicate typed correlation") {
		t.Fatalf("duplicate correlation metadata error = %v", err)
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
	if event.EventID == "" || event.BackendRunID != "run-1" || event.Kind != EventFailed || event.Status != StatusFailed || event.Error == nil || event.Error.Code != "invalid_event_stream" {
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
