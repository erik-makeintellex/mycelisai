package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
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

// Confirm tokens are bound to the purpose they were minted for. Since A2b the
// purpose, bound-subject digest and minting principal are durable columns on
// confirm_tokens, written at mint; routes decide from the column, and the
// server-authored proof (resolved_intent, scope_validation) stays a second
// guard. A token of the wrong kind, of unknown (legacy NULL) purpose, bound to
// another principal or another blueprint is rejected before it is consumed, so
// it stays valid for its real path.

const (
	chatActionResolvedIntent = "chat-action"
	codeTokenWrongPurpose    = "token_wrong_purpose"
	codeTokenAlreadyUsed     = "token_already_used"
	codeTokenPurposeUnknown  = "token_purpose_unknown"
	codeBlueprintMismatch    = "blueprint_mismatch"
)

var (
	errTokenWrongPurpose = errors.New("confirm token was not issued for this action")
	// errTokenAlreadyUsed is returned when a concurrent confirm won the token.
	errTokenAlreadyUsed = errors.New("confirm token was already used")
	// errTokenPurposeUnknown refuses legacy tokens minted before A2b (Q4-A).
	errTokenPurposeUnknown = errors.New("confirm token predates purpose binding")
	// errBlueprintMismatch refuses a commit body that is not the negotiated blueprint.
	errBlueprintMismatch = errors.New("blueprint differs from the negotiated proposal")
	// errApproverRequired refuses a tier-2 commit by a non-approver (A2b C1).
	errApproverRequired = errors.New("this proposal needs admin approval")
)

// confirmTokenPurpose names the durable purpose a route accepts and the
// proof-content check that must also hold.
type confirmTokenPurpose struct {
	column string
	check  func(resolvedIntent string, scope *protocol.ScopeValidation) bool
}

var (
	blueprintResources = []string{"missions", "teams", "service_manifests"}
	groupResources     = []string{"collaboration_groups", "team_channels"}
)

// blueprintCommitPurpose accepts only tokens minted by intent negotiation for
// a mission blueprint. Chat proposals (including posture-raised ones), group
// approvals, schedule proposals, and invocation tokens are rejected.
var blueprintCommitPurpose = confirmTokenPurpose{column: tokenPurposeMissionBlueprint, check: func(resolvedIntent string, scope *protocol.ScopeValidation) bool {
	intent := strings.TrimSpace(resolvedIntent)
	if intent == "" || intent == chatActionResolvedIntent || intent == invocation.CapabilityID ||
		strings.HasPrefix(intent, "groups.") || strings.HasPrefix(intent, "schedule cadence proposal") {
		return false
	}
	// Tier-2 blueprints are allowed here; the approver gate runs at consume.
	// Posture-raised (chat) approvals never belong to a blueprint.
	postureRaised := scope != nil && scope.Approval != nil && scope.Approval.ApprovalReason == approvalReasonOutcomePosture
	return scope != nil && !postureRaised && len(scope.PlannedToolCalls) == 0 &&
		slices.Equal(scope.AffectedResources, blueprintResources)
}}

// groupMutationPurpose accepts only the group-approval token minted for this
// exact operation and group.
func groupMutationPurpose(op, groupID string) confirmTokenPurpose {
	want := "groups." + op
	if groupID != "" {
		want += "." + groupID
	}
	return confirmTokenPurpose{column: tokenPurposeGroupMutation, check: func(resolvedIntent string, scope *protocol.ScopeValidation) bool {
		return resolvedIntent == want && scope != nil && !requiresApprover(scope) &&
			len(scope.PlannedToolCalls) == 0 && slices.Equal(scope.AffectedResources, groupResources)
	}}
}

// checkTokenPurpose decides from the durable column: NULL (legacy, minted
// before A2b) is refused, and another purpose is the wrong route.
func checkTokenPurpose(row confirmTokenRow, want string) error {
	switch {
	case row.Purpose == "":
		return errTokenPurposeUnknown
	case row.Purpose != want:
		return errTokenWrongPurpose
	}
	return nil
}

// consumeConfirmTokenFor checks existence, expiry, and purpose, then consumes
// the token only if this call wins the single-use update.
func (s *AdminServer) consumeConfirmTokenFor(token string, purpose confirmTokenPurpose) (string, error) {
	proofID, _, err := s.consumeBoundConfirmToken(nil, token, purpose, nil, nil)
	return proofID, err
}

// consumeProposerTokenFor is consumeConfirmTokenFor for proposer-bound tokens
// (A2b Q3): only the minting principal of r, or an approver, may consume it.
// bind, when set, checks the stored binding digest before consumption. floor,
// when set, is a server-recomputed scope whose tier also applies (the higher
// of stored and recomputed wins). Tier 2 needs an approver (errApproverRequired).
func (s *AdminServer) consumeProposerTokenFor(r *http.Request, token string, purpose confirmTokenPurpose, bind func(confirmTokenRow) error, floor *protocol.ScopeValidation) (string, confirmAuthority, error) {
	if r == nil {
		return "", confirmAuthority{}, errConfirmerNotProposer
	}
	return s.consumeBoundConfirmToken(r, token, purpose, bind, floor)
}

