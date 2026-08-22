package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func requireQAFixtureManagement(w http.ResponseWriter, r *http.Request) bool {
	if !qaFixtureManagementEnabled() {
		respondAPIError(w, "QA fixture management is disabled", http.StatusNotFound)
		return false
	}
	_, ok := requireRootAdminScope(w, r, "testing:fixtures")
	return ok
}

// qaFixtureScopeFromRequest resolves the optional test-only ownership scope.
// A missing header keeps normal product requests completely outside QA cleanup.
func (s *AdminServer) qaFixtureScopeFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	scopeID := strings.TrimSpace(r.Header.Get(qaFixtureScopeHeader))
	if scopeID == "" {
		return "", true
	}
	if !requireQAFixtureManagement(w, r) {
		return "", false
	}
	if err := s.validateOpenQAFixtureScope(r.Context(), scopeID); err != nil {
		respondAPIError(w, "QA fixture scope is not open", http.StatusConflict)
		return "", false
	}
	return scopeID, true
}

func (s *AdminServer) HandleCreateQAFixtureScope(w http.ResponseWriter, r *http.Request) {
	if !requireQAFixtureManagement(w, r) {
		return
	}
	var req qaFixtureScopeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	owner, err := normalizeQAFixtureIdentity("owner_ref", req.OwnerRef)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	execution, err := normalizeQAFixtureIdentity("execution_ref", req.ExecutionRef)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	expiresAt, err := qaFixtureExpiry(req.TTLSeconds)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	scope, err := s.createQAFixtureScope(r.Context(), owner, execution, expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		respondAPIError(w, "A fixture scope already exists for this owner and execution", http.StatusConflict)
		return
	}
	if err != nil {
		respondAPIError(w, "Failed to create fixture scope: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(scope))
}

func (s *AdminServer) HandleAddQAFixtureResources(w http.ResponseWriter, r *http.Request) {
	if !requireQAFixtureManagement(w, r) {
		return
	}
	scopeID := strings.TrimSpace(r.PathValue("id"))
	var req qaFixtureResourcesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	resources := make([]qaFixtureResource, 0, len(req.Resources))
	for _, raw := range req.Resources {
		resource, err := normalizeQAFixtureResource(raw)
		if err != nil {
			respondAPIError(w, err.Error(), http.StatusBadRequest)
			return
		}
		resources = append(resources, resource)
	}
	if scopeID == "" || len(resources) == 0 {
		respondAPIError(w, "scope id and at least one resource are required", http.StatusBadRequest)
		return
	}
	owner, execution, err := normalizeQAFixtureOwnership(req.OwnerRef, req.ExecutionRef)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	registered, err := s.addQAFixtureResources(r.Context(), scopeID, owner, execution, resources)
	if !respondQAFixtureStoreError(w, err) {
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"scope_id":  scopeID,
		"resources": registered,
	}))
}

func (s *AdminServer) HandlePurgeQAFixtureScope(w http.ResponseWriter, r *http.Request) {
	if !requireQAFixtureManagement(w, r) {
		return
	}
	scopeID := strings.TrimSpace(r.PathValue("id"))
	var req qaFixturePurgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	owner, execution, err := normalizeQAFixtureOwnership(req.OwnerRef, req.ExecutionRef)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.OwnerRef = owner
	req.ExecutionRef = execution
	result, err := s.purgeQAFixtureScope(r.Context(), scopeID, req)
	if !respondQAFixtureStoreError(w, err) {
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(result))
}

func respondQAFixtureStoreError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return true
	}
	switch {
	case errors.Is(err, sql.ErrNoRows):
		respondAPIError(w, "Fixture scope not found", http.StatusNotFound)
	case errors.Is(err, errQAFixtureScopeMismatch):
		respondAPIError(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, errQAFixtureScopeClosed):
		respondAPIError(w, err.Error(), http.StatusConflict)
	case errors.Is(err, errQAFixtureResourceUnowned):
		respondAPIError(w, err.Error(), http.StatusConflict)
	default:
		respondAPIError(w, "Fixture management failed: "+err.Error(), http.StatusServiceUnavailable)
	}
	return false
}
