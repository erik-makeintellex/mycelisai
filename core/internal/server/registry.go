package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/registry"
)

// GET /api/v1/registry/templates
func (s *AdminServer) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templates, err := s.Registry.ListTemplates(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, templates)
}

// POST /api/v1/registry/templates
func (s *AdminServer) handleRegisterTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tmpl registry.ConnectorTemplate
	if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if err := s.Registry.RegisterTemplate(r.Context(), tmpl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, map[string]string{"status": "registered"})
}

// POST /api/v1/teams/{id}/connectors
func (s *AdminServer) handleInstallConnector(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract Team ID
	// Path: /api/v1/teams/{id}/connectors
	// IDs are UUIDs (36 chars)
	// prefix: /api/v1/teams/
	prefixLen := len("/api/v1/teams/")
	path := r.URL.Path
	if len(path) < prefixLen+36 {
		http.Error(w, "Invalid Path", http.StatusBadRequest)
		return
	}
	teamIDStr := path[prefixLen : prefixLen+36]

	if _, err := uuid.Parse(teamIDStr); err != nil {
		http.Error(w, "Invalid Team UUID", http.StatusBadRequest)
		return
	}

	var req struct {
		TemplateID uuid.UUID       `json:"template_id"`
		Name       string          `json:"name"`
		Config     json.RawMessage `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	respondInstallConnectorError(w, s.Registry.InstallConnector(r.Context(), req.TemplateID, req.Config))
}

// respondInstallConnectorError maps install outcomes to normalized envelopes.
// A valid install is a 501 connector_deployment_unavailable blocker: no
// deployer exists, so nothing is recorded and nothing claims to provision.
func respondInstallConnectorError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, registry.ErrConnectorDeploymentUnavailable):
		respondGovernanceError(w, http.StatusNotImplemented,
			"Connector deployment is not available yet; nothing was installed.",
			"connector_deployment_unavailable", "Use MCP servers or configured providers for external data until connector deployment ships.")
	case errors.Is(err, registry.ErrTemplateNotFound):
		respondAPIError(w, "connector template not found", http.StatusNotFound)
	case errors.Is(err, registry.ErrInvalidConnectorConfig):
		respondAPIError(w, err.Error(), http.StatusBadRequest)
	case err != nil:
		respondAPIError(w, "connector install failed", http.StatusInternalServerError)
	default:
		respondAPIError(w, "connector install returned no outcome", http.StatusInternalServerError)
	}
}

// GET /api/v1/teams/{id}/wiring
func (s *AdminServer) handleGetWiring(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract Team ID
	// Path: /api/v1/teams/{id}/wiring
	path := r.URL.Path
	prefix := "/api/v1/teams/"
	suffix := "/wiring"
	if len(path) < len(prefix)+36+len(suffix) {
		http.Error(w, "Invalid Path", http.StatusBadRequest)
		return
	}

	// id start at len(prefix)
	idStr := path[len(prefix) : len(prefix)+36]
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid Team UUID", http.StatusBadRequest)
		return
	}

	graph, err := s.Registry.GetWiring(r.Context(), teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, graph)
}
