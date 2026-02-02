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

	return r, nil
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
		Temperature: 0.7, // TODO: Load from Profile config (requires extending schema)
		MaxTokens:   2048,
	}

	return adapter.Infer(ctx, req.Prompt, opts)
}

// Deprecated: Infer is alias for InferWithContract
func (r *Router) Infer(req InferRequest) (*InferResponse, error) {
	return r.InferWithContract(context.Background(), req)
}
