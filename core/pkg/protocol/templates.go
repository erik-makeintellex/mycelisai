package protocol

import (
	"encoding/json"
	"time"
)

// ── CE-1: Orchestration Templates as First-Class Primitive ──────────

// TemplateID identifies an orchestration template.
type TemplateID string

const (
	TemplateChatToAnswer   TemplateID = "chat-to-answer"
	TemplateChatToProposal TemplateID = "chat-to-proposal"
)

// ExecutionMode is the user-visible mode shown in the UI ribbon.
type ExecutionMode string

const (
	ModeAnswer   ExecutionMode = "answer"
	ModeProposal ExecutionMode = "proposal"
)

// TemplateSpec is the machine-readable definition of an orchestration template.
type TemplateSpec struct {
	ID              TemplateID    `json:"id"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Mode            ExecutionMode `json:"mode"`
	RequiresConfirm bool          `json:"requires_confirm"`
	MutatesState    bool          `json:"mutates_state"`
	GovernanceLevel string        `json:"governance_level"` // "minimal" or "full"
}

// TemplateRegistry holds all registered templates (in-memory for V1).
var TemplateRegistry = map[TemplateID]TemplateSpec{
	TemplateChatToAnswer: {
		ID:              TemplateChatToAnswer,
		Name:            "Chat-to-Answer",
		Description:     "Read-only truth retrieval with provenance",
		Mode:            ModeAnswer,
		RequiresConfirm: false,
		MutatesState:    false,
		GovernanceLevel: "minimal",
	},
	TemplateChatToProposal: {
		ID:              TemplateChatToProposal,
		Name:            "Chat-to-Proposal",
		Description:     "State mutation with explicit intent-proof discipline",
		Mode:            ModeProposal,
		RequiresConfirm: true,
		MutatesState:    true,
		GovernanceLevel: "full",
	},
}

// ── Confirm Tokens ──────────────────────────────────────────────────

// ConfirmToken is a single-use token that gates state mutation.
// Generated when a proposal is ready, consumed on commit. 15-minute TTL.
type ConfirmToken struct {
	Token         string     `json:"token"`
	IntentProofID string     `json:"intent_proof_id"`
	TemplateID    TemplateID `json:"template_id"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     time.Time  `json:"expires_at"`
}

// ── Intent Proofs ───────────────────────────────────────────────────

// IntentProof is the complete proof bundle for a Chat-to-Proposal execution.
// Generated BEFORE execution. No proof = no commit.
type IntentProof struct {
	ID               string           `json:"id"`
	TemplateID       TemplateID       `json:"template_id"`
	ResolvedIntent   string           `json:"resolved_intent"`
	UserConfirmToken string           `json:"user_confirmation_token,omitempty"`
	PermissionCheck  string           `json:"permission_check"` // "pass" or "fail"
	PolicyDecision   string           `json:"policy_decision"`  // "allow", "deny", "require_approval"
	ScopeValidation  *ScopeValidation `json:"scope_validation,omitempty"`
	AuditEventID     string           `json:"audit_event_id"`
	MissionID        string           `json:"mission_id,omitempty"`
	Status           string           `json:"status"`     // "pending", "confirmed", "denied"
	CreatedAt        time.Time        `json:"created_at"`
	ConfirmedAt      *time.Time       `json:"confirmed_at,omitempty"`
}

// ScopeValidation tracks what resources a proposal will affect.
type ScopeValidation struct {
	Tools             []string `json:"tools"`
	AffectedResources []string `json:"affected_resources"` // e.g. ["missions", "teams", "service_manifests"]
	RiskLevel         string   `json:"risk_level"`         // "low", "medium", "high"
}

// ── Answer Provenance ───────────────────────────────────────────────

// AnswerProvenance is the minimal proof for Chat-to-Answer responses.
// No confirm token, no scope validation — just audit linkage.
type AnswerProvenance struct {
	ResolvedIntent  string   `json:"resolved_intent"`   // "answer"
	PermissionCheck string   `json:"permission_check"`  // "pass"
	PolicyDecision  string   `json:"policy_decision"`   // "allow"
	AuditEventID    string   `json:"audit_event_id"`
	ConsultChain    []string `json:"consult_chain,omitempty"`
}

// ── Negotiate Response ──────────────────────────────────────────────

// NegotiateResponse is the enriched response from POST /api/v1/intent/negotiate.
// CE-1 wraps the blueprint with proof and confirm token.
type NegotiateResponse struct {
	Blueprint    *MissionBlueprint `json:"blueprint"`
	IntentProof  *IntentProof      `json:"intent_proof,omitempty"`
	ConfirmToken *ConfirmToken     `json:"confirm_token,omitempty"`
	Template     *TemplateSpec     `json:"template,omitempty"`
}

// ── Commit Request Extension ────────────────────────────────────────

// CommitRequest extends MissionBlueprint with an optional confirm token.
type CommitRequest struct {
	MissionBlueprint
	ConfirmToken string `json:"confirm_token"`
}

// MarshalJSON implements custom marshaling to flatten blueprint fields.
func (cr CommitRequest) MarshalJSON() ([]byte, error) {
	type Alias CommitRequest
	return json.Marshal(struct {
		Alias
	}{Alias: Alias(cr)})
}
