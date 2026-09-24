package invocation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (s *Store) Claim(ctx context.Context, invocationID string) (string, int64, error) {
	if s == nil || s.db == nil {
		return "", 0, fmt.Errorf("invocation database unavailable")
	}
	id, err := parsedUUID(invocationID)
	if err != nil {
		return "", 0, ErrInvalid
	}
	owner := uuid.NewString()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", 0, err
	}
	defer tx.Rollback()
	var generation int64
	err = tx.QueryRowContext(ctx, `
		UPDATE execution_invocations
		SET state='claimed',owner_token=$2,generation=generation+1,attempt=attempt+1,
		    lease_until=clock_timestamp()+INTERVAL '15 seconds',updated_at=clock_timestamp()
		WHERE id=$1 AND (state='ready' OR (state='claimed' AND lease_until<clock_timestamp()))
		RETURNING generation`, id, owner).Scan(&generation)
	if err == sql.ErrNoRows {
		return "", 0, ErrConflict
	}
	if err != nil {
		return "", 0, err
	}
	if err := auditTx(ctx, tx, id, "counting_invocation_claimed", map[string]any{"generation": generation}); err != nil {
		return "", 0, err
	}
	if err := tx.Commit(); err != nil {
		return "", 0, err
	}
	return owner, generation, nil
}

