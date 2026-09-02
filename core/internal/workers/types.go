package workers

import (
	"context"
	"time"
)

type BackendKind string
type RunStatus string
type EventKind string
type ApprovalDecision string
type Protocol string

const (
	BackendCentral       BackendKind = "central"
	BackendFrameworkRuns BackendKind = "framework_runs"

	StatusAccepted       RunStatus = "accepted"
	StatusRunning        RunStatus = "running"
	StatusApprovalNeeded RunStatus = "approval_needed"
	StatusCompleted      RunStatus = "completed"
	StatusFailed         RunStatus = "failed"
	StatusCancelled      RunStatus = "cancelled"

	EventAccepted       EventKind = "accepted"
	EventProgress       EventKind = "progress"
	EventApprovalNeeded EventKind = "approval_needed"
	EventCompleted      EventKind = "completed"
	EventFailed         EventKind = "failed"
	EventCancelled      EventKind = "cancelled"

	DecisionApprove ApprovalDecision = "approve"
	DecisionDeny    ApprovalDecision = "deny"

	ProtocolRunsAPI Protocol = "runs_api"
)

// WorkerBackend is the single execution interface used by agentry.
type WorkerBackend interface {
	CreateRun(context.Context, WorkerRunRequest) (WorkerRunHandle, error)
	StreamRunEvents(context.Context, string) (<-chan WorkerEvent, error)
	GetRun(context.Context, string) (WorkerRunHandle, error)
	StopRun(context.Context, string) error
	SubmitApproval(context.Context, string, WorkerApprovalDecision) error
	GetCapabilities(context.Context) (WorkerCapabilities, error)
	HealthCheck(context.Context) (WorkerHealth, error)
}

// RunFinalizer is implemented by backends that allow Mycelis to mark a run
// terminal after local governed execution has completed.
type RunFinalizer interface {
	CompleteRun(context.Context, string, WorkerResult) error
	FailRun(context.Context, string, *WorkerError) error
}

// GovernedControlBackend is the strict external control contract. Core owns
// command identity and supplies the worker snapshot version used for CAS.
type GovernedControlBackend interface {
	StopRunCommand(context.Context, string, WorkerStopCommand) (WorkerControlReceipt, error)
	SubmitApprovalCommand(context.Context, string, WorkerApprovalDecision) (WorkerControlReceipt, error)
}

type WorkerConfig struct {
	Backend           BackendKind   `json:"backend" yaml:"backend"`
	BaseURL           string        `json:"base_url,omitempty" yaml:"base_url,omitempty"`
	APIKeySecretRef   string        `json:"api_key_secret_ref,omitempty" yaml:"api_key_secret_ref,omitempty"`
	CapabilitiesPath  string        `json:"capabilities_endpoint,omitempty" yaml:"capabilities_endpoint,omitempty"`
	HealthPath        string        `json:"health_endpoint,omitempty" yaml:"health_endpoint,omitempty"`
	PreferredProtocol Protocol      `json:"preferred_protocol,omitempty" yaml:"preferred_protocol,omitempty"`
	ApprovalMode      string        `json:"approval_mode,omitempty" yaml:"approval_mode,omitempty"`
	EventStreamMode   string        `json:"event_stream_mode,omitempty" yaml:"event_stream_mode,omitempty"`
	TimeoutPolicy     TimeoutPolicy `json:"timeout_policy,omitempty" yaml:"timeout_policy,omitempty"`
}

type TimeoutPolicy struct {
	ConnectMS int `json:"connect_ms,omitempty" yaml:"connect_ms,omitempty"`
	RunMS     int `json:"run_ms,omitempty" yaml:"run_ms,omitempty"`
	StreamMS  int `json:"stream_ms,omitempty" yaml:"stream_ms,omitempty"`
}

type WorkerRunRequest struct {
	RunID             string            `json:"run_id"`
	OrgID             string            `json:"org_id,omitempty"`
	ProjectID         string            `json:"project_id,omitempty"`
	UserID            string            `json:"user_id,omitempty"`
	RequestedBy       string            `json:"requested_by,omitempty"`
	Intent            string            `json:"intent"`
	Instructions      string            `json:"instructions,omitempty"`
	Input             map[string]any    `json:"input,omitempty"`
	RequiredProtocols []Protocol        `json:"required_protocols,omitempty"`
	RequiredFeatures  []string          `json:"required_features,omitempty"`
	Correlation       WorkerCorrelation `json:"correlation"`
	Metadata          map[string]any    `json:"metadata,omitempty"`
}

// WorkerCorrelation is the non-secret control-plane identity carried across a
// worker boundary. Mycelis owns the only run identity on the wire.
type WorkerCorrelation struct {
	RunID               string `json:"run_id"`
	IntentProofID       string `json:"intent_proof_id"`
	ExecutionContractID string `json:"execution_contract_id"`
	TeamID              string `json:"team_id,omitempty"`
	WorkItemID          string `json:"work_item_id"`
	OutcomeID           string `json:"outcome_id,omitempty"`
	IdempotencyKey      string `json:"idempotency_key"`
	SourceKind          string `json:"source_kind"`
	SourceChannel       string `json:"source_channel"`
	PayloadKind         string `json:"payload_kind"`
	GraphRevision       string `json:"graph_revision"`
}

