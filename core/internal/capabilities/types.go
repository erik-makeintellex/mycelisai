package capabilities

import "time"

const ManifestVersion = "capability_manifest.v1"

// Manifest describes a runtime capability available to Soma, agents, MCP, or
// operator-controlled host actions.
type Manifest struct {
	ID                  string         `json:"id"`
	Version             string         `json:"version"`
	DisplayName         string         `json:"display_name"`
	Kind                string         `json:"kind"`
	Source              string         `json:"source"`
	Status              string         `json:"status"`
	RiskClass           string         `json:"risk_class"`
	Description         string         `json:"description,omitempty"`
	ToolRefs            []string       `json:"tool_refs,omitempty"`
	DefaultAllowedRoles []string       `json:"default_allowed_roles,omitempty"`
	AuditRequired       bool           `json:"audit_required"`
	ApprovalRequired    bool           `json:"approval_required"`
	Metadata            map[string]any `json:"metadata,omitempty"`
	DerivedAt           time.Time      `json:"derived_at"`
}

type Snapshot struct {
	GeneratedAt time.Time  `json:"generated_at"`
	Count       int        `json:"count"`
	Manifests   []Manifest `json:"manifests"`
}