func (s *AdminServer) consumeBoundConfirmToken(proposer *http.Request, token string, purpose confirmTokenPurpose, bind func(confirmTokenRow) error, floor *protocol.ScopeValidation) (string, confirmAuthority, error) {
	db := s.getDB()
	if db == nil {
		return "", confirmAuthority{}, errDBUnavailable
	}
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return "", confirmAuthority{}, errInvalidToken
	}
	var resolvedIntent string
	var row confirmTokenRow
	var consumed bool
	var expiresAt time.Time
	var scopeJSON []byte
	err = db.QueryRow(
		`SELECT t.intent_proof_id, t.consumed, t.expires_at, p.resolved_intent, p.scope_validation,
		        COALESCE(t.purpose, ''), COALESCE(t.binding_digest, ''), COALESCE(t.minted_by, '')
		 FROM confirm_tokens t JOIN intent_proofs p ON p.id = t.intent_proof_id WHERE t.token = $1`,
		tokenUUID,
	).Scan(&row.ProofID, &consumed, &expiresAt, &resolvedIntent, &scopeJSON, &row.Purpose, &row.BindingDigest, &row.MintedBy)
	if err != nil {
		return "", confirmAuthority{}, errTokenNotFound
	}
	if consumed {
		return "", confirmAuthority{}, errTokenAlreadyUsed
	}
	if time.Now().After(expiresAt) {
		return "", confirmAuthority{}, errTokenExpired
	}
	if err := checkTokenPurpose(row, purpose.column); err != nil {
		return "", confirmAuthority{}, err
	}
	scope := &protocol.ScopeValidation{}
	if len(scopeJSON) > 0 && json.Unmarshal(scopeJSON, scope) != nil {
		return "", confirmAuthority{}, errTokenWrongPurpose
	}
	if purpose.check == nil || !purpose.check(resolvedIntent, scope) {
		return "", confirmAuthority{}, errTokenWrongPurpose
	}
	var authority confirmAuthority
	if proposer != nil {
		effective := scope
		if floor != nil && requiresApprover(floor) && !requiresApprover(scope) {
			effective = floor
		}
		switch {
		case requiresApprover(effective) && !isApprover(IdentityFromContext(proposer.Context())):
			return "", authority, errApproverRequired
		case !requiresApprover(effective) && !confirmerMayConfirmOwn(proposer, row.MintedBy):
			return "", authority, errConfirmerNotProposer
		}
		authority = resolveConfirmAuthority(proposer, effective, row.MintedBy)
	}
	if bind != nil {
		if err := bind(row); err != nil {
			return "", confirmAuthority{}, err
		}
	}
	result, err := db.Exec(
		`UPDATE confirm_tokens SET consumed = TRUE, consumed_at = $1 WHERE token = $2 AND consumed = FALSE`,
		time.Now(), tokenUUID,
	)
	if err != nil {
		return "", confirmAuthority{}, err
	}
	if n, err := result.RowsAffected(); err != nil || n != 1 {
		return "", confirmAuthority{}, errTokenAlreadyUsed
	}
	return row.ProofID, authority, nil
}

// blueprintBinding refuses a commit whose body is not exactly the negotiated
// blueprint (A2b item 6). A token without a stored digest is refused.
func blueprintBinding(bp *protocol.MissionBlueprint) func(confirmTokenRow) error {
	digest := blueprintDigest(bp)
	return func(row confirmTokenRow) error {
		if row.BindingDigest == "" || digest == "" ||
			subtle.ConstantTimeCompare([]byte(row.BindingDigest), []byte(digest)) != 1 {
			return errBlueprintMismatch
		}
		return nil
	}
}

// confirmTokenErrorCode maps token errors to normalized codes.
func confirmTokenErrorCode(err error) string {
	switch {
	case errors.Is(err, errTokenAlreadyUsed), errors.Is(err, errTokenConsumed):
		return codeTokenAlreadyUsed
	case errors.Is(err, errTokenWrongPurpose):
		return codeTokenWrongPurpose
	case errors.Is(err, errConfirmerNotProposer):
		return codeConfirmerNotProposer
	case errors.Is(err, errTokenPurposeUnknown):
		return codeTokenPurposeUnknown
	case errors.Is(err, errBlueprintMismatch):
		return codeBlueprintMismatch
	}
	return "invalid_confirm_token"
}

// respondConfirmTokenError answers a rejected token: 409 when it was already
// used, 403 when the confirmer is not the proposer, otherwise invalidStatus.
// Codes are normalized; nothing was consumed unless this caller won the token.
func respondConfirmTokenError(w http.ResponseWriter, err error, invalidStatus int) {
	status, action := invalidStatus, ""
	switch confirmTokenErrorCode(err) {
	case codeTokenAlreadyUsed:
		status, action = http.StatusConflict, "Refresh the conversation; this proposal was already confirmed"
	case codeConfirmerNotProposer:
		status, action = http.StatusForbidden, "Ask the person who proposed this to confirm it, or ask an admin. The proposal is still valid."
	case codeTokenPurposeUnknown:
		status, action = http.StatusConflict, "Ask Soma to propose this again"
	case codeBlueprintMismatch:
		status, action = http.StatusConflict, "Negotiate again, or commit the proposed blueprint unchanged"
	}
	respondGovernanceError(w, status, "invalid confirm_token: "+err.Error(), confirmTokenErrorCode(err), action)
}

// blueprintDigest binds a mission_blueprint token to the negotiated blueprint
// (A2b item 6): hex sha256 of its JSON encoding. Commit recomputes it from the
// decoded body the same way, so any edit changes the digest.
func blueprintDigest(bp *protocol.MissionBlueprint) string {
	if bp == nil {
		return ""
	}
	data, err := json.Marshal(bp)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
