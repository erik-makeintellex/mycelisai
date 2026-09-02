package workers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func WorkerBackendError(code, message string, recoverable bool) *WorkerError {
	return &WorkerError{Code: code, Message: message, Recoverable: recoverable}
}

func (e *WorkerError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func statusError(prefix string, res *http.Response) error {
	var envelope map[string]any
	if json.NewDecoder(res.Body).Decode(&envelope) == nil && exactKeys(envelope, "error") {
		if raw := mapValue(envelope["error"]); validErrorMap(raw) {
			return errorFromMap(raw)
		}
	}
	return WorkerBackendError("backend_http_error",
		fmt.Sprintf("%s: upstream returned HTTP %d", prefix, res.StatusCode), res.StatusCode >= 500)
}

func runHandleFromMap(raw map[string]any, backend BackendKind, protocol Protocol) (WorkerRunHandle, error) {
	if !exactKeys(raw, "run_id", "correlation", "status", "version", "created_at", "updated_at", "approval", "result", "error", "usage", "metadata") {
		return WorkerRunHandle{}, invalidShape("run snapshot")
	}
	runID := stringValue(raw["run_id"])
	if strings.TrimSpace(runID) == "" {
		return WorkerRunHandle{}, WorkerBackendError("invalid_backend_response", "External worker backend returned no run_id.", true)
	}
	status := normalizeRunStatus(stringValue(raw["status"]))
	if status == "" {
		return WorkerRunHandle{}, WorkerBackendError("invalid_backend_response", "External worker backend returned an invalid run status.", true)
	}
	createdAt, err := requiredWireTime(raw["created_at"])
	if err != nil {
		return WorkerRunHandle{}, invalidTimestamp("run created_at")
	}
	updatedAt, err := requiredWireTime(raw["updated_at"])
	if err != nil {
		return WorkerRunHandle{}, invalidTimestamp("run updated_at")
	}
	handle := WorkerRunHandle{
		RunID: runID, Backend: backend, Status: status, Protocol: protocol,
		CreatedAt: createdAt, UpdatedAt: updatedAt, Metadata: mapValue(raw["metadata"]),
		Version: int64Value(raw["version"]), Correlation: correlationFromMap(mapValue(raw["correlation"])),
	}
	if handle.Version < 1 || handle.Correlation.RunID != handle.RunID || !completeCorrelation(handle.Correlation) {
		return WorkerRunHandle{}, WorkerBackendError("invalid_backend_response", "External worker returned an invalid run version or correlation.", false)
	}
	if approval := mapValue(raw["approval"]); approval != nil {
		if !validApprovalMap(approval) {
			return WorkerRunHandle{}, invalidShape("run approval")
		}
		handle.Approval = approvalFromMap(approval)
	}
	if result := mapValue(raw["result"]); result != nil {
		handle.Result, err = resultFromMap(result, runID)
		if err != nil {
			return WorkerRunHandle{}, err
		}
	}
	if errMap := mapValue(raw["error"]); errMap != nil {
		if !validErrorMap(errMap) {
			return WorkerRunHandle{}, invalidShape("run error")
		}
		handle.Error = errorFromMap(errMap)
	}
	if err := validateLifecycleFields(handle.Status, handle.Approval, handle.Result, handle.Error); err != nil {
		return WorkerRunHandle{}, err
	}
	return handle, nil
}

func eventFromMap(raw map[string]any, expectedRunID string, backend BackendKind) (WorkerEvent, error) {
	if !exactKeys(raw, "event_id", "sequence", "version", "run_id", "correlation", "kind", "status", "message", "timestamp", "approval", "result", "error", "usage", "metadata") {
		return WorkerEvent{}, invalidShape("event")
	}
	status := normalizeRunStatus(stringValue(raw["status"]))
	kind := normalizeEventKind(stringValue(raw["kind"]))
	if status == "" || kind == "" || kind != kindFromStatus(status) {
		return WorkerEvent{}, WorkerBackendError("invalid_backend_response", "External worker returned inconsistent event kind or status.", true)
	}
	event := WorkerEvent{
		EventID: stringValue(raw["event_id"]), RunID: stringValue(raw["run_id"]), Backend: backend,
		Sequence: int64Value(raw["sequence"]), Version: int64Value(raw["version"]),
		Correlation: correlationFromMap(mapValue(raw["correlation"])), Kind: kind, Status: status,
		Message: stringValue(raw["message"]), Metadata: mapValue(raw["metadata"]),
	}
	if strings.TrimSpace(event.EventID) == "" {
		return WorkerEvent{}, WorkerBackendError("invalid_backend_response", "External worker returned an event without event_id.", true)
	}
	if event.Sequence <= 0 || event.Version < 1 {
		return WorkerEvent{}, WorkerBackendError("invalid_backend_response", "External worker event requires a positive sequence and version.", true)
	}
	if event.RunID != expectedRunID {
		return WorkerEvent{}, WorkerBackendError("run_identity_mismatch", "External worker event did not preserve the authoritative Mycelis run_id.", false)
	}
	if event.Correlation.RunID != event.RunID || !completeCorrelation(event.Correlation) {
		return WorkerEvent{}, WorkerBackendError("invalid_backend_response", "External worker event correlation is incomplete or mismatched.", false)
	}
	timestamp, err := requiredWireTime(raw["timestamp"])
	if err != nil {
		return WorkerEvent{}, invalidTimestamp("event timestamp")
	}
	event.Timestamp = timestamp
	if approval := mapValue(raw["approval"]); approval != nil {
		if !validApprovalMap(approval) {
			return WorkerEvent{}, invalidShape("event approval")
		}
		event.Approval = approvalFromMap(approval)
	}
	if kind == EventApprovalNeeded && (event.Approval == nil || strings.TrimSpace(event.Approval.ID) == "") {
		return WorkerEvent{}, WorkerBackendError("invalid_backend_response", "External worker requested approval without an approval id.", true)
	}
	if result := mapValue(raw["result"]); result != nil {
		event.Result, err = resultFromMap(result, event.RunID)
		if err != nil {
			return WorkerEvent{}, err
		}
	}
	if errMap := mapValue(raw["error"]); errMap != nil {
		if !validErrorMap(errMap) {
			return WorkerEvent{}, invalidShape("event error")
		}
		event.Error = errorFromMap(errMap)
	}
	if err := validateLifecycleFields(event.Status, event.Approval, event.Result, event.Error); err != nil {
		return WorkerEvent{}, err
	}
	return event, nil
}
