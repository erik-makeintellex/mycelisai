package protocol

import "time"

// Deployment kinds returned by GET /api/v1/deployments. A deployment is
// anything currently defined to run: a team, a group, or a department an
// organization plans to launch.
const (
	DeploymentKindStandingTeam      = "standing_team"
	DeploymentKindMissionTeam       = "mission_team"
	DeploymentKindGroup             = "group"
	DeploymentKindPlannedDepartment = "planned_department"
)

// DeploymentKinds lists every deployment kind in display order.
var DeploymentKinds = []string{
	DeploymentKindStandingTeam,
	DeploymentKindMissionTeam,
	DeploymentKindGroup,
	DeploymentKindPlannedDepartment,
}

// Deployment lifecycle states.
const (
	DeploymentStateRunning  = "running"
	DeploymentStateIdle     = "idle"
	DeploymentStatePlanned  = "planned"
	DeploymentStatePaused   = "paused"
	DeploymentStateArchived = "archived"
	DeploymentStateExpired  = "expired"
	DeploymentStateError    = "error"
)

// DeploymentStates lists every lifecycle state accepted by the ?state= filter.
var DeploymentStates = []string{
	DeploymentStateRunning,
	DeploymentStateIdle,
	DeploymentStatePlanned,
	DeploymentStatePaused,
	DeploymentStateArchived,
	DeploymentStateExpired,
	DeploymentStateError,
}

// Deployment scope kinds.
const (
	DeploymentScopeOrganization = "organization"
	DeploymentScopeMission      = "mission"
	DeploymentScopeGlobal       = "global"
)

// Deployment source status values reported per source adapter.
const (
	DeploymentSourceOK          = "ok"
	DeploymentSourceDegraded    = "degraded"
	DeploymentSourceUnavailable = "unavailable"
	DeploymentSourceRestricted  = "restricted"
)

// Provider origins for a deployment member.
const (
	ProviderOriginManifest = "manifest" // explicit provider on the (policy-resolved) manifest
	ProviderOriginProfile  = "profile"  // routed through the cognitive profile binding
	ProviderOriginFallback = "fallback" // router fell back to another available engine
	ProviderOriginDeclared = "declared" // router offline; declared value only
	ProviderOriginPlanned  = "planned"  // not running yet; organization AI engine profile
)

// DeploymentScope says where a deployment lives.
type DeploymentScope struct {
	Kind string `json:"kind"` // organization | mission | global
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// DeploymentTool is one internal tool, MCP tool, or group capability.
type DeploymentTool struct {
	Name   string `json:"name"`
	Source string `json:"source"` // internal | mcp | capability
}

// DeploymentMember is one agent, team, coordinator, or person in a deployment.
// Provider and Model are ids only; secrets never appear here.
type DeploymentMember struct {
	ID              string           `json:"id"`
	Role            string           `json:"role"`
	Profile         string           `json:"profile,omitempty"`
	Provider        string           `json:"provider,omitempty"`
	Model           string           `json:"model,omitempty"`
	ProviderOrigin  string           `json:"provider_origin,omitempty"`
	ProviderStatus  string           `json:"provider_status,omitempty"` // available | unavailable code | unknown | planned
	EngineSummary   string           `json:"engine_summary,omitempty"`
	Status          string           `json:"status,omitempty"` // offline | idle | busy | error
	LastHeartbeatAt *time.Time       `json:"last_heartbeat_at,omitempty"`
	Tools           []DeploymentTool `json:"tools"`
}

// DeploymentOutputsSummary counts retained outputs linked to a deployment.
type DeploymentOutputsSummary struct {
	Retained     int            `json:"retained"`
	ByStatus     map[string]int `json:"by_status"`
	ProofStatus  string         `json:"proof_status"` // not_recorded until outputs carry proof envelopes
	LastOutputAt *time.Time     `json:"last_output_at,omitempty"`
}

// DeploymentOutputItem describes a recent retained output linked to a deployment.
type DeploymentOutputItem struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	FilePath    string     `json:"file_path,omitempty"`
	ContentType string     `json:"content_type,omitempty"`
	Bytes       int64      `json:"bytes,omitempty"`
	Status      string     `json:"status,omitempty"`
	DownloadURL string     `json:"download_url,omitempty"`
	ViewURL     string     `json:"view_url,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

// DeploymentConfigRef points at the config view resolved by RC-004F.
type DeploymentConfigRef struct {
	Kind string `json:"kind"` // team | group | organization
	ID   string `json:"id"`
	Path string `json:"path"`
}

// DeploymentRecord is one row on the Deployments page.
type DeploymentRecord struct {
	ID                string                    `json:"id"`
	Kind              string                    `json:"kind"`
	Name              string                    `json:"name"`
	Purpose           string                    `json:"purpose,omitempty"`
	WhatRuns          string                    `json:"what_runs"`
	Scope             DeploymentScope           `json:"scope"`
	Origin            string                    `json:"origin"`
	RuntimeID         string                    `json:"runtime_id,omitempty"`
	LifecycleState    string                    `json:"lifecycle_state"`
	Since             *time.Time                `json:"since,omitempty"`
	ExpiresAt         *time.Time                `json:"expires_at,omitempty"`
	LastActivity      *time.Time                `json:"last_activity,omitempty"`
	Members           []DeploymentMember        `json:"members"`
	Tools             []DeploymentTool          `json:"tools"`
	InputChannels     []string                  `json:"input_channels,omitempty"`
	OutputChannels    []string                  `json:"output_channels,omitempty"`
	WorkMode          string                    `json:"work_mode,omitempty"`
	ApprovalPolicyRef string                    `json:"approval_policy_ref,omitempty"`
	ApprovalPosture   string                    `json:"approval_posture,omitempty"`
	OutputsSummary    *DeploymentOutputsSummary `json:"outputs_summary,omitempty"`
	RecentOutputs     []DeploymentOutputItem    `json:"recent_outputs,omitempty"`
	ConfigRef         DeploymentConfigRef       `json:"config_ref"`
}

// DeploymentSourceStatus reports how one source adapter fared.
type DeploymentSourceStatus struct {
	Source string `json:"source"` // standing_teams | mission_teams | groups | organizations | outputs
	Status string `json:"status"` // ok | degraded | unavailable | restricted
	Detail string `json:"detail,omitempty"`
	Count  int    `json:"count"`
}

// DeploymentList is the data payload of GET /api/v1/deployments.
type DeploymentList struct {
	Deployments []DeploymentRecord       `json:"deployments"`
	Sources     []DeploymentSourceStatus `json:"sources"`
	GeneratedAt time.Time                `json:"generated_at"`
}

// DeploymentDetail is the data payload of GET /api/v1/deployments/{kind}/{id}.
type DeploymentDetail struct {
	Deployment DeploymentRecord         `json:"deployment"`
	Sources    []DeploymentSourceStatus `json:"sources"`
}

// IsDeploymentKind reports whether kind is a known deployment kind.
func IsDeploymentKind(kind string) bool { return deploymentListContains(DeploymentKinds, kind) }

// IsDeploymentState reports whether state is a known lifecycle state.
func IsDeploymentState(state string) bool { return deploymentListContains(DeploymentStates, state) }

func deploymentListContains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
