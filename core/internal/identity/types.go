package identity

import "time"

const (
	DefaultTenantID = "default"

	AccountStatusActive  = "active"
	AccountStatusPaused  = "paused"
	AccountStatusDeleted = "deleted"

	UserStatusActive   = "active"
	UserStatusInvited  = "invited"
	UserStatusDisabled = "disabled"
	UserStatusDeleted  = "deleted"

	SessionStatusActive  = "active"
	SessionStatusRevoked = "revoked"
	SessionStatusExpired = "expired"
)

type Account struct {
	ID        string         `json:"id" db:"id"`
	TenantID  string         `json:"tenant_id" db:"tenant_id"`
	Slug      string         `json:"slug" db:"slug"`
	Name      string         `json:"name" db:"name"`
	Status    string         `json:"status" db:"status"`
	Settings  map[string]any `json:"settings" db:"settings"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

type User struct {
	ID              string     `json:"id" db:"id"`
	AccountID       string     `json:"account_id" db:"account_id"`
	Username        string     `json:"username" db:"username"`
	Email           string     `json:"email" db:"email"`
	DisplayName     string     `json:"display_name" db:"display_name"`
	Status          string     `json:"status" db:"status"`
	ExternalSubject string     `json:"external_subject" db:"external_subject"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

type Group struct {
	ID          string    `json:"id" db:"id"`
	AccountID   string    `json:"account_id" db:"account_id"`
	Key         string    `json:"key" db:"key"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Role struct {
	ID          string    `json:"id" db:"id"`
	AccountID   string    `json:"account_id" db:"account_id"`
	Key         string    `json:"key" db:"key"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Scope       string    `json:"scope" db:"scope"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type RolePermission struct {
	RoleID        string    `json:"role_id" db:"role_id"`
	PermissionKey string    `json:"permission_key" db:"permission_key"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type OrgMembership struct {
	ID        string     `json:"id" db:"id"`
	AccountID string     `json:"account_id" db:"account_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	GroupID   string     `json:"group_id" db:"group_id"`
	RoleID    string     `json:"role_id" db:"role_id"`
	Status    string     `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

type AuthProvider struct {
	ID        string         `json:"id" db:"id"`
	AccountID string         `json:"account_id" db:"account_id"`
	Type      string         `json:"type" db:"type"`
	Issuer    string         `json:"issuer" db:"issuer"`
	ClientID  string         `json:"client_id" db:"client_id"`
	Config    map[string]any `json:"config" db:"config"`
	Enabled   bool           `json:"enabled" db:"enabled"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

type Session struct {
	ID           string         `json:"id" db:"id"`
	AccountID    string         `json:"account_id" db:"account_id"`
	UserID       string         `json:"user_id" db:"user_id"`
	ProviderID   string         `json:"provider_id" db:"provider_id"`
	TokenHash    string         `json:"token_hash" db:"token_hash"`
	Status       string         `json:"status" db:"status"`
	Metadata     map[string]any `json:"metadata" db:"metadata"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
	ExpiresAt    time.Time      `json:"expires_at" db:"expires_at"`
	LastSeenAt   *time.Time     `json:"last_seen_at,omitempty" db:"last_seen_at"`
	RevokedAt    *time.Time     `json:"revoked_at,omitempty" db:"revoked_at"`
	RevokedBy    string         `json:"revoked_by" db:"revoked_by"`
	RevokeReason string         `json:"revoke_reason" db:"revoke_reason"`
}

type UserProfile struct {
	AccountID string         `json:"account_id" db:"account_id"`
	UserID    string         `json:"user_id" db:"user_id"`
	Profile   map[string]any `json:"profile" db:"profile"`
	Context   map[string]any `json:"context" db:"context"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

type AuditEvent struct {
	ID             string         `json:"id" db:"id"`
	AccountID      string         `json:"account_id" db:"account_id"`
	ActorUserID    string         `json:"actor_user_id" db:"actor_user_id"`
	ActorSessionID string         `json:"actor_session_id" db:"actor_session_id"`
	EventType      string         `json:"event_type" db:"event_type"`
	TargetKind     string         `json:"target_kind" db:"target_kind"`
	TargetID       string         `json:"target_id" db:"target_id"`
	SourceKind     string         `json:"source_kind" db:"source_kind"`
	SourceChannel  string         `json:"source_channel" db:"source_channel"`
	Payload        map[string]any `json:"payload" db:"payload"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
}

type UserContext struct {
	Account     Account        `json:"account"`
	User        User           `json:"user"`
	Profile     UserProfile    `json:"profile"`
	Roles       []Role         `json:"roles"`
	Groups      []Group        `json:"groups"`
	Permissions []string       `json:"permissions"`
	Metadata    map[string]any `json:"metadata"`
}
