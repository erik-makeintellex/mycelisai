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

func TestHermesAPIBackendDiscoversCapabilitiesAndCreatesRun(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/v1/capabilities":
			writeJSON(t, w, map[string]any{
				"healthy":               true,
				"supported_protocols":   []string{"runs_api", "responses_api"},
				"supports_events":       true,
				"supports_cancellation": true,
				"supports_approvals":    true,
				"features":              []string{"tools", "skills"},
			})
		case "/v1/runs":
			writeJSON(t, w, map[string]any{"run_id": "hermes-run-1", "status": "accepted"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	backend := newTestHermesBackend(t, server.URL)
	handle, err := backend.CreateRun(context.Background(), WorkerRunRequest{Intent: "build output"})
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if handle.RunID != "hermes-run-1" || handle.Protocol != ProtocolRunsAPI {
		t.Fatalf("handle = %+v", handle)
	}
	if authHeader != "Bearer test-secret" {
		t.Fatalf("expected bearer auth from secret resolver")
	}
}

func TestHermesAPIBackendStreamsEventsAndNormalizesApproval(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/runs/run-1/events" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"kind\":\"approval_needed\",\"approval\":{\"id\":\"approval-1\",\"kind\":\"command\",\"summary\":\"Run command\",\"risk_level\":\"high\"}}\n\n"))
		_, _ = w.Write([]byte("data: {\"status\":\"completed\",\"result\":{\"summary\":\"done\"}}\n\n"))
	}))
	defer server.Close()

	backend := newTestHermesBackend(t, server.URL)
	events, err := backend.StreamRunEvents(context.Background(), "run-1")
	if err != nil {
		t.Fatalf("StreamRunEvents: %v", err)
	}
	var approvalSeen, completedSeen bool
	for event := range events {
		if event.Kind == EventApprovalNeeded && event.Approval != nil && event.Approval.ID == "approval-1" {
			approvalSeen = true
		}
		if event.Kind == EventCompleted && event.Result != nil && event.Result.Summary == "done" {
			completedSeen = true
		}
	}
	if !approvalSeen || !completedSeen {
		t.Fatalf("approvalSeen=%v completedSeen=%v", approvalSeen, completedSeen)
	}
}

func TestHermesAPIBackendStopAndApprovalUseNormalizedEndpoints(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		writeJSON(t, w, map[string]any{"ok": true})
	}))
	defer server.Close()

	backend := newTestHermesBackend(t, server.URL)
	if err := backend.StopRun(context.Background(), "run-1"); err != nil {
		t.Fatalf("StopRun: %v", err)
	}
	if err := backend.SubmitApproval(context.Background(), "run-1", WorkerApprovalDecision{ApprovalID: "approval-1", Decision: DecisionApprove}); err != nil {
		t.Fatalf("SubmitApproval: %v", err)
	}
	want := []string{"/v1/runs/run-1/stop", "/v1/runs/run-1/approvals/approval-1"}
	for i, path := range want {
		if paths[i] != path {
			t.Fatalf("paths[%d] = %q, want %q", i, paths[i], path)
		}
	}
}

func newTestHermesBackend(t *testing.T, baseURL string) *HermesAPIBackend {
	t.Helper()
	backend, err := NewHermesAPIBackend(WorkerConfig{
		Backend:         BackendHermesAPI,
		BaseURL:         baseURL,
		APIKeySecretRef: "secret://hermes/api",
	}, mapSecretResolver{"secret://hermes/api": "test-secret"})
	if err != nil {
		t.Fatalf("NewHermesAPIBackend: %v", err)
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
