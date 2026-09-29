package invocation

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
)

type membershipRow struct {
	id, roleID, status, version string
	expiresAt                   sql.NullTime
}

type roleRow struct {
	id, accountID, scope, version, permissionVersion string
	allowed                                          bool
}

// lockAuthority uses the same fixed lock order for proposal, confirmation,
// admission and execution-start. The preliminary user read only finds account;
// the locked user row must agree or admission fails.
func lockAuthority(ctx context.Context, tx *sql.Tx, userID, groupID string) (authority, error) {
	userID, err := parsedUUID(userID)
	if err != nil {
		return authority{}, ErrDenied
	}
	groupID, err = parsedUUID(groupID)
	if err != nil {
		return authority{}, ErrDenied
	}
	var accountID sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT account_id::text FROM users WHERE id=$1`, userID).Scan(&accountID); err != nil || !accountID.Valid {
		return authority{}, ErrDenied
	}
	var accountState, accountVersion string
	if err := tx.QueryRowContext(ctx, `SELECT status,xmin::text FROM accounts WHERE id=$1 FOR UPDATE`, accountID.String).Scan(&accountState, &accountVersion); err != nil || accountState != "active" {
		return authority{}, ErrDenied
	}
	var lockedAccount, userState, userVersion string
	if err := tx.QueryRowContext(ctx, `SELECT account_id::text,status,xmin::text FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&lockedAccount, &userState, &userVersion); err != nil || lockedAccount != accountID.String || userState != "active" {
		return authority{}, ErrDenied
	}
	var groupAccount, groupVersion string
	if err := tx.QueryRowContext(ctx, `SELECT account_id::text,xmin::text FROM groups WHERE id=$1 FOR UPDATE`, groupID).Scan(&groupAccount, &groupVersion); err != nil || groupAccount != accountID.String {
		return authority{}, ErrDenied
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT id::text,role_id::text,status,expires_at,xmin::text
		FROM org_memberships
		WHERE account_id=$1 AND user_id=$2 AND group_id=$3
		ORDER BY id FOR UPDATE`, accountID.String, userID, groupID)
	if err != nil {
		return authority{}, err
	}
	memberships := []membershipRow{}
	for rows.Next() {
		var m membershipRow
		if err := rows.Scan(&m.id, &m.roleID, &m.status, &m.expiresAt, &m.version); err != nil {
			rows.Close()
			return authority{}, err
		}
		memberships = append(memberships, m)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return authority{}, err
	}
	if len(memberships) == 0 {
		return authority{}, ErrDenied
	}
	roleIDs := make([]string, 0, len(memberships))
	seen := map[string]bool{}
	for _, m := range memberships {
		if !seen[m.roleID] {
			roleIDs = append(roleIDs, m.roleID)
			seen[m.roleID] = true
		}
	}
	sort.Strings(roleIDs)
	roles := map[string]roleRow{}
	for _, id := range roleIDs {
		var r roleRow
		var roleAccount sql.NullString
		err := tx.QueryRowContext(ctx, `SELECT id::text,account_id::text,scope,xmin::text FROM roles WHERE id=$1 FOR UPDATE`, id).
			Scan(&r.id, &roleAccount, &r.scope, &r.version)
		if err != nil {
			return authority{}, ErrDenied
		}
		if roleAccount.Valid {
			r.accountID = roleAccount.String
			if r.accountID != accountID.String {
				return authority{}, ErrDenied
			}
		} else if r.scope != "system" {
			return authority{}, ErrDenied
		}
		roles[id] = r
	}
	for _, id := range roleIDs {
		var permission, permissionVersion string
		err := tx.QueryRowContext(ctx, `SELECT permission_key,xmin::text FROM role_permissions WHERE role_id=$1 AND permission_key=$2 FOR UPDATE`, id, Permission).Scan(&permission, &permissionVersion)
		if err == nil && permission == Permission {
			r := roles[id]
			r.allowed = true
			r.permissionVersion = permissionVersion
			roles[id] = r
		} else if err != nil && err != sql.ErrNoRows {
			return authority{}, err
		}
	}
	now, err := dbNow(ctx, tx)
	if err != nil {
		return authority{}, err
	}
	var selected *authority
	for _, m := range memberships {
		r := roles[m.roleID]
		if m.status != "active" || !r.allowed || (m.expiresAt.Valid && !m.expiresAt.Time.After(now)) {
			continue
		}
		a := authority{
			AccountID: accountID.String, AccountState: accountState, AccountVersion: accountVersion,
			UserID: userID, UserState: userState, UserVersion: userVersion,
			GroupID: groupID, GroupVersion: groupVersion,
			MembershipID: m.id, MemberState: m.status, MemberVersion: m.version,
			RoleID: r.id, RoleScope: r.scope, RoleVersion: r.version,
			Permission: Permission, PermissionVersion: r.permissionVersion,
		}
		if m.expiresAt.Valid {
			a.MemberExpiry = m.expiresAt.Time.UTC()
		}
		if selected != nil {
			return authority{}, ErrDenied
		}
		selected = &a
	}
	if selected != nil {
		return *selected, nil
	}
	return authority{}, ErrDenied
}

func lockBinding(ctx context.Context, tx *sql.Tx) (binding, error) {
	var capabilityID, status string
	var metadata []byte
	err := tx.QueryRowContext(ctx, `
		SELECT capability_id,status,metadata::text FROM capability_manifests
		WHERE id=$1 FOR UPDATE`, CapabilityID).Scan(&capabilityID, &status, &metadata)
	if err != nil || capabilityID != CapabilityID || (status != "available" && status != "enabled") {
		return binding{}, ErrDenied
	}
	var document struct {
		InvocationBinding json.RawMessage `json:"invocation_binding"`
	}
	if err := json.Unmarshal(metadata, &document); err != nil || len(document.InvocationBinding) == 0 {
		return binding{}, ErrDenied
	}
	var b binding
	if err := json.Unmarshal(document.InvocationBinding, &b); err != nil {
		return binding{}, ErrDenied
	}
	if err := validBinding(b); err != nil {
		return binding{}, err
	}
	return b, nil
}
