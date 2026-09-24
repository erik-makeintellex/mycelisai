package invocation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type grantRow struct {
	Grant
	MembershipID, AuthorityDigest, BindingDigest, InputDigest string
	Authority                                                 authority
	Binding                                                   binding
	Input                                                     Input
	RevokedAt                                                 sql.NullTime
	Reserved                                                  int
}

// dbNow is read after authority/receipt row locks, so a lock wait cannot
// preserve a transaction-start timestamp past a grant or lease deadline.
func dbNow(ctx context.Context, tx *sql.Tx) (time.Time, error) {
	var now time.Time
	err := tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now)
	return now.UTC(), err
}

func grantCoordinates(ctx context.Context, db *sql.DB, grantID string) (string, string, error) {
	var userID, groupID string
	if err := db.QueryRowContext(ctx, `SELECT user_id::text,group_id::text FROM execution_effect_grants WHERE id=$1`, grantID).Scan(&userID, &groupID); err != nil {
		return "", "", ErrDenied
	}
	return userID, groupID, nil
}

func loadGrantLocked(ctx context.Context, tx *sql.Tx, grantID string) (grantRow, error) {
	var g grantRow
	var authorityJSON, bindingJSON, inputJSON []byte
	err := tx.QueryRowContext(ctx, `
		SELECT id::text,digest,intent_proof_id::text,execution_contract_id::text,
		user_id::text,account_id::text,group_id::text,membership_id::text,
		capability_id,authority_snapshot::text,authority_digest,binding_snapshot::text,
		binding_digest,input_snapshot::text,input_digest,unit_budget,expires_at
		FROM execution_effect_grants WHERE id=$1 FOR UPDATE`, grantID).
		Scan(&g.ID, &g.Digest, &g.ProofID, &g.ContractID, &g.UserID,
			&g.AccountID, &g.GroupID, &g.MembershipID, &g.CapabilityID,
			&authorityJSON, &g.AuthorityDigest, &bindingJSON, &g.BindingDigest,
			&inputJSON, &g.InputDigest, &g.Budget, &g.ExpiresAt)
	if err != nil {
		return grantRow{}, ErrDenied
	}
	if json.Unmarshal(authorityJSON, &g.Authority) != nil || json.Unmarshal(bindingJSON, &g.Binding) != nil || json.Unmarshal(inputJSON, &g.Input) != nil {
		return grantRow{}, ErrDenied
	}
	g.ExpiresAt = g.ExpiresAt.UTC()
	content := grantContent{g.ID, g.ProofID, g.ContractID, g.UserID, g.AccountID,
		g.GroupID, g.MembershipID, g.CapabilityID, g.AuthorityDigest,
		g.BindingDigest, g.InputDigest, g.Authority, g.Binding, g.Input,
		g.Budget, g.ExpiresAt}
	if digest(content) != g.Digest {
		return grantRow{}, fmt.Errorf("%w: immutable grant digest mismatch", ErrDenied)
	}
	return g, nil
}

