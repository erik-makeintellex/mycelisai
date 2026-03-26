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

type executionProviderResolution struct {
	Profile         string
	ProviderID      string
	Provider        ProviderConfig
	Code            string
	Summary         string
	FallbackApplied bool
	Available       bool
}

func (r *Router) ExecutionAvailability(profile string, explicitProvider string) ExecutionAvailability {
	availability := ExecutionAvailability{
		Available:         false,
		Profile:           strings.TrimSpace(profile),
		RecommendedAction: "Open Settings and verify that at least one AI Engine is enabled and reachable for Soma.",
		SetupRequired:     true,
		SetupPath:         DefaultExecutionSetupPath,
	}

	resolution := r.resolveExecutionProvider(profile, explicitProvider)
	availability.Profile = resolution.Profile
	availability.ProviderID = resolution.ProviderID
	availability.ModelID = strings.TrimSpace(resolution.Provider.ModelID)
	availability.FallbackApplied = resolution.FallbackApplied
	availability.Available = resolution.Available
	availability.Code = resolution.Code
	availability.Summary = resolution.Summary
	if resolution.Available {
		availability.SetupRequired = false
		availability.RecommendedAction = ""
		availability.SetupPath = ""
		if resolution.FallbackApplied {
			availability.SetupRequired = true
			availability.RecommendedAction = "Review AI Engine Settings if you want a different default for Soma."
			availability.SetupPath = DefaultExecutionSetupPath
		}
	}
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

func (r *Router) resolveExecutionProvider(profile string, explicitProvider string) executionProviderResolution {
	resolution := executionProviderResolution{
		Profile: strings.TrimSpace(profile),
	}

	if r == nil || r.Config == nil {
		resolution.Code = ExecutionRouterUnavailable
		resolution.Summary = "Soma cannot run because the cognitive router is offline."
		return resolution
	}

	requestedProviderID := strings.TrimSpace(explicitProvider)
	if requestedProviderID == "" {
		if resolution.Profile == "" {
			resolution.Profile = DefaultExecutionProfileName
		}
		requestedProviderID = strings.TrimSpace(r.Config.Profiles[resolution.Profile])
		if requestedProviderID == "" {
			fallbackID := r.preferredFallbackProviderID()
			if fallbackID == "" {
				resolution.Code = ExecutionNoProviders
				resolution.Summary = "Soma does not have any available AI Engines configured for chat."
				return resolution
			}
			fallbackProvider := r.Config.Providers[fallbackID]
			resolution.ProviderID = fallbackID
			resolution.Provider = fallbackProvider
			resolution.Code = ExecutionAvailable
			resolution.Summary = "Soma needed a default AI Engine binding and will use the available local fallback."
			resolution.FallbackApplied = true
			resolution.Available = true
			return resolution
		}
	}

	provider, ok := r.Config.Providers[requestedProviderID]
	if ok && r.providerConfiguredForExecution(requestedProviderID) {
		resolution.ProviderID = requestedProviderID
		resolution.Provider = provider
		resolution.Code = ExecutionAvailable
		resolution.Summary = "Soma has an available cognitive engine."
		resolution.Available = true
		return resolution
	}
	if !ok && r.Adapters != nil && r.Adapters[requestedProviderID] != nil {
		resolution.ProviderID = requestedProviderID
		resolution.Code = ExecutionAvailable
		resolution.Summary = "Soma has an available cognitive engine."
		resolution.Available = true
		return resolution
	}

	resolution.ProviderID = requestedProviderID
	if !ok {
		resolution.Code = ExecutionProviderMissing
		resolution.Summary = "Soma is routed to an AI Engine provider that is not configured."
		return resolution
	}

	if fallbackID, fallbackProvider, applied := r.executionFallbackProvider(requestedProviderID); applied {
		resolution.ProviderID = fallbackID
		resolution.Provider = fallbackProvider
		resolution.Code = ExecutionAvailable
		resolution.Summary = "Soma will use the available fallback AI Engine because the configured default is not executable."
		resolution.FallbackApplied = true
		resolution.Available = true
		return resolution
	}

	resolution.Provider = provider
	switch {
	case !provider.Enabled:
		resolution.Code = ExecutionProviderDisabled
		resolution.Summary = "Soma is routed to an AI Engine that is configured but disabled."
	case strings.TrimSpace(provider.ModelID) == "":
		resolution.Code = ExecutionModelMissing
		resolution.Summary = "Soma is routed to an AI Engine without a model configured."
	default:
		resolution.Code = ExecutionProviderOffline
		resolution.Summary = "Soma is routed to an AI Engine that is not available at runtime."
	}
	return resolution
}

func (r *Router) executionFallbackProvider(excludeProviderID string) (string, ProviderConfig, bool) {
	fallbackID := r.preferredFallbackProviderID()
	if fallbackID == "" || fallbackID == strings.TrimSpace(excludeProviderID) {
		return "", ProviderConfig{}, false
	}
	fallbackProvider, ok := r.Config.Providers[fallbackID]
	if !ok {
		return "", ProviderConfig{}, false
	}
	return fallbackID, fallbackProvider, true
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
