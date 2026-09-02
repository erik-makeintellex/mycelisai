package workers

import (
	"fmt"
	"strings"
	"time"
)

func capabilitiesFromMap(raw map[string]any, backend BackendKind) WorkerCapabilities {
	return WorkerCapabilities{
		Backend: backend, Healthy: boolValue(raw["healthy"]),
		SupportedProtocols:   protocolsFromAny(raw["supported_protocols"]),
		SupportsEvents:       boolValue(raw["supports_events"]),
		SupportsCancellation: boolValue(raw["supports_cancellation"]),
		SupportsApprovals:    boolValue(raw["supports_approvals"]),
		SupportsUsage:        boolValue(raw["supports_usage"]),
		Features:             stringSlice(raw["features"]), Raw: raw,
	}
}

func approvalFromMap(raw map[string]any) *WorkerApprovalRequest {
	return &WorkerApprovalRequest{
		ID: stringValue(raw["id"]), Kind: stringValue(raw["kind"]),
		Summary: stringValue(raw["summary"]), RiskLevel: stringValue(raw["risk_level"]),
		RequestedAction: stringValue(raw["requested_action"]), Metadata: mapValue(raw["metadata"]),
	}
}

func resultFromMap(raw map[string]any, runID string) (*WorkerResult, error) {
	if !exactKeys(raw, "summary", "outputs", "metadata", "finished_at") {
		return nil, invalidShape("result")
	}
	finishedAt, err := requiredWireTime(raw["finished_at"])
	if err != nil {
		return nil, invalidTimestamp("result finished_at")
	}
	outputs, err := outputsFromMap(raw, runID)
	if err != nil {
		return nil, err
	}
	result := &WorkerResult{
		Summary: stringValue(raw["summary"]), Outputs: outputs,
		Metadata: mapValue(raw["metadata"]), FinishedAt: finishedAt,
	}
	if !candidateMetadata(result.Metadata) {
		return nil, invalidShape("candidate result authority")
	}
	return result, nil
}

