package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

// Provider routes touch provider config and adapters only through the
// Router's locked accessors (ConfigSnapshot, ProviderSnapshot,
// StoreProviderConfig, AdapterSnapshot, AddProvider, UpdateProvider,
// RemoveProvider). Every mutating route, and the probe, requires root admin
// + cognitive:write and is audited first (routing_mutation_authority.go).

// GET /api/v1/brains — list all providers with enriched metadata + health status.
func (s *AdminServer) HandleListBrains(w http.ResponseWriter, r *http.Request) {
	cfg := s.Cognitive.ConfigSnapshot()
	if cfg == nil {
		respondJSON(w, map[string]any{"ok": true, "data": []BrainEntry{}})
		return
	}

	full := cognitiveFullView(r)
	entries := make([]any, 0, len(cfg.Providers))
	for id, prov := range cfg.Providers {
		prov = cognitive.NormalizeProviderTokenDefaults(prov)
		entry := brainEntryFromProvider(id, prov, s.brainStatus(r.Context(), id, prov.Enabled))
		if full {
			entries = append(entries, entry)
			continue
		}
		entries = append(entries, brainSummary{ID: entry.ID, Type: entry.Type, Enabled: entry.Enabled,
			Location: entry.Location, DataBoundary: entry.DataBoundary, Status: entry.Status})
	}

	respondJSON(w, map[string]any{"ok": true, "data": entries})
}

// brainSummary is the GET /api/v1/brains row for callers without the full
// cognitive view (S6e): no endpoint, model, or policy detail.
type brainSummary struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Enabled      bool   `json:"enabled"`
	Location     string `json:"location"`
	DataBoundary string `json:"data_boundary"`
	Status       string `json:"status"`
}

