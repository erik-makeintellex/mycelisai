package invocation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/trust"
	"github.com/mycelis/core/pkg/protocol"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Propose(ctx context.Context, userID string, p Proposal) (Proposed, error) {
	if s == nil || s.db == nil {
		return Proposed{}, fmt.Errorf("invocation database unavailable")
	}
	if p.Budget < 1 || p.Budget > maxBudget {
		return Proposed{}, ErrInvalid
	}
	input, err := normalizedInput(Input{Counter: p.Counter})
	if err != nil {
		return Proposed{}, err
	}
	userID, err = parsedUUID(userID)
	if err != nil {
		return Proposed{}, ErrDenied
	}
	groupID, err := parsedUUID(p.GroupID)
	if err != nil {
		return Proposed{}, ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Proposed{}, err
	}
	defer tx.Rollback()
	a, err := lockAuthority(ctx, tx, userID, groupID)
	if err != nil {
		return Proposed{}, err
	}
	b, err := lockBinding(ctx, tx)
	if err != nil {
		return Proposed{}, err
	}
	now, err := dbNow(ctx, tx)
	if err != nil {
		return Proposed{}, err
	}
	confirmDeadline := now.Add(15 * time.Minute).Truncate(time.Microsecond)
	grantDeadline := now.Add(time.Hour).Truncate(time.Microsecond)
	bound := boundary{
		Version: "counting.v1", UserID: a.UserID, AccountID: a.AccountID,
		GroupID: a.GroupID, MembershipID: a.MembershipID,
		Permission: Permission, CapabilityID: CapabilityID,
		Authority: a, AuthorityDigest: digest(a), Binding: b,
		BindingDigest: digest(b), Input: input, InputDigest: digest(input),
		Budget: p.Budget, ExpiresAt: grantDeadline,
	}
	scope := map[string]any{
		"tools": []string{CapabilityID}, "affected_resources": []string{"group:" + groupID},
		"risk_level": "low", boundaryKey: bound,
	}
	proofID, tokenID := uuid.NewString(), uuid.NewString()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO intent_proofs
		(id,template_id,resolved_intent,permission_check,policy_decision,scope_validation,status,expires_at)
		VALUES ($1,$2,$3,'pass','require_approval',$4::jsonb,'pending',$5)`,
		proofID, string(protocol.TemplateChatToProposal), CapabilityID, string(jsonBytes(scope)), confirmDeadline)
	if err != nil {
		return Proposed{}, err
	}
	contractID, err := trust.UpsertContract(ctx, tx, trust.ContractInput{
		IntentProofID: proofID, TemplateID: protocol.TemplateChatToProposal,
		ResolvedIntent: CapabilityID,
	})
	if err != nil {
		return Proposed{}, err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO confirm_tokens(token,intent_proof_id,template_id,consumed,expires_at)
		VALUES($1,$2,$3,FALSE,$4)`, tokenID, proofID, string(protocol.TemplateChatToProposal), confirmDeadline)
	if err != nil {
		return Proposed{}, err
	}
	if err := tx.Commit(); err != nil {
		return Proposed{}, err
	}
	return Proposed{ProofID: proofID, ContractID: contractID, ConfirmToken: tokenID, ExpiresAt: confirmDeadline}, nil
}

// HandlesToken inspects the persisted key, not its value. A malformed counting
// boundary must enter Confirm and fail closed instead of reaching legacy tools.
func (s *Store) HandlesToken(ctx context.Context, token string) (bool, error) {
	if s == nil || s.db == nil {
		return false, fmt.Errorf("invocation database unavailable")
	}
	token, err := parsedUUID(token)
	if err != nil {
		return false, nil
	}
	var present bool
	err = s.db.QueryRowContext(ctx, `
		SELECT p.resolved_intent=$2 OR COALESCE(jsonb_typeof(p.scope_validation)='object' AND p.scope_validation ? 'invocation_boundary',FALSE)
		FROM confirm_tokens t JOIN intent_proofs p ON p.id=t.intent_proof_id WHERE t.token=$1`, token, CapabilityID).Scan(&present)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return present, err
}

type grantContent struct {
	ID, ProofID, ContractID, UserID, AccountID, GroupID, MembershipID string
	CapabilityID, AuthorityDigest, BindingDigest, InputDigest         string
	Authority                                                         authority
	Binding                                                           binding
	Input                                                             Input
	Budget                                                            int
	ExpiresAt                                                         time.Time
}

