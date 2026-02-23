package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

// ── CE-1: Template Engine — Confirm Tokens, Intent Proofs, Audit ────

const confirmTokenTTL = 15 * time.Minute

// createAuditEvent inserts a log_entry as an audit record and returns its UUID.
// Reuses the existing log_entries table with level='audit'.
func (s *AdminServer) createAuditEvent(templateID protocol.TemplateID, source, message string, ctx map[string]any) (string, error) {
	db := s.getDB()
	if db == nil {
		return "", nil // graceful: audit is non-blocking
	}

	id := uuid.New()
	traceID := uuid.New().String()
	contextJSON, _ := json.Marshal(ctx)

	_, err := db.Exec(
		`INSERT INTO log_entries (id, trace_id, timestamp, level, source, intent, message, context)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, traceID, time.Now(), "audit", source, string(templateID), message, contextJSON,
	)
	if err != nil {
		log.Printf("CE-1: audit event insert failed: %v", err)
		return "", err
	}

	return id.String(), nil
}

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

	_, err := db.Exec(
		`INSERT INTO intent_proofs (id, template_id, resolved_intent, permission_check, policy_decision, scope_validation, audit_event_id, status, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, string(templateID), intent, "pass", "allow", scopeJSON, auditUUID, "pending", expiresAt,
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
		PolicyDecision:  "allow",
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

	// Mark consumed atomically
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

// ── Token validation errors ─────────────────────────────────────────

type tokenError string

func (e tokenError) Error() string { return string(e) }

const (
	errDBUnavailable tokenError = "database not available"
	errInvalidToken  tokenError = "invalid token format"
	errTokenNotFound tokenError = "token not found"
	errTokenConsumed tokenError = "token already consumed"
	errTokenExpired  tokenError = "token expired"
)

// ── API Handlers ────────────────────────────────────────────────────

// GET /api/v1/templates — list all registered orchestration templates.
func (s *AdminServer) handleListTemplatesAPI(w http.ResponseWriter, r *http.Request) {
	templates := make([]protocol.TemplateSpec, 0, len(protocol.TemplateRegistry))
	for _, t := range protocol.TemplateRegistry {
		templates = append(templates, t)
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(templates))
}

// GET /api/v1/intent/proof/{id} — retrieve a specific intent proof.
func (s *AdminServer) handleGetIntentProof(w http.ResponseWriter, r *http.Request) {
	proofID := r.PathValue("id")
	db := s.getDB()
	if db == nil {
		respondAPIError(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		respondAPIError(w, "invalid proof ID", http.StatusBadRequest)
		return
	}

	var proof protocol.IntentProof
	var scopeJSON []byte
	var auditEventID *string
	var missionID *string
	var confirmedAt *time.Time

	err = db.QueryRow(
		`SELECT id, template_id, resolved_intent, user_confirmation_token, permission_check,
		        policy_decision, scope_validation, audit_event_id, mission_id, status, created_at, confirmed_at
		 FROM intent_proofs WHERE id = $1`,
		proofUUID,
	).Scan(
		&proof.ID, &proof.TemplateID, &proof.ResolvedIntent, &proof.UserConfirmToken,
		&proof.PermissionCheck, &proof.PolicyDecision, &scopeJSON, &auditEventID,
		&missionID, &proof.Status, &proof.CreatedAt, &confirmedAt,
	)
	if err != nil {
		respondAPIError(w, "intent proof not found", http.StatusNotFound)
		return
	}

	if len(scopeJSON) > 0 {
		var scope protocol.ScopeValidation
		if json.Unmarshal(scopeJSON, &scope) == nil {
			proof.ScopeValidation = &scope
		}
	}
	if auditEventID != nil {
		proof.AuditEventID = *auditEventID
	}
	if missionID != nil {
		proof.MissionID = *missionID
	}
	if confirmedAt != nil {
		proof.ConfirmedAt = confirmedAt
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(proof))
}

// POST /api/v1/intent/confirm-action — confirm a chat-based proposal action.
// Validates the confirm token, marks the intent proof as confirmed, creates audit trail.
// This is the lightweight counterpart to intent/commit (which requires a full blueprint).
func (s *AdminServer) HandleConfirmAction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConfirmToken string `json:"confirm_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if req.ConfirmToken == "" {
		respondAPIError(w, "Missing confirm_token", http.StatusBadRequest)
		return
	}

	// Validate + consume the single-use token
	proofID, err := s.validateConfirmToken(req.ConfirmToken)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Mark the intent proof as confirmed (no mission ID for chat proposals)
	s.confirmChatProof(proofID)

	// Create audit event for the confirmation
	auditID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal, "confirm-action",
		"Chat proposal confirmed by user",
		map[string]any{"proof_id": proofID},
	)

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"confirmed":      true,
		"proof_id":       proofID,
		"audit_event_id": auditID,
		"run_id":         nil, // populated by /api/v1/intent/commit for blueprint activations
	}))
}

// confirmChatProof updates a proof's status to confirmed for chat-based proposals.
// Unlike confirmIntentProof, this doesn't require a mission ID.
func (s *AdminServer) confirmChatProof(proofID string) {
	db := s.getDB()
	if db == nil {
		return
	}

	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		return
	}

	_, err = db.Exec(
		`UPDATE intent_proofs SET status = 'confirmed', confirmed_at = $1 WHERE id = $2`,
		time.Now(), proofUUID,
	)
	if err != nil {
		log.Printf("CE-1: confirm chat proof update failed: %v", err)
	}
}

// ── Scope Validation Helpers ────────────────────────────────────────

// buildScopeFromBlueprint extracts scope validation metadata from a blueprint.
func buildScopeFromBlueprint(bp *protocol.MissionBlueprint) *protocol.ScopeValidation {
	toolSet := make(map[string]bool)
	for _, team := range bp.Teams {
		for _, agent := range team.Agents {
			for _, tool := range agent.Tools {
				toolSet[tool] = true
			}
		}
	}

	tools := make([]string, 0, len(toolSet))
	for t := range toolSet {
		tools = append(tools, t)
	}

	totalAgents := 0
	for _, team := range bp.Teams {
		totalAgents += len(team.Agents)
	}

	risk := "low"
	if totalAgents > 5 || len(bp.Teams) > 2 {
		risk = "medium"
	}
	if totalAgents > 10 || len(bp.Teams) > 4 {
		risk = "high"
	}

	return &protocol.ScopeValidation{
		Tools:             tools,
		AffectedResources: []string{"missions", "teams", "service_manifests"},
		RiskLevel:         risk,
	}
}