type WorkerRunHandle struct {
	RunID       string                 `json:"run_id"`
	Backend     BackendKind            `json:"backend"`
	Status      RunStatus              `json:"status"`
	Protocol    Protocol               `json:"protocol,omitempty"`
	Version     int64                  `json:"version"`
	Correlation WorkerCorrelation      `json:"correlation"`
	CreatedAt   time.Time              `json:"created_at,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at,omitempty"`
	Approval    *WorkerApprovalRequest `json:"approval,omitempty"`
	Result      *WorkerResult          `json:"result,omitempty"`
	Error       *WorkerError           `json:"error,omitempty"`
	AuditRecord *WorkerAuditRecord     `json:"audit_record,omitempty"`
	Usage       *WorkerUsage           `json:"usage,omitempty"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
}

type WorkerEvent struct {
	EventID     string                 `json:"event_id"`
	RunID       string                 `json:"run_id"`
	Backend     BackendKind            `json:"backend"`
	Sequence    int64                  `json:"sequence"`
	Version     int64                  `json:"version"`
	Correlation WorkerCorrelation      `json:"correlation"`
	Kind        EventKind              `json:"kind"`
	Status      RunStatus              `json:"status,omitempty"`
	Message     string                 `json:"message,omitempty"`
	Approval    *WorkerApprovalRequest `json:"approval,omitempty"`
	Result      *WorkerResult          `json:"result,omitempty"`
	Error       *WorkerError           `json:"error,omitempty"`
	Audit       *WorkerAuditRecord     `json:"audit,omitempty"`
	Usage       *WorkerUsage           `json:"usage,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
}

type WorkerApprovalRequest struct {
	ID              string         `json:"id"`
	Kind            string         `json:"kind"`
	Summary         string         `json:"summary"`
	RiskLevel       string         `json:"risk_level"`
	RequestedAction string         `json:"requested_action"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
}

type WorkerApprovalDecision struct {
	ApprovalID      string           `json:"approval_id"`
	Decision        ApprovalDecision `json:"decision"`
	CommandID       string           `json:"command_id"`
	ExpectedVersion int64            `json:"expected_version"`
	ActorID         string           `json:"actor_id"`
	Reason          string           `json:"reason,omitempty"`
	Metadata        map[string]any   `json:"metadata,omitempty"`
}

type WorkerStopCommand struct {
	CommandID       string         `json:"command_id"`
	ExpectedVersion int64          `json:"expected_version"`
	ActorID         string         `json:"actor_id"`
	Reason          string         `json:"reason,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

type WorkerControlReceipt struct {
	CommandID string       `json:"command_id"`
	RunID     string       `json:"run_id"`
	Kind      string       `json:"kind"`
	State     string       `json:"state"`
	Version   int64        `json:"version"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Error     *WorkerError `json:"error,omitempty"`
	Replayed  bool         `json:"-"`
}

type WorkerResult struct {
	Summary    string         `json:"summary,omitempty"`
	Outputs    []WorkerOutput `json:"outputs,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	FinishedAt time.Time      `json:"finished_at,omitempty"`
}

type WorkerOutput struct {
	ID          string         `json:"id,omitempty"`
	Kind        string         `json:"kind"`
	Name        string         `json:"name,omitempty"`
	URI         string         `json:"uri,omitempty"`
	ContentType string         `json:"content_type,omitempty"`
	SizeBytes   int64          `json:"size_bytes"`
	SHA256      string         `json:"sha256"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type WorkerError struct {
	Code        string         `json:"code"`
	Message     string         `json:"message"`
	Recoverable bool           `json:"recoverable"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type WorkerCapabilities struct {
	Backend              BackendKind    `json:"backend"`
	Healthy              bool           `json:"healthy"`
	SupportedProtocols   []Protocol     `json:"supported_protocols"`
	SupportsEvents       bool           `json:"supports_events"`
	SupportsCancellation bool           `json:"supports_cancellation"`
	SupportsApprovals    bool           `json:"supports_approvals"`
	SupportsUsage        bool           `json:"supports_usage"`
	Features             []string       `json:"features,omitempty"`
	Raw                  map[string]any `json:"raw,omitempty"`
}

type WorkerUsage struct {
	InputTokens  int64   `json:"input_tokens,omitempty"`
	OutputTokens int64   `json:"output_tokens,omitempty"`
	DurationMS   int64   `json:"duration_ms,omitempty"`
	CostEstimate float64 `json:"cost_estimate,omitempty"`
	Currency     string  `json:"currency,omitempty"`
}

type WorkerAuditRecord struct {
	RunID        string         `json:"run_id"`
	Backend      BackendKind    `json:"backend"`
	ActorID      string         `json:"actor_id,omitempty"`
	PolicyID     string         `json:"policy_id,omitempty"`
	DecisionPath []string       `json:"decision_path,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type WorkerHealth struct {
	Backend   BackendKind    `json:"backend"`
	Healthy   bool           `json:"healthy"`
	Message   string         `json:"message,omitempty"`
	CheckedAt time.Time      `json:"checked_at"`
	Raw       map[string]any `json:"raw,omitempty"`
}
