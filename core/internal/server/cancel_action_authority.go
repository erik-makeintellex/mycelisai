package server

import (
	"database/sql"
	"net/http"

	"github.com/google/uuid"
)

// Cancel-action ownership (S7b item 3): only the proposer (the principal that
// minted a confirm token for the proof) or an approver holding
// approvals:decide may cancel. Authority comes from the stored token and the
// authenticated identity, never from the request body.

var cancellerNotProposerCopy = roleBlockerText{
	User: blockerText{"Only the person who asked for this, or an admin, can cancel it.",
		"Ask them to cancel it, or ask an admin." + nothingRan},
	Admin: blockerText{Action: "Cancelling someone else's proposal needs approvals:decide on your admin account." + nothingRan},
}

// cancellerMayCancelOrRespond loads the proof's minters and writes 404, 401,
// or 403 unless the caller may cancel. It runs before any write.
func (s *AdminServer) cancellerMayCancelOrRespond(w http.ResponseWriter, r *http.Request, db *sql.DB, proofID uuid.UUID) bool {
	rows, err := db.QueryContext(r.Context(),
		`SELECT COALESCE(t.minted_by, '') FROM intent_proofs p LEFT JOIN confirm_tokens t ON t.intent_proof_id = p.id WHERE p.id = $1`, proofID)
	if err != nil {
		respondAPIError(w, "failed to load proposal ownership", http.StatusInternalServerError)
		return false
	}
	defer rows.Close()
	found, owner := false, false
	for rows.Next() {
		var mintedBy string
		if err := rows.Scan(&mintedBy); err != nil {
			respondAPIError(w, "failed to load proposal ownership", http.StatusInternalServerError)
			return false
		}
		found = true
		owner = owner || confirmerIsMinter(r, mintedBy)
	}
	if err := rows.Err(); err != nil {
		respondAPIError(w, "failed to load proposal ownership", http.StatusInternalServerError)
		return false
	}
	if !found {
		respondAPIError(w, "intent proof not found", http.StatusNotFound)
		return false
	}
	if owner || isApprover(IdentityFromContext(r.Context())) {
		return true
	}
	status := http.StatusForbidden
	if IdentityFromContext(r.Context()) == nil {
		status = http.StatusUnauthorized
	}
	respondBlockerText(w, r, status, codeCancellerNotProposer, cancellerNotProposerCopy,
		"only the proposer or an approver may cancel this proposal", nil)
	return false
}