// PUT /api/v1/brains/{id}/toggle — enable or disable a provider.
func (s *AdminServer) HandleToggleBrain(w http.ResponseWriter, r *http.Request) {
	if !requireRoutingWriter(w, r) {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing provider ID", http.StatusBadRequest)
		return
	}
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondError(w, "Cognitive system offline", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	unlock := lockRoutingWrite()
	defer unlock()

	prov, ok := s.Cognitive.ProviderSnapshot(id)
	if !ok {
		respondError(w, "Provider not found", http.StatusNotFound)
		return
	}
	if prov.Enabled && !req.Enabled && s.rejectIfProviderBound(w, id) {
		return
	}
	if _, ok := s.auditRoutingMutation(w, r, auditProviderToggled, "Cognitive provider toggled", map[string]any{
		"provider_id": id, "enabled": req.Enabled, "previous_enabled": prov.Enabled,
	}); !ok {
		return
	}

	prov.Enabled = req.Enabled
	s.Cognitive.StoreProviderConfig(id, prov)

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
	if !requireRoutingWriter(w, r) {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing provider ID", http.StatusBadRequest)
		return
	}
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondError(w, "Cognitive system offline", http.StatusServiceUnavailable)
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

	unlock := lockRoutingWrite()
	defer unlock()

	prov, ok := s.Cognitive.ProviderSnapshot(id)
	if !ok {
		respondError(w, "Provider not found", http.StatusNotFound)
		return
	}
	if _, ok := s.auditRoutingMutation(w, r, auditProviderPolicyUpdated, "Cognitive provider policy updated", map[string]any{
		"provider_id": id, "usage_policy": req.UsagePolicy, "token_budget_profile": req.TokenBudgetProfile,
		"max_output_tokens": req.MaxOutputTokens, "roles_allowed": req.RolesAllowed,
	}); !ok {
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
	s.Cognitive.StoreProviderConfig(id, prov)

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
	if !requireRoutingWriter(w, r) {
		return
	}
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
	if req.APIKey != "" {
		respondError(w, "raw api_key values are not accepted; use api_key_env or a deployment secret reference", http.StatusBadRequest)
		return
	}

	unlock := lockRoutingWrite()
	defer unlock()

	if _, exists := s.Cognitive.ProviderSnapshot(req.ID); exists {
		respondError(w, "provider id already exists", http.StatusConflict)
		return
	}

	cfg := providerConfigFromBrainRequest(req, true)
	if _, ok := s.auditRoutingMutation(w, r, auditProviderAdded, "Cognitive provider added", brainAuditFields(req.ID, cfg)); !ok {
		return
	}

	if err := s.Cognitive.AddProvider(req.ID, cfg); err != nil {
		log.Printf("AddProvider %s failed: %v", req.ID, err)
		respondError(w, "Failed to add provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{"ok": true, "data": brainEntryFromProvider(req.ID, cfg, s.brainStatus(r.Context(), req.ID, cfg.Enabled))})
}

// PUT /api/v1/brains/{id} — update a provider's full configuration and hot-reload it.
func (s *AdminServer) HandleUpdateBrain(w http.ResponseWriter, r *http.Request) {
	if !requireRoutingWriter(w, r) {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing provider ID", http.StatusBadRequest)
		return
	}
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondError(w, "Cognitive system offline", http.StatusServiceUnavailable)
		return
	}

	var req brainUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if req.APIKey != "" {
		respondError(w, "raw api_key values are not accepted; use api_key_env or a deployment secret reference", http.StatusBadRequest)
		return
	}

	unlock := lockRoutingWrite()
	defer unlock()

	if _, exists := s.Cognitive.ProviderSnapshot(id); !exists {
		respondError(w, "Provider not found", http.StatusNotFound)
		return
	}
	cfg := providerConfigFromBrainRequest(req, false)
	// A full update that leaves a bound provider non-executable (disabled
	// or without a model) would strand the profiles routed to it.
	if !providerConfigExecutable(cfg) && s.rejectIfProviderBound(w, id) {
		return
	}
	if _, ok := s.auditRoutingMutation(w, r, auditProviderUpdated, "Cognitive provider updated", brainAuditFields(id, cfg)); !ok {
		return
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
	if !requireRoutingWriter(w, r) {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing provider ID", http.StatusBadRequest)
		return
	}
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		respondError(w, "Cognitive system offline", http.StatusServiceUnavailable)
		return
	}

	unlock := lockRoutingWrite()
	defer unlock()

	snapshot := s.Cognitive.ConfigSnapshot()
	if _, exists := snapshot.Providers[id]; !exists {
		respondError(w, "Provider not found", http.StatusNotFound)
		return
	}
	// Guard: refuse to delete the last remaining provider
	if len(snapshot.Providers) <= 1 {
		respondError(w, "Cannot delete the last provider — at least one must remain", http.StatusConflict)
		return
	}
	if s.rejectIfProviderBound(w, id) {
		return
	}
	if _, ok := s.auditRoutingMutation(w, r, auditProviderDeleted, "Cognitive provider deleted", map[string]any{"provider_id": id}); !ok {
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
// Gated and audited because it causes outbound egress to the provider.
func (s *AdminServer) HandleProbeBrain(w http.ResponseWriter, r *http.Request) {
	if !requireRoutingWriter(w, r) {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing provider ID", http.StatusBadRequest)
		return
	}
	if s.Cognitive == nil {
		respondError(w, "Cognitive system offline", http.StatusServiceUnavailable)
		return
	}

	adapter, ok := s.Cognitive.AdapterSnapshot(id)
	if !ok {
		respondError(w, "Provider not found or not initialized", http.StatusNotFound)
		return
	}
	if _, ok := s.auditRoutingMutation(w, r, auditProviderProbed, "Cognitive provider probed", map[string]any{"provider_id": id}); !ok {
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

// brainAuditFields describes a provider write for the audit record. It never
// carries a secret or a secret reference.
func brainAuditFields(id string, cfg cognitive.ProviderConfig) map[string]any {
	return map[string]any{
		"provider_id":   id,
		"type":          cfg.Type,
		"model_id":      cfg.ModelID,
		"location":      cfg.Location,
		"data_boundary": cfg.DataBoundary,
		"usage_policy":  cfg.UsagePolicy,
		"enabled":       cfg.Enabled,
	}
}
