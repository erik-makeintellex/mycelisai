package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

// GET /api/v1/templates - list registered templates.
func (s *AdminServer) handleListTemplatesAPI(w http.ResponseWriter, r *http.Request) {
	view := strings.TrimSpace(r.URL.Query().Get("view"))
	if view == "starter" || view == "organization-starters" {
		templates, err := s.loadOrganizationStarterTemplates()
		if err != nil {
			respondAPIError(w, "failed to load organization starter templates", http.StatusInternalServerError)
			return
		}
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(templates))
		return
	}

	templates := make([]protocol.TemplateSpec, 0, len(protocol.TemplateRegistry))
	for _, t := range protocol.TemplateRegistry {
		templates = append(templates, t)
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(templates))
}

// GET /api/v1/intent/proof/{id} - retrieve a specific intent proof.
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
