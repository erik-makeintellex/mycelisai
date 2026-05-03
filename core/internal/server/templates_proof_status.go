package server

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// confirmChatProofTx updates a proof's status to confirmed for chat-based proposals.
// Unlike confirmIntentProof, this doesn't require a mission ID.
func (s *AdminServer) confirmChatProofTx(tx *sql.Tx, proofID string) error {
	if tx == nil {
		return errDBUnavailable
	}
	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`UPDATE intent_proofs SET status = 'confirmed', confirmed_at = $1 WHERE id = $2`,
		time.Now(), proofUUID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *AdminServer) failChatProofTx(tx *sql.Tx, proofID string) error {
	if tx == nil {
		return errDBUnavailable
	}
	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE intent_proofs SET status = 'failed' WHERE id = $1`, proofUUID)
	return err
}
