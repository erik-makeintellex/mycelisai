package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/configview"
	"github.com/mycelis/core/pkg/protocol"
)

// Template clone (RC-004G): POST /api/v1/config/templates/{id}/clone.
// Creates an operator template from a file bundle, a file starter, or
// another operator template. The first call validates the clone and returns
// 202 with a preview, an intent proof, and a confirm token bound to the
// clone document hash; the retry with the same body and confirm_token
// stores it in operator_templates.

type cloneTemplateRequest struct {
	NewID        string         `json:"new_id,omitempty"`
	Name         string         `json:"name"`
	Overrides    map[string]any `json:"overrides,omitempty"`
	ConfirmToken string         `json:"confirm_token,omitempty"`
}

var (
	operatorTemplateIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}$`)
	slugSeparators            = regexp.MustCompile(`[^a-z0-9]+`)
)

const maxCloneNameLen = 80

// POST /api/v1/config/templates/{id}/clone
func (s *AdminServer) HandleCloneTemplate(w http.ResponseWriter, r *http.Request) {
	sourceID := strings.TrimSpace(r.PathValue("id"))
	identity, ok := s.authorizeConfigMutation(w, r, configKindTemplate)
	if !ok {
		return
	}
	var req cloneTemplateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxConfigDocumentBytes)).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > maxCloneNameLen {
		respondAPIError(w, fmt.Sprintf("name is required and must be at most %d characters", maxCloneNameLen), http.StatusBadRequest)
		return
	}
	newID := strings.TrimSpace(req.NewID)
	if newID == "" {
		newID = "custom-" + strings.Trim(slugSeparators.ReplaceAllString(strings.ToLower(req.Name), "-"), "-")
	}
	if !operatorTemplateIDPattern.MatchString(newID) {
		respondAPIError(w, "new_id must be 2-63 lowercase letters, digits, or dashes", http.StatusBadRequest)
		return
	}
	if sourceID == "" {
		respondAPIError(w, "source template id is required", http.StatusBadRequest)
		return
	}
	if s.getDB() == nil {
		respondAPIError(w, "governed mutations require the database-backed approval flow", http.StatusServiceUnavailable)
		return
	}

	source, err := s.templateConfigDocument(r, sourceID)
	if err == nil {
		err = s.ensureTemplateIDFree(r.Context(), newID)
	}
	if err != nil {
		respondConfigError(w, "template-clone", configKindTemplate, sourceID, err)
		return
	}
	doc, err := buildCloneDocument(source.doc, newID, req.Name, req.Overrides)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if violations := configWriteBackViolations(doc); len(violations) > 0 {
		respondWriteBackViolations(w, violations)
		return
	}
	canonical, sha, err := canonicalConfigJSON(doc)
	if err != nil {
		respondAPIError(w, "clone document could not be encoded", http.StatusBadRequest)
		return
	}
	if _, verr := bootstrap.ValidateStarterDocument(canonical, s.starterPolicyGroups()); verr != nil {
		respondAPIJSON(w, http.StatusBadRequest, protocol.APIResponse{
			Error: "clone failed validation: " + verr.Error(),
			Data:  map[string]any{"validation_errors": []string{verr.Error()}},
		})
		return
	}

	intent := governedIntent("template.clone", configKindTemplate, newID, sha)
	details := map[string]any{"source_template_id": sourceID, "name": req.Name, "source": source.source}
	if strings.TrimSpace(req.ConfirmToken) == "" {
		s.respondClonePreview(w, newID, sha, intent, doc, details)
		return
	}
	rec := configMutationRecord{
		EventType: protocol.EventTemplateCloned, AuditMessage: "Cloned template " + sourceID + " as " + newID,
		Kind: configKindTemplate, ID: newID, DocumentSHA: sha, Actor: identity.UserID, Details: details,
	}
	write := operatorTemplateWrite{
		ID: newID, Name: req.Name, SourceTemplateID: sourceID, Document: canonical,
		DocumentSHA256: sha, CreatedBy: identity.UserID,
	}
	result, err := s.commitConfigMutation(r.Context(), strings.TrimSpace(req.ConfirmToken), intent, &rec,
		func(ctx context.Context, tx *sql.Tx, auditID, proofID string) error {
			write.AuditEventID, write.IntentProofID = auditID, proofID
			err := insertOperatorTemplateTx(ctx, tx, write)
			if errors.Is(err, errOperatorTemplateExists) {
				return configConflict("template id already exists: " + newID)
			}
			return err
		})
	if err != nil {
		respondConfigError(w, "template-clone", configKindTemplate, newID, err)
		return
	}
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(map[string]any{
		"cloned": true, "template": map[string]any{
			"id": newID, "name": req.Name, "source": starterSourceOperator,
			"source_template_id": sourceID, "document_sha256": sha,
		},
		"intent_proof_id": rec.ProofID, "audit_event_id": result.AuditEventID,
		"run_id": result.RunID, "mission_event_id": result.MissionEventID,
	}))
}

func (s *AdminServer) respondClonePreview(w http.ResponseWriter, newID, sha, intent string, doc map[string]any, details map[string]any) {
	rendered, err := configview.Render(doc, configview.FormatYAML)
	if err != nil {
		respondAPIError(w, "clone document could not be rendered", http.StatusInternalServerError)
		return
	}
	audit := map[string]any{"id": newID, "document_sha256": sha}
	for key, value := range details {
		audit[key] = value
	}
	s.respondGovernedConfirmation(w, governedConfirmRequest{
		Intent:       intent,
		AuditSource:  auditSourceTemplateClonePreview,
		AuditMessage: "Template clone previewed; awaiting confirmation",
		AuditContext: audit,
		Scope: &protocol.ScopeValidation{
			AffectedResources: []string{"operator_templates:" + newID},
			RiskLevel:         "medium",
		},
		Extra: map[string]any{"preview": map[string]any{
			"kind": configKindTemplate, "id": newID, "source_template_id": details["source_template_id"],
			"document_sha256": sha, "format": rendered.Format, "document": rendered.Document,
			"redactions": rendered.Redactions, "valid": true, "validation_errors": []string{},
		}},
	})
}

// ensureTemplateIDFree reports 409 when id is a file template or an
// operator template already.
func (s *AdminServer) ensureTemplateIDFree(ctx context.Context, id string) error {
	fileDoc, err := s.fileTemplateDocument(id)
	if err != nil {
		return err
	}
	if fileDoc != nil {
		return configConflict("template id already exists: " + id)
	}
	row, err := s.getOperatorTemplate(ctx, id)
	if err != nil {
		return operatorTemplateStoreError(err)
	}
	if row != nil {
		return configConflict("template id already exists: " + id)
	}
	return nil
}

// buildCloneDocument copies the source bundle, deep-merges overrides, and
// sets id, name, and the starter label to the clone's identity.
func buildCloneDocument(source any, newID, name string, overrides map[string]any) (map[string]any, error) {
	doc, err := normalizeConfigObject(source)
	if err != nil {
		return nil, err
	}
	if len(overrides) > 0 {
		clean, err := normalizeConfigObject(overrides)
		if err != nil {
			return nil, fmt.Errorf("overrides: %v", err)
		}
		if v, ok := clean["id"]; ok && fmt.Sprint(v) != newID {
			return nil, fmt.Errorf("overrides cannot change id; use new_id")
		}
		mergeConfigMaps(doc, clean)
	}
	doc["id"] = newID
	doc["name"] = name
	starter, _ := doc["starter"].(map[string]any)
	if starter == nil {
		starter = map[string]any{}
	}
	starter["label"] = name
	doc["starter"] = starter
	return doc, nil
}

// mergeConfigMaps deep-merges src into dst: maps merge, everything else
// (including lists) replaces.
func mergeConfigMaps(dst, src map[string]any) {
	for key, value := range src {
		if srcMap, ok := value.(map[string]any); ok {
			if dstMap, ok := dst[key].(map[string]any); ok {
				mergeConfigMaps(dstMap, srcMap)
				continue
			}
		}
		dst[key] = value
	}
}
