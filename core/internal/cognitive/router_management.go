package cognitive

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// buildAdapter constructs an LLMProvider for the given provider config.
// Extracted from NewRouter init logic so it can be reused by hot-reload methods.
func (r *Router) buildAdapter(id string, cfg ProviderConfig) (LLMProvider, error) {
	inputType := cfg.Type
	if inputType == "" {
		inputType = cfg.Driver
	}
	switch inputType {
	case "openai", "openai_compatible":
		return NewOpenAIAdapter(cfg)
	case "anthropic":
		return NewAnthropicAdapter(cfg)
	case "google":
		return NewGoogleAdapter(cfg)
	case "ollama":
		cfg.Type = "openai_compatible"
		return NewOpenAIAdapter(cfg)
	default:
		return nil, fmt.Errorf("unknown provider type %q for provider %q", inputType, id)
	}
}

// AddProvider hot-injects a new provider without requiring a restart.
// Builds and probes the adapter, then persists the updated config.
func (r *Router) AddProvider(id string, cfg ProviderConfig) error {
	cfg = NormalizeProviderTokenDefaults(cfg)
	adapter, err := r.buildAdapter(id, cfg)
	if err != nil {
		return fmt.Errorf("build adapter: %w", err)
	}
	r.mu.Lock()
	if r.Config.Providers == nil {
		r.Config.Providers = make(map[string]ProviderConfig)
	}
	r.Config.Providers[id] = cfg
	r.Adapters[id] = adapter
	r.mu.Unlock()
	return r.SaveConfig()
}

// UpdateProvider replaces an existing adapter in-place without restart.
// If api_key is empty, the existing key is preserved.
func (r *Router) UpdateProvider(id string, cfg ProviderConfig) error {
	r.mu.RLock()
	existing, exists := r.Config.Providers[id]
	r.mu.RUnlock()
	if !exists {
		return fmt.Errorf("provider %q not found", id)
	}
	if cfg.AuthKey == "" {
		cfg.AuthKey = existing.AuthKey
	}
	if cfg.AuthKeyEnv == "" {
		cfg.AuthKeyEnv = existing.AuthKeyEnv
	}
	cfg = NormalizeProviderTokenDefaults(cfg)
	adapter, err := r.buildAdapter(id, cfg)
	if err != nil {
		return fmt.Errorf("build adapter: %w", err)
	}
	r.mu.Lock()
	r.Config.Providers[id] = cfg
	r.Adapters[id] = adapter
	r.mu.Unlock()
	return r.SaveConfig()
}

// RemoveProvider deletes a provider and its adapter, then persists.
func (r *Router) RemoveProvider(id string) error {
	r.mu.Lock()
	delete(r.Config.Providers, id)
	delete(r.Adapters, id)
	r.mu.Unlock()
	return r.SaveConfig()
}

// SaveConfig persists the current BrainConfig back to the YAML file.
// Only writes providers (without secrets) and profiles — safe for runtime updates.
func (r *Router) SaveConfig() error {
	if r.ConfigPath == "" {
		return fmt.Errorf("no config path set — cannot persist")
	}

	data, err := yaml.Marshal(r.redactedConfigForPersistence())
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(r.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	log.Printf("Cognitive config saved to %s", r.ConfigPath)
	return nil
}

// redactedConfigForPersistence returns a shallow copy of r.Config with every
// provider's raw AuthKey cleared. Secrets resolve at runtime only from
// AuthKeyEnv / deployment env overrides (MYCELIS_PROVIDER_<ID>_API_KEY) and
// must never round-trip into cognitive.yaml, a tracked file — including when
// an env override or a UI/API save populated AuthKey in memory.
func (r *Router) redactedConfigForPersistence() *BrainConfig {
	if r.Config == nil {
		return nil
	}
	out := &BrainConfig{
		Profiles:         r.Config.Profiles,
		ProfileFallbacks: r.Config.ProfileFallbacks,
		Media:            r.Config.Media,
	}
	if r.Config.Providers != nil {
		out.Providers = make(map[string]ProviderConfig, len(r.Config.Providers))
		for id, cfg := range r.Config.Providers {
			cfg.AuthKey = ""
			out.Providers[id] = cfg
		}
	}
	return out
}
