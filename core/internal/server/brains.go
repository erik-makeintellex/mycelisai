package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

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