func lockGrantState(ctx context.Context, tx *sql.Tx, grantID string) (int, sql.NullTime, error) {
	var reserved int
	var revoked sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT reserved_units,revoked_at FROM execution_effect_grant_state WHERE grant_id=$1 FOR UPDATE`, grantID).
		Scan(&reserved, &revoked)
	if err != nil {
		return 0, sql.NullTime{}, ErrDenied
	}
	return reserved, revoked, nil
}

func (s *Store) Admit(ctx context.Context, userID, grantID, grantDigest, idempotencyKey string, input Input) (Invocation, bool, error) {
	if s == nil || s.db == nil {
		return Invocation{}, false, fmt.Errorf("invocation database unavailable")
	}
	userID, err := parsedUUID(userID)
	if err != nil {
		return Invocation{}, false, ErrDenied
	}
	grantID, err = parsedUUID(grantID)
	if err != nil {
		return Invocation{}, false, ErrInvalid
	}
	input, err = normalizedInput(input)
	if err != nil {
		return Invocation{}, false, err
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" || len(idempotencyKey) > 128 || strings.ContainsAny(idempotencyKey, "\r\n\t") {
		return Invocation{}, false, ErrInvalid
	}
	ownerID, groupID, err := grantCoordinates(ctx, s.db, grantID)
	if err != nil || ownerID != userID {
		return Invocation{}, false, ErrDenied
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Invocation{}, false, err
	}
	defer tx.Rollback()
	a, err := lockAuthority(ctx, tx, userID, groupID)
	if err != nil {
		return Invocation{}, false, err
	}
	g, err := loadGrantLocked(ctx, tx, grantID)
	if err != nil {
		return Invocation{}, false, err
	}
	if g.UserID != userID || g.AccountID != a.AccountID || g.GroupID != a.GroupID || g.Digest != grantDigest || g.AuthorityDigest != digest(a) || g.AuthorityDigest != digest(g.Authority) || g.InputDigest != digest(g.Input) || g.BindingDigest != digest(g.Binding) || g.CapabilityID != CapabilityID {
		return Invocation{}, false, ErrDenied
	}
	_, revoked, err := lockGrantState(ctx, tx, grantID)
	if err != nil {
		return Invocation{}, false, err
	}
	// Duplicate receipts are historical evidence, even if the grant has since
	// expired or been revoked. They cannot reserve or call the adapter again.
	var existingID string
	err = tx.QueryRowContext(ctx, `SELECT id::text FROM execution_invocations WHERE grant_id=$1 AND idempotency_key=$2`, grantID, idempotencyKey).Scan(&existingID)
	if err == nil {
		if g.InputDigest != digest(input) {
			return Invocation{}, false, ErrConflict
		}
		inv, err := loadInvocation(ctx, tx, existingID)
		if err != nil {
			return Invocation{}, false, err
		}
		if err := tx.Commit(); err != nil {
			return Invocation{}, false, err
		}
		return inv, true, nil
	}
	if err != sql.ErrNoRows {
		return Invocation{}, false, err
	}
	b, err := lockBinding(ctx, tx)
	if err != nil || digest(b) != g.BindingDigest {
		return Invocation{}, false, ErrDenied
	}
	now, err := dbNow(ctx, tx)
	if err != nil {
		return Invocation{}, false, err
	}
	if revoked.Valid || !g.ExpiresAt.After(now) || (!a.MemberExpiry.IsZero() && !a.MemberExpiry.After(now)) || g.InputDigest != digest(input) {
		return Invocation{}, false, ErrDenied
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE execution_effect_grant_state SET reserved_units=reserved_units+1
		WHERE grant_id=$1 AND revoked_at IS NULL
		AND reserved_units < (SELECT unit_budget FROM execution_effect_grants WHERE id=$1 AND expires_at>clock_timestamp())`, grantID)
	if err != nil {
		return Invocation{}, false, err
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return Invocation{}, false, ErrDenied
	}
	invID := uuid.NewString()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO execution_invocations
		(id,grant_id,account_id,user_id,group_id,capability_id,grant_digest,
		 authority_snapshot,authority_digest,binding_snapshot,binding_digest,
		 input_snapshot,input_digest,idempotency_key,state)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9,$10::jsonb,$11,$12::jsonb,$13,$14,'ready')`,
		invID, grantID, a.AccountID, userID, groupID, CapabilityID, g.Digest,
		string(jsonBytes(a)), digest(a), string(jsonBytes(b)), digest(b),
		string(jsonBytes(input)), digest(input), idempotencyKey)
	if err != nil {
		return Invocation{}, false, err
	}
	if err := auditTx(ctx, tx, invID, "counting_invocation_admitted", map[string]any{"grant_id": grantID, "actor_user_id": userID, "group_id": groupID}); err != nil {
		return Invocation{}, false, err
	}
	inv, err := loadInvocation(ctx, tx, invID)
	if err != nil {
		return Invocation{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Invocation{}, false, err
	}
	return inv, false, nil
}

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func loadInvocation(ctx context.Context, db queryRower, id string) (Invocation, error) {
	var inv Invocation
	var input, result, reconciliation []byte
	err := db.QueryRowContext(ctx, `
		SELECT id::text,grant_id::text,grant_digest,user_id::text,account_id::text,
		group_id::text,capability_id,input_snapshot::text,idempotency_key,state,
		generation,attempt,result::text,reconciliation::text,created_at,updated_at
		FROM execution_invocations WHERE id=$1`, id).
		Scan(&inv.ID, &inv.GrantID, &inv.GrantDigest, &inv.UserID,
			&inv.AccountID, &inv.GroupID, &inv.CapabilityID, &input,
			&inv.IdempotencyKey, &inv.State, &inv.Generation, &inv.Attempt,
			&result, &reconciliation, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		return Invocation{}, ErrDenied
	}
	if json.Unmarshal(input, &inv.Input) != nil {
		return Invocation{}, ErrDenied
	}
	inv.Result = result
	inv.Reconciliation = reconciliation
	return inv, nil
}

func (s *Store) Get(ctx context.Context, userID, invocationID string) (Invocation, error) {
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
	inv, err := loadInvocation(ctx, s.db, invocationID)
	if err != nil || inv.UserID != userID {
		return Invocation{}, ErrDenied
	}
	return inv, nil
}

func auditTx(ctx context.Context, tx *sql.Tx, traceID, event string, fields map[string]any) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO log_entries(trace_id,timestamp,level,source,intent,message,context)
		VALUES($1,NOW(),'audit','core.invocation',$2,$2,$3::jsonb)`, traceID, event, string(jsonBytes(fields)))
	return err
}
