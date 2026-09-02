package protocol

import (
	"encoding/json"
	"time"
)

type Status string
type EventKind string

const (
	StatusAccepted       Status = "accepted"
	StatusRunning        Status = "running"
	StatusApprovalNeeded Status = "approval_needed"
	StatusCompleted      Status = "completed"
	StatusFailed         Status = "failed"
	StatusCancelled      Status = "cancelled"

	EventAccepted       EventKind = "accepted"
	EventProgress       EventKind = "progress"
	EventApprovalNeeded EventKind = "approval_needed"
	EventCompleted      EventKind = "completed"
	EventFailed         EventKind = "failed"
	EventCancelled      EventKind = "cancelled"
)

type Correlation struct {
	RunID               string `json:"run_id"`
	IntentProofID       string `json:"intent_proof_id"`
	ExecutionContractID string `json:"execution_contract_id"`
	TeamID              string `json:"team_id,omitempty"`
	OutcomeID           string `json:"outcome_id,omitempty"`
	WorkItemID          string `json:"work_item_id"`
	IdempotencyKey      string `json:"idempotency_key"`
	SourceKind          string `json:"source_kind"`
	SourceChannel       string `json:"source_channel"`
	PayloadKind         string `json:"payload_kind"`
	GraphRevision       string `json:"graph_revision"`
}

type CreateRequest struct {
	RunID             string         `json:"run_id"`
	Intent            string         `json:"intent"`
	Correlation       Correlation    `json:"correlation"`
	OrgID             string         `json:"org_id"`
	ProjectID         string         `json:"project_id"`
	UserID            string         `json:"user_id"`
	RequestedBy       string         `json:"requested_by"`
	Instructions      string         `json:"instructions"`
	Input             map[string]any `json:"input"`
	RequiredProtocols []string       `json:"required_protocols"`
	RequiredFeatures  []string       `json:"required_features"`
	Metadata          map[string]any `json:"metadata"`
}

type Run struct {
	RunID       string         `json:"run_id"`
	Correlation Correlation    `json:"correlation"`
	Status      Status         `json:"status"`
	Version     uint64         `json:"version"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Approval    *Approval      `json:"approval,omitempty"`
	Result      *Result        `json:"result,omitempty"`
	Error       *Error         `json:"error,omitempty"`
	Usage       *Usage         `json:"usage,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type Event struct {
	EventID     string         `json:"event_id"`
	Sequence    uint64         `json:"sequence"`
	Version     uint64         `json:"version"`
	RunID       string         `json:"run_id"`
	Correlation Correlation    `json:"correlation"`
	Kind        EventKind      `json:"kind"`
	Status      Status         `json:"status"`
	Timestamp   time.Time      `json:"timestamp"`
	Message     string         `json:"message,omitempty"`
	Approval    *Approval      `json:"approval,omitempty"`
	Result      *Result        `json:"result,omitempty"`
	Error       *Error         `json:"error,omitempty"`
	Usage       *Usage         `json:"usage,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type Approval struct {
	ID              string         `json:"id"`
	Kind            string         `json:"kind"`
	Summary         string         `json:"summary"`
	RiskLevel       string         `json:"risk_level"`
	RequestedAction string         `json:"requested_action"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
}

type Result struct {
	Summary    string         `json:"summary"`
	Outputs    []Output       `json:"outputs"`
	Metadata   map[string]any `json:"metadata"`
	FinishedAt time.Time      `json:"finished_at"`
}

type Output struct {
	ID          string         `json:"id"`
	Kind        string         `json:"kind"`
	Name        string         `json:"name,omitempty"`
	URI         string         `json:"uri,omitempty"`
	ContentType string         `json:"content_type"`
	SizeBytes   int64          `json:"size_bytes"`
	SHA256      string         `json:"sha256"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type Error struct {
	Code        string         `json:"code"`
	Message     string         `json:"message"`
	Recoverable bool           `json:"recoverable"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type Usage struct {
	InputTokens  int64   `json:"input_tokens,omitempty"`
	OutputTokens int64   `json:"output_tokens,omitempty"`
	DurationMS   int64   `json:"duration_ms,omitempty"`
	CostEstimate float64 `json:"cost_estimate,omitempty"`
	Currency     string  `json:"currency,omitempty"`
}

type StopRequest struct {
	CommandID       string         `json:"command_id"`
	ExpectedVersion uint64         `json:"expected_version"`
	ActorID         string         `json:"actor_id"`
	Reason          string         `json:"reason,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

type ApprovalDecisionRequest struct {
	ApprovalID      string         `json:"approval_id"`
	Decision        string         `json:"decision"`
	CommandID       string         `json:"command_id"`
	ExpectedVersion uint64         `json:"expected_version"`
	ActorID         string         `json:"actor_id"`
	Reason          string         `json:"reason,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

type ControlReceipt struct {
	CommandID string    `json:"command_id"`
	RunID     string    `json:"run_id"`
	Kind      string    `json:"kind"`
	State     string    `json:"state"`
	Version   uint64    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Error     *Error    `json:"error,omitempty"`
}

type ExecutorOutcome struct {
	Status   Status         `json:"status"`
	Message  string         `json:"message,omitempty"`
	Approval *Approval      `json:"approval,omitempty"`
	Result   *Result        `json:"result,omitempty"`
	Error    *Error         `json:"error,omitempty"`
	Usage    *Usage         `json:"usage,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type Capabilities struct {
	Healthy              bool           `json:"healthy"`
	Backend              string         `json:"backend"`
	SupportedProtocols   []string       `json:"supported_protocols"`
	SupportsEvents       bool           `json:"supports_events"`
	SupportsCancellation bool           `json:"supports_cancellation"`
	SupportsApprovals    bool           `json:"supports_approvals"`
	SupportsUsage        bool           `json:"supports_usage"`
	ProductionReady      bool           `json:"production_ready"`
	Features             []string       `json:"features"`
	Metadata             map[string]any `json:"metadata,omitempty"`
}

type ErrorEnvelope struct {
	Error Error `json:"error"`
}

func Clone[T any](value T) T {
	raw, _ := json.Marshal(value)
	var cloned T
	_ = json.Unmarshal(raw, &cloned)
	return cloned
}
