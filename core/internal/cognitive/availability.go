package cognitive

import "strings"

var defaultExecutionProfiles = []string{
	"admin",
	"chat",
	"architect",
	"coder",
	"creative",
	"overseer",
	"sentry",
}

const (
	ExecutionAvailable          = "available"
	ExecutionNoProviders        = "no_provider_available"
	ExecutionProfileUnbound     = "profile_unbound"
	ExecutionProviderMissing    = "provider_missing"
	ExecutionProviderDisabled   = "provider_disabled"
	ExecutionProviderOffline    = "provider_uninitialized"
	ExecutionModelMissing       = "model_missing"
	ExecutionRouterUnavailable  = "router_unavailable"
	DefaultExecutionSetupPath   = "/settings"
	DefaultExecutionProfileName = "chat"
)

func (r *Router) profileAvailability(profile string) ExecutionAvailability {
	return r.ExecutionAvailability(profile, "")
}

func (r *Router) ExecutionAvailability(profile string, explicitProvider string) ExecutionAvailability {
	availability := ExecutionAvailability{
		Available:         false,
		Profile:           strings.TrimSpace(profile),
		RecommendedAction: "Open Settings and verify that at least one AI Engine is enabled and reachable for Soma.",
		SetupRequired:     true,
		SetupPath:         DefaultExecutionSetupPath,
	}

	if r == nil || r.Config == nil {
		availability.Code = ExecutionRouterUnavailable
		availability.Summary = "Soma cannot run because the cognitive router is offline."
		return availability
	}

	providerID := strings.TrimSpace(explicitProvider)
	if providerID == "" {
		if profile == "" {
			profile = DefaultExecutionProfileName
			availability.Profile = profile
		}
		providerID = strings.TrimSpace(r.Config.Profiles[profile])
		if providerID == "" {
			if r.preferredFallbackProviderID() == "" {
				availability.Code = ExecutionNoProviders
				availability.Summary = "Soma does not have any available AI Engines configured for chat."
				return availability
			}
			availability.Code = ExecutionProfileUnbound
			availability.Summary = "Soma does not have an AI Engine profile bound for chat."
			if fallbackID := r.preferredFallbackProviderID(); fallbackID != "" {
				provider := r.Config.Providers[fallbackID]
				availability.ProviderID = fallbackID
				availability.ModelID = provider.ModelID
				availability.Summary = "Soma needed a default AI Engine binding and will use the available local fallback."
				availability.RecommendedAction = "Review AI Engine Settings if you want a different default for Soma."
				availability.FallbackApplied = true
				availability.Available = true
				availability.Code = ExecutionAvailable
				return availability
			}
			return availability
		}
	}

	availability.ProviderID = providerID

	provider, ok := r.Config.Providers[providerID]
	if !ok {
		availability.Code = ExecutionProviderMissing
		availability.Summary = "Soma is routed to an AI Engine provider that is not configured."
		return availability
	}

	availability.ModelID = strings.TrimSpace(provider.ModelID)
	if !provider.Enabled {
		availability.Code = ExecutionProviderDisabled
		availability.Summary = "Soma is routed to an AI Engine that is configured but disabled."
		return availability
	}
	if availability.ModelID == "" {
		availability.Code = ExecutionModelMissing
		availability.Summary = "Soma is routed to an AI Engine without a model configured."
		return availability
	}
	if r.Adapters == nil || r.Adapters[providerID] == nil {
		availability.Code = ExecutionProviderOffline
		availability.Summary = "Soma is routed to an AI Engine that is not available at runtime."
		return availability
	}

	availability.Available = true
	availability.Code = ExecutionAvailable
	availability.Summary = "Soma has an available cognitive engine."
	availability.SetupRequired = false
	availability.RecommendedAction = ""
	availability.SetupPath = ""
	return availability
}

func (r *Router) EnsureDefaultProfileBindings() map[string]string {
	if r == nil || r.Config == nil {
		return nil
	}
	fallbackID := r.preferredFallbackProviderID()
	if fallbackID == "" {
		return nil
	}
	if r.Config.Profiles == nil {
		r.Config.Profiles = make(map[string]string)
	}

	rebound := make(map[string]string)
	for _, profile := range defaultExecutionProfiles {
		current := strings.TrimSpace(r.Config.Profiles[profile])
		if r.providerConfiguredForExecution(current) {
			continue
		}
		r.Config.Profiles[profile] = fallbackID
		rebound[profile] = fallbackID
	}
	if len(rebound) == 0 {
		return nil
	}
	return rebound
}

func (r *Router) providerConfiguredForExecution(providerID string) bool {
	if r == nil || r.Config == nil || providerID == "" {
		return false
	}
	provider, ok := r.Config.Providers[providerID]
	if !ok || !provider.Enabled || strings.TrimSpace(provider.ModelID) == "" {
		return false
	}
	if r.Adapters == nil {
		return false
	}
	return r.Adapters[providerID] != nil
}

func (r *Router) preferredFallbackProviderID() string {
	if r == nil || r.Config == nil || len(r.Config.Providers) == 0 {
		return ""
	}

	candidates := []string{
		"ollama",
		"emergency-ollama",
		"local-ollama-dev",
		"local-sovereign",
		"lmstudio",
	}
	for _, candidate := range candidates {
		if r.providerConfiguredForExecution(candidate) {
			return candidate
		}
	}

	for providerID, provider := range r.Config.Providers {
		if provider.Location == "remote" {
			continue
		}
		if r.providerConfiguredForExecution(providerID) {
			return providerID
		}
	}

	for providerID := range r.Config.Providers {
		if r.providerConfiguredForExecution(providerID) {
			return providerID
		}
	}

	return ""
}
