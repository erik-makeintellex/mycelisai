package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var safeID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)
var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

var canonicalSourceKinds = map[string]bool{
	"workspace_ui": true, "web_api": true, "automation_trigger": true,
	"scheduler": true, "sensor": true, "iot": true, "internal_tool": true,
	"mcp": true, "system": true,
}

var correlationMetadataKeys = map[string]bool{
	"run_id": true, "correlation_id": true, "correlation": true,
	"intent_proof_id": true, "execution_contract_id": true, "team_id": true,
	"outcome_id": true, "work_item_id": true, "idempotency_key": true,
	"source_kind": true, "source_channel": true, "payload_kind": true,
	"graph_revision": true, "_framework_resume": true,
}

func ValidateExternalID(name, value string) error { return validateID(name, value, true) }

func NormalizeCreate(request *CreateRequest) {
	if request.Input == nil {
		request.Input = map[string]any{}
	}
	if request.Metadata == nil {
		request.Metadata = map[string]any{}
	}
	if request.RequiredProtocols == nil {
		request.RequiredProtocols = []string{}
	}
	if request.RequiredFeatures == nil {
		request.RequiredFeatures = []string{}
	}
}

func ValidateCreate(request CreateRequest) error {
	if err := validateID("run_id", request.RunID, true); err != nil {
		return err
	}
	if request.Intent == "" || request.Intent != strings.TrimSpace(request.Intent) || len(request.Intent) > 16_384 {
		return errors.New("intent must be canonical non-empty text of at most 16384 bytes")
	}
	if len(request.Instructions) > 65_536 {
		return errors.New("instructions exceeds 65536 bytes")
	}
	for name, value := range map[string]string{
		"org_id": request.OrgID, "project_id": request.ProjectID,
		"user_id": request.UserID, "requested_by": request.RequestedBy,
	} {
		if len(value) > 256 {
			return fmt.Errorf("%s exceeds 256 bytes", name)
		}
	}
	if len(request.RequiredProtocols) > 32 || len(request.RequiredFeatures) > 256 {
		return errors.New("required protocol or feature list exceeds its limit")
	}
	if err := validateCorrelation(request.Correlation, request.RunID); err != nil {
		return err
	}
	for _, item := range request.RequiredProtocols {
		if item != "runs_api" {
			return errors.New("required_protocols contains an unsupported protocol")
		}
	}
	for key := range request.Metadata {
		if correlationMetadataKeys[key] {
			return fmt.Errorf("metadata duplicates reserved field %q", key)
		}
	}
	if containsRawSecret(request.Input) || containsRawSecret(request.Metadata) {
		return errors.New("request contains raw secret material; use a secret reference")
	}
	return nil
}

func validateCorrelation(correlation Correlation, runID string) error {
	if correlation.RunID != runID {
		return errors.New("correlation.run_id must exactly match run_id")
	}
	for name, value := range map[string]string{
		"correlation.run_id":                correlation.RunID,
		"correlation.intent_proof_id":       correlation.IntentProofID,
		"correlation.execution_contract_id": correlation.ExecutionContractID,
		"correlation.team_id":               correlation.TeamID,
		"correlation.outcome_id":            correlation.OutcomeID,
		"correlation.work_item_id":          correlation.WorkItemID,
		"correlation.idempotency_key":       correlation.IdempotencyKey,
		"correlation.source_kind":           correlation.SourceKind,
		"correlation.source_channel":        correlation.SourceChannel,
		"correlation.payload_kind":          correlation.PayloadKind,
		"correlation.graph_revision":        correlation.GraphRevision,
	} {
		required := name != "correlation.team_id" && name != "correlation.outcome_id"
		if err := validateID(name, value, required); err != nil {
			return err
		}
	}
	if !canonicalSourceKinds[correlation.SourceKind] {
		return errors.New("correlation.source_kind is not canonical")
	}
	if correlation.PayloadKind != "command" {
		return errors.New("correlation.payload_kind must be command")
	}
	return nil
}

