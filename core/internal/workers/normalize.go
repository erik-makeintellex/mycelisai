package workers

import (
	"fmt"
	"net/http"
	"strings"
	"time"
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
	return WorkerBackendError(
		"backend_http_error",
		fmt.Sprintf("%s: upstream returned HTTP %d", prefix, res.StatusCode),
		true,
	)
}

func runHandleFromMap(raw map[string]any, backend BackendKind, protocol Protocol) (WorkerRunHandle, error) {
	now := time.Now().UTC()
	runID := stringValue(raw["run_id"])
	if strings.TrimSpace(runID) == "" {
		return WorkerRunHandle{}, WorkerBackendError("invalid_backend_response", "External worker backend returned no run_id.", true)
	}
	status := normalizeRunStatus(stringValue(raw["status"]))
	if status == "" {
		return WorkerRunHandle{}, WorkerBackendError("invalid_backend_response", "External worker backend returned an invalid run status.", true)
	}
	handle := WorkerRunHandle{
		RunID:     runID,
		Backend:   backend,
		Status:    status,
		Protocol:  protocol,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  mapValue(raw["metadata"]),
	}
	if approval := mapValue(raw["approval"]); approval != nil {
		handle.Approval = approvalFromMap(approval)
	}
	if result := mapValue(raw["result"]); result != nil {
		handle.Result = resultFromMap(result)
	}
	if errMap := mapValue(raw["error"]); errMap != nil {
		handle.Error = errorFromMap(errMap)
	}
	return handle, nil
}

func eventFromMap(raw map[string]any, fallbackRunID string, backend BackendKind) (WorkerEvent, error) {
	status := normalizeRunStatus(stringValue(raw["status"]))
	if status == "" {
		return WorkerEvent{}, WorkerBackendError("invalid_backend_response", "External worker returned an invalid event status.", true)
	}
	kind := normalizeEventKind(stringValue(raw["kind"]))
	if kind == "" {
		return WorkerEvent{}, WorkerBackendError("invalid_backend_response", "External worker returned an invalid event kind.", true)
	}
	if kind != kindFromStatus(status) {
		return WorkerEvent{}, WorkerBackendError("invalid_backend_response", "External worker returned inconsistent event kind and status.", true)
	}
	event := WorkerEvent{
		EventID:      stringValue(raw["event_id"]),
		RunID:        stringValue(raw["run_id"]),
		BackendRunID: fallbackRunID,
		Backend:      backend,
		Kind:         kind,
		Status:       status,
		Message:      stringValue(raw["message"]),
		Timestamp:    time.Now().UTC(),
		Metadata:     mapValue(raw["metadata"]),
	}
	if strings.TrimSpace(event.EventID) == "" {
		return WorkerEvent{}, WorkerBackendError("invalid_backend_response", "External worker returned an event without event_id.", true)
	}
	if strings.TrimSpace(event.RunID) == "" {
		return WorkerEvent{}, WorkerBackendError("invalid_backend_response", "External worker returned an event without run_id.", true)
	}
	if event.RunID != fallbackRunID {
		return WorkerEvent{}, WorkerBackendError("run_identity_mismatch", "External worker event did not preserve the authoritative Mycelis run_id.", false)
	}
	if approval := mapValue(raw["approval"]); approval != nil {
		event.Approval = approvalFromMap(approval)
	}
	if kind == EventApprovalNeeded && (event.Approval == nil || strings.TrimSpace(event.Approval.ID) == "") {
		return WorkerEvent{}, WorkerBackendError("invalid_backend_response", "External worker requested approval without an approval id.", true)
	}
	if result := mapValue(raw["result"]); result != nil {
		event.Result = resultFromMap(result)
	}
	if errMap := mapValue(raw["error"]); errMap != nil {
		event.Error = errorFromMap(errMap)
	}
	return event, nil
}

func capabilitiesFromMap(raw map[string]any, backend BackendKind) WorkerCapabilities {
	return WorkerCapabilities{
		Backend:              backend,
		Healthy:              boolValue(raw["healthy"]),
		SupportedProtocols:   protocolsFromAny(raw["supported_protocols"]),
		SupportsEvents:       boolValue(raw["supports_events"]),
		SupportsCancellation: boolValue(raw["supports_cancellation"]),
		SupportsApprovals:    boolValue(raw["supports_approvals"]),
		SupportsUsage:        boolValue(raw["supports_usage"]),
		Features:             stringSlice(raw["features"]),
		Raw:                  raw,
	}
}

