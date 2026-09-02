package workers

import (
	"regexp"
	"strings"
)

var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func exactKeys(raw map[string]any, allowed ...string) bool {
	if raw == nil {
		return false
	}
	set := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		set[key] = true
	}
	for key := range raw {
		if !set[key] {
			return false
		}
	}
	return true
}

func invalidShape(subject string) error {
	return WorkerBackendError("invalid_backend_response", "External worker returned an invalid "+subject+".", false)
}

func validApprovalMap(raw map[string]any) bool {
	if !exactKeys(raw, "id", "kind", "summary", "risk_level", "requested_action", "metadata", "expires_at") {
		return false
	}
	for _, key := range []string{"id", "kind", "summary", "risk_level", "requested_action"} {
		if strings.TrimSpace(stringValue(raw[key])) == "" {
			return false
		}
	}
	return true
}

func validErrorMap(raw map[string]any) bool {
	if !exactKeys(raw, "code", "message", "recoverable", "metadata") {
		return false
	}
	return strings.TrimSpace(stringValue(raw["code"])) != "" &&
		strings.TrimSpace(stringValue(raw["message"])) != "" && raw["recoverable"] != nil
}

func candidateMetadata(metadata map[string]any) bool {
	return stringValue(metadata["completion_authority"]) == "candidate" &&
		boolValue(metadata["requires_core_validation"]) && !boolValue(metadata["verified"])
}

func validateLifecycleFields(
	status RunStatus,
	approval *WorkerApprovalRequest,
	result *WorkerResult,
	workerError *WorkerError,
) error {
	valid := false
	switch status {
	case StatusAccepted, StatusRunning, StatusCancelled:
		valid = approval == nil && result == nil && workerError == nil
	case StatusApprovalNeeded:
		valid = approval != nil && result == nil && workerError == nil
	case StatusCompleted:
		valid = approval == nil && result != nil && workerError == nil
	case StatusFailed:
		valid = approval == nil && result == nil && workerError != nil
	}
	if !valid {
		return invalidShape("run lifecycle fields")
	}
	return nil
}
