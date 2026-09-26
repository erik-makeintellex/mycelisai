package cognitive

import (
	"fmt"
	"strings"
)

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

	if r == nil || r.Config == nil {
		availability.Code = ExecutionRouterUnavailable
		availability.Summary = "Soma cannot run because the cognitive router is offline."
		return availability
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	resolution := r.resolveExecutionProviderLocked(profile, explicitProvider)
	if !resolution.Available && strings.TrimSpace(explicitProvider) == "" {
		if hint := profileRecoveryHint(r.Config, resolution.Profile, resolution.ProviderID, resolution.Code); hint != "" {
			availability.RecommendedAction = hint
		}
	}
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

// EnsureDefaultProfileBindings binds profiles that have no executable
// provider at all. It never picks an arbitrary available provider: a profile
// is only bound when the operator has configured an explicit
// ProfileFallbacks entry for it. Profiles with no explicit configuration stay
// unbound and resolve through the normal unavailable/blocker path.
func (r *Router) EnsureDefaultProfileBindings() map[string]string {
	if r == nil || r.Config == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Config.Profiles == nil {
		r.Config.Profiles = make(map[string]string)
	}

	rebound := make(map[string]string)
	for _, profile := range defaultExecutionProfiles {
		if fallbackID, ok := r.applyExplicitFallbackLocked(profile); ok {
			rebound[profile] = fallbackID
		}
	}
	if len(rebound) == 0 {
		return nil
	}
	return rebound
}

// applyExplicitFallbackLocked rebinds one profile whose bound provider is
// not executable to its first explicit same-boundary ProfileFallbacks
// candidate. The caller holds r.mu for writing.
func (r *Router) applyExplicitFallbackLocked(profile string) (string, bool) {
	current := strings.TrimSpace(r.Config.Profiles[profile])
	if r.providerConfiguredForExecution(current) {
		return "", false
	}
	fallbackID, _, ok := r.executionFallbackCandidate(profile, current, r.primaryBoundaryForProfile(current))
	if !ok {
		return "", false
	}
	r.Config.Profiles[profile] = fallbackID
	markProfileSource(r.Config, profile, ProfileSourceFallback)
	return fallbackID, true
}

// profileRecoveryHint names the misrouted profile and its remedies for a
// profile-routed request whose provider cannot run. It never names an
// endpoint or secret.
func profileRecoveryHint(config *BrainConfig, profile, providerID, code string) string {
	switch code {
	case ExecutionProviderDisabled, ExecutionProviderMissing, ExecutionModelMissing, ExecutionProviderOffline:
	default:
		return ""
	}
	origin := config.ProfileOverrideOrigins[profile]
	switch {
	case origin == ProfileOriginEnv:
		return fmt.Sprintf("Profile %s is pinned to provider %s by MYCELIS_PROFILE_%s_PROVIDER: remove or change it in .env.compose and recreate Core, or enable provider %s.", profile, providerID, strings.ToUpper(profile), providerID)
	case origin != "" || config.ProfileSource(profile) == ProfileSourceOverride:
		return fmt.Sprintf("Reset the %s override to root (DELETE /api/v1/cognitive/profiles/%s/override) or enable provider %s.", profile, profile, providerID)
	default:
		return fmt.Sprintf("Enable provider %s for profile %s, or point root_provider at an enabled provider.", providerID, profile)
	}
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
	if r == nil || r.Config == nil {
		return r.resolveExecutionProviderLocked(profile, explicitProvider)
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resolveExecutionProviderLocked(profile, explicitProvider)
}

// resolveExecutionProviderLocked is resolveExecutionProvider for a caller
// that already holds r.mu (read or write).
func (r *Router) resolveExecutionProviderLocked(profile string, explicitProvider string) executionProviderResolution {
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
			// No primary provider at all: fail closed to the local_only
			// boundary (see primaryBoundaryForProfile) rather than allowing
			// any fallback candidate regardless of boundary.
			fallbackID, fallbackProvider, ok := r.executionFallbackCandidate(resolution.Profile, "", r.primaryBoundaryForProfile(""))
			if !ok {
				resolution.Code = ExecutionNoProviders
				resolution.Summary = "Soma does not have any available AI Engines configured for chat."
				return resolution
			}
			resolution.ProviderID = fallbackID
			resolution.Provider = fallbackProvider
			resolution.Code = ExecutionAvailable
			resolution.Summary = "Soma used the operator-configured fallback AI Engine because the profile has no primary binding."
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

	if fallbackID, fallbackProvider, applied := r.executionFallbackCandidate(resolution.Profile, requestedProviderID, normalizedDataBoundary(provider.DataBoundary)); applied {
		resolution.ProviderID = fallbackID
		resolution.Provider = fallbackProvider
		resolution.Code = ExecutionAvailable
		resolution.Summary = "Soma used the operator-configured same-boundary fallback AI Engine because the configured default is not executable."
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

// primaryBoundaryForProfile resolves the normalized data boundary that a
// fallback candidate must match for profile's currently configured provider
// (providerID). When providerID is blank, or does not resolve to a
// configured provider, there is no primary boundary to compare against, so
// this fails closed to DataBoundaryLocalOnly rather than treating "no
// primary" as "boundary check does not apply". An empty/unknown DataBoundary
// on a resolved provider is likewise normalized to DataBoundaryLocalOnly —
// see normalizedDataBoundary.
func (r *Router) primaryBoundaryForProfile(providerID string) string {
	id := strings.TrimSpace(providerID)
	if id == "" || r == nil || r.Config == nil {
		return DataBoundaryLocalOnly
	}
	provider, ok := r.Config.Providers[id]
	if !ok {
		return DataBoundaryLocalOnly
	}
	return normalizedDataBoundary(provider.DataBoundary)
}

// executionFallbackCandidate returns the first configured, executable
// fallback provider for profile from the operator's explicit
// ProfileFallbacks list. It never substitutes a provider that was not
// explicitly listed for this profile, and it always enforces the data
// boundary: only a candidate whose normalized DataBoundary equals the
// caller-supplied primaryBoundary (also normalized here defensively) is
// eligible — local_only never reaches a leaves_org provider, including when
// either side's DataBoundary is empty/unset (empty is never treated as a
// wildcard match). An empty/omitted ProfileFallbacks list (the default)
// returns false: no fallback exists and the caller must fail closed.
// Cross-boundary entries are rejected earlier, at startup, by
// validateProfileFallbackBoundaries — this function's boundary check is a
// second, independent guard against ever acting on one.
func (r *Router) executionFallbackCandidate(profile string, excludeProviderID string, primaryBoundary string) (string, ProviderConfig, bool) {
	if r == nil || r.Config == nil || len(r.Config.ProfileFallbacks) == 0 {
		return "", ProviderConfig{}, false
	}
	exclude := strings.TrimSpace(excludeProviderID)
	requiredBoundary := normalizedDataBoundary(primaryBoundary)
	for _, candidateID := range r.Config.ProfileFallbacks[strings.TrimSpace(profile)] {
		candidateID = strings.TrimSpace(candidateID)
		if candidateID == "" || candidateID == exclude {
			continue
		}
		candidate, ok := r.Config.Providers[candidateID]
		if !ok {
			continue
		}
		if normalizedDataBoundary(candidate.DataBoundary) != requiredBoundary {
			continue
		}
		if !r.providerConfiguredForExecution(candidateID) {
			continue
		}
		return candidateID, candidate, true
	}
	return "", ProviderConfig{}, false
}

// validateProfileFallbackBoundaries rejects a ProfileFallbacks configuration
// where any listed candidate's normalized data boundary differs from the
// profile's primary configured provider's normalized data boundary (empty
// treated as DataBoundaryLocalOnly on both sides — see
// normalizedDataBoundary). This is a hard config error surfaced at startup
// (NewRouter returns it), not a silent skip: an operator must fix a
// cross-boundary fallback entry rather than have Core quietly ignore it.
// A candidate ID with no matching entry in Providers is left for
// executionFallbackCandidate to skip at request time; there is no boundary
// to compare, so it is not a boundary violation here.
func validateProfileFallbackBoundaries(config *BrainConfig) error {
	if config == nil || len(config.ProfileFallbacks) == 0 {
		return nil
	}
	for profile, candidates := range config.ProfileFallbacks {
		primaryID := strings.TrimSpace(config.Profiles[profile])
		primaryBoundary := DataBoundaryLocalOnly
		if primaryID != "" {
			if primaryProvider, ok := config.Providers[primaryID]; ok {
				primaryBoundary = normalizedDataBoundary(primaryProvider.DataBoundary)
			}
		}
		for _, rawCandidateID := range candidates {
			candidateID := strings.TrimSpace(rawCandidateID)
			if candidateID == "" {
				continue
			}
			candidateProvider, ok := config.Providers[candidateID]
			if !ok {
				continue
			}
			candidateBoundary := normalizedDataBoundary(candidateProvider.DataBoundary)
			if candidateBoundary != primaryBoundary {
				return fmt.Errorf(
					"profile_fallbacks[%q] lists provider %q with data_boundary %q, but profile %q's primary provider %q has data_boundary %q; every fallback entry must share the primary provider's data boundary (local_only must never list a leaves_org provider)",
					profile, candidateID, candidateBoundary, profile, primaryID, primaryBoundary,
				)
			}
		}
	}
	return nil
}