func ValidateStop(request StopRequest) error {
	if err := validateID("command_id", request.CommandID, true); err != nil {
		return err
	}
	if request.ExpectedVersion < 1 {
		return errors.New("expected_version must be at least 1")
	}
	if err := validateID("actor_id", request.ActorID, true); err != nil {
		return err
	}
	return validateControlMetadata(request.Metadata)
}

func ValidateApprovalDecision(request ApprovalDecisionRequest) error {
	for name, value := range map[string]string{
		"approval_id": request.ApprovalID,
		"command_id":  request.CommandID,
		"actor_id":    request.ActorID,
	} {
		if err := validateID(name, value, true); err != nil {
			return err
		}
	}
	if request.ExpectedVersion < 1 {
		return errors.New("expected_version must be at least 1")
	}
	if request.Decision != "approve" && request.Decision != "deny" {
		return errors.New("decision must be approve or deny")
	}
	return validateControlMetadata(request.Metadata)
}

func ValidateOutcome(outcome *ExecutorOutcome, runID string) error {
	if outcome == nil {
		return errors.New("executor returned no outcome")
	}
	if containsRawSecret(outcome.Metadata) || containsReservedAuthority(outcome.Metadata) {
		return errors.New("executor outcome metadata contains secret or reserved authority fields")
	}
	if outcome.Usage != nil && (outcome.Usage.InputTokens < 0 || outcome.Usage.OutputTokens < 0 ||
		outcome.Usage.DurationMS < 0 || outcome.Usage.CostEstimate < 0) {
		return errors.New("executor usage cannot be negative")
	}
	switch outcome.Status {
	case StatusRunning:
		if outcome.Approval != nil || outcome.Result != nil || outcome.Error != nil {
			return errors.New("running outcome contains terminal fields")
		}
	case StatusApprovalNeeded:
		if outcome.Approval == nil || outcome.Result != nil || outcome.Error != nil {
			return errors.New("approval_needed outcome requires only approval")
		}
		if err := validateApproval(*outcome.Approval); err != nil {
			return err
		}
	case StatusCompleted:
		if outcome.Result == nil || outcome.Approval != nil || outcome.Error != nil {
			return errors.New("completed outcome requires only result")
		}
		if err := ForceAndValidateCandidate(outcome.Result, runID); err != nil {
			return err
		}
	case StatusFailed:
		if outcome.Error == nil || outcome.Approval != nil || outcome.Result != nil {
			return errors.New("failed outcome requires only error")
		}
		if strings.TrimSpace(outcome.Error.Code) == "" || strings.TrimSpace(outcome.Error.Message) == "" ||
			containsRawSecret(outcome.Error.Metadata) {
			return errors.New("failed outcome error is incomplete or unsafe")
		}
	case StatusCancelled:
		if outcome.Approval != nil || outcome.Result != nil || outcome.Error != nil {
			return errors.New("cancelled outcome contains terminal fields")
		}
	default:
		return errors.New("executor returned unsupported status")
	}
	return nil
}

func ForceAndValidateCandidate(result *Result, runID string) error {
	if result == nil {
		return errors.New("candidate result is required")
	}
	if result.FinishedAt.IsZero() || result.Outputs == nil || containsRawSecret(result.Metadata) {
		return errors.New("candidate result requires finished_at, outputs, and safe metadata")
	}
	result.Metadata = cloneMap(result.Metadata)
	result.Metadata["completion_authority"] = "candidate"
	result.Metadata["requires_core_validation"] = true
	result.Metadata["verified"] = false
	for index := range result.Outputs {
		output := &result.Outputs[index]
		if err := validateID("output.id", output.ID, true); err != nil {
			return err
		}
		if output.Kind == "" || output.ContentType == "" {
			return errors.New("candidate output kind and content_type are required")
		}
		if containsRawSecret(output.Metadata) || containsReservedAuthority(output.Metadata) {
			return errors.New("candidate output metadata contains secret or reserved authority fields")
		}
		expectedURI := "candidate://" + runID + "/" + output.ID
		if output.URI != expectedURI || strings.Contains(output.URI, "..") {
			return errors.New("candidate output uri must exactly bind its run and output id")
		}
		if output.SizeBytes < 0 || !sha256Pattern.MatchString(output.SHA256) {
			return errors.New("candidate output requires nonnegative size_bytes and lowercase sha256")
		}
		output.Metadata = cloneMap(output.Metadata)
		output.Metadata["completion_authority"] = "candidate"
		output.Metadata["requires_core_validation"] = true
		output.Metadata["verified"] = false
	}
	return nil
}

