package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

// GET /api/v1/cognitive/status
// Returns health and configuration of all cognitive engines (text + configured media).
func (s *AdminServer) HandleCognitiveStatus(w http.ResponseWriter, r *http.Request) {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondJSON(w, map[string]any{"text": map[string]string{"status": "offline"}, "media": map[string]string{"status": "offline"}})
		return
	}

	type engineStatus struct {
		Status       string `json:"status"`
		Endpoint     string `json:"endpoint,omitempty"`
		Model        string `json:"model,omitempty"`
		ProviderID   string `json:"provider_id,omitempty"`
		ProviderType string `json:"provider_type,omitempty"`
		Location     string `json:"location,omitempty"`
		DataBoundary string `json:"data_boundary,omitempty"`
		UsagePolicy  string `json:"usage_policy,omitempty"`
		Configured   bool   `json:"configured,omitempty"`
		Enabled      *bool  `json:"enabled,omitempty"`
		Detail       string `json:"detail,omitempty"`
	}

	// text.status comes from the chat profile's effective provider only;
	// every profile carries its own availability, code, and probe result.
	// Disabled providers are never probed.
	text, routes, routeHealth, overlayError := s.cognitiveProfileRouteStatus(r.Context())
	cfg := s.Cognitive.ConfigSnapshot()
	media := &engineStatus{Status: "offline"}

	// Probe media engine
	if cfg.Media != nil {
		provider := cfg.Media.EffectiveProvider()
		mediaEnabled := provider.IsEnabled()
		media = &engineStatus{
			Status:       "offline",
			Endpoint:     provider.Endpoint,
			Model:        provider.ModelID,
			ProviderID:   provider.ProviderID,
			ProviderType: provider.Type,
			Location:     provider.Location,
			DataBoundary: provider.DataBoundary,
			UsagePolicy:  provider.UsagePolicy,
			Configured:   cfg.Media.IsConfigured(),
			Enabled:      &mediaEnabled,
		}

		if !mediaEnabled {
			media.Status = "disabled"
		} else if provider.Location == cognitive.DefaultMediaRemoteLocation {
			media.Status = "configured"
			media.Detail = "Hosted media provider is configured; live provider health is checked during generation."
		} else if strings.TrimSpace(provider.Endpoint) != "" {
			healthURL := strings.TrimSuffix(provider.Endpoint, "/v1") + "/health"
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()

			httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
			resp, err := http.DefaultClient.Do(httpReq)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					media.Status = "online"
				}
			}
		}
	}

	if !cognitiveFullView(r) {
		respondJSON(w, narrowCognitiveStatus(text, media.Status, routes, routeHealth, overlayError))
		return
	}

	// Surface the effective root provider (root_provider/MYCELIS_ROOT_PROVIDER),
	// its model id, and every profile's effective route (source, availability,
	// code, override origin, probe result) so operators and live tests can
	// prove what Core actually routes to, without exposing secrets.
	response := map[string]any{
		"text":                 text,
		"media":                media,
		"profiles":             routes,
		"profile_route_health": routeHealth,
		"overlay_error":        overlayError,
	}
	if rootID := strings.TrimSpace(cfg.RootProvider); rootID != "" {
		response["root_provider"] = rootID
		if rootProvider, ok := cfg.Providers[rootID]; ok {
			response["root_provider_model"] = rootProvider.ModelID
		}
	}

	respondJSON(w, response)
}

// POST /api/v1/cognitive/infer
func (s *AdminServer) handleInfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Cognitive == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	var req cognitive.InferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	resp, err := s.Cognitive.Infer(req)
	if err != nil {
		log.Printf("Inference Failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, resp)
}

// GET /api/v1/cognitive/config
// Returns the current Cognitive Configuration (Profiles + Providers)
func (s *AdminServer) HandleCognitiveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Cognitive == nil || s.Cognitive.Config == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	// The config snapshot carries provider endpoints and override detail, so
	// only root admin with cognitive:read or cognitive:write may read it (S6e).
	if IdentityFromContext(r.Context()) == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	if !cognitiveFullView(r) {
		respondAPIError(w, "Root admin with cognitive:read required", http.StatusForbidden)
		return
	}

	// Return a locked snapshot of the config struct
	respondJSON(w, s.Cognitive.ConfigSnapshot())
}

// PUT /api/v1/cognitive/profiles and DELETE .../profiles/{profile}/override
// live in cognitive_profile_overrides.go.

// PUT /api/v1/cognitive/providers/{id}
// Updates a provider's configuration (endpoint, model_id, api_key_env).
// Root admin + cognitive:write. It updates config and YAML only; it does not
// rebuild the adapter.
func (s *AdminServer) HandleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, cognitiveWriteScope); !ok {
		return
	}
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	providerID := r.PathValue("id")
	if providerID == "" {
		http.Error(w, "Missing provider ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Endpoint  string `json:"endpoint,omitempty"`
		ModelID   string `json:"model_id,omitempty"`
		APIKey    string `json:"api_key,omitempty"`     // Rejected; kept only for explicit legacy payload errors.
		APIKeyEnv string `json:"api_key_env,omitempty"` // Env var name (persisted to YAML)
		Type      string `json:"type,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.APIKey) != "" {
		http.Error(w, "raw api_key values are not accepted; use api_key_env or a deployment secret reference", http.StatusBadRequest)
		return
	}

	// Serialize the read-modify-write with every other routing mutation so a
	// stale snapshot can never undo a brains toggle or slip past the
	// provider_bound check (lock order: routing_mutation_authority.go).
	unlock := lockRoutingWrite()
	defer unlock()

	previous, existed := s.Cognitive.ProviderSnapshot(providerID)
	existing := previous
	if !existed {
		// Create new provider
		existing = cognitive.ProviderConfig{}
	}

	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.Endpoint != "" {
		existing.Endpoint = req.Endpoint
	}
	if req.ModelID != "" {
		existing.ModelID = req.ModelID
	}
	if req.APIKeyEnv != "" {
		existing.AuthKeyEnv = req.APIKeyEnv
	}
	// Same rule as PUT /brains/{id}: an update that leaves a bound provider
	// non-executable (disabled or blank model) would strand its profiles.
	if existed && !providerConfigExecutable(existing) && s.rejectIfProviderBound(w, providerID) {
		return
	}
	auditFields := brainAuditFields(providerID, existing)
	auditFields["created"] = !existed
	auditFields["endpoint_changed"] = existed && existing.Endpoint != previous.Endpoint
	auditFields["auth_ref_changed"] = existed && existing.AuthKeyEnv != previous.AuthKeyEnv
	if _, ok := s.auditRoutingMutation(w, r, auditProviderConfigUpdated, "Cognitive provider config updated", auditFields); !ok {
		return
	}
	s.Cognitive.StoreProviderConfig(providerID, existing)

	// Persist to YAML (AuthKey/AuthKeyEnv are json:"-" so won't leak)
	if err := s.Cognitive.SaveConfig(); err != nil {
		log.Printf("Failed to persist cognitive config: %v", err)
		s.recordRoutingMutationFailure(r, auditProviderConfigUpdated, "Cognitive provider config applied but not persisted", map[string]any{"provider_id": providerID})
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	log.Printf("Provider '%s' updated: model=%s", providerID, existing.ModelID)

	// Return sanitized provider info (no secrets)
	respondJSON(w, map[string]any{
		"id":         providerID,
		"type":       existing.Type,
		"endpoint":   existing.Endpoint,
		"model_id":   existing.ModelID,
		"configured": existing.AuthKey != "" || existing.AuthKeyEnv != "",
	})
}
