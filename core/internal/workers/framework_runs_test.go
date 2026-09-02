package workers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mapSecretResolver map[string]string

func (m mapSecretResolver) ResolveSecret(_ context.Context, ref string) (string, error) {
	return m[ref], nil
}

func TestFrameworkRunsBackendExposesNormalizedLifecycle(t *testing.T) {
	var authHeaders []string
	var paths []string
	var createPayload map[string]any
	var approvalPayload, stopPayload map[string]any
	const createdAt = "2026-09-02T12:00:00.123456Z"
	const updatedAt = "2026-09-02T12:00:01.654321Z"
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
			writeJSON(t, w, map[string]any{"run_id": "mycelis-run-1", "status": "accepted", "version": 1, "correlation": createPayload["correlation"], "created_at": createdAt, "updated_at": createdAt})
		case "/v1/runs/mycelis-run-1":
			writeJSON(t, w, map[string]any{"run_id": "mycelis-run-1", "status": "running", "version": 2, "correlation": createPayload["correlation"], "created_at": createdAt, "updated_at": updatedAt})
		case "/v1/runs/mycelis-run-1/events":
			w.Header().Set("Content-Type", "text/event-stream")
			correlation := `{"run_id":"mycelis-run-1","intent_proof_id":"proof-1","execution_contract_id":"contract-1","team_id":"","outcome_id":"","work_item_id":"work-1","idempotency_key":"dispatch-1","source_kind":"web_api","source_channel":"api.intent.confirm-action","payload_kind":"command","graph_revision":"graph-v1"}`
			_, _ = w.Write([]byte("id: 1\ndata: {\"event_id\":\"approval-event-1\",\"sequence\":1,\"version\":2,\"run_id\":\"mycelis-run-1\",\"correlation\":" + correlation + ",\"kind\":\"approval_needed\",\"status\":\"approval_needed\",\"timestamp\":\"" + updatedAt + "\",\"approval\":{\"id\":\"approval-1\",\"kind\":\"tool\",\"summary\":\"Run command\",\"risk_level\":\"low\",\"requested_action\":\"invoke\"}}\n\n"))
			_, _ = w.Write([]byte("id: 2\ndata: {\"event_id\":\"completed-event-1\",\"sequence\":2,\"version\":3,\"run_id\":\"mycelis-run-1\",\"correlation\":" + correlation + ",\"kind\":\"completed\",\"status\":\"completed\",\"timestamp\":\"" + updatedAt + "\",\"result\":{\"summary\":\"done\",\"finished_at\":\"" + updatedAt + "\",\"metadata\":{\"completion_authority\":\"candidate\",\"requires_core_validation\":true,\"verified\":false},\"outputs\":[{\"id\":\"output-1\",\"kind\":\"reference\",\"uri\":\"candidate://mycelis-run-1/output-1\",\"content_type\":\"application/json\",\"size_bytes\":42,\"sha256\":\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\",\"metadata\":{\"completion_authority\":\"candidate\",\"requires_core_validation\":true,\"verified\":false}}]}}\n\n"))
		case "/v1/runs/mycelis-run-1/approvals/approval-1":
			_ = json.NewDecoder(r.Body).Decode(&approvalPayload)
			w.WriteHeader(http.StatusAccepted)
			writeJSON(t, w, controlReceipt("approval-command-1", "approve"))
		case "/v1/runs/mycelis-run-1/stop":
			_ = json.NewDecoder(r.Body).Decode(&stopPayload)
			writeJSON(t, w, controlReceipt("stop-command-1", "stop"))
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
	if err != nil || handle.RunID != "mycelis-run-1" || handle.Version != 1 || handle.Status != StatusAccepted || handle.Backend != BackendFrameworkRuns || handle.CreatedAt.Format(time.RFC3339Nano) != createdAt {
		t.Fatalf("CreateRun = %#v, %v", handle, err)
	}
	correlation, _ := createPayload["correlation"].(map[string]any)
	if createPayload["run_id"] != "mycelis-run-1" || correlation["run_id"] != "mycelis-run-1" || correlation["intent_proof_id"] != "proof-1" || correlation["execution_contract_id"] != "contract-1" || correlation["work_item_id"] != "work-1" || correlation["idempotency_key"] != "dispatch-1" || correlation["source_kind"] != "web_api" || correlation["source_channel"] != "api.intent.confirm-action" || correlation["payload_kind"] != "command" || correlation["graph_revision"] != "graph-v1" {
		t.Fatalf("create run correlation = %#v", createPayload)
	}
	runID := handle.RunID
	handle, err = backend.GetRun(context.Background(), runID)
	if err != nil || handle.Status != StatusRunning {
		t.Fatalf("GetRun = %#v, %v", handle, err)
	}
	events, err := backend.StreamRunEvents(context.Background(), runID)
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
			if event.Result.Outputs[0].SizeBytes != 42 || event.Result.Outputs[0].SHA256 == "" || event.Timestamp.Format(time.RFC3339Nano) != updatedAt {
				t.Fatalf("candidate/timestamp = %#v", event)
			}
			completedSeen = true
		}
	}
	if !approvalSeen || !completedSeen {
		t.Fatalf("approvalSeen=%v completedSeen=%v", approvalSeen, completedSeen)
	}
	approval, err := backend.SubmitApprovalCommand(context.Background(), runID, WorkerApprovalDecision{
		ApprovalID: "approval-1", Decision: DecisionApprove, CommandID: "approval-command-1",
		ExpectedVersion: 2, ActorID: "operator-1",
	})
	if err != nil || approval.Replayed {
		t.Fatalf("SubmitApprovalCommand: %#v, %v", approval, err)
	}
	stop, err := backend.StopRunCommand(context.Background(), runID, WorkerStopCommand{
		CommandID: "stop-command-1", ExpectedVersion: 3, ActorID: "operator-1",
	})
	if err != nil || !stop.Replayed {
		t.Fatalf("StopRunCommand: %#v, %v", stop, err)
	}
	if approvalPayload["command_id"] != "approval-command-1" || approvalPayload["expected_version"] != float64(2) || stopPayload["command_id"] != "stop-command-1" {
		t.Fatalf("strict control payloads = approval %#v, stop %#v", approvalPayload, stopPayload)
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
	if _, err := backend.SubmitApprovalCommand(context.Background(), "run-1", WorkerApprovalDecision{ApprovalID: "approval-1", Decision: "later"}); err == nil {
		t.Fatal("expected invalid approval decision to fail closed")
	}
}

