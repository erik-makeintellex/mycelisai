package server

import (
	"context"
	"regexp"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

var validProviderID = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

// BrainEntry is the enriched provider info returned by GET /api/v1/brains.
type BrainEntry struct {
	ID                 string   `json:"id"`
	Type               string   `json:"type"`
	Endpoint           string   `json:"endpoint,omitempty"`
	ModelID            string   `json:"model_id"`
	Location           string   `json:"location"`
	DataBoundary       string   `json:"data_boundary"`
	UsagePolicy        string   `json:"usage_policy"`
	TokenBudgetProfile string   `json:"token_budget_profile"`
	MaxOutputTokens    int      `json:"max_output_tokens"`
	RolesAllowed       []string `json:"roles_allowed"`
	Enabled            bool     `json:"enabled"`
	Status             string   `json:"status"`
}

type brainUpsertRequest struct {
	ID                 string   `json:"id"`
	Type               string   `json:"type"`
	Endpoint           string   `json:"endpoint"`
	ModelID            string   `json:"model_id"`
	APIKey             string   `json:"api_key"`
	Location           string   `json:"location"`
	DataBoundary       string   `json:"data_boundary"`
	UsagePolicy        string   `json:"usage_policy"`
	TokenBudgetProfile string   `json:"token_budget_profile"`
	MaxOutputTokens    int      `json:"max_output_tokens"`
	RolesAllowed       []string `json:"roles_allowed"`
	Enabled            bool     `json:"enabled"`
}

func providerConfigFromBrainRequest(req brainUpsertRequest, applyDefaults bool) cognitive.ProviderConfig {
	cfg := cognitive.ProviderConfig{
		Type:               req.Type,
		Endpoint:           req.Endpoint,
		ModelID:            req.ModelID,
		AuthKey:            req.APIKey,
		Location:           req.Location,
		DataBoundary:       req.DataBoundary,
		UsagePolicy:        req.UsagePolicy,
		TokenBudgetProfile: req.TokenBudgetProfile,
		MaxOutputTokens:    req.MaxOutputTokens,
		RolesAllowed:       req.RolesAllowed,
		Enabled:            req.Enabled,
	}
	if applyDefaults {
		applyBrainProviderDefaults(&cfg)
	}
	return cognitive.NormalizeProviderTokenDefaults(cfg)
}

func applyBrainProviderDefaults(cfg *cognitive.ProviderConfig) {
	if cfg.Location == "" {
		cfg.Location = "local"
	}
	if cfg.DataBoundary == "" {
		cfg.DataBoundary = "local_only"
	}
	if cfg.UsagePolicy == "" {
		cfg.UsagePolicy = "local_first"
	}
}

func brainEntryFromProvider(id string, prov cognitive.ProviderConfig, status string) BrainEntry {
	applyBrainProviderDefaults(&prov)
	return BrainEntry{
		ID:                 id,
		Type:               prov.Type,
		Endpoint:           prov.Endpoint,
		ModelID:            prov.ModelID,
		Location:           prov.Location,
		DataBoundary:       prov.DataBoundary,
		UsagePolicy:        prov.UsagePolicy,
		TokenBudgetProfile: prov.TokenBudgetProfile,
		MaxOutputTokens:    prov.MaxOutputTokens,
		RolesAllowed:       prov.RolesAllowed,
		Enabled:            prov.Enabled,
		Status:             status,
	}
}

func (s *AdminServer) brainStatus(ctx context.Context, id string, enabled bool) string {
	if !enabled {
		return "disabled"
	}
	adapter, ok := s.Cognitive.Adapters[id]
	if !ok {
		return "offline"
	}
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	alive, _ := adapter.Probe(probeCtx)
	cancel()
	if alive {
		return "online"
	}
	return "offline"
}
