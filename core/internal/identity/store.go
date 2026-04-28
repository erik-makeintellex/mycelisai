package identity

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateAccount(ctx context.Context, account Account) (*Account, error) {
	if s.db == nil {
		return nil, fmt.Errorf("identity store: database not available")
	}
	account = normalizeAccount(account)
	settings, err := marshalMap(account.Settings)
	if err != nil {
		return nil, fmt.Errorf("identity store: marshal account settings: %w", err)
	}
	row := s.db.QueryRowContext(ctx, `
		INSERT INTO accounts (id, tenant_id, slug, name, status, settings)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		ON CONFLICT (slug) DO UPDATE
		SET tenant_id = EXCLUDED.tenant_id,
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			settings = EXCLUDED.settings,
			updated_at = NOW()
		RETURNING id, tenant_id, slug, name, status, settings::text, created_at, updated_at
	`, account.ID, account.TenantID, account.Slug, account.Name, account.Status, settings)
	return scanAccount(row)
}

func (s *Store) CreateUser(ctx context.Context, user User) (*User, error) {
	if s.db == nil {
		return nil, fmt.Errorf("identity store: database not available")
	}
	user = normalizeUser(user)
	row := s.db.QueryRowContext(ctx, `
		INSERT INTO users
			(id, account_id, username, email, display_name, status, external_subject)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, NULLIF($7, ''))
		ON CONFLICT (id) DO UPDATE
		SET account_id = EXCLUDED.account_id,
			username = EXCLUDED.username,
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			status = EXCLUDED.status,
			external_subject = EXCLUDED.external_subject,
			updated_at = NOW()
		RETURNING id, account_id, username, COALESCE(email, ''), display_name, status,
			COALESCE(external_subject, ''), created_at, updated_at, last_login_at
	`, user.ID, user.AccountID, user.Username, user.Email, user.DisplayName, user.Status, user.ExternalSubject)
	return scanUser(row)
}

func (s *Store) CreateSession(ctx context.Context, session Session) (*Session, error) {
	if s.db == nil {
		return nil, fmt.Errorf("identity store: database not available")
	}
	session = normalizeSession(session)
	metadata, err := marshalMap(session.Metadata)
	if err != nil {
		return nil, fmt.Errorf("identity store: marshal session metadata: %w", err)
	}
	row := s.db.QueryRowContext(ctx, `
		INSERT INTO sessions
			(id, account_id, user_id, provider_id, token_hash, status, metadata, expires_at)
		VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, $6, $7::jsonb, $8)
		RETURNING id, account_id, user_id, COALESCE(provider_id::text, ''), token_hash, status,
			metadata::text, created_at, updated_at, expires_at, last_seen_at,
			revoked_at, COALESCE(revoked_by::text, ''), revoke_reason
	`, session.ID, session.AccountID, session.UserID, session.ProviderID,
		session.TokenHash, session.Status, metadata, session.ExpiresAt)
	return scanSession(row)
}

func (s *Store) AssignRole(ctx context.Context, membership OrgMembership) (*OrgMembership, error) {
	if s.db == nil {
		return nil, fmt.Errorf("identity store: database not available")
	}
	membership = normalizeMembership(membership)
	row := s.db.QueryRowContext(ctx, `
		WITH valid_role AS (SELECT id FROM roles WHERE id = $5 AND (account_id IS NULL OR account_id = $2))
		INSERT INTO org_memberships (id, account_id, user_id, group_id, role_id, status, expires_at)
		SELECT $1, $2, $3, NULLIF($4, '')::uuid, id, $6, $7 FROM valid_role
		ON CONFLICT (account_id, user_id, role_id, COALESCE(group_id, '00000000-0000-0000-0000-000000000000'::uuid))
		DO UPDATE SET status = EXCLUDED.status, expires_at = EXCLUDED.expires_at, updated_at = NOW()
		RETURNING id, account_id, user_id, COALESCE(group_id::text, ''), role_id,
			status, created_at, updated_at, expires_at
	`, membership.ID, membership.AccountID, membership.UserID, membership.GroupID,
		membership.RoleID, membership.Status, membership.ExpiresAt)
	return scanMembership(row)
}

func (s *Store) GetUserContext(ctx context.Context, accountID, userID string) (*UserContext, error) {
	if s.db == nil {
		return nil, fmt.Errorf("identity store: database not available")
	}
	accountID, userID = strings.TrimSpace(accountID), strings.TrimSpace(userID)
	if accountID == "" || userID == "" {
		return nil, fmt.Errorf("identity store: account_id and user_id are required")
	}
	account, user, err := s.getAccountUser(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}
	profile, err := s.getProfile(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}
	roles, groups, err := s.listRolesAndGroups(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}
	permissions, err := s.listPermissions(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}
	return &UserContext{
		Account: *account, User: *user, Profile: profile,
		Roles: roles, Groups: groups, Permissions: permissions,
		Metadata: map[string]any{"source": "identity_store"},
	}, nil
}

