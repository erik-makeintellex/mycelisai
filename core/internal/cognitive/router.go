package cognitive

import (
	"context"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Router manages model selection and inference via Adapters
type Router struct {
	Config   *BrainConfig
	Adapters map[string]LLMProvider
}

// NewRouter loads configuration and initializes adapters
func NewRouter(configPath string) (*Router, error) {
	// 1. Load Config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read brain config: %w", err)
	}

	var config BrainConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse brain config: %w", err)
	}

	// 2. Dynamic Overrides (Docker Support)
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		if !strings.HasPrefix(host, "http") {
			host = "http://" + host
		}
		// Patch all local_ollama endpoints
		for k, v := range config.Providers {
			if k == "local_ollama" || v.Type == "openai_compatible" {
				v.Endpoint = host + "/v1" // Standardize on /v1 for adapter
				config.Providers[k] = v
			}
		}
	}

	r := &Router{
		Config:   &config,
		Adapters: make(map[string]LLMProvider),
	}

	// 3. Initialize Adapters
	for id, pConfig := range config.Providers {
		var adapter LLMProvider
		var err error

		switch pConfig.Type {
		case "openai", "openai_compatible":
			adapter, err = NewOpenAIAdapter(pConfig)
		case "anthropic":
			adapter, err = NewAnthropicAdapter(pConfig)
		case "google":
			adapter, err = NewGoogleAdapter(pConfig)
		default:
			err = fmt.Errorf("unknown provider type: %s", pConfig.Type)
		}

		if err != nil {
			// Log but don't crash? For now, we allow partial failures except for critical ones.
			fmt.Printf("⚠️ Failed to init provider %s: %v\n", id, err)
			continue
		}
		r.Adapters[id] = adapter
	}

	// 4. Discovery & Grading
	r.AutoConfigure(context.Background())

	return r, nil
}

// AutoConfigure probes providers and re-routes profiles if necessary
func (r *Router) AutoConfigure(ctx context.Context) {
	sd := NewServiceDiscovery(r.Adapters)
	discoveryResults := sd.DiscoverAll(ctx)

	// Log Discovery Results
	fmt.Println("--- Cognitive Discovery Report ---")
	for id, res := range discoveryResults {
		// Grade the model based on Config ModelID (we don't have it in res yet, need to look up)
		modelID := r.Config.Providers[id].ModelID
		tier := GradeModel(modelID)

		status := "✅ Online"
		if !res.Healthy {
			status = fmt.Sprintf("❌ Offline (%v)", res.Error)
		}
		fmt.Printf("[%s] Model: %s (Tier %s) -> %s\n", id, modelID, tier, status)
	}
	fmt.Println("----------------------------------")

	// Auto-Config (Profiles)
	// For each profile, check if its provider is healthy. If not, fallback.
	for profileName, providerID := range r.Config.Profiles {
		res, exists := discoveryResults[providerID]
		if !exists {
			fmt.Printf("⚠️ Profile '%s' points to unknown provider '%s'\n", profileName, providerID)
			continue
		}

		if !res.Healthy {
			fmt.Printf("⚠️ Profile '%s' provider '%s' is DOWN. Attempting fallback...\n", profileName, providerID)

			// Simple Fallback: Find FIRST healthy provider
			// Improvement: Find healthy provider with matching Tier requirement (TODO)
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

// InferWithContract executes the request against the configured profile/provider
func (r *Router) InferWithContract(ctx context.Context, req InferRequest) (*InferResponse, error) {

	// 1. Resolve Profile -> ProviderID
	providerID, ok := r.Config.Profiles[req.Profile]
	if !ok {
		// Fallback to first available or error? use 'sentry' default?
		providerID = "sentry" // Safe default
		if _, exists := r.Config.Profiles["sentry"]; !exists {
			return nil, fmt.Errorf("profile '%s' not found and no sentry fallback", req.Profile)
		}
	}

	// 2. Get Adapter
	adapter, ok := r.Adapters[providerID]
	if !ok {
		return nil, fmt.Errorf("provider '%s' (for profile '%s') is not initialized", providerID, req.Profile)
	}

	// 3. Execute
	// Defaults for options
	opts := InferOptions{
		Temperature: 0.7, // TODO: Load from Profile config
		MaxTokens:   2048,
	}

	resp, err := adapter.Infer(ctx, req.Prompt, opts)
	if err != nil {
		// --- Runtime Self-Recovery ---
		// If inference fails, we should check if the provider is still healthy.
		// If dead, we trigger AutoConfigure and retry ONCE.
		fmt.Printf("⚠️ Inference failed on '%s': %v. Attempting Self-Recovery...\n", providerID, err)

		// 1. Probe specific provider to confirm death (avoid jitter)
		if healthy, _ := adapter.Probe(ctx); !healthy {
			fmt.Printf("❌ Provider '%s' confirmed DEAD. Re-calibrating Matrix...\n", providerID)

			// 2. Trigger Auto-Config (Heal)
			r.AutoConfigure(ctx)

			// 3. Retry on NEW provider
			newProviderID, ok := r.Config.Profiles[req.Profile]
			if !ok {
				return nil, fmt.Errorf("profile lost during recovery")
			}

			if newProviderID == providerID {
				return nil, fmt.Errorf("recovery failed: no alternative provider found (stuck on %s)", providerID)
			}

			newAdapter, ok := r.Adapters[newProviderID]
			if !ok {
				return nil, fmt.Errorf("recovery failed: new provider %s not init", newProviderID)
			}

			fmt.Printf("✅ Optimized to '%s'. Retrying request...\n", newProviderID)
			return newAdapter.Infer(ctx, req.Prompt, opts)
		}
		// If healthy (e.g. context timeout or API error), simple error return
		return nil, err
	}

	return resp, nil
}

// Deprecated: Infer is alias for InferWithContract
func (r *Router) Infer(req InferRequest) (*InferResponse, error) {
	return r.InferWithContract(context.Background(), req)
}
