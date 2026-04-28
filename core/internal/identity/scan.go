package identity

import (
	"database/sql"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanAccount(row scanner) (*Account, error) {
	var account Account
	var settings string
	if err := row.Scan(&account.ID, &account.TenantID, &account.Slug, &account.Name,
		&account.Status, &settings, &account.CreatedAt, &account.UpdatedAt); err != nil {
		return nil, err
	}
	account.Settings = mapFromJSON(settings)
	return &account, nil
}

func scanUser(row scanner) (*User, error) {
	var user User
	var lastLogin sql.NullTime
	err := row.Scan(&user.ID, &user.AccountID, &user.Username, &user.Email,
		&user.DisplayName, &user.Status, &user.ExternalSubject,
		&user.CreatedAt, &user.UpdatedAt, &lastLogin)
	if err != nil {
		return nil, err
	}
	if lastLogin.Valid {
		user.LastLoginAt = &lastLogin.Time
	}
	return &user, nil
}

func scanAccountUser(row scanner, account *Account, user *User) error {
	var settings string
	var lastLogin sql.NullTime
	err := row.Scan(&account.ID, &account.TenantID, &account.Slug, &account.Name,
		&account.Status, &settings, &account.CreatedAt, &account.UpdatedAt,
		&user.ID, &user.AccountID, &user.Username, &user.Email, &user.DisplayName,
		&user.Status, &user.ExternalSubject, &user.CreatedAt, &user.UpdatedAt, &lastLogin)
	if err != nil {
		return err
	}
	account.Settings = mapFromJSON(settings)
	if lastLogin.Valid {
		user.LastLoginAt = &lastLogin.Time
	}
	return nil
}

func scanMembership(row scanner) (*OrgMembership, error) {
	var membership OrgMembership
	var expires sql.NullTime
	err := row.Scan(&membership.ID, &membership.AccountID, &membership.UserID,
		&membership.GroupID, &membership.RoleID, &membership.Status,
		&membership.CreatedAt, &membership.UpdatedAt, &expires)
	if err != nil {
		return nil, err
	}
	if expires.Valid {
		membership.ExpiresAt = &expires.Time
	}
	return &membership, nil
}

func scanSession(row scanner) (*Session, error) {
	var session Session
	var metadata string
	var lastSeen, revoked sql.NullTime
	err := row.Scan(&session.ID, &session.AccountID, &session.UserID,
		&session.ProviderID, &session.TokenHash, &session.Status, &metadata,
		&session.CreatedAt, &session.UpdatedAt, &session.ExpiresAt, &lastSeen,
		&revoked, &session.RevokedBy, &session.RevokeReason)
	if err != nil {
		return nil, err
	}
	session.Metadata = mapFromJSON(metadata)
	if lastSeen.Valid {
		session.LastSeenAt = &lastSeen.Time
	}
	if revoked.Valid {
		session.RevokedAt = &revoked.Time
	}
	return &session, nil
}

func scanProfile(row scanner) (*UserProfile, error) {
	var profile UserProfile
	var profileJSON, contextJSON string
	err := row.Scan(&profile.AccountID, &profile.UserID, &profileJSON,
		&contextJSON, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		return nil, err
	}
	profile.Profile = mapFromJSON(profileJSON)
	profile.Context = mapFromJSON(contextJSON)
	return &profile, nil
}

func scanAuditEvent(row scanner) (*AuditEvent, error) {
	var event AuditEvent
	var payload string
	err := row.Scan(&event.ID, &event.AccountID, &event.ActorUserID,
		&event.ActorSessionID, &event.EventType, &event.TargetKind,
		&event.TargetID, &event.SourceKind, &event.SourceChannel,
		&payload, &event.CreatedAt)
	if err != nil {
		return nil, err
	}
	event.Payload = mapFromJSON(payload)
	return &event, nil
}

func scanRoleGroupRows(rows *sql.Rows) ([]Role, []Group, error) {
	roles := []Role{}
	groups := []Group{}
	seenRoles, seenGroups := map[string]bool{}, map[string]bool{}
	for rows.Next() {
		var role Role
		var group Group
		var groupCreated, groupUpdated sql.NullTime
		err := rows.Scan(&role.ID, &role.AccountID, &role.Key, &role.Name,
			&role.Description, &role.Scope, &role.CreatedAt, &role.UpdatedAt,
			&group.ID, &group.AccountID, &group.Key, &group.Name,
			&group.Description, &groupCreated, &groupUpdated)
		if err != nil {
			return nil, nil, err
		}
		if !seenRoles[role.ID] {
			roles = append(roles, role)
			seenRoles[role.ID] = true
		}
		if group.ID != "" && !seenGroups[group.ID] {
			if groupCreated.Valid {
				group.CreatedAt = groupCreated.Time
			}
			if groupUpdated.Valid {
				group.UpdatedAt = groupUpdated.Time
			}
			groups = append(groups, group)
			seenGroups[group.ID] = true
		}
	}
	return roles, groups, rows.Err()
}