func outputsFromMap(raw map[string]any, runID string) ([]WorkerOutput, error) {
	items, ok := raw["outputs"].([]any)
	if !ok {
		return nil, invalidShape("candidate outputs")
	}
	outputs := make([]WorkerOutput, 0, len(items))
	for _, item := range items {
		output, err := workerOutputFromMap(mapValue(item), runID)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func workerOutputFromMap(raw map[string]any, runID string) (WorkerOutput, error) {
	if !exactKeys(raw, "id", "kind", "name", "uri", "content_type", "size_bytes", "sha256", "metadata") {
		return WorkerOutput{}, invalidShape("candidate output")
	}
	output := WorkerOutput{
		ID: stringValue(raw["id"]), Kind: stringValue(raw["kind"]),
		Name: stringValue(raw["name"]), URI: stringValue(raw["uri"]),
		ContentType: stringValue(raw["content_type"]), SizeBytes: int64Value(raw["size_bytes"]),
		SHA256: stringValue(raw["sha256"]), Metadata: mapValue(raw["metadata"]),
	}
	if output.ID == "" || output.Kind == "" || output.ContentType == "" || output.SizeBytes < 0 ||
		!sha256Pattern.MatchString(output.SHA256) ||
		!strings.HasPrefix(output.URI, "candidate://"+runID+"/") || strings.Contains(output.URI, "..") ||
		!candidateMetadata(output.Metadata) {
		return WorkerOutput{}, invalidShape("candidate output")
	}
	return output, nil
}

func controlReceiptFromMap(raw map[string]any, runID, commandID string, replayed bool) (WorkerControlReceipt, error) {
	if !exactKeys(raw, "command_id", "run_id", "kind", "state", "version", "created_at", "updated_at", "error") {
		return WorkerControlReceipt{}, invalidShape("control receipt")
	}
	receipt := WorkerControlReceipt{
		CommandID: stringValue(raw["command_id"]), RunID: stringValue(raw["run_id"]),
		Kind: stringValue(raw["kind"]), State: stringValue(raw["state"]),
		Version: int64Value(raw["version"]), Replayed: replayed,
	}
	if receipt.CommandID != commandID || receipt.RunID != runID || receipt.Version < 1 {
		return WorkerControlReceipt{}, WorkerBackendError("invalid_backend_response", "External worker returned a mismatched control receipt.", false)
	}
	if receipt.Kind != "stop" && receipt.Kind != "approve" && receipt.Kind != "deny" {
		return WorkerControlReceipt{}, WorkerBackendError("invalid_backend_response", "External worker returned an invalid control kind.", false)
	}
	if receipt.State != "pending" && receipt.State != "applied" && receipt.State != "failed" {
		return WorkerControlReceipt{}, WorkerBackendError("invalid_backend_response", "External worker returned an invalid control state.", false)
	}
	var err error
	receipt.CreatedAt, err = requiredWireTime(raw["created_at"])
	if err != nil {
		return WorkerControlReceipt{}, invalidTimestamp("control created_at")
	}
	receipt.UpdatedAt, err = requiredWireTime(raw["updated_at"])
	if err != nil {
		return WorkerControlReceipt{}, invalidTimestamp("control updated_at")
	}
	if rawError := mapValue(raw["error"]); rawError != nil {
		if !validErrorMap(rawError) {
			return WorkerControlReceipt{}, invalidShape("control receipt error")
		}
		receipt.Error = errorFromMap(rawError)
	}
	return receipt, nil
}

func requiredWireTime(value any) (time.Time, error) {
	raw := stringValue(value)
	if raw == "" {
		return time.Time{}, fmt.Errorf("timestamp is required")
	}
	return time.Parse(time.RFC3339Nano, raw)
}

func invalidTimestamp(field string) error {
	return WorkerBackendError("invalid_backend_response", "External worker returned an invalid or missing "+field+".", false)
}

func errorFromMap(raw map[string]any) *WorkerError {
	return &WorkerError{
		Code: stringValue(raw["code"]), Message: stringValue(raw["message"]),
		Recoverable: boolValue(raw["recoverable"]), Metadata: mapValue(raw["metadata"]),
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
	switch RunStatus(value) {
	case StatusAccepted, StatusRunning, StatusApprovalNeeded, StatusCompleted, StatusFailed, StatusCancelled:
		return RunStatus(value)
	default:
		return ""
	}
}

func normalizeEventKind(value string) EventKind {
	switch EventKind(value) {
	case EventAccepted, EventProgress, EventApprovalNeeded, EventCompleted, EventFailed, EventCancelled:
		return EventKind(value)
	default:
		return ""
	}
}

func protocolsFromAny(value any) []Protocol {
	values := stringSlice(value)
	out := make([]Protocol, 0, len(values))
	for _, value := range values {
		if Protocol(value) == ProtocolRunsAPI {
			out = append(out, Protocol(value))
		}
	}
	return out
}

func stringValue(value any) string {
	valueString, _ := value.(string)
	return valueString
}

func int64Value(value any) int64 {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int64:
		return typed
	case float64:
		if typed == float64(int64(typed)) {
			return int64(typed)
		}
	}
	return -1
}

func correlationFromMap(raw map[string]any) WorkerCorrelation {
	return WorkerCorrelation{
		RunID: stringValue(raw["run_id"]), IntentProofID: stringValue(raw["intent_proof_id"]),
		ExecutionContractID: stringValue(raw["execution_contract_id"]), TeamID: stringValue(raw["team_id"]),
		WorkItemID: stringValue(raw["work_item_id"]), OutcomeID: stringValue(raw["outcome_id"]),
		IdempotencyKey: stringValue(raw["idempotency_key"]), SourceKind: stringValue(raw["source_kind"]),
		SourceChannel: stringValue(raw["source_channel"]), PayloadKind: stringValue(raw["payload_kind"]),
		GraphRevision: stringValue(raw["graph_revision"]),
	}
}

func completeCorrelation(c WorkerCorrelation) bool {
	for _, value := range []string{c.RunID, c.IntentProofID, c.ExecutionContractID, c.WorkItemID,
		c.IdempotencyKey, c.SourceKind, c.SourceChannel, c.PayloadKind, c.GraphRevision} {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func boolValue(value any) bool {
	valueBool, _ := value.(bool)
	return valueBool
}

func stringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		values, _ := value.([]string)
		return values
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
	out, _ := value.(map[string]any)
	return out
}
