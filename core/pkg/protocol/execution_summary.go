package protocol

type ExecutionShape string

const (
	ExecutionShapeDirectSoma       ExecutionShape = "direct_soma"
	ExecutionShapeGuidedProposal   ExecutionShape = "guided_proposal"
	ExecutionShapeToolAssistedWork ExecutionShape = "tool_assisted_work"
	ExecutionShapeTeamExecution    ExecutionShape = "team_execution"
	ExecutionShapeAutomation       ExecutionShape = "automation"
	ExecutionShapePluginCapability ExecutionShape = "plugin_capability"
)

type ExecutionStatus string

const (
	ExecutionStatusProposed  ExecutionStatus = "proposed"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusBlocked   ExecutionStatus = "blocked"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

type CapabilityUseKind string

const (
	CapabilityUseTool       CapabilityUseKind = "tool"
	CapabilityUseTeam       CapabilityUseKind = "team"
	CapabilityUseMCP        CapabilityUseKind = "mcp"
	CapabilityUseAutomation CapabilityUseKind = "automation"
	CapabilityUsePlugin     CapabilityUseKind = "plugin"
)

type ExecutionSummary struct {
	Intent        ExecutionIntent        `json:"intent"`
	Understanding ExecutionUnderstanding `json:"understanding"`
	Execution     ExecutionState         `json:"execution"`
	CapabilityUse []CapabilityUse        `json:"capability_use,omitempty"`
	Outputs       []ExecutionOutput      `json:"outputs,omitempty"`
	Proof         ExecutionProof         `json:"proof,omitempty"`
	AuditRecovery AuditRecovery          `json:"audit_recovery,omitempty"`
	NextStep      *ExecutionNextStep     `json:"next_step,omitempty"`
}

type ExecutionIntent struct {
	Original string `json:"original,omitempty"`
	Resolved string `json:"resolved"`
}

type ExecutionUnderstanding struct {
	Summary     string   `json:"summary"`
	Assumptions []string `json:"assumptions,omitempty"`
}

type ExecutionState struct {
	Shape   ExecutionShape  `json:"shape"`
	Status  ExecutionStatus `json:"status"`
	Summary string          `json:"summary"`
}

type CapabilityUse struct {
	ID     string            `json:"id"`
	Label  string            `json:"label,omitempty"`
	Kind   CapabilityUseKind `json:"kind"`
	Reason string            `json:"reason,omitempty"`
	Risk   string            `json:"risk,omitempty"`
}

type ExecutionOutput struct {
	ID       string `json:"id,omitempty"`
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	Summary  string `json:"summary,omitempty"`
	Href     string `json:"href,omitempty"`
	Retained *bool  `json:"retained,omitempty"`
}

type ExecutionProof struct {
	RunID         string `json:"run_id,omitempty"`
	AuditEventID  string `json:"audit_event_id,omitempty"`
	IntentProofID string `json:"intent_proof_id,omitempty"`
	Verified      *bool  `json:"verified,omitempty"`
	Href          string `json:"href,omitempty"`
}

type AuditRecovery struct {
	ApprovalStatus string `json:"approval_status,omitempty"`
	RecoveryState  string `json:"recovery_state,omitempty"`
	Blocker        string `json:"blocker,omitempty"`
	Retryable      *bool  `json:"retryable,omitempty"`
}

type ExecutionNextStep struct {
	Label  string `json:"label"`
	Action string `json:"action,omitempty"`
	Href   string `json:"href,omitempty"`
}
