package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// Config apply (RC-004G): POST /api/v1/config/{kind}/{id}/apply. Operator
// templates and collaboration groups are applicable; standing-team manifests
// and organizations stay view-only. Every apply is a dry run first: without
// confirm_token the endpoint validates, diffs, and returns 202 with an
// intent proof and a confirm token bound to the document hash. The retry
// with that token applies exactly the previewed document.

type configApplyRequest struct {
	Document     string `json:"document"`
	Format       string `json:"format"`
	ConfirmToken string `json:"confirm_token,omitempty"`
}

// configApplyPreview is the dry-run result shown before any apply.
type configApplyPreview struct {
	Kind             string     `json:"kind"`
	ID               string     `json:"id"`
	DocumentSHA256   string     `json:"document_sha256"`
	Diff             configDiff `json:"diff"`
	Valid            bool       `json:"valid"`
	ValidationErrors []string   `json:"validation_errors"`
}

// configApplyTarget is a loaded, validated apply candidate for one kind.
type configApplyTarget struct {
	current any
	// diffAfter is what current is compared with; nil means proposed.
	diffAfter    any
	proposed     map[string]any
	validation   []string
	auditDetails map[string]any
	// apply writes the change inside the confirm transaction.
	apply func(ctx context.Context, tx *sql.Tx, auditID, proofID string) error
	// result loads what the caller gets back after commit.
	result func(ctx context.Context) (any, error)
}

func (t *configApplyTarget) diffSubject() any {
	if t.diffAfter != nil {
		return t.diffAfter
	}
	return t.proposed
}

var configApplyViewOnlyKinds = map[string]string{
	configKindTeam:         "standing-team manifests are view-only; they change through bundle files and restart",
	configKindOrganization: "organization config is view-only here; use the organization settings endpoints",
}

func configConflict(msg string) error {
	return &configViewError{status: http.StatusConflict, msg: msg}
}

// authorizeConfigMutation requires the root admin role with config:write for
// every clone and apply, and additionally groups:write for groups (the same
// scope /api/v1/groups mutations require).
func (s *AdminServer) authorizeConfigMutation(w http.ResponseWriter, r *http.Request, kind string) (*RequestIdentity, bool) {
	identity, ok := requireRootAdminScope(w, r, "config:write")
	if !ok {
		return nil, false
	}
	if kind == configKindGroup && !hasScope(identity, "groups:write") {
		respondAPIError(w, "Missing required scope: groups:write", http.StatusForbidden)
		return nil, false
	}
	return identity, true
}

// POST /api/v1/config/{kind}/{id}/apply
func (s *AdminServer) HandleApplyConfig(w http.ResponseWriter, r *http.Request) {
	segment := strings.TrimSpace(r.PathValue("kind"))
	id := strings.TrimSpace(r.PathValue("id"))
	kind, ok := configViewKinds[segment]
	if !ok || id == "" {
		respondAPIError(w, "unknown config kind or id: expected groups or templates", http.StatusBadRequest)
		return
	}
	if reason, viewOnly := configApplyViewOnlyKinds[kind]; viewOnly {
		respondAPIError(w, reason, http.StatusConflict)
		return
	}
	identity, ok := s.authorizeConfigMutation(w, r, kind)
	if !ok {
		return
	}
	var req configApplyRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxConfigDocumentBytes)).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	doc, err := parseConfigDocument(req.Document, req.Format)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if violations := configWriteBackViolations(doc); len(violations) > 0 {
		respondWriteBackViolations(w, violations)
		return
	}

	var target *configApplyTarget
	switch kind {
	case configKindGroup:
		target, err = s.groupApplyTarget(r.Context(), id, doc)
	case configKindTemplate:
		target, err = s.templateApplyTarget(r.Context(), id, doc)
	}
	if err != nil {
		respondConfigError(w, "config-apply", kind, id, err)
		return
	}
	_, sha, err := canonicalConfigJSON(target.proposed)
	if err != nil {
		respondAPIError(w, "document could not be encoded", http.StatusBadRequest)
		return
	}
	preview := configApplyPreview{
		Kind: kind, ID: id, DocumentSHA256: sha,
		Diff:             diffConfigDocuments(target.current, target.diffSubject()),
		Valid:            len(target.validation) == 0,
		ValidationErrors: append([]string{}, target.validation...),
	}
	intent := governedIntent("config.apply", kind, id, sha)
	if strings.TrimSpace(req.ConfirmToken) == "" {
		s.respondApplyPreview(w, preview, intent, target)
		return
	}
	if !preview.Valid {
		respondAPIJSON(w, http.StatusBadRequest, protocol.APIResponse{Error: "document failed validation", Data: map[string]any{"preview": preview}})
		return
	}
	rec := configMutationRecord{
		EventType: protocol.EventConfigApplied, AuditMessage: "Applied " + kind + " config " + id,
		Kind: kind, ID: id, DocumentSHA: sha, Actor: identity.UserID, Details: target.auditDetails,
	}
	result, err := s.commitConfigMutation(r.Context(), strings.TrimSpace(req.ConfirmToken), intent, &rec, target.apply)
	if err != nil {
		respondConfigError(w, "config-apply", kind, id, err)
		return
	}
	var applied any
	if target.result != nil {
		if applied, err = target.result(r.Context()); err != nil {
			log.Printf("[config-apply] reload %s/%s after commit: %v", kind, id, err)
		}
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"applied": true, "kind": kind, "id": id, "document_sha256": sha,
		"diff": preview.Diff, "intent_proof_id": rec.ProofID,
		"audit_event_id": result.AuditEventID, "run_id": result.RunID,
		"mission_event_id": result.MissionEventID, "result": applied,
	}))
}