func (s *Store) Confirm(ctx context.Context, userID, token string) (Grant, error) {
	if s == nil || s.db == nil {
		return Grant{}, fmt.Errorf("invocation database unavailable")
	}
	userID, err := parsedUUID(userID)
	if err != nil {
		return Grant{}, ErrDenied
	}
	token, err = parsedUUID(token)
	if err != nil {
		return Grant{}, ErrInvalid
	}
	// This pre-read determines lock targets. Every byte and status is re-read
	// under row locks below before a grant can be committed.
	var proofID string
	var untrustedScope []byte
	err = s.db.QueryRowContext(ctx, `
		SELECT p.id::text,p.scope_validation::text FROM confirm_tokens t
		JOIN intent_proofs p ON p.id=t.intent_proof_id WHERE t.token=$1`, token).
		Scan(&proofID, &untrustedScope)
	if err != nil {
		return Grant{}, ErrDenied
	}
	preBound, err := parseBoundary(untrustedScope)
	if err != nil || preBound.UserID != userID {
		return Grant{}, ErrDenied
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Grant{}, err
	}
	defer tx.Rollback()
	a, err := lockAuthority(ctx, tx, userID, preBound.GroupID)
	if err != nil {
		return Grant{}, err
	}
	var scope []byte
	var status, templateID, resolvedIntent, decision string
	var proofExpiry sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT scope_validation::text,status,template_id,resolved_intent,policy_decision,expires_at
		FROM intent_proofs WHERE id=$1 FOR UPDATE`, proofID).
		Scan(&scope, &status, &templateID, &resolvedIntent, &decision, &proofExpiry)
	if err != nil || status != "pending" || templateID != string(protocol.TemplateChatToProposal) || resolvedIntent != CapabilityID || decision != "require_approval" || !proofExpiry.Valid {
		return Grant{}, ErrDenied
	}
	bound, err := parseBoundary(scope)
	if err != nil || digest(bound) != digest(preBound) || bound.UserID != userID || bound.AccountID != a.AccountID || bound.GroupID != a.GroupID || bound.MembershipID != a.MembershipID || bound.AuthorityDigest != digest(a) || bound.AuthorityDigest != digest(bound.Authority) || bound.Permission != Permission || bound.CapabilityID != CapabilityID || bound.InputDigest != digest(bound.Input) || bound.BindingDigest != digest(bound.Binding) || bound.Budget < 1 || bound.Budget > maxBudget {
		return Grant{}, ErrDenied
	}
	if _, err := normalizedInput(bound.Input); err != nil {
		return Grant{}, ErrDenied
	}
	var contractID, contractStatus, contractTemplate, contractIntent string
	err = tx.QueryRowContext(ctx, `
		SELECT id::text,status,template_id,resolved_intent FROM execution_contracts WHERE intent_proof_id=$1 FOR UPDATE`, proofID).
		Scan(&contractID, &contractStatus, &contractTemplate, &contractIntent)
	if err != nil || contractStatus != "proposed" || contractTemplate != templateID || contractIntent != resolvedIntent {
		return Grant{}, ErrDenied
	}
	var tokenProofID, tokenTemplateID string
	var consumed bool
	var tokenExpiry time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT intent_proof_id::text,template_id,consumed,expires_at
		FROM confirm_tokens WHERE token=$1 FOR UPDATE`, token).
		Scan(&tokenProofID, &tokenTemplateID, &consumed, &tokenExpiry)
	if err != nil || tokenProofID != proofID || tokenTemplateID != templateID || consumed {
		return Grant{}, ErrDenied
	}
	currentBinding, err := lockBinding(ctx, tx)
	if err != nil || digest(currentBinding) != bound.BindingDigest {
		return Grant{}, ErrDenied
	}
	now, err := dbNow(ctx, tx)
	if err != nil {
		return Grant{}, err
	}
	if !proofExpiry.Time.After(now) || !tokenExpiry.After(now) || !bound.ExpiresAt.After(now) || (!a.MemberExpiry.IsZero() && !a.MemberExpiry.After(now)) {
		return Grant{}, ErrDenied
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE confirm_tokens SET consumed=TRUE,consumed_at=NOW()
		WHERE token=$1 AND consumed=FALSE AND expires_at>clock_timestamp()`, token)
	if err != nil {
		return Grant{}, err
	}
	changed, err := result.RowsAffected()
	if err != nil || changed != 1 {
		return Grant{}, ErrDenied
	}
	grantID := uuid.NewString()
	content := grantContent{grantID, proofID, contractID, a.UserID, a.AccountID, a.GroupID, a.MembershipID,
		CapabilityID, bound.AuthorityDigest, bound.BindingDigest, bound.InputDigest,
		a, currentBinding, bound.Input, bound.Budget, bound.ExpiresAt}
	grantDigest := digest(content)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO execution_effect_grants
		(id,intent_proof_id,execution_contract_id,account_id,user_id,group_id,membership_id,
		 capability_id,authority_snapshot,authority_digest,binding_snapshot,binding_digest,
		 input_snapshot,input_digest,unit_budget,expires_at,digest)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10,$11::jsonb,$12,$13::jsonb,$14,$15,$16,$17)`,
		grantID, proofID, contractID, a.AccountID, a.UserID, a.GroupID, a.MembershipID,
		CapabilityID, string(jsonBytes(a)), bound.AuthorityDigest, string(jsonBytes(currentBinding)),
		bound.BindingDigest, string(jsonBytes(bound.Input)), bound.InputDigest, bound.Budget, bound.ExpiresAt, grantDigest)
	if err != nil {
		return Grant{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO execution_effect_grant_state(grant_id) VALUES($1)`, grantID); err != nil {
		return Grant{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE intent_proofs SET status='confirmed',confirmed_at=NOW() WHERE id=$1`, proofID); err != nil {
		return Grant{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE execution_contracts SET status='confirmed',confirmed_at=NOW(),updated_at=NOW() WHERE id=$1`, contractID); err != nil {
		return Grant{}, err
	}
	if err = auditTx(ctx, tx, grantID, "counting_grant_accepted", map[string]any{"proof_id": proofID, "actor_user_id": userID, "group_id": a.GroupID}); err != nil {
		return Grant{}, err
	}
	if err = tx.Commit(); err != nil {
		return Grant{}, err
	}
	return Grant{ID: grantID, Digest: grantDigest, ProofID: proofID, ContractID: contractID, CapabilityID: CapabilityID,
		UserID: userID, AccountID: a.AccountID, GroupID: a.GroupID, Budget: bound.Budget, ExpiresAt: bound.ExpiresAt}, nil
}

func parseBoundary(scope []byte) (boundary, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(scope, &object); err != nil {
		return boundary{}, ErrDenied
	}
	raw, ok := object[boundaryKey]
	if !ok || len(raw) == 0 || strings.TrimSpace(string(raw)) == "null" {
		return boundary{}, ErrDenied
	}
	var b boundary
	if err := json.Unmarshal(raw, &b); err != nil || b.Version != "counting.v1" {
		return boundary{}, ErrDenied
	}
	return b, nil
}