// Start is the last authority check before the network boundary. Its commit
// records the point after which any missing response is treated as uncertain.
func (s *Store) Start(ctx context.Context, invocationID, owner string, generation int64) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("invocation database unavailable")
	}
	id, err := parsedUUID(invocationID)
	if err != nil {
		return ErrInvalid
	}
	owner, err = parsedUUID(owner)
	if err != nil || generation < 1 {
		return ErrInvalid
	}
	var grantID, userID, groupID string
	err = s.db.QueryRowContext(ctx, `SELECT grant_id::text,user_id::text,group_id::text FROM execution_invocations WHERE id=$1`, id).
		Scan(&grantID, &userID, &groupID)
	if err != nil {
		return ErrDenied
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	a, authErr := lockAuthority(ctx, tx, userID, groupID)
	if authErr != nil {
		_ = tx.Rollback()
		_ = s.failClaimNoEffect(ctx, id, owner, generation, "current_authority_denied")
		return ErrDenied
	}
	g, grantErr := loadGrantLocked(ctx, tx, grantID)
	if grantErr != nil {
		_ = tx.Rollback()
		_ = s.failClaimNoEffect(ctx, id, owner, generation, "grant_missing")
		return ErrDenied
	}
	_, revoked, stateErr := lockGrantState(ctx, tx, grantID)
	if stateErr != nil {
		_ = tx.Rollback()
		return stateErr
	}
	b, bindingErr := lockBinding(ctx, tx)
	// A missing binding still leaves this an authorized pre-call failure.
	if bindingErr != nil {
		_ = tx.Rollback()
		_ = s.failClaimNoEffect(ctx, id, owner, generation, "binding_missing")
		return ErrDenied
	}
	var state, currentOwner, currentGrantID, currentUserID, currentGroupID, currentAccountID string
	var grantDigest, authorityDigest, bindingDigest, inputDigest string
	var authorityJSON, bindingJSON, inputJSON []byte
	var currentGeneration int64
	var lease sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT state,COALESCE(owner_token::text,''),generation,lease_until,
		grant_id::text,user_id::text,group_id::text,account_id::text,
		grant_digest,authority_snapshot::text,authority_digest,
		binding_snapshot::text,binding_digest,input_snapshot::text,input_digest
		FROM execution_invocations WHERE id=$1 FOR UPDATE`, id).
		Scan(&state, &currentOwner, &currentGeneration, &lease,
			&currentGrantID, &currentUserID, &currentGroupID, &currentAccountID,
			&grantDigest, &authorityJSON, &authorityDigest,
			&bindingJSON, &bindingDigest, &inputJSON, &inputDigest)
	if err != nil {
		_ = tx.Rollback()
		return ErrDenied
	}
	now, err := dbNow(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if state != "claimed" || currentOwner != owner || currentGeneration != generation || !lease.Valid || !lease.Time.After(now) {
		_ = tx.Rollback()
		return ErrConflict
	}
	var invAuthority authority
	var invBinding binding
	var invInput Input
	validSnapshots := json.Unmarshal(authorityJSON, &invAuthority) == nil &&
		json.Unmarshal(bindingJSON, &invBinding) == nil &&
		json.Unmarshal(inputJSON, &invInput) == nil &&
		digest(invAuthority) == authorityDigest && digest(invBinding) == bindingDigest && digest(invInput) == inputDigest
	if revoked.Valid || !g.ExpiresAt.After(now) || (!a.MemberExpiry.IsZero() && !a.MemberExpiry.After(now)) || g.AuthorityDigest != digest(a) || g.AuthorityDigest != authorityDigest || g.BindingDigest != digest(b) || g.BindingDigest != bindingDigest || g.InputDigest != inputDigest || g.Digest != grantDigest || !validSnapshots || currentGrantID != g.ID || currentUserID != g.UserID || currentGroupID != g.GroupID || currentAccountID != g.AccountID {
		_, err = tx.ExecContext(ctx, `UPDATE execution_invocations SET state='failed_before_effect',result=$2::jsonb,updated_at=NOW() WHERE id=$1`, id, `{"reason":"authority_or_binding_changed"}`)
		if err == nil {
			err = auditTx(ctx, tx, id, "counting_invocation_denied_at_start", map[string]any{"actor_user_id": userID})
		}
		if err == nil {
			err = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
		if err != nil {
			return err
		}
		return ErrDenied
	}
	var memberExpiry any
	if !a.MemberExpiry.IsZero() {
		memberExpiry = a.MemberExpiry
	}
	change, err := tx.ExecContext(ctx, `UPDATE execution_invocations SET state='executing',executing_at=clock_timestamp(),updated_at=clock_timestamp() WHERE id=$1 AND lease_until>clock_timestamp() AND EXISTS (SELECT 1 FROM execution_effect_grants WHERE id=$2 AND expires_at>clock_timestamp()) AND ($3::timestamptz IS NULL OR $3::timestamptz>clock_timestamp())`, id, grantID, memberExpiry)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	rows, err := change.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if rows != 1 {
		_ = tx.Rollback()
		return ErrConflict
	}
	if err := auditTx(ctx, tx, id, "counting_invocation_executing", map[string]any{"generation": generation}); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *Store) failClaimNoEffect(ctx context.Context, id, owner string, generation int64, reason string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result := map[string]any{"reason": reason}
	change, err := tx.ExecContext(ctx, `
		UPDATE execution_invocations SET state='failed_before_effect',result=$4::jsonb,updated_at=NOW()
		WHERE id=$1 AND owner_token=$2 AND generation=$3 AND state='claimed'`,
		id, owner, generation, string(jsonBytes(result)))
	if err != nil {
		return err
	}
	rows, err := change.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 1 {
		if err := auditTx(ctx, tx, id, "counting_invocation_failed_before_effect", result); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Execute(ctx context.Context, invocationID string) (Invocation, error) {
	owner, generation, err := s.Claim(ctx, invocationID)
	if errors.Is(err, ErrConflict) {
		return loadInvocation(ctx, s.db, invocationID)
	}
	if err != nil {
		return Invocation{}, err
	}
	if err := s.Start(ctx, invocationID, owner, generation); err != nil {
		inv, loadErr := loadInvocation(ctx, s.db, invocationID)
		if loadErr != nil {
			return Invocation{}, err
		}
		return inv, err
	}
	var bindingJSON, inputJSON []byte
	err = s.db.QueryRowContext(ctx, `
		SELECT g.binding_snapshot::text,g.input_snapshot::text
		FROM execution_invocations i JOIN execution_effect_grants g ON g.id=i.grant_id
		WHERE i.id=$1`, invocationID).
		Scan(&bindingJSON, &inputJSON)
	if err != nil {
		_ = s.finish(ctx, invocationID, owner, generation, "unknown_effect", map[string]any{"reason": "binding_read_failed"})
		return loadInvocation(ctx, s.db, invocationID)
	}
	var b binding
	var input Input
	if json.Unmarshal(bindingJSON, &b) != nil || json.Unmarshal(inputJSON, &input) != nil || validBinding(b) != nil {
		_ = s.finish(ctx, invocationID, owner, generation, "unknown_effect", map[string]any{"reason": "binding_decode_failed"})
		return loadInvocation(ctx, s.db, invocationID)
	}
	result, callErr := invokeCounting(ctx, b, invocationID, input)
	if callErr != nil {
		_ = s.finish(ctx, invocationID, owner, generation, "unknown_effect", map[string]any{"reason": "adapter_response_uncertain"})
		return loadInvocation(ctx, s.db, invocationID)
	}
	state := "observed"
	if result.ObservedCount == 1 && result.InvocationID == invocationID {
		state = "verified"
	}
	if err := s.finish(ctx, invocationID, owner, generation, state, result); err != nil {
		return Invocation{}, err
	}
	return loadInvocation(ctx, s.db, invocationID)
}

func (s *Store) finish(ctx context.Context, invocationID, owner string, generation int64, state string, evidence any) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var currentState, currentOwner string
	var currentGeneration int64
	var lease sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT state,COALESCE(owner_token::text,''),generation,lease_until
		FROM execution_invocations WHERE id=$1 FOR UPDATE`, invocationID).
		Scan(&currentState, &currentOwner, &currentGeneration, &lease)
	if err != nil {
		return ErrDenied
	}
	if currentOwner != owner || currentGeneration != generation {
		return ErrConflict
	}
	now, err := dbNow(ctx, tx)
	if err != nil {
		return err
	}
	if currentState != "executing" || !lease.Valid || !lease.Time.After(now) {
		if currentState == "executing" && (!lease.Valid || !lease.Time.After(now)) {
			if _, err := tx.ExecContext(ctx, `UPDATE execution_invocations SET state='unknown_effect',updated_at=NOW() WHERE id=$1`, invocationID); err != nil {
				return err
			}
		}
		if err := auditTx(ctx, tx, invocationID, "counting_late_observation", map[string]any{"generation": generation, "proposed_state": state, "evidence": evidence}); err != nil {
			return err
		}
		return tx.Commit()
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE execution_invocations SET state=$2,result=$3::jsonb,observed_at=NOW(),updated_at=NOW()
		WHERE id=$1`, invocationID, state, string(jsonBytes(evidence)))
	if err != nil {
		return err
	}
	if err := auditTx(ctx, tx, invocationID, "counting_invocation_"+state, map[string]any{"generation": generation}); err != nil {
		return err
	}
	return tx.Commit()
}

// RecoverExpired never reclaims an executing call. It only makes uncertainty
// durable; expired claimed rows remain eligible for a fresh fenced claim.
func (s *Store) RecoverExpired(ctx context.Context) (int64, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("invocation database unavailable")
	}
	result, err := s.db.ExecContext(ctx, `
		WITH targets AS (
			SELECT id FROM execution_invocations
			WHERE state='executing' AND lease_until<clock_timestamp()
			ORDER BY updated_at LIMIT 100 FOR UPDATE SKIP LOCKED
		), changed AS (
			UPDATE execution_invocations SET state='unknown_effect',updated_at=NOW(),
			result='{"reason":"executing_lease_expired"}'::jsonb
			WHERE id IN (SELECT id FROM targets) RETURNING id
		)
		INSERT INTO log_entries(trace_id,timestamp,level,source,intent,message,context)
		SELECT id::text,NOW(),'audit','core.invocation','counting_invocation_unknown_effect',
		       'counting_invocation_unknown_effect','{"reason":"executing_lease_expired"}'::jsonb
		FROM changed`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
