package server

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

// Governed confirmation is the two-call approval pattern shared by groups
// and config clone/apply (RC-004G):
//
//  1. The first call records an audit entry, persists an intent proof whose
//     resolved_intent names exactly what will change, mints a single-use,
//     expiring confirm token bound to that proof, and answers 202.
//  2. The retry carries confirm_token. The mutation transaction consumes the
//     token and checks that its proof was minted for the same intent before
//     writing anything, so the token is spent only if the mutation commits.

// errConfirmIntentMismatch means a token was minted for a different intent,
// for example a document that changed after its preview.
var errConfirmIntentMismatch = errors.New("confirm_token was issued for a different request; request a new preview")

// governedConfirmRequest describes one governed mutation awaiting approval.
type governedConfirmRequest struct {
	// Intent is the resolved intent the token is bound to. Include every
	// value that must not change between preview and confirm (for example a
	// document hash).
	Intent       string
	AuditSource  string
	AuditMessage string
	AuditContext map[string]any
	Scope        *protocol.ScopeValidation
	// Extra fields are merged into the 202 payload (preview, ids).
	Extra map[string]any
}

// governedIntent builds a resolved intent bound to a document hash.
func governedIntent(action, kind, id, documentSHA string) string {
	return fmt.Sprintf("%s.%s.%s@sha256:%s", action, kind, id, documentSHA)
}

// respondGovernedConfirmation mints the intent proof and confirm token for
// req and writes the 202 approval response, or a 503 when the approval
// records cannot be written. It always writes a response.
func (s *AdminServer) respondGovernedConfirmation(w http.ResponseWriter, req governedConfirmRequest) {
	if s.getDB() == nil {
		respondAPIError(w, "governed mutations require the database-backed approval flow", http.StatusServiceUnavailable)
		return
	}
	auditEventID, _ := s.createAuditEvent(protocol.TemplateChatToProposal, req.AuditSource, req.AuditMessage, req.AuditContext)
	proof, err := s.createIntentProof(protocol.TemplateChatToProposal, req.Intent, req.Scope, auditEventID)
	if err != nil || proof == nil {
		respondAPIError(w, "failed to create approval proof", http.StatusServiceUnavailable)
		return
	}
	token, err := s.generateConfirmToken(proof.ID, protocol.TemplateChatToProposal)
	if err != nil || token == nil {
		respondAPIError(w, "failed to create confirm token", http.StatusServiceUnavailable)
		return
	}
	payload := map[string]any{
		"requires_approval": true,
		"intent_proof":      proof,
		"confirm_token":     token,
		"audit_event_id":    auditEventID,
	}
	for key, value := range req.Extra {
		if _, reserved := payload[key]; !reserved {
			payload[key] = value
		}
	}
	respondAPIJSON(w, http.StatusAccepted, protocol.NewAPISuccess(payload))
}

// consumeGovernedConfirmationTx consumes token inside tx and verifies that
// its intent proof was minted for expectedIntent. It returns the proof id.
// On errConfirmIntentMismatch the caller must roll back, which leaves the
// token unspent for the request it was issued for.
func (s *AdminServer) consumeGovernedConfirmationTx(tx *sql.Tx, token, expectedIntent string) (string, error) {
	proofID, err := s.consumeConfirmTokenTx(tx, token)
	if err != nil {
		return "", err
	}
	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		return "", errInvalidToken
	}
	var intent string
	if err := tx.QueryRow(`SELECT resolved_intent FROM intent_proofs WHERE id = $1`, proofUUID).Scan(&intent); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errTokenNotFound
		}
		return "", err
	}
	if intent != expectedIntent {
		return "", errConfirmIntentMismatch
	}
	return proofID, nil
}

// governedConfirmStatus maps a consume error to its HTTP status: token
// problems are 400, a token for another request is 409, anything else 500.
func governedConfirmStatus(err error) int {
	var tokenErr tokenError
	switch {
	case errors.Is(err, errConfirmIntentMismatch):
		return http.StatusConflict
	case errors.As(err, &tokenErr):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