func approvalFromMap(raw map[string]any) *WorkerApprovalRequest {
	return &WorkerApprovalRequest{
		ID:              stringValue(raw["id"]),
		Kind:            stringValue(raw["kind"]),
		Summary:         stringValue(raw["summary"]),
		RiskLevel:       stringValue(raw["risk_level"]),
		RequestedAction: stringValue(raw["requested_action"]),
		Metadata:        mapValue(raw["metadata"]),
	}
}

func resultFromMap(raw map[string]any) *WorkerResult {
	return &WorkerResult{
		Summary:    stringValue(raw["summary"]),
		Outputs:    outputsFromMap(raw),
		Metadata:   mapValue(raw["metadata"]),
		FinishedAt: time.Now().UTC(),
	}
}

func outputsFromMap(raw map[string]any) []WorkerOutput {
	return workerOutputsFromAny(raw["outputs"])
}

func workerOutputsFromAny(value any) []WorkerOutput {
	switch items := value.(type) {
	case []WorkerOutput:
		return items
	case []map[string]any:
		outputs := make([]WorkerOutput, 0, len(items))
		for _, item := range items {
			if output := workerOutputFromMap(item); output != nil {
				outputs = append(outputs, *output)
			}
		}
		return outputs
	case []any:
		outputs := make([]WorkerOutput, 0, len(items))
		for _, item := range items {
			if output := workerOutputFromAny(item); output != nil {
				outputs = append(outputs, *output)
			}
		}
		return outputs
	default:
		return nil
	}
}

func workerOutputFromAny(value any) *WorkerOutput {
	if output, ok := value.(WorkerOutput); ok {
		return &output
	}
	if raw := mapValue(value); raw != nil {
		return workerOutputFromMap(raw)
	}
	return nil
}

func workerOutputFromMap(raw map[string]any) *WorkerOutput {
	output := WorkerOutput{
		ID:          stringValue(raw["id"]),
		Kind:        stringValue(raw["kind"]),
		Name:        stringValue(raw["name"]),
		URI:         stringValue(raw["uri"]),
		ContentType: stringValue(raw["content_type"]),
		Metadata:    mapValue(raw["metadata"]),
	}
	if output.Kind == "" {
		output.Kind = "reference"
	}
	if output.ID == "" && output.Name == "" && output.URI == "" {
		return nil
	}
	return &output
}

func errorFromMap(raw map[string]any) *WorkerError {
	return &WorkerError{
		Code:        stringValue(raw["code"]),
		Message:     stringValue(raw["message"]),
		Recoverable: boolValue(raw["recoverable"]),
		Metadata:    mapValue(raw["metadata"]),
	}
}

func kindFromStatus(status RunStatus) EventKind {
	switch status {
	case StatusAccepted:
		return EventAccepted
	case StatusApprovalNeeded:
		return EventApprovalNeeded
	case StatusCompleted:
		return EventCompleted
	case StatusFailed:
		return EventFailed
	case StatusCancelled:
		return EventCancelled
	default:
		return EventProgress
	}
}

func normalizeRunStatus(value string) RunStatus {
	switch value {
	case "accepted":
		return StatusAccepted
	case "running":
		return StatusRunning
	case "approval_needed":
		return StatusApprovalNeeded
	case "completed":
		return StatusCompleted
	case "failed":
		return StatusFailed
	case "cancelled":
		return StatusCancelled
	default:
		return ""
	}
}

func normalizeEventKind(value string) EventKind {
	switch value {
	case "accepted":
		return EventAccepted
	case "progress":
		return EventProgress
	case "approval_needed":
		return EventApprovalNeeded
	case "completed":
		return EventCompleted
	case "failed":
		return EventFailed
	case "cancelled":
		return EventCancelled
	default:
		return ""
	}
}

func protocolsFromAny(value any) []Protocol {
	values := stringSlice(value)
	out := make([]Protocol, 0, len(values))
	for _, value := range values {
		switch Protocol(value) {
		case ProtocolRunsAPI:
			out = append(out, Protocol(value))
		}
	}
	return out
}

func stringValue(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func boolValue(value any) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

func stringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		if values, ok := value.([]string); ok {
			return values
		}
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if value := stringValue(item); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func mapValue(value any) map[string]any {
	if value == nil {
		return nil
	}
	if out, ok := value.(map[string]any); ok {
		return out
	}
	return nil
}
