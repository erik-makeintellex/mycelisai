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

type ExecutionRunClass string

const (
	ExecutionRunClassLinked ExecutionRunClass = "run_linked"
	ExecutionRunClassNoRun  ExecutionRunClass = "no_run"
)

type ExecutionProofClass string

const (
	ExecutionProofClassAuditOnly  ExecutionProofClass = "audit_only"
	ExecutionProofClassIntentOnly ExecutionProofClass = "intent_proof"
	ExecutionProofClassRunAudit   ExecutionProofClass = "run_and_audit"
	ExecutionProofClassGuidance   ExecutionProofClass = "guidance_only"
)

type ExecutionContractStatus string

const (
	ExecutionContractStatusProposed  ExecutionContractStatus = "proposed"
	ExecutionContractStatusConfirmed ExecutionContractStatus = "confirmed"
	ExecutionContractStatusCompleted ExecutionContractStatus = "completed"
	ExecutionContractStatusFailed    ExecutionContractStatus = "failed"
)

type ProofArtifactStatus string

const (
	ProofArtifactStatusSuccess  ProofArtifactStatus = "success"
	ProofArtifactStatusFailure  ProofArtifactStatus = "failure"
	ProofArtifactStatusDegraded ProofArtifactStatus = "degraded"
)

type TrustValidationSource string

const (
	TrustValidationSourceIntentProof   TrustValidationSource = "intent_proof"
	TrustValidationSourceConfirmAction TrustValidationSource = "confirm_action"
)

type TrustEvidenceStrength string

const (
	TrustEvidenceStrengthIntentOnly TrustEvidenceStrength = "intent_only"
	TrustEvidenceStrengthRunAudit   TrustEvidenceStrength = "run_audit"
	TrustEvidenceStrengthDegraded   TrustEvidenceStrength = "degraded"
)

type TrustProofQuality string

const (
	TrustProofQualityProposed TrustProofQuality = "proposed"
	TrustProofQualityVerified TrustProofQuality = "verified"
	TrustProofQualityFailed   TrustProofQuality = "failed"
)

type ExecutionRetentionClass string

const (
	ExecutionRetentionClassRetained    ExecutionRetentionClass = "retained"
	ExecutionRetentionClassNonRetained ExecutionRetentionClass = "non_retained"
	ExecutionRetentionClassExternalRef ExecutionRetentionClass = "external_reference"
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
	ContractID    string                 `json:"contract_id,omitempty"`
	ProofID       string                 `json:"proof_id,omitempty"`
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
	ID             string                  `json:"id,omitempty"`
	Kind           string                  `json:"kind"`
	Title          string                  `json:"title"`
	Summary        string                  `json:"summary,omitempty"`
	Href           string                  `json:"href,omitempty"`
	Entrypoint     string                  `json:"entrypoint,omitempty"`
	Folder         string                  `json:"folder,omitempty"`
	Files          []string                `json:"files,omitempty"`
	Validation     string                  `json:"validation,omitempty"`
	Retained       *bool                   `json:"retained,omitempty"`
	RetentionClass ExecutionRetentionClass `json:"retention_class,omitempty"`
}

type ExecutionProof struct {
	ContractID     string              `json:"contract_id,omitempty"`
	ProofID        string              `json:"proof_id,omitempty"`
	RunID          string              `json:"run_id,omitempty"`
	RunClass       ExecutionRunClass   `json:"run_class,omitempty"`
	NoRunReason    string              `json:"no_run_reason,omitempty"`
	ProofClass     ExecutionProofClass `json:"proof_class,omitempty"`
	AuditEventID   string              `json:"audit_event_id,omitempty"`
	IntentProofID  string              `json:"intent_proof_id,omitempty"`
	ExchangeItemID string              `json:"exchange_item_id,omitempty"`
	Verified       *bool               `json:"verified,omitempty"`
	Href           string              `json:"href,omitempty"`
}

type AuditRecovery struct {
	ApprovalStatus string                `json:"approval_status,omitempty"`
	RecoveryState  string                `json:"recovery_state,omitempty"`
	Blocker        string                `json:"blocker,omitempty"`
	Retryable      *bool                 `json:"retryable,omitempty"`
	Degradation    *ExecutionDegradation `json:"degradation,omitempty"`
}

type ExecutionDegradation struct {
	Code              string `json:"code,omitempty"`
	WhatFailed        string `json:"what_failed,omitempty"`
	TrustedState      string `json:"trusted_state,omitempty"`
	InvalidatedProof  string `json:"invalidated_proof,omitempty"`
	SafeContinuation  string `json:"safe_continuation,omitempty"`
	RequiresAttention bool   `json:"requires_attention,omitempty"`
}

type ExecutionNextStep struct {
	Label  string `json:"label"`
	Action string `json:"action,omitempty"`
	Href   string `json:"href,omitempty"`
}

type ExecutionContractRecord struct {
	ID                    string                  `json:"id"`
	IntentProofID         string                  `json:"intent_proof_id,omitempty"`
	RunID                 string                  `json:"run_id,omitempty"`
	TemplateID            TemplateID              `json:"template_id"`
	ResolvedIntent        string                  `json:"resolved_intent,omitempty"`
	ExecutionShape        ExecutionShape          `json:"execution_shape"`
	Status                ExecutionContractStatus `json:"status"`
	ExecutionStatus       ExecutionStatus         `json:"execution_status"`
	ValidationSource      TrustValidationSource   `json:"validation_source"`
	EvidenceStrength      TrustEvidenceStrength   `json:"evidence_strength"`
	ProofQuality          TrustProofQuality       `json:"proof_quality"`
	LatestProofArtifactID string                  `json:"latest_proof_artifact_id,omitempty"`
	AuditEventID          string                  `json:"audit_event_id,omitempty"`
	OutputRefs            []ExecutionOutput       `json:"output_refs,omitempty"`
	AuditRefs             []map[string]string     `json:"audit_refs,omitempty"`
	ReviewLineage         []map[string]string     `json:"review_lineage,omitempty"`
	Degradation           map[string]any          `json:"degradation,omitempty"`
	Recovery              map[string]any          `json:"recovery,omitempty"`
	CreatedAt             string                  `json:"created_at,omitempty"`
	ConfirmedAt           string                  `json:"confirmed_at,omitempty"`
	CompletedAt           string                  `json:"completed_at,omitempty"`
	UpdatedAt             string                  `json:"updated_at,omitempty"`
}

type ProofArtifactRecord struct {
	ID               string                `json:"id"`
	ContractID       string                `json:"contract_id,omitempty"`
	IntentProofID    string                `json:"intent_proof_id,omitempty"`
	RunID            string                `json:"run_id,omitempty"`
	ArtifactKind     string                `json:"artifact_kind"`
	Status           ProofArtifactStatus   `json:"status"`
	ProofClass       ExecutionProofClass   `json:"proof_class"`
	ValidationSource TrustValidationSource `json:"validation_source"`
	EvidenceStrength TrustEvidenceStrength `json:"evidence_strength"`
	ProofQuality     TrustProofQuality     `json:"proof_quality"`
	OutputRefs       []ExecutionOutput     `json:"output_refs,omitempty"`
	AuditRefs        []map[string]string   `json:"audit_refs,omitempty"`
	ReviewLineage    []map[string]string   `json:"review_lineage,omitempty"`
	Degradation      map[string]any        `json:"degradation,omitempty"`
	Recovery         map[string]any        `json:"recovery,omitempty"`
	Payload          map[string]any        `json:"payload,omitempty"`
	CreatedAt        string                `json:"created_at,omitempty"`
}
