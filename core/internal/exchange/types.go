package exchange

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type FieldDefinition struct {
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	SemanticMeaning string    `json:"semantic_meaning"`
	Indexed         bool      `json:"indexed"`
	Visibility      string    `json:"visibility"`
	UsageContexts   []string  `json:"usage_contexts"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
}

type SchemaDefinition struct {
	ID                   string    `json:"id"`
	Label                string    `json:"label"`
	Description          string    `json:"description"`
	RequiredFields       []string  `json:"required_fields"`
	OptionalFields       []string  `json:"optional_fields"`
	RequiredCapabilities []string  `json:"required_capabilities"`
	CreatedAt            time.Time `json:"created_at,omitempty"`
}

type CapabilityDefinition struct {
	ID                  string    `json:"id"`
	Label               string    `json:"label"`
	Source              string    `json:"source"`
	RiskClass           string    `json:"risk_class"`
	DefaultAllowedRoles []string  `json:"default_allowed_roles"`
	AuditRequired       bool      `json:"audit_required"`
	ApprovalRequired    bool      `json:"approval_required"`
	Description         string    `json:"description"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
}

type ChannelParticipant struct {
	Role     string `json:"role"`
	CanRead  bool   `json:"can_read"`
	CanWrite bool   `json:"can_write"`
}

type Channel struct {
	ID              uuid.UUID            `json:"id"`
	Name            string               `json:"name"`
	Type            string               `json:"type"`
	Owner           string               `json:"owner"`
	Participants    []ChannelParticipant `json:"participants"`
	Reviewers       []string             `json:"reviewers"`
	SchemaID        string               `json:"schema_id"`
	RetentionPolicy string               `json:"retention_policy"`
	Visibility      string               `json:"visibility"`
	SensitivityClass string              `json:"sensitivity_class"`
	Description     string               `json:"description"`
	Metadata        json.RawMessage      `json:"metadata"`
	CreatedAt       time.Time            `json:"created_at"`
}

type Thread struct {
	ID            uuid.UUID       `json:"id"`
	ChannelID     uuid.UUID       `json:"channel_id"`
	ChannelName   string          `json:"channel_name,omitempty"`
	ThreadType    string          `json:"thread_type"`
	Title         string          `json:"title"`
	Status        string          `json:"status"`
	Participants  []string        `json:"participants"`
	AllowedReviewers []string     `json:"allowed_reviewers"`
	EscalationRights []string     `json:"escalation_rights"`
	ContinuityKey string          `json:"continuity_key,omitempty"`
	CreatedBy     string          `json:"created_by"`
	Metadata      json.RawMessage `json:"metadata"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type ExchangeItem struct {
	ID          uuid.UUID       `json:"id"`
	ChannelID   uuid.UUID       `json:"channel_id"`
	ChannelName string          `json:"channel_name,omitempty"`
	SchemaID    string          `json:"schema_id"`
	Payload     json.RawMessage `json:"payload"`
	CreatedBy   string          `json:"created_by"`
	AddressedTo string          `json:"addressed_to,omitempty"`
	ThreadID    *uuid.UUID      `json:"thread_id,omitempty"`
	Visibility  string          `json:"visibility"`
	SensitivityClass string     `json:"sensitivity_class"`
	SourceRole  string          `json:"source_role"`
	SourceTeam  string          `json:"source_team,omitempty"`
	TargetRole  string          `json:"target_role,omitempty"`
	TargetTeam  string          `json:"target_team,omitempty"`
	AllowedConsumers []string   `json:"allowed_consumers"`
	CapabilityID string         `json:"capability_id,omitempty"`
	TrustClass  string          `json:"trust_class"`
	ReviewRequired bool         `json:"review_required"`
	Metadata    json.RawMessage `json:"metadata"`
	Summary     string          `json:"summary"`
	CreatedAt   time.Time       `json:"created_at"`
}

type PublishInput struct {
	ChannelName string
	SchemaID    string
	Payload     map[string]any
	CreatedBy   string
	AddressedTo string
	ThreadID    *uuid.UUID
	Visibility  string
	SensitivityClass string
	SourceRole  string
	SourceTeam  string
	TargetRole  string
	TargetTeam  string
	AllowedConsumers []string
	CapabilityID string
	TrustClass  string
	ReviewRequired bool
	Metadata    map[string]any
	Summary     string
}

type CreateThreadInput struct {
	ChannelName   string
	ThreadType    string
	Title         string
	Status        string
	Participants  []string
	AllowedReviewers []string
	EscalationRights []string
	ContinuityKey string
	CreatedBy     string
	Metadata      map[string]any
}

type SearchHit struct {
	Item      ExchangeItem `json:"item"`
	Distance  float64      `json:"distance"`
	ChannelID string       `json:"channel_id,omitempty"`
}

type EmbedFunc func(ctx context.Context, content string) ([]float64, error)

type Actor struct {
	UserID string
	Role   string
	Team   string
	Scopes []string
}
