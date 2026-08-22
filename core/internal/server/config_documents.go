package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

type configDocumentInput struct {
	Document       *protocol.ConfigDocument        `json:"document,omitempty"`
	Content        string                          `json:"content,omitempty"`
	Format         string                          `json:"format,omitempty"`
	Path           string                          `json:"path,omitempty"`
	OperatorValues protocol.MinimumSufficientBrief `json:"operator_values,omitempty"`
	PolicyValues   protocol.MinimumSufficientBrief `json:"policy_values,omitempty"`
}

func (s *AdminServer) HandleConfigDocumentDryRun(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "config_documents:read"); !ok {
		return
	}
	input, document, ok := decodeConfigDocumentInput(w, r)
	if !ok {
		return
	}
	dryRun := protocol.DryRunConfigDocument(document)
	response := map[string]any{"dry_run": dryRun}
	if dryRun.Valid {
		compiled, err := configdocuments.CompileDocument(document, input.OperatorValues, input.PolicyValues)
		if err != nil {
			respondAPIError(w, err.Error(), http.StatusBadRequest)
			return
		}
		response["compiled"] = compiled
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(response))
}

func (s *AdminServer) HandleCreateConfigDocument(w http.ResponseWriter, r *http.Request) {
	identity, ok := requireRootAdminScope(w, r, "config_documents:write")
	if !ok {
		return
	}
	_, document, ok := decodeConfigDocumentInput(w, r)
	if !ok {
		return
	}
	store, ok := s.configDocumentStore(w)
	if !ok {
		return
	}
	record, err := store.StoreRevision(r.Context(), "default", identity.UserID, document)
	if err != nil {
		respondConfigDocumentError(w, err)
		return
	}
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(record))
}

func (s *AdminServer) HandleListConfigDocuments(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "config_documents:read"); !ok {
		return
	}
	store, ok := s.configDocumentStore(w)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
	records, err := store.List(r.Context(), "default", configdocuments.ListFilter{
		Kind:       protocol.ConfigDocumentKind(r.URL.Query().Get("kind")),
		ScopeKind:  protocol.ConfigDocumentScopeKind(r.URL.Query().Get("scope_kind")),
		ScopeRef:   r.URL.Query().Get("scope_ref"),
		DocumentID: r.URL.Query().Get("document_id"),
		Limit:      limit,
	})
	if err != nil {
		respondConfigDocumentError(w, err)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(records))
}

func (s *AdminServer) HandleGetConfigDocument(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "config_documents:read"); !ok {
		return
	}
	store, ok := s.configDocumentStore(w)
	if !ok {
		return
	}
	record, err := store.GetRevision(r.Context(), "default", r.PathValue("recordId"))
	if err != nil {
		respondConfigDocumentError(w, err)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(record))
}

func (s *AdminServer) HandleCompileConfigDocument(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "config_documents:read"); !ok {
		return
	}
	store, ok := s.configDocumentStore(w)
	if !ok {
		return
	}
	record, err := store.GetRevision(r.Context(), "default", r.PathValue("recordId"))
	if err != nil {
		respondConfigDocumentError(w, err)
		return
	}
	var input struct {
		OperatorValues protocol.MinimumSufficientBrief `json:"operator_values,omitempty"`
		PolicyValues   protocol.MinimumSufficientBrief `json:"policy_values,omitempty"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil && !errors.Is(err, io.EOF) {
			respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
	}
	compiled, err := configdocuments.CompileDocument(record.Document, input.OperatorValues, input.PolicyValues)
	if err != nil {
		respondConfigDocumentError(w, err)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(compiled))
}

func (s *AdminServer) HandleActivateConfigDocument(w http.ResponseWriter, r *http.Request) {
	identity, ok := requireRootAdminScope(w, r, "config_documents:write")
	if !ok {
		return
	}
	action := configdocuments.ActivationAction(r.PathValue("action"))
	if action != configdocuments.ActivationActionActivate && action != configdocuments.ActivationActionRollback {
		respondAPIError(w, "Unsupported config document action", http.StatusBadRequest)
		return
	}
	store, ok := s.configDocumentStore(w)
	if !ok {
		return
	}
	auditID, _ := s.createAuditEvent(protocol.TemplateChatToProposal, "config-document-"+string(action), "Configuration revision "+string(action), attachActorIdentity(map[string]any{
		"actor":         "operator",
		"user":          auditUserLabelFromRequest(r),
		"action":        "config_document_" + string(action),
		"result_status": "requested",
		"resource":      r.PathValue("recordId"),
	}, r))
	result, err := store.ActivateRevision(r.Context(), "default", r.PathValue("recordId"), identity.UserID, auditID, action)
	if err != nil {
		respondConfigDocumentError(w, err)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(result))
}

func decodeConfigDocumentInput(w http.ResponseWriter, r *http.Request) (configDocumentInput, protocol.ConfigDocument, bool) {
	var input configDocumentInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		respondAPIError(w, "Invalid config document request", http.StatusBadRequest)
		return input, protocol.ConfigDocument{}, false
	}
	selected := 0
	if input.Document != nil {
		selected++
	}
	if strings.TrimSpace(input.Content) != "" {
		selected++
	}
	if strings.TrimSpace(input.Path) != "" {
		selected++
	}
	if selected != 1 {
		respondAPIError(w, "Provide exactly one of document, content, or path", http.StatusBadRequest)
		return input, protocol.ConfigDocument{}, false
	}
	if input.Document != nil {
		return input, *input.Document, true
	}
	var document protocol.ConfigDocument
	var err error
	if strings.TrimSpace(input.Path) != "" {
		document, err = configdocuments.LoadDocumentFile(configdocuments.ConfiguredRoot(), input.Path)
	} else {
		document, err = configdocuments.ParseDocument([]byte(input.Content), input.Format)
	}
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return input, protocol.ConfigDocument{}, false
	}
	return input, document, true
}

func (s *AdminServer) configDocumentStore(w http.ResponseWriter) (*configdocuments.Store, bool) {
	if s.DB == nil {
		respondAPIError(w, "Configuration database unavailable", http.StatusServiceUnavailable)
		return nil, false
	}
	return configdocuments.NewStore(s.DB), true
}

func respondConfigDocumentError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, configdocuments.ErrRevisionNotFound), errors.Is(err, sql.ErrNoRows):
		status = http.StatusNotFound
	case strings.Contains(err.Error(), "database not available"):
		status = http.StatusServiceUnavailable
	}
	respondAPIError(w, err.Error(), status)
}
