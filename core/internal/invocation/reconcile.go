package invocation

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (s *Store) Revoke(ctx context.Context, userID, grantID, reason string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("invocation database unavailable")
	}
	userID, err := parsedUUID(userID)
	if err != nil {
		return ErrDenied
	}
	grantID, err = parsedUUID(grantID)
	if err != nil {
		return ErrInvalid
	}
	owner, groupID, err := grantCoordinates(ctx, s.db, grantID)
	if err != nil || owner != userID {
		return ErrDenied
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	a, err := lockAuthority(ctx, tx, userID, groupID)
	if err != nil {
		return err
	}
	g, err := loadGrantLocked(ctx, tx, grantID)
	if err != nil || g.UserID != userID || g.AccountID != a.AccountID {
		return ErrDenied
	}
	_, revoked, err := lockGrantState(ctx, tx, grantID)
	if err != nil {
		return err
	}
	if !revoked.Valid {
		_, err = tx.ExecContext(ctx, `
			UPDATE execution_effect_grant_state SET revoked_at=NOW(),revoked_by=$2,revoke_reason=$3
			WHERE grant_id=$1 AND revoked_at IS NULL`, grantID, userID, strings.TrimSpace(reason))
		if err != nil {
			return err
		}
		if err := auditTx(ctx, tx, grantID, "counting_grant_revoked", map[string]any{"actor_user_id": userID, "reason": reason}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Reconcile(ctx context.Context, userID, invocationID, observation string, evidence Evidence) (Invocation, error) {
	if s == nil || s.db == nil {
		return Invocation{}, fmt.Errorf("invocation database unavailable")
	}
	userID, err := parsedUUID(userID)
	if err != nil {
		return Invocation{}, ErrDenied
	}
	invocationID, err = parsedUUID(invocationID)
	if err != nil {
		return Invocation{}, ErrInvalid
	}
	if observation != "committed" && observation != "not_committed" && observation != "still_unknown" {
		return Invocation{}, ErrInvalid
	}
	if evidence.InvocationID != invocationID || strings.TrimSpace(evidence.Source) == "" || strings.TrimSpace(evidence.Summary) == "" {
		return Invocation{}, ErrInvalid
	}
	if observation == "committed" && (evidence.ObservedCount == nil || *evidence.ObservedCount < 1) {
		return Invocation{}, ErrInvalid
	}
	if observation == "not_committed" && (evidence.ObservedCount == nil || *evidence.ObservedCount != 0) {
		return Invocation{}, ErrInvalid
	}
	var groupID, invocationUser string
	err = s.db.QueryRowContext(ctx, `SELECT group_id::text,user_id::text FROM execution_invocations WHERE id=$1`, invocationID).
		Scan(&groupID, &invocationUser)
	if err != nil || invocationUser != userID {
		return Invocation{}, ErrDenied
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Invocation{}, err
	}
	defer tx.Rollback()
	a, err := lockAuthority(ctx, tx, userID, groupID)
	if err != nil {
		return Invocation{}, err
	}
	var grantID string
	err = tx.QueryRowContext(ctx, `SELECT grant_id::text FROM execution_invocations WHERE id=$1`, invocationID).Scan(&grantID)
	if err != nil {
		return Invocation{}, ErrDenied
	}
	g, err := loadGrantLocked(ctx, tx, grantID)
	if err != nil || g.UserID != userID || g.AccountID != a.AccountID || g.GroupID != groupID {
		return Invocation{}, ErrDenied
	}
	if _, _, err := lockGrantState(ctx, tx, grantID); err != nil {
		return Invocation{}, err
	}
	var state string
	err = tx.QueryRowContext(ctx, `SELECT state FROM execution_invocations WHERE id=$1 FOR UPDATE`, invocationID).Scan(&state)
	if err != nil || (state != "unknown_effect" && state != "observed") {
		return Invocation{}, ErrConflict
	}
	next := "unknown_effect"
	if observation != "still_unknown" {
		next = "reconciled"
	}
	record := map[string]any{
		"observation": observation, "evidence": evidence,
		"actor_user_id": userID, "recorded_at": time.Now().UTC(),
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE execution_invocations SET state=$2,reconciliation=$3::jsonb,updated_at=NOW()
		WHERE id=$1`, invocationID, next, string(jsonBytes(record)))
	if err != nil {
		return Invocation{}, err
	}
	if err := auditTx(ctx, tx, invocationID, "counting_invocation_reconciled", record); err != nil {
		return Invocation{}, err
	}
	inv, err := loadInvocation(ctx, tx, invocationID)
	if err != nil {
		return Invocation{}, err
	}
	if err := tx.Commit(); err != nil {
		return Invocation{}, err
	}
	return inv, nil
}