func (s *Store) RecordAuditEvent(ctx context.Context, event AuditEvent) (*AuditEvent, error) {
	if s.db == nil {
		return nil, fmt.Errorf("identity store: database not available")
	}
	event = normalizeAuditEvent(event)
	payload, err := marshalMap(event.Payload)
	if err != nil {
		return nil, fmt.Errorf("identity store: marshal audit payload: %w", err)
	}
	row := s.db.QueryRowContext(ctx, `
		INSERT INTO identity_audit_events
			(id, account_id, actor_user_id, actor_session_id, event_type,
			 target_kind, target_id, source_kind, source_channel, payload)
		VALUES ($1, $2, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, $5,
			$6, $7, $8, $9, $10::jsonb)
		RETURNING id, account_id, COALESCE(actor_user_id::text, ''),
			COALESCE(actor_session_id::text, ''), event_type, target_kind,
			target_id, source_kind, source_channel, payload::text, created_at
	`, event.ID, event.AccountID, event.ActorUserID, event.ActorSessionID,
		event.EventType, event.TargetKind, event.TargetID, event.SourceKind,
		event.SourceChannel, payload)
	return scanAuditEvent(row)
}

func (s *Store) getAccountUser(ctx context.Context, accountID, userID string) (*Account, *User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT a.id, a.tenant_id, a.slug, a.name, a.status, a.settings::text, a.created_at, a.updated_at,
			u.id, u.account_id, u.username, COALESCE(u.email, ''), u.display_name, u.status,
			COALESCE(u.external_subject, ''), u.created_at, u.updated_at, u.last_login_at
		FROM users u
		JOIN accounts a ON a.id = u.account_id
		WHERE u.account_id = $1 AND u.id = $2 AND u.status = 'active' AND a.status = 'active'
	`, accountID, userID)
	var account Account
	var user User
	if err := scanAccountUser(row, &account, &user); err != nil {
		return nil, nil, fmt.Errorf("identity store: get user context: %w", err)
	}
	return &account, &user, nil
}

func (s *Store) getProfile(ctx context.Context, accountID, userID string) (UserProfile, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT account_id, user_id, profile::text, context::text, created_at, updated_at
		FROM user_profiles
		WHERE account_id = $1 AND user_id = $2
	`, accountID, userID)
	profile, err := scanProfile(row)
	if err == sql.ErrNoRows {
		return UserProfile{AccountID: accountID, UserID: userID, Profile: map[string]any{}, Context: map[string]any{}}, nil
	}
	if err != nil {
		return UserProfile{}, fmt.Errorf("identity store: get profile: %w", err)
	}
	return *profile, nil
}

func (s *Store) listRolesAndGroups(ctx context.Context, accountID, userID string) ([]Role, []Group, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.id, COALESCE(r.account_id::text, ''), r.key, r.name, r.description, r.scope, r.created_at, r.updated_at,
			COALESCE(g.id::text, ''), COALESCE(g.account_id::text, ''), COALESCE(g.key, ''),
			COALESCE(g.name, ''), COALESCE(g.description, ''), g.created_at, g.updated_at
		FROM org_memberships m
		JOIN roles r ON r.id = m.role_id
		LEFT JOIN groups g ON g.id = m.group_id
		WHERE m.account_id = $1 AND m.user_id = $2 AND m.status = 'active'
			AND (m.expires_at IS NULL OR m.expires_at > NOW()) AND (r.account_id IS NULL OR r.account_id = m.account_id) AND (g.id IS NULL OR g.account_id = m.account_id)
		ORDER BY r.key, g.key
	`, accountID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("identity store: list roles: %w", err)
	}
	defer rows.Close()
	return scanRoleGroupRows(rows)
}

func (s *Store) listPermissions(ctx context.Context, accountID, userID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT rp.permission_key
		FROM org_memberships m
		JOIN roles r ON r.id = m.role_id
		JOIN role_permissions rp ON rp.role_id = m.role_id
		WHERE m.account_id = $1 AND m.user_id = $2 AND m.status = 'active'
			AND (m.expires_at IS NULL OR m.expires_at > NOW()) AND (r.account_id IS NULL OR r.account_id = m.account_id)
		ORDER BY rp.permission_key
	`, accountID, userID)
	if err != nil {
		return nil, fmt.Errorf("identity store: list permissions: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		out = append(out, permission)
	}
	if out == nil {
		out = []string{}
	}
	return out, rows.Err()
}

func normalizeAccount(account Account) Account {
	account.ID = defaultString(account.ID, uuid.NewString())
	account.TenantID = defaultString(account.TenantID, DefaultTenantID)
	account.Status = defaultString(account.Status, AccountStatusActive)
	if account.Settings == nil {
		account.Settings = map[string]any{}
	}
	return account
}

func normalizeUser(user User) User {
	user.ID = defaultString(user.ID, uuid.NewString())
	user.Status = defaultString(user.Status, UserStatusActive)
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	return user
}

func normalizeSession(session Session) Session {
	session.ID = defaultString(session.ID, uuid.NewString())
	session.Status = defaultString(session.Status, SessionStatusActive)
	if session.Metadata == nil {
		session.Metadata = map[string]any{}
	}
	return session
}

func normalizeMembership(membership OrgMembership) OrgMembership {
	membership.ID = defaultString(membership.ID, uuid.NewString())
	membership.Status = defaultString(membership.Status, UserStatusActive)
	return membership
}

func normalizeAuditEvent(event AuditEvent) AuditEvent {
	event.ID = defaultString(event.ID, uuid.NewString())
	event.SourceKind = defaultString(event.SourceKind, "system")
	event.SourceChannel = defaultString(event.SourceChannel, "identity.store")
	if event.Payload == nil {
		event.Payload = map[string]any{}
	}
	return event
}
