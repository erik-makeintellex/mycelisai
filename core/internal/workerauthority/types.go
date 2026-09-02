package workerauthority

import (
	"encoding/json"
	"errors"
)

const (
	BackendFrameworkRuns = "framework_runs"
	ProtocolRunsAPI      = "runs_api"

	BindingCreatePending     = "create_pending"
	BindingBound             = "bound"
	BindingReconcileRequired = "reconcile_required"
	BindingTerminal          = "terminal"

	ApprovalPending   = "pending"
	ApprovalDecided   = "decided"
	ApprovalExpired   = "expired"
	ApprovalWithdrawn = "withdrawn"

	CommandStaged       = "staged"
	CommandPending      = "pending"
	CommandAcknowledged = "acknowledged"
	CommandFailed       = "failed"
	CommandUncertain    = "uncertain"
)

var (
	ErrUnavailable  = errors.New("worker authority store unavailable")
	ErrConflict     = errors.New("worker authority identity conflict")
	ErrStaleVersion = errors.New("worker authority version is stale")
	ErrEventGap     = errors.New("worker event sequence is not contiguous")
)

type Binding struct {
	RunID, IntentProofID, ExecutionContractID             string
	TeamID, WorkItemID, OutcomeID, IdempotencyKey         string
	GraphRevision, SourceKind, SourceChannel, PayloadKind string
	RequestDigest, State, RunStatus                       string
	ServiceVersion, LastEventSequence, CursorVersion      int64
}

type EventReceipt struct {
	ID, RunID, EventID, EventKind, PayloadDigest    string
	Sequence, ServiceVersion, ExpectedCursorVersion int64
	RunStatus                                       string
	Correlation                                     Correlation
	NormalizedPayload                               json.RawMessage
}

type Correlation struct {
	RunID, IntentProofID, ExecutionContractID string
	TeamID, WorkItemID, OutcomeID             string
	IdempotencyKey, GraphRevision             string
	SourceKind, SourceChannel, PayloadKind    string
}

type ApprovalRequest struct {
	ID, RunID, ApprovalID, RequestDigest string
	Kind, Summary, RiskLevel, Action     string
	State, Decision, DecidedBy, Reason   string
	Version                              int64
}

type ControlCommand struct {
	CommandID, RunID, ApprovalRequestID, IdempotencyKey string
	Kind, PayloadDigest, State                          string
	ExpectedServiceVersion, Version                     int64
	Payload                                             json.RawMessage
}
