package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/conversationtemplates"
	"github.com/mycelis/core/pkg/protocol"
)

type conversationTemplateCreateRequest struct {
	Name                 string                                   `json:"name"`
	Description          string                                   `json:"description,omitempty"`
	Scope                protocol.ConversationTemplateScope       `json:"scope"`
	CreatorKind          protocol.ConversationTemplateCreatorKind `json:"creator_kind,omitempty"`
	Status               protocol.ConversationTemplateStatus      `json:"status,omitempty"`
	TemplateBody         string                                   `json:"template_body"`
	Variables            map[string]any                           `json:"variables,omitempty"`
	OutputContract       map[string]any                           `json:"output_contract,omitempty"`
	RecommendedTeamShape map[string]any                           `json:"recommended_team_shape,omitempty"`
	ModelRoutingHint     map[string]any                           `json:"model_routing_hint,omitempty"`
	GovernanceTags       []string                                 `json:"governance_tags,omitempty"`
}

type conversationTemplatePatchRequest struct {
	Name                 *string                                   `json:"name,omitempty"`
	Description          *string                                   `json:"description,omitempty"`
	Scope                *protocol.ConversationTemplateScope       `json:"scope,omitempty"`
	CreatorKind          *protocol.ConversationTemplateCreatorKind `json:"creator_kind,omitempty"`
	Status               *protocol.ConversationTemplateStatus      `json:"status,omitempty"`
	TemplateBody         *string                                   `json:"template_body,omitempty"`
	Variables            map[string]any                            `json:"variables,omitempty"`
	OutputContract       map[string]any                            `json:"output_contract,omitempty"`
	RecommendedTeamShape map[string]any                            `json:"recommended_team_shape,omitempty"`
	ModelRoutingHint     map[string]any                            `json:"model_routing_hint,omitempty"`
	GovernanceTags       []string                                  `json:"governance_tags,omitempty"`
}

type conversationTemplateInstantiateRequest struct {
	Variables map[string]any `json:"variables,omitempty"`
}

// GET /api/v1/conversation-templates
func (s *AdminServer) HandleListConversationTemplates(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "conversation_templates:read"); !ok {
		return
	}
	store, ok := s.conversationTemplateStore(w)
	if !ok {
		return
	}
	items, err := store.List(r.Context(), conversationtemplates.ListFilter{
		TenantID: "default",
		Scope:    r.URL.Query().Get("scope"),
		Status:   r.URL.Query().Get("status"),
		Limit:    parseLimit(r.URL.Query().Get("limit"), 50),
	})
	if err != nil {
		respondAPIError(w, "Failed to list conversation templates: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(items))
}

// POST /api/v1/conversation-templates
func (s *AdminServer) HandleCreateConversationTemplate(w http.ResponseWriter, r *http.Request) {
	identity, ok := requireRootAdminScope(w, r, "conversation_templates:write")
	if !ok {
		return
	}
	store, ok := s.conversationTemplateStore(w)
	if !ok {
		return
	}
	var req conversationTemplateCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	tpl := protocol.ConversationTemplate{
		TenantID:             "default",
		Name:                 req.Name,
		Description:          req.Description,
		Scope:                req.Scope,
		CreatedBy:            identity.UserID,
		CreatorKind:          req.CreatorKind,
		Status:               req.Status,
		TemplateBody:         req.TemplateBody,
		Variables:            req.Variables,
		OutputContract:       req.OutputContract,
		RecommendedTeamShape: req.RecommendedTeamShape,
		ModelRoutingHint:     req.ModelRoutingHint,
		GovernanceTags:       req.GovernanceTags,
	}
	created, err := store.Create(r.Context(), tpl)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(created))
}

// GET /api/v1/conversation-templates/{id}
func (s *AdminServer) HandleGetConversationTemplate(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "conversation_templates:read"); !ok {
		return
	}
	store, ok := s.conversationTemplateStore(w)
	if !ok {
		return
	}
	tpl, err := store.Get(r.Context(), strings.TrimSpace(r.PathValue("id")))
	if errors.Is(err, sql.ErrNoRows) {
		respondAPIError(w, "Conversation template not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, "Failed to load conversation template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(tpl))
}

// PATCH /api/v1/conversation-templates/{id}
func (s *AdminServer) HandleUpdateConversationTemplate(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "conversation_templates:write"); !ok {
		return
	}
	store, ok := s.conversationTemplateStore(w)
	if !ok {
		return
	}
	tpl, err := store.Get(r.Context(), strings.TrimSpace(r.PathValue("id")))
	if errors.Is(err, sql.ErrNoRows) {
		respondAPIError(w, "Conversation template not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, "Failed to load conversation template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var req conversationTemplatePatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	applyConversationTemplatePatch(tpl, req)
	updated, err := store.Update(r.Context(), *tpl)
	if errors.Is(err, sql.ErrNoRows) {
		respondAPIError(w, "Conversation template not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

// POST /api/v1/conversation-templates/{id}/instantiate
func (s *AdminServer) HandleInstantiateConversationTemplate(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "conversation_templates:use"); !ok {
		return
	}
	store, ok := s.conversationTemplateStore(w)
	if !ok {
		return
	}
	var req conversationTemplateInstantiateRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
	}
	instantiation, err := store.Instantiate(r.Context(), strings.TrimSpace(r.PathValue("id")), req.Variables)
	if errors.Is(err, sql.ErrNoRows) {
		respondAPIError(w, "Conversation template not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, "Failed to instantiate conversation template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(instantiation))
}

func (s *AdminServer) conversationTemplateStore(w http.ResponseWriter) (*conversationtemplates.Store, bool) {
	if s.DB == nil {
		respondAPIError(w, "Conversation templates database unavailable", http.StatusServiceUnavailable)
		return nil, false
	}
	return conversationtemplates.NewStore(s.DB), true
}

func applyConversationTemplatePatch(tpl *protocol.ConversationTemplate, req conversationTemplatePatchRequest) {
	if req.Name != nil {
		tpl.Name = *req.Name
	}
	if req.Description != nil {
		tpl.Description = *req.Description
	}
	if req.Scope != nil {
		tpl.Scope = *req.Scope
	}
	if req.CreatorKind != nil {
		tpl.CreatorKind = *req.CreatorKind
	}
	if req.Status != nil {
		tpl.Status = *req.Status
	}
	if req.TemplateBody != nil {
		tpl.TemplateBody = *req.TemplateBody
	}
	if req.Variables != nil {
		tpl.Variables = req.Variables
	}
	if req.OutputContract != nil {
		tpl.OutputContract = req.OutputContract
	}
	if req.RecommendedTeamShape != nil {
		tpl.RecommendedTeamShape = req.RecommendedTeamShape
	}
	if req.ModelRoutingHint != nil {
		tpl.ModelRoutingHint = req.ModelRoutingHint
	}
	if req.GovernanceTags != nil {
		tpl.GovernanceTags = req.GovernanceTags
	}
}