func ValidateEvent(event Event) error {
	if event.Sequence < 1 || event.Version < 1 || event.EventID == "" || event.RunID == "" || event.Timestamp.IsZero() {
		return errors.New("event identity, sequence, version, correlation, and timestamp are required")
	}
	pairs := map[EventKind]Status{
		EventAccepted: StatusAccepted, EventProgress: StatusRunning,
		EventApprovalNeeded: StatusApprovalNeeded, EventCompleted: StatusCompleted,
		EventFailed: StatusFailed, EventCancelled: StatusCancelled,
	}
	if pairs[event.Kind] != event.Status {
		return errors.New("event kind/status pair is invalid")
	}
	if event.Correlation.RunID != event.RunID {
		return errors.New("event correlation does not match run")
	}
	if err := validateCorrelation(event.Correlation, event.RunID); err != nil {
		return fmt.Errorf("event %w", err)
	}
	if event.Kind == EventApprovalNeeded && event.Approval == nil {
		return errors.New("approval event requires approval")
	}
	if event.Kind == EventCompleted && event.Result == nil {
		return errors.New("completed event requires result")
	}
	if event.Kind == EventFailed && event.Error == nil {
		return errors.New("failed event requires error")
	}
	return nil
}

func Digest(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:]), nil
}

func validateID(name, value string, required bool) error {
	if value == "" && !required {
		return nil
	}
	if !safeID.MatchString(value) {
		return fmt.Errorf("%s must match the canonical identifier pattern", name)
	}
	return nil
}

func validateApproval(approval Approval) error {
	if err := validateID("approval.id", approval.ID, true); err != nil {
		return err
	}
	if approval.Kind == "" || approval.Summary == "" || approval.RiskLevel == "" || approval.RequestedAction == "" {
		return errors.New("approval fields are incomplete")
	}
	if containsRawSecret(approval.Metadata) {
		return errors.New("approval metadata contains raw secret material")
	}
	return nil
}

func validateControlMetadata(metadata map[string]any) error {
	for key := range metadata {
		if correlationMetadataKeys[key] || key == "command_id" || key == "expected_version" || key == "approval_id" {
			return fmt.Errorf("metadata duplicates reserved field %q", key)
		}
	}
	if containsRawSecret(metadata) {
		return errors.New("control metadata contains raw secret material")
	}
	return nil
}

func containsRawSecret(value any) bool {
	switch item := value.(type) {
	case map[string]any:
		for key, child := range item {
			name := strings.ToLower(strings.TrimSpace(key))
			sensitive := name == "password" || name == "token" || name == "access_token" ||
				name == "bearer_token" || name == "api_key" || name == "secret" ||
				name == "client_secret" || name == "authorization" || name == "credential"
			if !strings.HasSuffix(name, "_ref") && sensitive && fmt.Sprint(child) != "" {
				return true
			}
			if containsRawSecret(child) {
				return true
			}
		}
	case []any:
		for _, child := range item {
			if containsRawSecret(child) {
				return true
			}
		}
	}
	return false
}

func containsReservedAuthority(metadata map[string]any) bool {
	for _, key := range []string{
		"completion_authority", "requires_core_validation", "verified",
		"execution_authority", "storage",
	} {
		if _, exists := metadata[key]; exists {
			return true
		}
	}
	return false
}

func cloneMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return Clone(value)
}
