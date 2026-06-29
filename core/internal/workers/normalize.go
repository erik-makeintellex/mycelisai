package workers

import (
	"fmt"
	"io"
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
	body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = res.Status
	}
	return WorkerBackendError("backend_http_error", prefix+": "+message, true)
}

func runHandleFromMap(raw map[string]any, backend BackendKind, protocol Protocol) WorkerRunHandle {
	now := time.Now().UTC()
	runID := stringValue(first(raw, "run_id", "id"))
	status := RunStatus(stringValue(raw["status"]))
	if status == "" {
		status = StatusAccepted
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
	return handle
}

func eventFromMap(raw map[string]any, fallbackRunID string, backend BackendKind) WorkerEvent {
	status := RunStatus(stringValue(raw["status"]))
	kind := EventKind(stringValue(first(raw, "kind", "type", "event")))
	if kind == "" {
		kind = kindFromStatus(status)
	}
	event := WorkerEvent{
		RunID:     stringValue(first(raw, "run_id", "id")),
		Backend:   backend,
		Kind:      kind,
		Status:    status,
		Message:   stringValue(raw["message"]),
		Timestamp: time.Now().UTC(),
		Metadata:  mapValue(raw["metadata"]),
	}
	if event.RunID == "" {
		event.RunID = fallbackRunID
	}
	if approval := mapValue(raw["approval"]); approval != nil {
		event.Approval = approvalFromMap(approval)
	}
	if result := mapValue(raw["result"]); result != nil {
		event.Result = resultFromMap(result)
	}
	if errMap := mapValue(raw["error"]); errMap != nil {
		event.Error = errorFromMap(errMap)
	}
	return event
}

func capabilitiesFromMap(raw map[string]any, backend BackendKind) WorkerCapabilities {
	return WorkerCapabilities{
		Backend:              backend,
		Healthy:              truthy(raw["healthy"]) || truthy(raw["ok"]),
		SupportedProtocols:   protocolsFromAny(first(raw, "supported_protocols", "protocols")),
		SupportsEvents:       truthy(first(raw, "supports_events", "events")),
		SupportsCancellation: truthy(first(raw, "supports_cancellation", "cancellation")),
		SupportsApprovals:    truthy(first(raw, "supports_approvals", "approvals")),
		SupportsUsage:        truthy(first(raw, "supports_usage", "usage")),
		Features:             stringSlice(raw["features"]),
		Raw:                  raw,
	}
}

func approvalFromMap(raw map[string]any) *WorkerApprovalRequest {
	return &WorkerApprovalRequest{
		ID:              stringValue(first(raw, "id", "approval_id")),
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
		Metadata:   mapValue(raw["metadata"]),
		FinishedAt: time.Now().UTC(),
	}
}

func errorFromMap(raw map[string]any) *WorkerError {
	return &WorkerError{
		Code:        stringValue(raw["code"]),
		Message:     stringValue(raw["message"]),
		Recoverable: truthy(raw["recoverable"]),
		Metadata:    mapValue(raw["metadata"]),
	}
}

func kindFromStatus(status RunStatus) EventKind {
	switch status {
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

func protocolsFromAny(value any) []Protocol {
	values := stringSlice(value)
	out := make([]Protocol, 0, len(values))
	for _, value := range values {
		switch Protocol(value) {
		case ProtocolRunsAPI, ProtocolResponsesAPI, ProtocolChatCompletion:
			out = append(out, Protocol(value))
		}
	}
	return out
}

func first(raw map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := raw[key]; ok {
			return value
		}
	}
	return nil
}

func stringValue(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func truthy(value any) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	if s, ok := value.(string); ok {
		return strings.EqualFold(s, "true") || strings.EqualFold(s, "ok") || strings.EqualFold(s, "healthy")
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
