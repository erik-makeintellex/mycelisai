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
	BackendCentral    BackendKind = "central"
	BackendHermesAPI  BackendKind = "hermes_api"
	BackendHermesLike BackendKind = "hermes_like"

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

	ProtocolRunsAPI        Protocol = "runs_api"
	ProtocolResponsesAPI   Protocol = "responses_api"
	ProtocolChatCompletion Protocol = "chat_completions"
	ProtocolUnknown        Protocol = "unknown"
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

type WorkerConfig struct {
	Backend            BackendKind   `json:"backend"`
	BaseURL            string        `json:"base_url,omitempty"`
	APIKeySecretRef    string        `json:"api_key_secret_ref,omitempty"`
	CapabilitiesPath   string        `json:"capabilities_endpoint,omitempty"`
	HealthPath         string        `json:"health_endpoint,omitempty"`
	PreferredProtocol  Protocol      `json:"preferred_protocol,omitempty"`
	SessionKeyStrategy string        `json:"session_key_strategy,omitempty"`
	ApprovalMode       string        `json:"approval_mode,omitempty"`
	EventStreamMode    string        `json:"event_stream_mode,omitempty"`
	TimeoutPolicy      TimeoutPolicy `json:"timeout_policy,omitempty"`
	ToolPolicy         ToolPolicy    `json:"tool_policy,omitempty"`
	FallbackBackend    BackendKind   `json:"fallback_backend,omitempty"`
}

type TimeoutPolicy struct {
	ConnectMS int `json:"connect_ms,omitempty"`
	RunMS     int `json:"run_ms,omitempty"`
	StreamMS  int `json:"stream_ms,omitempty"`
}

type ToolPolicy struct {
	AllowNetwork bool     `json:"allow_network"`
	AllowFiles   bool     `json:"allow_files"`
	AllowBrowser bool     `json:"allow_browser"`
	AllowedTools []string `json:"allowed_tools,omitempty"`
}

type WorkerRunRequest struct {
	OrgID             string         `json:"org_id,omitempty"`
	ProjectID         string         `json:"project_id,omitempty"`
	UserID            string         `json:"user_id,omitempty"`
	RequestedBy       string         `json:"requested_by,omitempty"`
	Intent            string         `json:"intent"`
	Instructions      string         `json:"instructions,omitempty"`
	Input             map[string]any `json:"input,omitempty"`
	RequiredProtocols []Protocol     `json:"required_protocols,omitempty"`
	RequiredFeatures  []string       `json:"required_features,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

type WorkerRunHandle struct {
	RunID       string                 `json:"run_id"`
	Backend     BackendKind            `json:"backend"`
	Status      RunStatus              `json:"status"`
	Protocol    Protocol               `json:"protocol,omitempty"`
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
	RunID     string                 `json:"run_id"`
	Backend   BackendKind            `json:"backend"`
	Kind      EventKind              `json:"kind"`
	Status    RunStatus              `json:"status,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Approval  *WorkerApprovalRequest `json:"approval,omitempty"`
	Result    *WorkerResult          `json:"result,omitempty"`
	Error     *WorkerError           `json:"error,omitempty"`
	Audit     *WorkerAuditRecord     `json:"audit,omitempty"`
	Usage     *WorkerUsage           `json:"usage,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]any         `json:"metadata,omitempty"`
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
	ApprovalID string           `json:"approval_id"`
	Decision   ApprovalDecision `json:"decision"`
	ActorID    string           `json:"actor_id,omitempty"`
	Reason     string           `json:"reason,omitempty"`
	Metadata   map[string]any   `json:"metadata,omitempty"`
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
