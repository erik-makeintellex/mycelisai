package protocol

// GovernanceProfileSnapshot captures the user-level governance posture that was
// active when Soma planned or approved an action.
type GovernanceProfileSnapshot struct {
	Role                 string `json:"role"`
	CostSensitivity      string `json:"cost_sensitivity"`
	ReviewStrictness     string `json:"review_strictness"`
	AutomationTolerance  string `json:"automation_tolerance"`
	EscalationPreference string `json:"escalation_preference"`
}

// ApprovalPolicy carries the approval decision and the thresholds that shaped
// it for a governed action.
type ApprovalPolicy struct {
	ApprovalRequired     bool                       `json:"approval_required"`
	ApprovalReason       string                     `json:"approval_reason,omitempty"`
	ApprovalMode         string                     `json:"approval_mode,omitempty"` // required | optional | auto_allowed
	CapabilityRisk       string                     `json:"capability_risk,omitempty"`
	CapabilityIDs        []string                   `json:"capability_ids,omitempty"`
	ExternalDataUse      bool                       `json:"external_data_use,omitempty"`
	EstimatedCost        float64                    `json:"estimated_cost,omitempty"`
	RequiredApproverRole string                     `json:"required_approver_role,omitempty"`
	ApprovalSteps        []string                   `json:"approval_steps,omitempty"`
	GovernanceProfile    *GovernanceProfileSnapshot `json:"governance_profile,omitempty"`
}

// AuditRecord is the normalized inspect-only audit response used by the UI.
type AuditRecord struct {
	ID             string         `json:"id"`
	TemplateID     string         `json:"template_id,omitempty"`
	Actor          string         `json:"actor"`
	User           string         `json:"user"`
	Action         string         `json:"action"`
	Timestamp      string         `json:"timestamp"`
	CapabilityUsed string         `json:"capability_used,omitempty"`
	ResultStatus   string         `json:"result_status"`
	ApprovalStatus string         `json:"approval_status,omitempty"`
	ApprovalReason string         `json:"approval_reason,omitempty"`
	RunID          string         `json:"run_id,omitempty"`
	IntentProofID  string         `json:"intent_proof_id,omitempty"`
	Resource       string         `json:"resource,omitempty"`
	Details        map[string]any `json:"details,omitempty"`
}
