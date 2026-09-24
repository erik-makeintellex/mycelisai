package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const TeamProvisionPermission = "framework_runs.provision_team"

var (
	ErrTeamOwnershipDenied   = errors.New("team ownership authority denied")
	ErrTeamOwnershipMissing  = errors.New("runtime team manifest missing")
	ErrTeamOwnershipConflict = errors.New("team ownership conflict")
	ErrTeamOwnershipInvalid  = errors.New("invalid team ownership request")
)

type TeamOwnership struct {
	TeamID         string     `json:"team_id"`
	GroupID        string     `json:"group_id"`
	ManifestDigest string     `json:"manifest_digest"`
	Revision       string     `json:"revision"`
	OwnerAccountID string     `json:"owner_account_id,omitempty"`
	OwnerGroupID   string     `json:"owner_group_id,omitempty"`
	ProvisionedBy  string     `json:"ownership_provisioned_by,omitempty"`
	ProvisionedAt  *time.Time `json:"ownership_provisioned_at,omitempty"`
	RevokedBy      string     `json:"ownership_revoked_by,omitempty"`
	RevokedAt      *time.Time `json:"ownership_revoked_at,omitempty"`
}

type TeamOwnershipStore struct{ db *sql.DB }

func NewTeamOwnershipStore(db *sql.DB) *TeamOwnershipStore { return &TeamOwnershipStore{db: db} }

type ownershipAuthority struct {
	actor, account, tenant, group string
	expiresAt                     time.Time
}

func ownershipUUID(value string) (string, error) {
	id, err := uuid.Parse(value)
	if err != nil || value != strings.TrimSpace(value) {
		return "", ErrTeamOwnershipInvalid
	}
	return id.String(), nil
}

func ownershipTeamID(value string) (string, error) {
	if value == "" || len(value) > 256 || value != strings.TrimSpace(value) {
		return "", ErrTeamOwnershipInvalid
	}
	return value, nil
}

func ownershipReason(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 500 {
		return "", ErrTeamOwnershipInvalid
	}
	return value, nil
}

func (s *TeamOwnershipStore) Inspect(ctx context.Context, actorID, teamID, groupID string) (TeamOwnership, error) {
	return s.change(ctx, "inspect", actorID, teamID, groupID, "", "", "")
}

func (s *TeamOwnershipStore) Bind(ctx context.Context, actorID, teamID, groupID, expectedDigest, expectedRevision, reason string) (TeamOwnership, error) {
	return s.change(ctx, "bind", actorID, teamID, groupID, expectedDigest, expectedRevision, reason)
}

func (s *TeamOwnershipStore) Revoke(ctx context.Context, actorID, teamID, groupID, expectedDigest, expectedRevision, reason string) (TeamOwnership, error) {
	return s.change(ctx, "revoke", actorID, teamID, groupID, expectedDigest, expectedRevision, reason)
}