func controlReceipt(commandID, kind string) map[string]any {
	return map[string]any{
		"command_id": commandID, "run_id": "mycelis-run-1", "kind": kind,
		"state": "applied", "version": 3,
		"created_at": "2026-09-02T12:00:01Z", "updated_at": "2026-09-02T12:00:02Z",
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
		writeJSON(t, w, strictRunSnapshot("different-run", "accepted", 1))
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
		writeJSON(t, w, strictRunSnapshot("different-run", "running", 1))
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
		writeJSON(t, w, strictRunSnapshot(" mycelis-run-1 ", "running", 1))
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
	if event.EventID == "" || event.RunID != "run-1" || event.Kind != EventFailed || event.Status != StatusFailed || event.Error == nil || event.Error.Code != "invalid_event_stream" {
		t.Fatalf("event = %#v", event)
	}
}

func TestFrameworkRunsEventStreamResumesAndRequiresMatchingDecimalCursor(t *testing.T) {
	for name, testCase := range map[string]struct {
		sseID       string
		wantFailure bool
	}{
		"resume":     {"3", false},
		"mismatch":   {"4", true},
		"nondecimal": {"event-3", true},
	} {
		t.Run(name, func(t *testing.T) {
			var lastEventID string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				lastEventID = r.Header.Get("Last-Event-ID")
				w.Header().Set("Content-Type", "text/event-stream")
				body := map[string]any{
					"event_id": "event-3", "sequence": 3, "version": 2,
					"run_id": "run-1", "correlation": testRawCorrelation("run-1"),
					"kind": "progress", "status": "running", "message": "working",
					"timestamp": "2026-09-02T12:00:03.123456Z",
				}
				encoded, _ := json.Marshal(body)
				_, _ = w.Write([]byte("id: " + testCase.sseID + "\ndata: " + string(encoded) + "\n\n"))
			}))
			defer server.Close()
			backend := newTestFrameworkRunsBackend(t, server.URL)
			events, err := backend.StreamRunEventsAfter(t.Context(), "run-1", 2)
			if err != nil {
				t.Fatal(err)
			}
			event := <-events
			if lastEventID != "2" {
				t.Fatalf("Last-Event-ID = %q", lastEventID)
			}
			failed := event.Error != nil && event.Error.Code == "invalid_event_stream"
			if failed != testCase.wantFailure {
				t.Fatalf("event = %#v, want failure %v", event, testCase.wantFailure)
			}
		})
	}
}

func TestFrameworkRunsControlRequiresCoreIdentityAndVersion(t *testing.T) {
	backend := newTestFrameworkRunsBackend(t, "https://workers.example.test")
	if _, err := backend.StopRunCommand(t.Context(), "run-1", WorkerStopCommand{}); err == nil {
		t.Fatal("missing Core control identity must fail before HTTP")
	}
	if err := backend.StopRun(t.Context(), "run-1"); err == nil {
		t.Fatal("legacy stop without command identity must fail closed")
	}
}

func strictRunSnapshot(runID, status string, version int) map[string]any {
	return map[string]any{
		"run_id": runID, "status": status, "version": version,
		"correlation": testRawCorrelation(runID),
		"created_at":  "2026-09-02T12:00:00Z", "updated_at": "2026-09-02T12:00:01Z",
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
