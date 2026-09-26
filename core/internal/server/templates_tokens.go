package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/trust"
	"github.com/mycelis/core/pkg/protocol"
)

// createIntentProof builds and persists a full intent proof bundle (status=pending).
func (s *AdminServer) createIntentProof(templateID protocol.TemplateID, intent string, scope *protocol.ScopeValidation, auditEventID string) (*protocol.IntentProof, error) {
	db := s.getDB()
	if db == nil {
		return nil, nil // graceful: proof is non-blocking if DB unavailable
	}

	id := uuid.New()
	expiresAt := time.Now().Add(confirmTokenTTL)

	var scopeJSON []byte
	if scope != nil {
		scopeJSON, _ = json.Marshal(scope)
	}

	var auditUUID *uuid.UUID
	if auditEventID != "" {
		parsed, err := uuid.Parse(auditEventID)
		if err == nil {
			auditUUID = &parsed
		}
	}

	policyDecision := "allow"
	if scope != nil && scope.Approval != nil && scope.Approval.ApprovalRequired {
		policyDecision = "require_approval"
	}

	_, err := db.Exec(
		`INSERT INTO intent_proofs (id, template_id, resolved_intent, permission_check, policy_decision, scope_validation, audit_event_id, status, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, string(templateID), intent, "pass", policyDecision, scopeJSON, auditUUID, "pending", expiresAt,
	)
	if err != nil {
		log.Printf("CE-1: intent proof insert failed: %v", err)
		return nil, err
	}

	contractID, err := trust.NewStore(db).UpsertContract(context.Background(), trust.ContractInput{
		IntentProofID:  id.String(),
		TemplateID:     templateID,
		ResolvedIntent: intent,
		AuditEventID:   auditEventID,
	})
	if err != nil {
		log.Printf("CE-1: execution contract upsert failed: %v", err)
		return nil, err
	}

	now := time.Now()
	return &protocol.IntentProof{
		ID:              id.String(),
		ContractID:      contractID,
		TemplateID:      templateID,
		ResolvedIntent:  intent,
		PermissionCheck: "pass",
		PolicyDecision:  policyDecision,
		ScopeValidation: scope,
		AuditEventID:    auditEventID,
		Status:          "pending",
		CreatedAt:       now,
	}, nil
}

// Durable confirm-token purposes (A2b, confirm_tokens.purpose).
const (
	tokenPurposeChatAction       = "chat_action"
	tokenPurposeMissionBlueprint = "mission_blueprint"
	tokenPurposeGroupMutation    = "group_mutation"
)

// confirmTokenMint is recorded on every token at mint: its purpose, the digest
// of the subject it is bound to (blueprints), and the minting principal.
type confirmTokenMint struct {
	Purpose       string
	BindingDigest string
	MintedBy      string
}

// generateConfirmToken creates a single-use token bound to an intent proof,
// its purpose and its minting principal. Token expires after confirmTokenTTL.
func (s *AdminServer) generateConfirmToken(proofID string, templateID protocol.TemplateID, mint confirmTokenMint) (*protocol.ConfirmToken, error) {
	db := s.getDB()
	if db == nil {
		return nil, nil
	}
	if strings.TrimSpace(mint.Purpose) == "" || strings.TrimSpace(mint.MintedBy) == "" {
		return nil, errTokenMintUnbound
	}

	token := uuid.New()
	now := time.Now()
	expiresAt := now.Add(confirmTokenTTL)

	_, err := db.Exec(
		`INSERT INTO confirm_tokens (token, intent_proof_id, template_id, expires_at, purpose, binding_digest, minted_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		token, proofID, string(templateID), expiresAt, mint.Purpose, sql.NullString{String: mint.BindingDigest, Valid: mint.BindingDigest != ""}, mint.MintedBy,
	)
	if err != nil {
		log.Printf("CE-1: confirm token insert failed: %v", err)
		return nil, err
	}

	return &protocol.ConfirmToken{
		Token:         token.String(),
		IntentProofID: proofID,
		TemplateID:    templateID,
		CreatedAt:     now,
		ExpiresAt:     expiresAt,
	}, nil
}

// confirmIntentProof updates a proof's status to confirmed after successful commit.
func (s *AdminServer) confirmIntentProof(proofID, missionID string) {
	db := s.getDB()
	if db == nil {
		return
	}

	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		return
	}
	missionUUID, err := uuid.Parse(missionID)
	if err != nil {
		return
	}

	_, err = db.Exec(
		`UPDATE intent_proofs SET status = 'confirmed', mission_id = $1, confirmed_at = $2 WHERE id = $3`,
		missionUUID, time.Now(), proofUUID,
	)
	if err != nil {
		log.Printf("CE-1: confirm intent proof update failed: %v", err)
	}
}

// confirmTokenRow is a token as recorded at mint (A2b). Empty strings stand for
// NULL columns (legacy tokens minted before A2b).
type confirmTokenRow struct {
	ProofID       string
	Purpose       string
	BindingDigest string
	MintedBy      string
}

// consumeConfirmTokenTx consumes a chat proposal token inside the caller's
// transaction; a rollback leaves the token unconsumed.
func (s *AdminServer) consumeConfirmTokenTx(tx *sql.Tx, token string) (confirmTokenRow, error) {
	if tx == nil {
		return confirmTokenRow{}, errDBUnavailable
	}
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return confirmTokenRow{}, errInvalidToken
	}

	var row confirmTokenRow
	var consumed bool
	var expiresAt time.Time

	err = tx.QueryRow(
		`SELECT intent_proof_id, consumed, expires_at, COALESCE(purpose, ''), COALESCE(binding_digest, ''), COALESCE(minted_by, '')
		 FROM confirm_tokens WHERE token = $1`,
		tokenUUID,
	).Scan(&row.ProofID, &consumed, &expiresAt, &row.Purpose, &row.BindingDigest, &row.MintedBy)
	if err == sql.ErrNoRows {
		return confirmTokenRow{}, errTokenNotFound
	} else if err != nil {
		return confirmTokenRow{}, err
	}
	if consumed {
		return confirmTokenRow{}, errTokenAlreadyUsed
	}
	if time.Now().After(expiresAt) {
		return confirmTokenRow{}, errTokenExpired
	}
	// A2b item 7: confirm-action accepts only chat proposal tokens. Blueprint,
	// group and legacy (NULL) tokens are refused before the UPDATE, so they
	// stay valid for their own route.
	if err := checkTokenPurpose(row, tokenPurposeChatAction); err != nil {
		return confirmTokenRow{}, err
	}

	result, err := tx.Exec(
		`UPDATE confirm_tokens SET consumed = TRUE, consumed_at = $1 WHERE token = $2 AND consumed = FALSE`,
		time.Now(), tokenUUID,
	)
	if err != nil {
		return confirmTokenRow{}, err
	}
	// Two concurrent confirms can both read consumed=false; only the one whose
	// update wins may execute (A2a C3).
	if n, err := result.RowsAffected(); err != nil || n != 1 {
		return confirmTokenRow{}, errTokenAlreadyUsed
	}

	return row, nil
}