func (s *TeamOwnershipStore) change(ctx context.Context, action, actorID, teamID, groupID, expectedDigest, expectedRevision, reason string) (TeamOwnership, error) {
	if s == nil || s.db == nil {
		return TeamOwnership{}, fmt.Errorf("team ownership: database unavailable")
	}
	var err error
	actorID, err = ownershipUUID(actorID)
	if err != nil {
		return TeamOwnership{}, ErrTeamOwnershipDenied
	}
	groupID, err = ownershipUUID(groupID)
	if err != nil {
		return TeamOwnership{}, err
	}
	teamID, err = ownershipTeamID(teamID)
	if err != nil {
		return TeamOwnership{}, err
	}
	if action != "inspect" {
		if expectedDigest == "" || len(expectedDigest) > 256 || expectedRevision == "" || len(expectedRevision) > 64 {
			return TeamOwnership{}, ErrTeamOwnershipInvalid
		}
		reason, err = ownershipReason(reason)
		if err != nil {
			return TeamOwnership{}, err
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TeamOwnership{}, err
	}
	defer tx.Rollback()
	authority, err := lockOwnershipAuthority(ctx, tx, actorID, groupID)
	if err != nil {
		return TeamOwnership{}, err
	}
	current, tenant, err := lockTeamOwnership(ctx, tx, authority.tenant, teamID, groupID)
	if err != nil {
		return TeamOwnership{}, err
	}
	if tenant != authority.tenant {
		return TeamOwnership{}, ErrTeamOwnershipDenied
	}
	if current.OwnerAccountID != "" && current.OwnerAccountID != authority.account {
		return TeamOwnership{}, ErrTeamOwnershipDenied
	}
	if !authority.expiresAt.IsZero() {
		var now time.Time
		if err := tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
			return TeamOwnership{}, err
		}
		if !authority.expiresAt.After(now) {
			return TeamOwnership{}, ErrTeamOwnershipDenied
		}
	}
	if action == "inspect" {
		if err := tx.Commit(); err != nil {
			return TeamOwnership{}, err
		}
		return current, nil
	}
	if current.ManifestDigest != expectedDigest || current.Revision != expectedRevision {
		return TeamOwnership{}, ErrTeamOwnershipConflict
	}
	before := current
	switch action {
	case "bind":
		if current.OwnerGroupID != "" && current.RevokedAt == nil {
			return TeamOwnership{}, ErrTeamOwnershipConflict
		}
		err = tx.QueryRowContext(ctx, `UPDATE runtime_team_manifests SET owner_account_id=$1,owner_group_id=$2,
			ownership_provisioned_by=$3,ownership_provisioned_at=clock_timestamp(),ownership_revoked_by=NULL,
			ownership_revoked_at=NULL,updated_at=clock_timestamp() WHERE tenant_id=$4 AND team_id=$5
			RETURNING xmin::text,ownership_provisioned_at`, authority.account, groupID, actorID, tenant, teamID).
			Scan(&current.Revision, &current.ProvisionedAt)
		current.OwnerAccountID, current.OwnerGroupID, current.ProvisionedBy = authority.account, groupID, actorID
		current.RevokedBy, current.RevokedAt = "", nil
	case "revoke":
		if current.OwnerAccountID != authority.account || current.OwnerGroupID != groupID || current.RevokedAt != nil {
			return TeamOwnership{}, ErrTeamOwnershipConflict
		}
		err = tx.QueryRowContext(ctx, `UPDATE runtime_team_manifests SET ownership_revoked_by=$1,
			ownership_revoked_at=clock_timestamp(),updated_at=clock_timestamp() WHERE tenant_id=$2 AND team_id=$3
			RETURNING xmin::text,ownership_revoked_at`, actorID, tenant, teamID).
			Scan(&current.Revision, &current.RevokedAt)
		current.RevokedBy = actorID
	default:
		return TeamOwnership{}, ErrTeamOwnershipInvalid
	}
	if err != nil {
		return TeamOwnership{}, err
	}
	if err := insertOwnershipAudit(ctx, tx, authority.account, actorID, action, reason, before, current); err != nil {
		return TeamOwnership{}, err
	}
	if err := tx.Commit(); err != nil {
		return TeamOwnership{}, err
	}
	return current, nil
}

