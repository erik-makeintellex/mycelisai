package server

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/mycelis/core/pkg/protocol"
)

// Proposer binding (A2b, owner Q3-A): a chat_action or mission_blueprint token
// that does not need an approver is confirmable only by the principal that
// minted it (confirm_tokens.minted_by) or by an approver. Authority comes from
// the stored token and the authenticated identity, never from the request body.

const codeConfirmerNotProposer = "confirmer_not_proposer"

var (
	errConfirmerNotProposer = errors.New("only the proposer or an approver may confirm this proposal")
	// errTokenMintUnbound refuses to mint a token without purpose or principal.
	errTokenMintUnbound = errors.New("confirm token mint requires a purpose and a minting principal")
)

// requestPrincipal is the authenticated principal of a request, or "" when the
// request carries no identity.
func requestPrincipal(r *http.Request) string {
	if r == nil || IdentityFromContext(r.Context()) == nil {
		return ""
	}
	return strings.TrimSpace(auditActorIDFromRequest(r))
}

// confirmerIsMinter reports whether the request principal minted the token.
func confirmerIsMinter(r *http.Request, mintedBy string) bool {
	principal := requestPrincipal(r)
	mintedBy = strings.TrimSpace(mintedBy)
	return principal != "" && mintedBy != "" &&
		subtle.ConstantTimeCompare([]byte(principal), []byte(mintedBy)) == 1
}

// confirmerMayConfirmOwn allows the minting principal or any approver.
func confirmerMayConfirmOwn(r *http.Request, mintedBy string) bool {
	return confirmerIsMinter(r, mintedBy) || (r != nil && isApprover(IdentityFromContext(r.Context())))
}

// respondConfirmerNotProposer writes the normalized proposer-binding blocker.
// Nothing ran and the token was not consumed.
func respondConfirmerNotProposer(w http.ResponseWriter, r *http.Request) {
	status := http.StatusForbidden
	if r == nil || IdentityFromContext(r.Context()) == nil {
		status = http.StatusUnauthorized
	}
	respondBlocker(w, r, status, codeConfirmerNotProposer, errConfirmerNotProposer.Error(), nil)
}

// Approval tiers (A2b item 1, owner Q1-A/Q2-A). The tier is evaluated at
// confirm time on the stored, server-authored scope, so it also covers proofs
// minted before A2b and no client field can lower it.
const (
	approverTierAuto     = 0 // approval not required: proposer
	approverTierSelf     = 1 // self-review: proposer (the minting principal)
	approverTierApprover = 2 // root admin + approvals:decide

	// approverCostCeiling is the largest profile MaxCost; above it the cost
	// gate is a hard approver floor instead of a user-tunable limit.
	approverCostCeiling = 5.0
)

// approverTier classifies a stored scope. Tier-2 conditions are checked first
// so the classifier fails toward the approver.
func approverTier(scope *protocol.ScopeValidation) (int, string) {
	if scope == nil || scope.Approval == nil {
		return approverTierAuto, "auto"
	}
	a := scope.Approval
	switch {
	case a.ApprovalReason == approvalReasonOutcomePosture || slices.Contains(a.ApprovalSteps, approvalStepRoleGate):
		return approverTierApprover, "policy"
	case tierRiskRank(a.CapabilityRisk) >= 3:
		return approverTierApprover, "capability_risk"
	case max(a.EstimatedCost, scope.EstimatedCost) > approverCostCeiling:
		return approverTierApprover, "cost"
	case !a.ApprovalRequired:
		return approverTierAuto, "auto"
	}
	return approverTierSelf, "self_review"
}

// tierRiskRank is approvalRank plus "critical", which ranks above high.
func tierRiskRank(risk string) int {
	if r := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(risk)), "-risk"); r == "critical" {
		return 4
	}
	return approvalRank(risk)
}

// applyApproverTier mirrors the confirm-time tier at mint so the proposal says
// "needs admin approval" before the click. It is monotone: it only ever adds
// the role gate and never lowers an approval.
func applyApproverTier(approval *protocol.ApprovalPolicy) *protocol.ApprovalPolicy {
	if approval == nil {
		return nil
	}
	if tier, _ := approverTier(&protocol.ScopeValidation{Approval: approval}); tier != approverTierApprover {
		return approval
	}
	raised := *approval
	raised.ApprovalSteps = withRoleGateStep(raised.ApprovalSteps)
	raised.ApprovalRequired = true
	raised.ApprovalMode = "required"
	raised.RequiredApproverRole = "admin"
	return &raised
}

// confirmAuthority is what the confirm-action audit records about who
// confirmed: tier, authority (proposer | approvals:decide) and self-approval.
type confirmAuthority struct {
	Tier         int
	Authority    string
	SelfApproved bool
}

// confirmAuthorities hands the authority resolved at the gate (keyed by intent
// proof) to the confirm-action audit written after commit. The audit removes
// the entry; a missing entry (async redelivery, restart) falls back to the
// tier-derived authority.
var confirmAuthorities sync.Map

func recordConfirmAuthority(r *http.Request, scope *protocol.ScopeValidation, tok confirmTokenRow) {
	if tok.ProofID != "" {
		confirmAuthorities.Store(tok.ProofID, resolveConfirmAuthority(r, scope, tok.MintedBy))
	}
}

func resolveConfirmAuthority(r *http.Request, scope *protocol.ScopeValidation, mintedBy string) confirmAuthority {
	tier, _ := approverTier(scope)
	self := confirmerIsMinter(r, mintedBy)
	authority := scopeApprovalsDecide
	if tier != approverTierApprover && self {
		authority = "proposer"
	}
	return confirmAuthority{Tier: tier, Authority: authority, SelfApproved: self}
}

// takeConfirmAuthority returns and clears the recorded authority for a proof.
func takeConfirmAuthority(proofID string, scope *protocol.ScopeValidation) confirmAuthority {
	if v, ok := confirmAuthorities.LoadAndDelete(proofID); ok {
		return v.(confirmAuthority)
	}
	tier, _ := approverTier(scope)
	if tier == approverTierApprover {
		return confirmAuthority{Tier: tier, Authority: scopeApprovalsDecide}
	}
	return confirmAuthority{Tier: tier, Authority: "proposer"}
}
