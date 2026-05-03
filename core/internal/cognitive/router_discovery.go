package cognitive

import (
	"context"
	"fmt"
)

// startupProbeProviderIDs returns the provider IDs that should be probed during startup.
// Policy:
//   - Always include default "ollama" when available.
//   - Include providers explicitly referenced by profiles.
//   - Do not probe unrelated declared adapters during startup.
func (r *Router) startupProbeProviderIDs() map[string]struct{} {
	include := make(map[string]struct{})

	if _, ok := r.Adapters["ollama"]; ok {
		include["ollama"] = struct{}{}
	}

	for _, providerID := range r.Config.Profiles {
		if providerID == "" {
			continue
		}
		if _, ok := r.Adapters[providerID]; ok {
			include[providerID] = struct{}{}
		}
	}

	return include
}

func (r *Router) autoConfigureWithProviders(ctx context.Context, providers map[string]LLMProvider) {
	sd := NewServiceDiscovery(providers)
	discoveryResults := sd.DiscoverAll(ctx)

	fmt.Println("--- Cognitive Discovery Report ---")
	for id, res := range discoveryResults {
		modelID := r.Config.Providers[id].ModelID
		tier := GradeModel(modelID)

		status := "✅ Online"
		if !res.Healthy {
			status = fmt.Sprintf("❌ Offline (%v)", res.Error)
		}
		fmt.Printf("[%s] Model: %s (Tier %s) -> %s\n", id, modelID, tier, status)
	}
	fmt.Println("----------------------------------")

	for profileName, providerID := range r.Config.Profiles {
		res, exists := discoveryResults[providerID]
		if !exists {
			continue
		}

		if !res.Healthy {
			fmt.Printf("⚠️ Profile '%s' provider '%s' is DOWN. Attempting fallback...\n", profileName, providerID)

			fallbackFound := false
			for fbID, fbRes := range discoveryResults {
				if fbRes.Healthy {
					fmt.Printf("🔄 Re-routing '%s' to '%s'\n", profileName, fbID)
					r.Config.Profiles[profileName] = fbID
					fallbackFound = true
					break
				}
			}
			if !fallbackFound {
				fmt.Printf("🔥 CRITICAL: No healthy providers found for '%s'.\n", profileName)
			}
		}
	}
}

// AutoConfigure probes providers and re-routes profiles if necessary.
func (r *Router) AutoConfigure(ctx context.Context) {
	r.autoConfigureWithProviders(ctx, r.Adapters)
}

// AutoConfigureStartup is a constrained startup calibration pass.
// It probes default Ollama plus any providers explicitly routed by profiles.
func (r *Router) AutoConfigureStartup(ctx context.Context) {
	include := r.startupProbeProviderIDs()
	if len(include) == 0 {
		return
	}

	scoped := make(map[string]LLMProvider)
	for id, adapter := range r.Adapters {
		if _, ok := include[id]; ok {
			scoped[id] = adapter
		}
	}
	if len(scoped) == 0 {
		return
	}

	r.autoConfigureWithProviders(ctx, scoped)
}
