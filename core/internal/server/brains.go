package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

// GET /api/v1/brains — list all providers with enriched metadata + health status.
func (s *AdminServer) HandleListBrains(w http.ResponseWriter, r *http.Request) {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondJSON(w, map[string]any{"ok": true, "data": []BrainEntry{}})
		return
	}

	entries := make([]BrainEntry, 0, len(s.Cognitive.Config.Providers))
	for id, prov := range s.Cognitive.Config.Providers {
		prov = cognitive.NormalizeProviderTokenDefaults(prov)
		entries = append(entries, brainEntryFromProvider(id, prov, s.brainStatus(r.Context(), id, prov.Enabled)))
	}

	respondJSON(w, map[string]any{"ok": true, "data": entries})
}

// PUT /api/v1/brains/{id}/toggle — enable or disable a provider.
func (s *AdminServer) HandleToggleBrain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing provider ID", http.StatusBadRequest)
		return
	}

	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondError(w, "Cognitive system offline", http.StatusServiceUnavailable)
		return
	}

	prov, ok := s.Cognitive.Config.Providers[id]
	if !ok {
		respondError(w, "Provider not found", http.StatusNotFound)
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	prov.Enabled = req.Enabled
	s.Cognitive.Config.Providers[id] = prov

	// Persist to YAML so toggle survives restart
	if err := s.Cognitive.SaveConfig(); err != nil {
		log.Printf("Failed to persist brain toggle for %s: %v", id, err)
		respondError(w, "Toggle applied but failed to persist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{"ok": true, "data": map[string]any{"id": id, "enabled": prov.Enabled}})
}

// PUT /api/v1/brains/{id}/policy — update usage policy for a provider.
func (s *AdminServer) HandleUpdateBrainPolicy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing provider ID", http.StatusBadRequest)
		return
	}

	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondError(w, "Cognitive system offline", http.StatusServiceUnavailable)
		return
	}

	prov, ok := s.Cognitive.Config.Providers[id]
	if !ok {
		respondError(w, "Provider not found", http.StatusNotFound)
		return
	}

	var req struct {
		UsagePolicy        string   `json:"usage_policy"`
		TokenBudgetProfile string   `json:"token_budget_profile"`
		MaxOutputTokens    int      `json:"max_output_tokens"`
		RolesAllowed       []string `json:"roles_allowed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if req.UsagePolicy != "" {
		prov.UsagePolicy = req.UsagePolicy
	}
	if req.TokenBudgetProfile != "" {
		prov.TokenBudgetProfile = req.TokenBudgetProfile
	}
	if req.MaxOutputTokens > 0 {
		prov.MaxOutputTokens = req.MaxOutputTokens
	}
	if len(req.RolesAllowed) > 0 {
		prov.RolesAllowed = req.RolesAllowed
	}
	prov = cognitive.NormalizeProviderTokenDefaults(prov)
	s.Cognitive.Config.Providers[id] = prov

	// Persist to YAML so policy survives restart
	if err := s.Cognitive.SaveConfig(); err != nil {
		log.Printf("Failed to persist brain policy for %s: %v", id, err)
		respondError(w, "Policy applied but failed to persist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{"ok": true, "data": map[string]any{
		"id":                   id,
		"usage_policy":         prov.UsagePolicy,
		"token_budget_profile": prov.TokenBudgetProfile,
		"max_output_tokens":    prov.MaxOutputTokens,
		"roles_allowed":        prov.RolesAllowed,
	}})
}

// POST /api/v1/brains — add a new provider and hot-inject it into the running router.
func (s *AdminServer) HandleAddBrain(w http.ResponseWriter, r *http.Request) {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondError(w, "Cognitive system offline", http.StatusServiceUnavailable)
		return
	}

	var req brainUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if req.ID == "" || !validProviderID.MatchString(req.ID) {
		respondError(w, "Invalid provider id: must be lowercase alphanumeric with dashes/underscores", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		respondError(w, "type is required", http.StatusBadRequest)
		return
	}
	if _, exists := s.Cognitive.Config.Providers[req.ID]; exists {
		respondError(w, "provider id already exists", http.StatusConflict)
		return
	}

	cfg := providerConfigFromBrainRequest(req, true)

	if err := s.Cognitive.AddProvider(req.ID, cfg); err != nil {
		log.Printf("AddProvider %s failed: %v", req.ID, err)
		respondError(w, "Failed to add provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{"ok": true, "data": brainEntryFromProvider(req.ID, cfg, s.brainStatus(r.Context(), req.ID, cfg.Enabled))})
}

// PUT /api/v1/brains/{id} — update a provider's full configuration and hot-reload it.
func (s *AdminServer) HandleUpdateBrain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing provider ID", http.StatusBadRequest)
		return
	}
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondError(w, "Cognitive system offline", http.StatusServiceUnavailable)
		return
	}
	if _, exists := s.Cognitive.Config.Providers[id]; !exists {
		respondError(w, "Provider not found", http.StatusNotFound)
		return
	}

	var req brainUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	cfg := providerConfigFromBrainRequest(req, false)

	if err := s.Cognitive.UpdateProvider(id, cfg); err != nil {
		log.Printf("UpdateProvider %s failed: %v", id, err)
		respondError(w, "Failed to update provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{"ok": true, "data": map[string]any{"id": id, "updated": true}})
}

// DELETE /api/v1/brains/{id} — remove a provider and its adapter.
func (s *AdminServer) HandleDeleteBrain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing provider ID", http.StatusBadRequest)
		return
	}
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondError(w, "Cognitive system offline", http.StatusServiceUnavailable)
		return
	}
	if _, exists := s.Cognitive.Config.Providers[id]; !exists {
		respondError(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Guard: refuse to delete the last remaining provider
	if len(s.Cognitive.Config.Providers) <= 1 {
		respondError(w, "Cannot delete the last provider — at least one must remain", http.StatusConflict)
		return
	}

	if err := s.Cognitive.RemoveProvider(id); err != nil {
		log.Printf("RemoveProvider %s failed: %v", id, err)
		respondError(w, "Failed to remove provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{"ok": true, "data": map[string]any{"id": id, "deleted": true}})
}

// POST /api/v1/brains/{id}/probe — live health check on a single provider.
func (s *AdminServer) HandleProbeBrain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing provider ID", http.StatusBadRequest)
		return
	}
	if s.Cognitive == nil {
		respondError(w, "Cognitive system offline", http.StatusServiceUnavailable)
		return
	}

	adapter, ok := s.Cognitive.Adapters[id]
	if !ok {
		respondError(w, "Provider not found or not initialized", http.StatusNotFound)
		return
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	alive, _ := adapter.Probe(ctx)
	cancel()
	latency := time.Since(start).Milliseconds()

	respondJSON(w, map[string]any{"ok": true, "data": map[string]any{
		"id":         id,
		"alive":      alive,
		"latency_ms": latency,
	}})
}
