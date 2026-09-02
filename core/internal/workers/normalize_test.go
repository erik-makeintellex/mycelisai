package workers

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestStatusErrorDoesNotExposeUpstreamBody(t *testing.T) {
	res := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader(`{"secret":"leak-me"}`)),
	}

	err := statusError("framework runs request", res)
	if strings.Contains(err.Error(), "leak-me") {
		t.Fatalf("error exposed upstream response body: %v", err)
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Fatalf("error = %q, want status code", err)
	}
}

func TestStatusErrorPreservesExactRunsAPIError(t *testing.T) {
	res := &http.Response{
		StatusCode: http.StatusConflict,
		Body: io.NopCloser(strings.NewReader(
			`{"error":{"code":"version_conflict","message":"Version is stale.","recoverable":true}}`,
		)),
	}
	err := statusError("framework runs request", res)
	workerErr, ok := err.(*WorkerError)
	if !ok || workerErr.Code != "version_conflict" || workerErr.Message != "Version is stale." || !workerErr.Recoverable {
		t.Fatalf("normalized Runs API error = %#v", err)
	}
}

func TestResultFromMapAcceptsExactRunsAPIOutputs(t *testing.T) {
	finishedAt := "2026-09-02T12:00:01.123456Z"
	result, err := resultFromMap(map[string]any{
		"summary":     "outputs ready",
		"finished_at": finishedAt,
		"metadata":    candidateAuthMetadata(),
		"outputs": []any{
			map[string]any{
				"id":           "artifact-1",
				"kind":         "file",
				"name":         "Run package",
				"uri":          "candidate://run-1/artifact-1",
				"content_type": "text/markdown",
				"size_bytes":   42,
				"sha256":       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"metadata":     candidateAuthMetadata(),
			},
		},
	}, "run-1")

	if err != nil || result == nil {
		t.Fatal("resultFromMap returned nil")
	}
	if len(result.Outputs) != 1 {
		t.Fatalf("outputs = %#v, want one exact output", result.Outputs)
	}
	if got := result.Outputs[0]; got.ID != "artifact-1" || got.Kind != "file" || got.Name != "Run package" || got.URI == "" || got.ContentType != "text/markdown" || got.SizeBytes != 42 || got.SHA256 == "" {
		t.Fatalf("outputs[0] = %#v", got)
	}
	if result.FinishedAt.Format(time.RFC3339Nano) != finishedAt {
		t.Fatalf("finished_at = %s", result.FinishedAt.Format(time.RFC3339Nano))
	}
}

func TestEventFromMapRequiresExactIdentityAndVocabulary(t *testing.T) {
	event, err := eventFromMap(map[string]any{
		"event_id":    "event-1",
		"run_id":      "run-1",
		"sequence":    1,
		"version":     1,
		"correlation": testRawCorrelation("run-1"),
		"timestamp":   "2026-09-02T12:00:01.123456Z",
		"kind":        "completed",
		"status":      "completed",
		"result": map[string]any{
			"summary": "team output ready", "finished_at": "2026-09-02T12:00:01.123456Z",
			"metadata": candidateAuthMetadata(),
			"outputs": []any{map[string]any{
				"id": "output-1", "kind": "file", "name": "Playable package",
				"uri": "candidate://run-1/output-1", "content_type": "text/html", "size_bytes": 42,
				"sha256":   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"metadata": candidateAuthMetadata(),
			}},
		},
	}, "run-1", BackendFrameworkRuns)
	if err != nil {
		t.Fatalf("eventFromMap: %v", err)
	}
	if event.Kind != EventCompleted {
		t.Fatalf("kind = %s, want %s", event.Kind, EventCompleted)
	}
	if event.Result == nil {
		t.Fatal("expected exact result outputs")
	}
	if len(event.Result.Outputs) != 1 {
		t.Fatalf("outputs = %#v, want one normalized output", event.Result.Outputs)
	}
	if got := event.Result.Outputs[0]; got.ID != "output-1" || got.Name != "Playable package" || got.URI != "candidate://run-1/output-1" {
		t.Fatalf("output = %#v", got)
	}
}

