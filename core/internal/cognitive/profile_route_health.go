package cognitive

import (
	"context"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

// Profile route health values, reported as profile_route_health by the
// cognitive status API.
const (
	ProfileRouteHealthOK       = "ok"
	ProfileRouteHealthDegraded = "degraded"
	ProfileRouteHealthFailed   = "failed"
)

// ProfileRouteProbeTimeout bounds each deduplicated provider probe.
const ProfileRouteProbeTimeout = 5 * time.Second

// ProfileRoute is one profile's effective route as Core would execute it.
// It carries no endpoint and no secret.
type ProfileRoute struct {
	ProviderID        string `json:"provider_id,omitempty"`
	ModelID           string `json:"model_id,omitempty"`
	Source            string `json:"source"`
	Available         bool   `json:"available"`
	Code              string `json:"code"`
	OverrideOrigin    string `json:"override_origin,omitempty"`
	DBRowPresent      bool   `json:"db_row_present,omitempty"`
	Reachable         *bool  `json:"reachable"`
	RecommendedAction string `json:"recommended_action,omitempty"`
	ExecutionProfile  bool   `json:"execution_profile"`

	adapter LLMProvider // set only when the provider is probeable
}

// ProfileRoutes resolves every execution profile plus every other bound
// profile under one read lock, without probing anything.
func (r *Router) ProfileRoutes() (map[string]ProfileRoute, bool) {
	out := make(map[string]ProfileRoute)
	if r == nil || r.Config == nil {
		return out, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for profile, binding := range r.Config.EffectiveProfileBindings() {
		resolution := r.resolveExecutionProviderLocked(profile, "")
		route := ProfileRoute{
			ProviderID:       binding.ProviderID,
			ModelID:          binding.ModelID,
			Source:           binding.Source,
			Available:        resolution.Available,
			Code:             resolution.Code,
			ExecutionProfile: IsExecutionProfile(profile),
		}
		if resolution.FallbackApplied {
			route.Source = ProfileSourceFallback
			route.ProviderID = resolution.ProviderID
			route.ModelID = strings.TrimSpace(resolution.Provider.ModelID)
		}
		state := r.profileOverrideStateLocked(profile)
		if state.Origin != "" && (route.Source == ProfileSourceOverride || state.Origin == ProfileOriginEnv) {
			route.OverrideOrigin = state.Origin
		}
		route.DBRowPresent = state.DBRowPresent
		if !resolution.Available {
			route.RecommendedAction = profileRecoveryHint(r.Config, profile, resolution.ProviderID, resolution.Code)
		} else if provider := r.Config.Providers[route.ProviderID]; probeableProvider(provider) {
			route.adapter = r.Adapters[route.ProviderID]
		}
		out[profile] = route
	}
	return out, r.Config.OverlayError
}

// probeableProvider limits route probes to local OpenAI-compatible engines;
// hosted providers and model gateways are never probed (no egress or cost).
func probeableProvider(provider ProviderConfig) bool {
	if provider.ModelGateway || strings.TrimSpace(provider.Endpoint) == "" {
		return false
	}
	providerType := strings.TrimSpace(provider.Type)
	if providerType == "" {
		providerType = strings.TrimSpace(provider.Driver)
	}
	return providerType == "openai_compatible" || providerType == "ollama"
}

// ProbeProfileRoutes probes each distinct probeable provider once, in
// parallel, with timeout per probe, and sets Reachable on every route bound
// to it. Routes that were not probed keep Reachable=nil. No lock is held
// while probing.
func ProbeProfileRoutes(ctx context.Context, routes map[string]ProfileRoute, timeout time.Duration) {
	adapters := make(map[string]LLMProvider)
	for _, route := range routes {
		if route.adapter != nil {
			adapters[route.ProviderID] = route.adapter
		}
	}
	results := make(map[string]bool, len(adapters))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for providerID, adapter := range adapters {
		wg.Add(1)
		go func(providerID string, adapter LLMProvider) {
			defer wg.Done()
			probeCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			alive, _ := adapter.Probe(probeCtx)
			mu.Lock()
			results[providerID] = alive
			mu.Unlock()
		}(providerID, adapter)
	}
	wg.Wait()
	for profile, route := range routes {
		if alive, ok := results[route.ProviderID]; ok && route.adapter != nil {
			reachable := alive
			route.Reachable = &reachable
			routes[profile] = route
		}
	}
}

// ProfileRouteHealth applies the severity table: an execution profile that
// is unavailable or unreachable fails; a fallback, an overlay error, or a
// DB row hidden under an env override degrades; anything wrong with a
// non-execution profile degrades.
func ProfileRouteHealth(routes map[string]ProfileRoute, overlayError bool) string {
	failed, degraded := false, overlayError
	for _, route := range routes {
		broken := !route.Available || (route.Reachable != nil && !*route.Reachable)
		switch {
		case broken && route.ExecutionProfile:
			failed = true
		case broken, route.Source == ProfileSourceFallback,
			route.DBRowPresent && route.OverrideOrigin == ProfileOriginEnv:
			degraded = true
		}
	}
	switch {
	case failed:
		return ProfileRouteHealthFailed
	case degraded:
		return ProfileRouteHealthDegraded
	default:
		return ProfileRouteHealthOK
	}
}

// warnMisroutedExecutionProfiles logs one WARN per execution profile that
// cannot run, naming the profile, provider, code, origin, and remedy. It
// never logs an endpoint or secret, and it never rebinds anything.
func (r *Router) warnMisroutedExecutionProfiles() {
	routes, _ := r.ProfileRoutes()
	names := make([]string, 0, len(routes))
	for profile := range routes {
		names = append(names, profile)
	}
	sort.Strings(names)
	for _, profile := range names {
		route := routes[profile]
		if !route.ExecutionProfile || route.Available {
			continue
		}
		origin := route.OverrideOrigin
		if origin == "" {
			origin = route.Source
		}
		log.Printf("WARN: cognitive profile %q routes to provider %q (%s, origin %s) and will fail closed per request. Remedy: %s",
			profile, route.ProviderID, route.Code, origin, route.RecommendedAction)
	}
}
