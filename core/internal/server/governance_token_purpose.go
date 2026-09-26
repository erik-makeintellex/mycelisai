package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/invocation"
	"github.com/mycelis/core/pkg/protocol"
)

// Confirm tokens are bound to the purpose they were minted for (A2a C2). The
// purpose is read from the server-authored intent proof (resolved_intent and
// scope_validation), so no schema change is needed. A token of the wrong kind
// is rejected before it is consumed, so it stays valid for its real path.
// Principal-bound tokens are A2b.

const (
	chatActionResolvedIntent = "chat-action"
	codeTokenWrongPurpose    = "token_wrong_purpose"
	codeTokenAlreadyUsed     = "token_already_used"
)

var (
	errTokenWrongPurpose = errors.New("confirm token was not issued for this action")
	// errTokenAlreadyUsed is returned when a concurrent confirm won the token.
	errTokenAlreadyUsed = errors.New("confirm token was already used")
)

// confirmTokenPurpose accepts or rejects a token from its proof.
type confirmTokenPurpose func(resolvedIntent string, scope *protocol.ScopeValidation) bool

var (
	blueprintResources = []string{"missions", "teams", "service_manifests"}
	groupResources     = []string{"collaboration_groups", "team_channels"}
)

// blueprintCommitPurpose accepts only tokens minted by intent negotiation for
// a mission blueprint. Chat proposals (including posture-raised ones), group
// approvals, schedule proposals, and invocation tokens are rejected.
func blueprintCommitPurpose(resolvedIntent string, scope *protocol.ScopeValidation) bool {
	intent := strings.TrimSpace(resolvedIntent)
	if intent == "" || intent == chatActionResolvedIntent || intent == invocation.CapabilityID ||
		strings.HasPrefix(intent, "groups.") || strings.HasPrefix(intent, "schedule cadence proposal") {
		return false
	}
	return scope != nil && !requiresApprover(scope) && len(scope.PlannedToolCalls) == 0 &&
		slices.Equal(scope.AffectedResources, blueprintResources)
}

// groupMutationPurpose accepts only the group-approval token minted for this
// exact operation and group.
func groupMutationPurpose(op, groupID string) confirmTokenPurpose {
	want := "groups." + op
	if groupID != "" {
		want += "." + groupID
	}
	return func(resolvedIntent string, scope *protocol.ScopeValidation) bool {
		return resolvedIntent == want && scope != nil && !requiresApprover(scope) &&
			len(scope.PlannedToolCalls) == 0 && slices.Equal(scope.AffectedResources, groupResources)
	}
}

// consumeConfirmTokenFor checks existence, expiry, and purpose, then consumes
// the token only if this call wins the single-use update.
func (s *AdminServer) consumeConfirmTokenFor(token string, purpose confirmTokenPurpose) (string, error) {
	db := s.getDB()
	if db == nil {
		return "", errDBUnavailable
	}
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return "", errInvalidToken
	}
	var proofID, resolvedIntent string
	var consumed bool
	var expiresAt time.Time
	var scopeJSON []byte
	err = db.QueryRow(
		`SELECT t.intent_proof_id, t.consumed, t.expires_at, p.resolved_intent, p.scope_validation
		 FROM confirm_tokens t JOIN intent_proofs p ON p.id = t.intent_proof_id WHERE t.token = $1`,
		tokenUUID,
	).Scan(&proofID, &consumed, &expiresAt, &resolvedIntent, &scopeJSON)
	if err != nil {
		return "", errTokenNotFound
	}
	if consumed {
		return "", errTokenAlreadyUsed
	}
	if time.Now().After(expiresAt) {
		return "", errTokenExpired
	}
	scope := &protocol.ScopeValidation{}
	if len(scopeJSON) > 0 && json.Unmarshal(scopeJSON, scope) != nil {
		return "", errTokenWrongPurpose
	}
	if purpose == nil || !purpose(resolvedIntent, scope) {
		return "", errTokenWrongPurpose
	}
	result, err := db.Exec(
		`UPDATE confirm_tokens SET consumed = TRUE, consumed_at = $1 WHERE token = $2 AND consumed = FALSE`,
		time.Now(), tokenUUID,
	)
	if err != nil {
		return "", err
	}
	if n, err := result.RowsAffected(); err != nil || n != 1 {
		return "", errTokenAlreadyUsed
	}
	return proofID, nil
}

// confirmTokenErrorCode maps token errors to normalized codes.
func confirmTokenErrorCode(err error) string {
	switch {
	case errors.Is(err, errTokenAlreadyUsed), errors.Is(err, errTokenConsumed):
		return codeTokenAlreadyUsed
	case errors.Is(err, errTokenWrongPurpose):
		return codeTokenWrongPurpose
	}
	return "invalid_confirm_token"
}

// respondConfirmTokenError answers a rejected token: 409 when it was already
// used, otherwise invalidStatus. Codes are normalized; nothing was consumed
// unless this caller won the token.
func respondConfirmTokenError(w http.ResponseWriter, err error, invalidStatus int) {
	status := invalidStatus
	if confirmTokenErrorCode(err) == codeTokenAlreadyUsed {
		status = http.StatusConflict
	}
	respondGovernanceError(w, status, "invalid confirm_token: "+err.Error(), confirmTokenErrorCode(err), "")
}