func TestLifecycleAndCandidateShapesFailClosed(t *testing.T) {
	base := map[string]any{
		"run_id": "run-1", "version": 2, "correlation": testRawCorrelation("run-1"),
		"created_at": "2026-09-02T12:00:00Z", "updated_at": "2026-09-02T12:00:01Z",
	}
	for name, mutation := range map[string]map[string]any{
		"completed without result": {"status": "completed"},
		"failed without error":     {"status": "failed"},
		"unknown response field":   {"status": "running", "legacy_id": "run-1"},
	} {
		raw := map[string]any{}
		for key, value := range base {
			raw[key] = value
		}
		for key, value := range mutation {
			raw[key] = value
		}
		if _, err := runHandleFromMap(raw, BackendFrameworkRuns, ProtocolRunsAPI); err == nil {
			t.Fatalf("%s must fail closed", name)
		}
	}
	invalid := map[string]any{
		"summary": "candidate", "finished_at": "2026-09-02T12:00:01Z",
		"metadata": candidateAuthMetadata(), "outputs": []any{map[string]any{
			"id": "output-1", "kind": "file", "uri": "candidate://run-1/output-1",
			"content_type": "text/plain", "size_bytes": -1, "sha256": "bad",
			"metadata": candidateAuthMetadata(),
		}},
	}
	if _, err := resultFromMap(invalid, "run-1"); err == nil {
		t.Fatal("invalid candidate size and digest must fail closed")
	}
}

func candidateAuthMetadata() map[string]any {
	return map[string]any{
		"completion_authority": "candidate", "requires_core_validation": true, "verified": false,
	}
}

func TestRunAndEventAliasesFailClosed(t *testing.T) {
	handle, err := runHandleFromMap(map[string]any{
		"run_id": "run-1", "status": "running", "version": 1,
		"correlation": testRawCorrelation("run-1"),
		"created_at":  "2026-09-02T12:00:00Z", "updated_at": "2026-09-02T12:00:01Z",
	}, BackendFrameworkRuns, ProtocolRunsAPI)
	if err != nil {
		t.Fatalf("runHandleFromMap: %v", err)
	}
	if handle.Status != StatusRunning {
		t.Fatalf("status = %q, want %q", handle.Status, StatusRunning)
	}
	if _, err := runHandleFromMap(map[string]any{"id": "run-1", "status": "in_progress"}, BackendFrameworkRuns, ProtocolRunsAPI); err == nil {
		t.Fatal("legacy run aliases must fail closed")
	}
	if _, err := eventFromMap(map[string]any{"event_id": "event-1", "run_id": "run-1", "type": "requires_action", "status": "requires_action"}, "run-1", BackendFrameworkRuns); err == nil {
		t.Fatal("legacy event aliases must fail closed")
	}
	for name, raw := range map[string]map[string]any{
		"event id":             {"run_id": "run-1", "kind": "progress", "status": "running"},
		"run id":               {"event_id": "event-1", "kind": "progress", "status": "running"},
		"kind/status mismatch": {"event_id": "event-1", "run_id": "run-1", "kind": "completed", "status": "running"},
		"approval details":     {"event_id": "event-1", "run_id": "run-1", "kind": "approval_needed", "status": "approval_needed"},
	} {
		if _, err := eventFromMap(raw, "run-1", BackendFrameworkRuns); err == nil {
			t.Fatalf("missing %s must fail closed", name)
		}
	}
}

func TestRunAndEventRejectMissingUpstreamTimestamps(t *testing.T) {
	run := map[string]any{
		"run_id": "run-1", "status": "running", "version": 1,
		"correlation": testRawCorrelation("run-1"),
		"created_at":  "2026-09-02T12:00:00Z", "updated_at": "2026-09-02T12:00:01Z",
	}
	delete(run, "created_at")
	if _, err := runHandleFromMap(run, BackendFrameworkRuns, ProtocolRunsAPI); err == nil {
		t.Fatal("missing run timestamp must fail closed")
	}
	event := map[string]any{
		"event_id": "event-1", "run_id": "run-1", "sequence": 1, "version": 1,
		"correlation": testRawCorrelation("run-1"), "kind": "progress", "status": "running",
	}
	if _, err := eventFromMap(event, "run-1", BackendFrameworkRuns); err == nil {
		t.Fatal("missing event timestamp must fail closed")
	}
}

func testRawCorrelation(runID string) map[string]any {
	return map[string]any{
		"run_id": runID, "intent_proof_id": "proof-1", "execution_contract_id": "contract-1",
		"work_item_id": "work-1", "idempotency_key": "dispatch-1", "source_kind": "web_api",
		"source_channel": "api.intent.confirm-action", "payload_kind": "command", "graph_revision": "graph-v1",
	}
}

func TestCapabilitiesRequireExactBooleanFields(t *testing.T) {
	caps := capabilitiesFromMap(map[string]any{"healthy": "true", "supports_events": "ok"}, BackendFrameworkRuns)
	if caps.Healthy || caps.SupportsEvents {
		t.Fatalf("string boolean aliases must not normalize: %#v", caps)
	}
}
