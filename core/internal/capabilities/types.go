package capabilities

import "time"

const ManifestVersion = "capability_manifest.v1"

// Manifest describes a runtime capability available to Soma, agents, MCP, or
// operator-controlled host actions.
type Manifest struct {
	ID                  string         `json:"id"`
	CapabilityID        string         `json:"capability_id"`
	Version             string         `json:"version"`
	ManifestVersion     string         `json:"manifest_version"`
	DisplayName         string         `json:"display_name"`
	Kind                string         `json:"kind"`
	Source              string         `json:"source"`
	Status              string         `json:"status"`
	Health              string         `json:"health"`
	RiskClass           string         `json:"risk_class"`
	Description         string         `json:"description,omitempty"`
	Purpose             string         `json:"purpose"`
	ToolRefs            []string       `json:"tool_refs,omitempty"`
	DefaultAllowedRoles []string       `json:"default_allowed_roles,omitempty"`
	AllowedRoles        []string       `json:"allowed_roles,omitempty"`
	AuditRequired       bool           `json:"audit_required"`
	ApprovalRequired    bool           `json:"approval_required"`
	ApprovalPosture     string         `json:"approval_posture"`
	InputSchemaRef      string         `json:"input_schema_ref"`
	OutputSchemaRef     string         `json:"output_schema_ref"`
	LastProbeStatus     string         `json:"last_probe_status"`
	LastProbeAt         time.Time      `json:"last_probe_at"`
	FailurePosture      string         `json:"failure_posture"`
	RecoveryPosture     string         `json:"recovery_posture"`
	AuditPolicy         string         `json:"audit_policy"`
	SecretRefPolicy     string         `json:"secret_ref_policy"`
	Owner               string         `json:"owner"`
	Metadata            map[string]any `json:"metadata,omitempty"`
	DerivedAt           time.Time      `json:"derived_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

type Snapshot struct {
	GeneratedAt time.Time  `json:"generated_at"`
	Count       int        `json:"count"`
	Manifests   []Manifest `json:"manifests"`
}
