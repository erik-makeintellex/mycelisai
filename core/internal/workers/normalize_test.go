package workers

import (
	"io"
	"net/http"
	"strings"
	"testing"
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

func TestResultFromMapAcceptsExactRunsAPIOutputs(t *testing.T) {
	result := resultFromMap(map[string]any{
		"summary": "outputs ready",
		"outputs": []any{
			map[string]any{
				"id":           "artifact-1",
				"kind":         "file",
				"name":         "Run package",
				"uri":          "/api/v1/workspace/files/view?path=generated%2Frun.md",
				"content_type": "text/markdown",
			},
		},
	})

	if result == nil {
		t.Fatal("resultFromMap returned nil")
	}
	if len(result.Outputs) != 1 {
		t.Fatalf("outputs = %#v, want one exact output", result.Outputs)
	}
	if got := result.Outputs[0]; got.ID != "artifact-1" || got.Kind != "file" || got.Name != "Run package" || got.URI == "" || got.ContentType != "text/markdown" {
		t.Fatalf("outputs[0] = %#v", got)
	}
}

func TestEventFromMapRequiresExactIdentityAndVocabulary(t *testing.T) {
	event, err := eventFromMap(map[string]any{
		"event_id": "event-1",
		"run_id":   "run-1",
		"kind":     "completed",
		"status":   "completed",
		"result": map[string]any{
			"summary": "team output ready",
			"outputs": []any{map[string]any{
				"id": "output-1", "kind": "file", "name": "Playable package", "uri": "generated/game/index.html",
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
	if got := event.Result.Outputs[0]; got.ID != "output-1" || got.Name != "Playable package" || got.URI != "generated/game/index.html" {
		t.Fatalf("output = %#v", got)
	}
}

func TestRunAndEventAliasesFailClosed(t *testing.T) {
	handle, err := runHandleFromMap(map[string]any{"run_id": "run-1", "status": "running"}, BackendFrameworkRuns, ProtocolRunsAPI)
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

func TestCapabilitiesRequireExactBooleanFields(t *testing.T) {
	caps := capabilitiesFromMap(map[string]any{"healthy": "true", "supports_events": "ok"}, BackendFrameworkRuns)
	if caps.Healthy || caps.SupportsEvents {
		t.Fatalf("string boolean aliases must not normalize: %#v", caps)
	}
}
