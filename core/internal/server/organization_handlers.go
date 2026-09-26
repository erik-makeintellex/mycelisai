package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// organizationsWriteScope gates organization create/update to root admin until
// the A2 route-scope matrix supersedes it (owner decision, S1).
const organizationsWriteScope = "organizations:write"

func (s *AdminServer) handleListOrganizations(w http.ResponseWriter, r *http.Request) {
	summaries, err := s.organizationStore().List(r.Context())
	if err != nil {
		respondOrganizationStoreError(w, err)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(summaries))
}

func (s *AdminServer) emitReviewLoopEvent(ctx context.Context, orgID string, eventKind ReviewLoopEventKind) {
	if _, err := s.triggerReviewLoopsForEvent(ctx, orgID, eventKind); err != nil {
		log.Printf("[review-loop-event] organization=%s event=%s skipped error=%v", orgID, eventKind, err)
	}
}

func (s *AdminServer) handleCreateOrganization(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, organizationsWriteScope); !ok {
		return
	}
	var req OrganizationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid organization create request", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Purpose = strings.TrimSpace(req.Purpose)
	req.TemplateID = strings.TrimSpace(req.TemplateID)

	if req.Name == "" {
		respondAPIError(w, "organization name is required", http.StatusBadRequest)
		return
	}
	if req.Purpose == "" {
		respondAPIError(w, "organization purpose is required", http.StatusBadRequest)
		return
	}
	if req.StartMode != OrganizationStartModeTemplate && req.StartMode != OrganizationStartModeEmpty {
		respondAPIError(w, "start_mode must be template or empty", http.StatusBadRequest)
		return
	}

	var template *OrganizationTemplateSummary
	if req.StartMode == OrganizationStartModeTemplate {
		if req.TemplateID == "" {
			respondAPIError(w, "template_id is required when start_mode is template", http.StatusBadRequest)
			return
		}
		resolved, err := s.resolveStarterTemplate(req.TemplateID)
		if err != nil {
			respondAPIError(w, "starter template not found", http.StatusNotFound)
			return
		}
		template = resolved
	}

	fixtureScopeID, ok := s.qaFixtureScopeFromRequest(w, r)
	if !ok {
		return
	}
	home := s.buildOrganizationHome(req, template)
	home.QAFixtureScopeID = fixtureScopeID
	err := s.withQAFixtureScopeLock(r.Context(), fixtureScopeID, func() error {
		if err := s.claimQAFixtureResourcesLocked(r.Context(), fixtureScopeID, []qaFixtureResource{{Kind: "organization", Ref: home.ID}}); err != nil {
			return err
		}
		saved, err := s.organizationStore().Save(r.Context(), home)
		if err != nil {
			return err
		}
		home = saved
		return nil
	})
	if errors.Is(err, ErrOrganizationConflict) || errors.Is(err, ErrOrganizationStoreUnavailable) {
		respondOrganizationStoreError(w, err)
		return
	}
	if err != nil {
		respondAPIError(w, "Failed to bind organization cleanup ownership", http.StatusServiceUnavailable)
		return
	}
	s.loopProfileStore().EnsureDefaults(home)
	s.emitReviewLoopEvent(r.Context(), home.ID, ReviewLoopEventOrganizationCreated)
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(home))
}

func (s *AdminServer) handleGetOrganizationHome(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	home, ok := s.loadOrganizationForRequest(w, r, id)
	if !ok {
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(normalizeOrganizationHome(home)))
}