func lockOwnershipAuthority(ctx context.Context, tx *sql.Tx, actorID, groupID string) (ownershipAuthority, error) {
	var a ownershipAuthority
	var accountID sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT account_id::text FROM users WHERE id=$1`, actorID).Scan(&accountID); err != nil || !accountID.Valid {
		return a, ErrTeamOwnershipDenied
	}
	var accountStatus string
	if err := tx.QueryRowContext(ctx, `SELECT tenant_id,status FROM accounts WHERE id=$1 FOR UPDATE`, accountID.String).Scan(&a.tenant, &accountStatus); err != nil || accountStatus != "active" {
		return a, ErrTeamOwnershipDenied
	}
	var lockedAccount, userStatus string
	if err := tx.QueryRowContext(ctx, `SELECT account_id::text,status FROM users WHERE id=$1 FOR UPDATE`, actorID).Scan(&lockedAccount, &userStatus); err != nil || lockedAccount != accountID.String || userStatus != "active" {
		return a, ErrTeamOwnershipDenied
	}
	var groupAccount string
	if err := tx.QueryRowContext(ctx, `SELECT account_id::text FROM groups WHERE id=$1 FOR UPDATE`, groupID).Scan(&groupAccount); err != nil || groupAccount != accountID.String {
		return a, ErrTeamOwnershipDenied
	}
	rows, err := tx.QueryContext(ctx, `SELECT id::text,role_id::text,group_id::text,status,expires_at FROM org_memberships
		WHERE account_id=$1 AND user_id=$2 ORDER BY id FOR UPDATE`, accountID.String, actorID)
	if err != nil {
		return a, err
	}
	type membership struct {
		roleID, status string
		groupID        sql.NullString
		expiry         sql.NullTime
	}
	members := []membership{}
	roleSet := map[string]bool{}
	for rows.Next() {
		var id string
		var m membership
		if err := rows.Scan(&id, &m.roleID, &m.groupID, &m.status, &m.expiry); err != nil {
			rows.Close()
			return a, err
		}
		members = append(members, m)
		roleSet[m.roleID] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return a, err
	}
	roleIDs := make([]string, 0, len(roleSet))
	for id := range roleSet {
		roleIDs = append(roleIDs, id)
	}
	sort.Strings(roleIDs)
	qualified := map[string]bool{}
	for _, id := range roleIDs {
		var roleAccount sql.NullString
		var scope string
		if err := tx.QueryRowContext(ctx, `SELECT account_id::text,scope FROM roles WHERE id=$1 FOR UPDATE`, id).Scan(&roleAccount, &scope); err != nil {
			return a, ErrTeamOwnershipDenied
		}
		if roleAccount.Valid || scope != "system" {
			continue
		}
		var permission string
		err := tx.QueryRowContext(ctx, `SELECT permission_key FROM role_permissions WHERE role_id=$1 AND permission_key=$2 FOR UPDATE`, id, TeamProvisionPermission).Scan(&permission)
		if err != nil && err != sql.ErrNoRows {
			return a, err
		}
		qualified[id] = err == nil && permission == TeamProvisionPermission
	}
	var now time.Time
	if err := tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return a, err
	}
	count := 0
	for _, m := range members {
		if !m.groupID.Valid && m.status == "active" && qualified[m.roleID] && (!m.expiry.Valid || m.expiry.Time.After(now)) {
			count++
			if m.expiry.Valid {
				a.expiresAt = m.expiry.Time
			}
		}
	}
	if count != 1 {
		return a, ErrTeamOwnershipDenied
	}
	a.actor, a.account, a.group = actorID, accountID.String, groupID
	return a, nil
}

func lockTeamOwnership(ctx context.Context, tx *sql.Tx, expectedTenant, teamID, groupID string) (TeamOwnership, string, error) {
	row := TeamOwnership{TeamID: teamID, GroupID: groupID}
	var tenant string
	var ownerAccount, ownerGroup, provisioner, revoker sql.NullString
	var provisionedAt, revokedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT tenant_id,manifest_digest,xmin::text,owner_account_id::text,
		owner_group_id::text,ownership_provisioned_by::text,ownership_provisioned_at,
		ownership_revoked_by::text,ownership_revoked_at FROM runtime_team_manifests
		WHERE tenant_id=$1 AND team_id=$2 FOR UPDATE`, expectedTenant, teamID).
		Scan(&tenant, &row.ManifestDigest, &row.Revision, &ownerAccount, &ownerGroup, &provisioner, &provisionedAt, &revoker, &revokedAt)
	if err == sql.ErrNoRows {
		return TeamOwnership{}, "", ErrTeamOwnershipMissing
	}
	if err != nil {
		return TeamOwnership{}, "", err
	}
	row.OwnerAccountID, row.OwnerGroupID = ownerAccount.String, ownerGroup.String
	row.ProvisionedBy, row.RevokedBy = provisioner.String, revoker.String
	if provisionedAt.Valid {
		t := provisionedAt.Time.UTC()
		row.ProvisionedAt = &t
	}
	if revokedAt.Valid {
		t := revokedAt.Time.UTC()
		row.RevokedAt = &t
	}
	return row, tenant, nil
}

func insertOwnershipAudit(ctx context.Context, tx *sql.Tx, accountID, actorID, action, reason string, before, after TeamOwnership) error {
	payload, err := json.Marshal(map[string]any{"reason": reason, "before": before, "after": after})
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO identity_audit_events
		(account_id,actor_user_id,event_type,target_kind,target_id,source_kind,source_channel,payload)
		VALUES($1,$2,$3,'runtime_team_manifest',$4,'web_api','core.team_ownership',$5::jsonb)`,
		accountID, actorID, "framework_runs.team_ownership_"+action, after.TeamID, payload)
	return err
}
