package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
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

	now := time.Now()
	return &protocol.IntentProof{
		ID:              id.String(),
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

// generateConfirmToken creates a single-use token bound to an intent proof.
// Token expires after confirmTokenTTL (15 minutes).
func (s *AdminServer) generateConfirmToken(proofID string, templateID protocol.TemplateID) (*protocol.ConfirmToken, error) {
	db := s.getDB()
	if db == nil {
		return nil, nil
	}

	token := uuid.New()
	now := time.Now()
	expiresAt := now.Add(confirmTokenTTL)

	_, err := db.Exec(
		`INSERT INTO confirm_tokens (token, intent_proof_id, template_id, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		token, proofID, string(templateID), expiresAt,
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

// validateConfirmToken checks that a token exists, is not consumed, and not expired.
// Returns the intent proof ID on success. Marks the token as consumed atomically.
func (s *AdminServer) validateConfirmToken(token string) (string, error) {
	db := s.getDB()
	if db == nil {
		return "", errDBUnavailable
	}

	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return "", errInvalidToken
	}

	var proofID string
	var consumed bool
	var expiresAt time.Time

	err = db.QueryRow(
		`SELECT intent_proof_id, consumed, expires_at FROM confirm_tokens WHERE token = $1`,
		tokenUUID,
	).Scan(&proofID, &consumed, &expiresAt)
	if err != nil {
		return "", errTokenNotFound
	}

	if consumed {
		return "", errTokenConsumed
	}
	if time.Now().After(expiresAt) {
		return "", errTokenExpired
	}

	_, err = db.Exec(
		`UPDATE confirm_tokens SET consumed = TRUE, consumed_at = $1 WHERE token = $2 AND consumed = FALSE`,
		time.Now(), tokenUUID,
	)
	if err != nil {
		log.Printf("CE-1: confirm token consume failed: %v", err)
		return "", err
	}

	return proofID, nil
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

func (s *AdminServer) consumeConfirmTokenTx(tx *sql.Tx, token string) (string, error) {
	if tx == nil {
		return "", errDBUnavailable
	}
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return "", errInvalidToken
	}

	var proofID string
	var consumed bool
	var expiresAt time.Time

	err = tx.QueryRow(
		`SELECT intent_proof_id, consumed, expires_at FROM confirm_tokens WHERE token = $1`,
		tokenUUID,
	).Scan(&proofID, &consumed, &expiresAt)
	if err == sql.ErrNoRows {
		return "", errTokenNotFound
	} else if err != nil {
		return "", err
	}
	if consumed {
		return "", errTokenConsumed
	}
	if time.Now().After(expiresAt) {
		return "", errTokenExpired
	}

	_, err = tx.Exec(
		`UPDATE confirm_tokens SET consumed = TRUE, consumed_at = $1 WHERE token = $2 AND consumed = FALSE`,
		time.Now(), tokenUUID,
	)
	if err != nil {
		return "", err
	}

	return proofID, nil
}
