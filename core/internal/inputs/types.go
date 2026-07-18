package inputs

import (
	"encoding/json"
	"time"
)

const (
	ScopeAll   = "all"
	ScopeGroup = "group"
	ScopeHost  = "host"

	AdapterAPI      = "api"
	AdapterWebhook  = "webhook"
	AdapterMCP      = "mcp"
	AdapterDevice   = "device"
	AdapterSensor   = "sensor"
	AdapterDatabase = "database"
	AdapterFile     = "file"

	AuthNone        = "none"
	AuthSecretRef   = "secret_ref"
	AuthBearerToken = "bearer_token"
	AuthAPIKey      = "api_key"
	AuthBasic       = "basic"

	BufferAppendLog      = "append_log"
	BufferLatestState    = "latest_state"
	BufferAppendLatest   = "append_with_latest"
	BufferWindowedRollup = "windowed_rollup"

	StatusAvailable = "available"
	StatusPaused    = "paused"
	StatusError     = "error"
)

// Source describes a governed ingress source that can feed Soma, teams, or outcomes.
type Source struct {
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	SourceType            string          `json:"source_type"`
	AdapterKind           string          `json:"adapter_kind"`
	ScopeKind             string          `json:"scope_kind"`
	ScopeRef              string          `json:"scope_ref,omitempty"`
	TargetOutcomeID       string          `json:"target_outcome_id,omitempty"`
	TargetGroupID         string          `json:"target_group_id,omitempty"`
	TargetHostID          string          `json:"target_host_id,omitempty"`
	AuthScheme            string          `json:"auth_scheme"`
	SecretRef             string          `json:"secret_ref,omitempty"`
	AllowedIngressSubject string          `json:"allowed_ingress_subject"`
	PayloadSchemaRef      string          `json:"payload_schema_ref,omitempty"`
	BufferMode            string          `json:"buffer_mode"`
	BufferPolicy          json.RawMessage `json:"buffer_policy,omitempty"`
	SensitivityClass      string          `json:"sensitivity_class"`
	TrustClass            string          `json:"trust_class"`
	Status                string          `json:"status"`
	Recovery              string          `json:"recovery,omitempty"`
	TenantID              string          `json:"tenant_id"`
	CreatedAt             time.Time       `json:"created_at,omitempty"`
	UpdatedAt             time.Time       `json:"updated_at,omitempty"`
}

type BufferEvent struct {
	EventID         string          `json:"event_id"`
	SourceID        string          `json:"source_id"`
	ChannelKey      string          `json:"channel_key"`
	Payload         json.RawMessage `json:"payload"`
	PayloadHash     string          `json:"payload_hash,omitempty"`
	SourceTimestamp *time.Time      `json:"source_timestamp,omitempty"`
	ReceivedAt      time.Time       `json:"received_at"`
	RunID           string          `json:"run_id,omitempty"`
	TeamID          string          `json:"team_id,omitempty"`
	AgentID         string          `json:"agent_id,omitempty"`
	SourceKind      string          `json:"source_kind"`
	SourceChannel   string          `json:"source_channel"`
	PayloadKind     string          `json:"payload_kind"`
	TenantID        string          `json:"tenant_id"`
}

type LatestValue struct {
	SourceID        string          `json:"source_id"`
	ChannelKey      string          `json:"channel_key"`
	EventID         string          `json:"event_id,omitempty"`
	Payload         json.RawMessage `json:"payload"`
	ReceivedAt      time.Time       `json:"received_at"`
	SourceTimestamp *time.Time      `json:"source_timestamp,omitempty"`
	TenantID        string          `json:"tenant_id"`
}

type WindowSummary struct {
	SourceID   string          `json:"source_id"`
	ChannelKey string          `json:"channel_key"`
	WindowKey  string          `json:"window_key"`
	Summary    string          `json:"summary"`
	Payload    json.RawMessage `json:"payload"`
	Count      int             `json:"count"`
	StartedAt  time.Time       `json:"started_at"`
	EndedAt    time.Time       `json:"ended_at"`
	TenantID   string          `json:"tenant_id"`
}

type BufferView struct {
	Mode    string          `json:"mode"`
	Source  Source          `json:"source"`
	Events  []BufferEvent   `json:"events,omitempty"`
	Latest  []LatestValue   `json:"latest,omitempty"`
	Windows []WindowSummary `json:"windows,omitempty"`
}