// respondApplyPreview answers the dry run. A valid document gets an intent
// proof and confirm token; an invalid one gets the preview with its errors
// and no token, so it cannot be confirmed.
func (s *AdminServer) respondApplyPreview(w http.ResponseWriter, preview configApplyPreview, intent string, target *configApplyTarget) {
	if !preview.Valid {
		_, _ = s.createAuditEvent(protocol.TemplateChatToProposal, auditSourceConfigApplyPreview,
			"Config apply preview blocked by validation",
			map[string]any{"kind": preview.Kind, "id": preview.ID, "document_sha256": preview.DocumentSHA256, "validation_errors": preview.ValidationErrors})
		respondAPIJSON(w, http.StatusAccepted, protocol.NewAPISuccess(map[string]any{
			"requires_approval": true, "blocked": true, "preview": preview,
			"intent_proof": nil, "confirm_token": nil,
		}))
		return
	}
	details := map[string]any{"kind": preview.Kind, "id": preview.ID, "document_sha256": preview.DocumentSHA256, "diff": preview.Diff}
	for key, value := range target.auditDetails {
		if _, reserved := details[key]; !reserved {
			details[key] = value
		}
	}
	s.respondGovernedConfirmation(w, governedConfirmRequest{
		Intent:       intent,
		AuditSource:  auditSourceConfigApplyPreview,
		AuditMessage: "Config apply previewed; awaiting confirmation",
		AuditContext: details,
		Scope: &protocol.ScopeValidation{
			AffectedResources: []string{configApplyResource(preview.Kind) + ":" + preview.ID},
			RiskLevel:         "high",
		},
		Extra: map[string]any{"preview": preview, "blocked": false},
	})
}

func configApplyResource(kind string) string {
	if kind == configKindGroup {
		return "collaboration_groups"
	}
	return "operator_templates"
}

// commitConfigMutation runs one confirmed mutation in a transaction:
// consume and bind the token, audit, mutate, and record the mission trail.
// rec.ProofID is set on success.
func (s *AdminServer) commitConfigMutation(ctx context.Context, token, intent string, rec *configMutationRecord,
	mutate func(ctx context.Context, tx *sql.Tx, auditID, proofID string) error) (configMutationResult, error) {
	db := s.getDB()
	if db == nil {
		return configMutationResult{}, configUnavailable("governed mutations require the database-backed approval flow")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return configMutationResult{}, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after commit
	proofID, err := s.consumeGovernedConfirmationTx(tx, token, intent)
	if err != nil {
		return configMutationResult{}, &configViewError{status: governedConfirmStatus(err), msg: "confirm_token rejected: " + err.Error()}
	}
	rec.ProofID = proofID
	auditID, err := auditConfigMutationTx(tx, *rec)
	if err != nil {
		return configMutationResult{}, err
	}
	if err := mutate(ctx, tx, auditID, proofID); err != nil {
		return configMutationResult{}, err
	}
	result, err := s.recordConfigMutationTx(tx, *rec, auditID)
	if err != nil {
		return configMutationResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return configMutationResult{}, err
	}
	s.publishConfigMutationSignal(*rec, result)
	return result, nil
}

func respondWriteBackViolations(w http.ResponseWriter, violations []string) {
	respondAPIJSON(w, http.StatusBadRequest, protocol.APIResponse{
		Error: "document contains redacted or secret values that cannot be written back",
		Data:  map[string]any{"write_back_violations": violations},
	})
}

// respondConfigError writes a configViewError with its status; other errors
// are logged and reported without driver detail.
func respondConfigError(w http.ResponseWriter, component, kind, id string, err error) {
	var viewErr *configViewError
	if errors.As(err, &viewErr) {
		respondAPIError(w, viewErr.msg, viewErr.status)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		respondAPIError(w, kind+" not found: "+id, http.StatusNotFound)
		return
	}
	if isOperatorTemplateTableMissing(err) {
		respondConfigError(w, component, kind, id, operatorTemplateStoreError(err))
		return
	}
	log.Printf("[%s] %s/%s: %v", component, kind, id, err)
	respondAPIError(w, "config mutation failed", http.StatusInternalServerError)
}
