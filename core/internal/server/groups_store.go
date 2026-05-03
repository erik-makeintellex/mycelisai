package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func scanGroupRow(scanner interface{ Scan(dest ...any) error }) (*CollaborationGroup, error) {
	var (
		g              CollaborationGroup
		allowedRaw     []byte
		memberRaw      []byte
		teamRaw        []byte
		expiry         sql.NullTime
		createdAuditID string
		updatedAuditID string
	)
	if err := scanner.Scan(
		&g.ID,
		&g.TenantID,
		&g.Name,
		&g.GoalStatement,
		&g.WorkMode,
		&allowedRaw,
		&memberRaw,
		&teamRaw,
		&g.CoordinatorProfile,
		&g.ApprovalPolicyRef,
		&g.Status,
		&g.CreatedBy,
		&expiry,
		&createdAuditID,
		&updatedAuditID,
		&g.CreatedAt,
		&g.UpdatedAt,
	); err != nil {
		return nil, err
	}
	var err error
	g.AllowedCapabilities, err = unmarshalStringList(allowedRaw)
	if err != nil {
		return nil, fmt.Errorf("decode allowed_capabilities: %w", err)
	}
	g.MemberUserIDs, err = unmarshalStringList(memberRaw)
	if err != nil {
		return nil, fmt.Errorf("decode member_user_ids: %w", err)
	}
	g.TeamIDs, err = unmarshalStringList(teamRaw)
	if err != nil {
		return nil, fmt.Errorf("decode team_ids: %w", err)
	}
	if expiry.Valid {
		ts := expiry.Time.UTC()
		g.Expiry = &ts
	}
	g.CreatedAuditEventID = strings.TrimSpace(createdAuditID)
	g.UpdatedAuditEventID = strings.TrimSpace(updatedAuditID)
	return &g, nil
}

func (s *AdminServer) listGroupsDB(ctx context.Context) ([]CollaborationGroup, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, tenant_id, name, goal_statement, work_mode,
		       allowed_capabilities, member_user_ids, team_ids,
		       coordinator_profile, approval_policy_ref, status, created_by,
		       expiry,
		       COALESCE(created_audit_event_id::text, ''),
		       COALESCE(updated_audit_event_id::text, ''),
		       created_at, updated_at
		FROM collaboration_groups
		WHERE tenant_id='default'
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CollaborationGroup, 0)
	for rows.Next() {
		g, scanErr := scanGroupRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, *g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AdminServer) getGroupDB(ctx context.Context, id string) (*CollaborationGroup, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	row := db.QueryRowContext(ctx, `
		SELECT id::text, tenant_id, name, goal_statement, work_mode,
		       allowed_capabilities, member_user_ids, team_ids,
		       coordinator_profile, approval_policy_ref, status, created_by,
		       expiry,
		       COALESCE(created_audit_event_id::text, ''),
		       COALESCE(updated_audit_event_id::text, ''),
		       created_at, updated_at
		FROM collaboration_groups
		WHERE id=$1 AND tenant_id='default'`, id)

	g, err := scanGroupRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return g, err
}

func (s *AdminServer) insertGroupDB(ctx context.Context, g *CollaborationGroup) error {
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	return db.QueryRowContext(ctx, `
		INSERT INTO collaboration_groups (
			id, tenant_id, name, goal_statement, work_mode,
			allowed_capabilities, member_user_ids, team_ids,
			coordinator_profile, approval_policy_ref, status,
			created_by, expiry, created_audit_event_id, updated_audit_event_id
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15
		)
		RETURNING created_at, updated_at`,
		g.ID, g.TenantID, g.Name, g.GoalStatement, g.WorkMode,
		marshalStringList(g.AllowedCapabilities),
		marshalStringList(g.MemberUserIDs),
		marshalStringList(g.TeamIDs),
		g.CoordinatorProfile, g.ApprovalPolicyRef, g.Status,
		g.CreatedBy, g.Expiry,
		parseAuditUUID(g.CreatedAuditEventID),
		parseAuditUUID(g.UpdatedAuditEventID),
	).Scan(&g.CreatedAt, &g.UpdatedAt)
}

func (s *AdminServer) updateGroupDB(ctx context.Context, id string, req createGroupRequest, updatedAuditEventID string) (*CollaborationGroup, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	res, err := db.ExecContext(ctx, `
		UPDATE collaboration_groups
		SET name=$2,
		    goal_statement=$3,
		    work_mode=$4,
		    allowed_capabilities=$5,
		    member_user_ids=$6,
		    team_ids=$7,
		    coordinator_profile=$8,
		    approval_policy_ref=$9,
		    expiry=$10,
		    updated_audit_event_id=$11,
		    updated_at=NOW()
		WHERE id=$1 AND tenant_id='default'`,
		id,
		strings.TrimSpace(req.Name),
		strings.TrimSpace(req.GoalStatement),
		strings.TrimSpace(req.WorkMode),
		marshalStringList(req.AllowedCapabilities),
		marshalStringList(req.MemberUserIDs),
		marshalStringList(req.TeamIDs),
		strings.TrimSpace(req.CoordinatorProfile),
		strings.TrimSpace(req.ApprovalPolicyRef),
		req.Expiry,
		parseAuditUUID(updatedAuditEventID),
	)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}
	return s.getGroupDB(ctx, id)
}

func (s *AdminServer) updateGroupStatusDB(ctx context.Context, id, status, updatedAuditEventID string) (*CollaborationGroup, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	res, err := db.ExecContext(ctx, `
		UPDATE collaboration_groups
		SET status=$2,
		    updated_audit_event_id=$3,
		    updated_at=NOW()
		WHERE id=$1 AND tenant_id='default'`,
		id,
		strings.TrimSpace(status),
		parseAuditUUID(updatedAuditEventID),
	)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}
	return s.getGroupDB(ctx, id)
}
