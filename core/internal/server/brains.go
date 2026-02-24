package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

// validProviderID matches alphanumeric + dash/underscore IDs.
var validProviderID = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

// BrainEntry is the enriched provider info returned by GET /api/v1/brains.
type BrainEntry struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Endpoint     string   `json:"endpoint,omitempty"`
	ModelID      string   `json:"model_id"`
	Location     string   `json:"location"`
	DataBoundary string   `json:"data_boundary"`
	UsagePolicy  string   `json:"usage_policy"`
	RolesAllowed []string `json:"roles_allowed"`
	Enabled      bool     `json:"enabled"`
	Status       string   `json:"status"` // "online" | "offline" | "disabled"
}

// GET /api/v1/brains — list all providers with enriched metadata + health status.
func (s *AdminServer) HandleListBrains(w http.ResponseWriter, r *http.Request) {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondJSON(w, map[string]any{"ok": true, "data": []BrainEntry{}})
		return
	}

	entries := make([]BrainEntry, 0, len(s.Cognitive.Config.Providers))
	for id, prov := range s.Cognitive.Config.Providers {
		entry := BrainEntry{
			ID:           id,
			Type:         prov.Type,
			Endpoint:     prov.Endpoint,
			ModelID:      prov.ModelID,
			Location:     prov.Location,
			DataBoundary: prov.DataBoundary,
			UsagePolicy:  prov.UsagePolicy,
			RolesAllowed: prov.RolesAllowed,
			Enabled:      prov.Enabled,
		}

		if entry.Location == "" {
			entry.Location = "local"
		}
		if entry.DataBoundary == "" {
			entry.DataBoundary = "local_only"
		}
		if entry.UsagePolicy == "" {
			entry.UsagePolicy = "local_first"
		}

		if !prov.Enabled {
			entry.Status = "disabled"
		} else if adapter, ok := s.Cognitive.Adapters[id]; ok {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			alive, _ := adapter.Probe(ctx)
			cancel()
			if alive {
				entry.Status = "online"
			} else {
				entry.Status = "offline"
			}
		} else {
			entry.Status = "offline"
		}

		entries = append(entries, entry)
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
		UsagePolicy  string   `json:"usage_policy"`
		RolesAllowed []string `json:"roles_allowed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if req.UsagePolicy != "" {
		prov.UsagePolicy = req.UsagePolicy
	}
	if len(req.RolesAllowed) > 0 {
		prov.RolesAllowed = req.RolesAllowed
	}
	s.Cognitive.Config.Providers[id] = prov

	// Persist to YAML so policy survives restart
	if err := s.Cognitive.SaveConfig(); err != nil {
		log.Printf("Failed to persist brain policy for %s: %v", id, err)
		respondError(w, "Policy applied but failed to persist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{"ok": true, "data": map[string]any{"id": id, "usage_policy": prov.UsagePolicy, "roles_allowed": prov.RolesAllowed}})
}

// brainUpsertRequest is the shared request body for add/update operations.
type brainUpsertRequest struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Endpoint     string   `json:"endpoint"`
	ModelID      string   `json:"model_id"`
	APIKey       string   `json:"api_key"`       // write-only — never returned
	Location     string   `json:"location"`
	DataBoundary string   `json:"data_boundary"`
	UsagePolicy  string   `json:"usage_policy"`
	RolesAllowed []string `json:"roles_allowed"`
	Enabled      bool     `json:"enabled"`
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

	cfg := cognitive.ProviderConfig{
		Type:         req.Type,
		Endpoint:     req.Endpoint,
		ModelID:      req.ModelID,
		AuthKey:      req.APIKey,
		Location:     req.Location,
		DataBoundary: req.DataBoundary,
		UsagePolicy:  req.UsagePolicy,
		RolesAllowed: req.RolesAllowed,
		Enabled:      req.Enabled,
	}
	if cfg.Location == "" {
		cfg.Location = "local"
	}
	if cfg.DataBoundary == "" {
		cfg.DataBoundary = "local_only"
	}
	if cfg.UsagePolicy == "" {
		cfg.UsagePolicy = "local_first"
	}

	if err := s.Cognitive.AddProvider(req.ID, cfg); err != nil {
		log.Printf("AddProvider %s failed: %v", req.ID, err)
		respondError(w, "Failed to add provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Probe immediately so the response includes a live status
	status := "offline"
	if adapter, ok := s.Cognitive.Adapters[req.ID]; ok && cfg.Enabled {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		alive, _ := adapter.Probe(ctx)
		cancel()
		if alive {
			status = "online"
		}
	} else if !cfg.Enabled {
		status = "disabled"
	}

	respondJSON(w, map[string]any{"ok": true, "data": BrainEntry{
		ID:           req.ID,
		Type:         cfg.Type,
		Endpoint:     cfg.Endpoint,
		ModelID:      cfg.ModelID,
		Location:     cfg.Location,
		DataBoundary: cfg.DataBoundary,
		UsagePolicy:  cfg.UsagePolicy,
		RolesAllowed: cfg.RolesAllowed,
		Enabled:      cfg.Enabled,
		Status:       status,
	}})
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

	cfg := cognitive.ProviderConfig{
		Type:         req.Type,
		Endpoint:     req.Endpoint,
		ModelID:      req.ModelID,
		AuthKey:      req.APIKey, // empty = keep existing (UpdateProvider handles this)
		Location:     req.Location,
		DataBoundary: req.DataBoundary,
		UsagePolicy:  req.UsagePolicy,
		RolesAllowed: req.RolesAllowed,
		Enabled:      req.Enabled,
	}

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
