package server

import (
	"encoding/json"
	"net/http"

	"github.com/mycelis/core/internal/codecontext"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) codeContextService() *codecontext.Service {
	if s.CodeContext != nil {
		return s.CodeContext
	}
	s.CodeContext = codecontext.NewService(codecontext.ConfigFromEnv())
	return s.CodeContext
}

func (s *AdminServer) HandleCodeContextSources(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "code_context:read"); !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		sources, err := s.codeContextService().ListSources(r.Context())
		if err != nil {
			respondAPIError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(sources))
	case http.MethodPost:
		if _, ok := requireRootAdminScope(w, r, "code_context:write"); !ok {
			return
		}
		var req codeContextSourceRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			respondAPIError(w, "Invalid code context source request", http.StatusBadRequest)
			return
		}
		input, err := req.input()
		if err != nil {
			respondAPIError(w, err.Error(), http.StatusBadRequest)
			return
		}
		source, err := s.codeContextService().RegisterSource(r.Context(), input)
		if err != nil {
			respondAPIError(w, err.Error(), http.StatusBadRequest)
			return
		}
		respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(source))
	default:
		respondAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *AdminServer) HandleCodeContextIndex(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "code_context:write"); !ok {
		return
	}
	var req struct {
		SourceID string `json:"source_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid code context index request", http.StatusBadRequest)
		return
	}
	resp, err := s.codeContextService().Index(r.Context(), req.SourceID)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(resp))
}

func (s *AdminServer) HandleCodeContextQuery(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "code_context:read"); !ok {
		return
	}
	var req codecontext.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid code context query request", http.StatusBadRequest)
		return
	}
	resp, err := s.codeContextService().Query(r.Context(), req)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(resp))
}

func (s *AdminServer) HandleCodeContextImpact(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "code_context:read"); !ok {
		return
	}
	var req codecontext.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid code context impact request", http.StatusBadRequest)
		return
	}
	resp, err := s.codeContextService().Impact(r.Context(), req)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(resp))
}

func (s *AdminServer) HandleCodeContextExplain(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "code_context:read"); !ok {
		return
	}
	var req codecontext.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid code context explain request", http.StatusBadRequest)
		return
	}
	resp, err := s.codeContextService().Explain(r.Context(), req)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(resp))
}

type codeContextSourceRequest struct {
	Document *protocol.ConfigDocument `json:"document,omitempty"`
	codecontext.SourceInput
}

func (r codeContextSourceRequest) input() (codecontext.SourceInput, error) {
	if r.Document == nil {
		return r.SourceInput, nil
	}
	compiled, err := configDocumentCodeContextSource(*r.Document)
	if err != nil {
		return codecontext.SourceInput{}, err
	}
	return compiled, nil
}

func configDocumentCodeContextSource(document protocol.ConfigDocument) (codecontext.SourceInput, error) {
	issues := protocol.ValidateConfigDocument(document)
	if len(issues) != 0 {
		return codecontext.SourceInput{}, configDocumentValidationRequestError{issue: issues[0]}
	}
	if document.Kind != protocol.ConfigDocumentKindCodeContextSource {
		return codecontext.SourceInput{}, configDocumentValidationRequestError{issue: protocol.ConfigDocumentValidationIssue{
			Code: "config.unsupported_kind", Field: "kind", Message: "document must be a CodeContextSource",
		}}
	}
	spec, err := protocol.DecodeCodeContextSourceSpec(document.Spec)
	if err != nil {
		return codecontext.SourceInput{}, err
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		return codecontext.SourceInput{}, err
	}
	return codecontext.SourceInput{
		ID:                spec.SourceID,
		Name:              document.Metadata.Name,
		SourceType:        spec.SourceType,
		RootPath:          spec.RootPath,
		ScopeKind:         string(document.Metadata.Scope.Kind),
		ScopeRef:          document.Metadata.Scope.Ref,
		IncludeGlobs:      spec.IncludeGlobs,
		ExcludeGlobs:      spec.ExcludeGlobs,
		Languages:         spec.Languages,
		ConfigDigest:      digest,
		ExtractionVersion: spec.ExtractionVersion,
		SensitivityClass:  spec.SensitivityClass,
		TrustClass:        spec.TrustClass,
	}, nil
}

type configDocumentValidationRequestError struct {
	issue protocol.ConfigDocumentValidationIssue
}

func (e configDocumentValidationRequestError) Error() string {
	return e.issue.Code + ": " + e.issue.Message
}
