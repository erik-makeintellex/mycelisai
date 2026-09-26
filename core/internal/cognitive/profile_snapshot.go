package cognitive

import "strings"

// ProfileProviderSnapshot resolves profile's effective provider under the
// router read lock and returns a copy of that provider's config. The
// provider ID matches the ProviderID that ProfileRoutes reports for the same
// profile: the bound provider, or the same-boundary fallback when one was
// applied. A profile ProfileRoutes does not list (blank, or neither an
// execution profile nor a bound profile) returns ("", ProviderConfig{}, false).
// ok is true only when the resolved provider ID has a provider config; the
// provider ID is returned even when ok is false. Callers may modify the
// returned config freely; it shares no memory with the live routing table.
func (r *Router) ProfileProviderSnapshot(profile string) (providerID string, cfg ProviderConfig, ok bool) {
	if r == nil || r.Config == nil {
		return "", ProviderConfig{}, false
	}
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return "", ProviderConfig{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, bound := r.Config.Profiles[profile]; !bound && !IsExecutionProfile(profile) {
		return "", ProviderConfig{}, false
	}
	providerID = strings.TrimSpace(r.Config.Profiles[profile])
	if resolution := r.resolveExecutionProviderLocked(profile, ""); resolution.FallbackApplied {
		providerID = resolution.ProviderID
	}
	if providerID == "" {
		return "", ProviderConfig{}, false
	}
	cfg, ok = r.Config.Providers[providerID]
	if !ok {
		return providerID, ProviderConfig{}, false
	}
	return providerID, cloneProviderConfig(cfg), true
}

// AdapterSnapshot returns provider id's runtime adapter under the read lock.
// Adapters are replaced, never mutated, by Add/Update/RemoveProvider, so the
// returned value is safe to use after the lock is released.
func (r *Router) AdapterSnapshot(id string) (LLMProvider, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.Adapters[id]
	return adapter, ok && adapter != nil
}

// cloneProviderConfig copies cfg including its reference-typed fields.
func cloneProviderConfig(cfg ProviderConfig) ProviderConfig {
	if cfg.RolesAllowed != nil {
		cfg.RolesAllowed = append([]string(nil), cfg.RolesAllowed...)
	}
	return cfg
}