func (s *AdminServer) handleUpdateOrganizationAIEngine(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, organizationsWriteScope); !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	var req OrganizationAIEngineUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid AI Engine update request", http.StatusBadRequest)
		return
	}

	profile, ok := lookupOrganizationAIEngineProfile(req.ProfileID)
	if !ok {
		respondAPIError(w, "profile_id must be one of the guided AI Engine options", http.StatusBadRequest)
		return
	}

	updated, err := s.organizationStore().Update(r.Context(), id, func(home OrganizationHomePayload) OrganizationHomePayload {
		home.AIEngineProfileID = string(profile.ID)
		home.AIEngineSettingsSummary = profile.Summary
		return normalizeOrganizationHome(home)
	})
	if err != nil {
		respondOrganizationStoreError(w, err)
		return
	}

	s.emitReviewLoopEvent(r.Context(), updated.ID, ReviewLoopEventOrganizationAIEngineChanged)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleUpdateResponseContract(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, organizationsWriteScope); !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	var req ResponseContractUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid Response Contract update request", http.StatusBadRequest)
		return
	}

	profile, ok := lookupResponseContractProfile(req.ProfileID)
	if !ok {
		respondAPIError(w, "profile_id must be one of the guided Response Style options", http.StatusBadRequest)
		return
	}

	updated, err := s.organizationStore().Update(r.Context(), id, func(home OrganizationHomePayload) OrganizationHomePayload {
		home.ResponseContractProfileID = string(profile.ID)
		home.ResponseContractSummary = profile.Summary
		return normalizeOrganizationHome(home)
	})
	if err != nil {
		respondOrganizationStoreError(w, err)
		return
	}

	s.emitReviewLoopEvent(r.Context(), updated.ID, ReviewLoopEventResponseContractChanged)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleGetOrganizationOutputModelRouting(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	home, ok := s.loadOrganizationForRequest(w, r, id)
	if !ok {
		return
	}
	home = normalizeOrganizationHome(home)

	catalog := synthesizeOutputModelCatalog(s.listLocalOllamaModelIDs())
	payload := OrganizationOutputModelRoutingPayload{
		RoutingMode:                home.OutputModelRoutingMode,
		DefaultModelID:             home.DefaultOutputModelID,
		DefaultModelSummary:        home.DefaultOutputModelSummary,
		Bindings:                   append([]OrganizationOutputModelBinding(nil), home.OutputModelBindings...),
		AvailableModels:            catalog,
		RecommendedModels:          recommendedOutputModels(catalog),
		ReviewCandidates:           outputModelReviewCandidates(catalog),
		HardwareSummary:            "Local-first self-hosted posture tuned for the current Ollama inventory and a 16GB-class GPU host.",
		ReviewPermissionPrompt:     "Ask the owner/admin before Soma reviews potential model behavior for a requested output or changes saved routing.",
		AutomaticSelectionCriteria: outputModelAutomaticSelectionCriteria(),
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(payload))
}

func (s *AdminServer) handleUpdateOrganizationOutputModelRouting(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, organizationsWriteScope); !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	var req OrganizationOutputModelRoutingUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid output model routing update request", http.StatusBadRequest)
		return
	}

	routingMode := normalizeOrganizationOutputModelRoutingMode(req.RoutingMode)
	defaultModelID := strings.TrimSpace(req.DefaultModelID)
	if defaultModelID == "" {
		respondAPIError(w, "default_model_id is required", http.StatusBadRequest)
		return
	}

	validOutputTypes := make(map[string]struct{}, len(canonicalOutputTypeIDs()))
	for _, outputTypeID := range canonicalOutputTypeIDs() {
		validOutputTypes[string(outputTypeID)] = struct{}{}
	}

	normalizedBindings := make([]OrganizationOutputModelBinding, 0, len(req.Bindings))
	for _, binding := range req.Bindings {
		outputTypeID := strings.TrimSpace(binding.OutputTypeID)
		if _, ok := validOutputTypes[outputTypeID]; !ok {
			respondAPIError(w, "bindings contain an unknown output_type_id", http.StatusBadRequest)
			return
		}
		modelID := strings.TrimSpace(binding.ModelID)
		if binding.UseOrganizationDefault {
			modelID = defaultModelID
		}
		if modelID == "" {
			respondAPIError(w, "bindings require a model_id unless they use the organization default", http.StatusBadRequest)
			return
		}
		normalizedBindings = append(normalizedBindings, OrganizationOutputModelBinding{
			OutputTypeID:    outputTypeID,
			OutputTypeLabel: outputTypeLabel(outputTypeID),
			ModelID:         modelID,
			ModelSummary:    outputModelLabel(modelID),
		})
	}

	updated, err := s.organizationStore().Update(r.Context(), id, func(home OrganizationHomePayload) OrganizationHomePayload {
		home = normalizeOrganizationHome(home)
		home.OutputModelRoutingMode = string(routingMode)
		home.DefaultOutputModelID = defaultModelID
		home.DefaultOutputModelSummary = outputModelLabel(defaultModelID)
		home.OutputModelBindings = normalizedOrganizationOutputModelBindings(normalizedBindings, defaultModelID)
		return normalizeOrganizationHome(home)
	})
	if err != nil {
		respondOrganizationStoreError(w, err)
		return
	}

	s.emitReviewLoopEvent(r.Context(), updated.ID, ReviewLoopEventOrganizationAIEngineChanged)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}
