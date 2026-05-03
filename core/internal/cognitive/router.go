package cognitive

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

// Router manages model selection and inference via Adapters.
// Phase 5.2: Tracks cumulative token usage for telemetry reporting.
type Router struct {
	Config     *BrainConfig
	ConfigPath string // path to cognitive.yaml for persistence
	Adapters   map[string]LLMProvider

	// mu guards concurrent reads/writes to Config.Providers and Adapters
	// during hot-reload operations (AddProvider, UpdateProvider, RemoveProvider).
	mu sync.RWMutex

	// Token telemetry — sliding window for rate calculation
	totalTokens  atomic.Int64 // cumulative tokens processed
	windowStart  atomic.Int64 // unix nanoseconds of window start
	windowTokens atomic.Int64 // tokens in current window
}

// RecordTokens adds to the cumulative and windowed token counters.
func (r *Router) RecordTokens(n int) {
	r.totalTokens.Add(int64(n))
	r.windowTokens.Add(int64(n))
}

// TokenRate returns estimated tokens/second over the current window.
// Resets the window every 60 seconds for fresh measurements.
func (r *Router) TokenRate() float64 {
	now := time.Now().UnixNano()
	start := r.windowStart.Load()

	// Initialize window on first call
	if start == 0 {
		r.windowStart.Store(now)
		return 0
	}

	elapsed := float64(now-start) / float64(time.Second)
	if elapsed <= 0 {
		return 0
	}

	tokens := float64(r.windowTokens.Load())
	rate := tokens / elapsed

	// Reset window every 60 seconds
	if elapsed > 60 {
		r.windowTokens.Store(0)
		r.windowStart.Store(now)
	}

	return rate
}

// NewRouter loads configuration and initializes adapters
func NewRouter(configPath string, db *sql.DB) (*Router, error) {
	// 1. Load YAML Config (Base / Fallback)
	var config BrainConfig

	// Optional: If file exists, load it. If not, ignore (if we have DB)
	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("WARN: Failed to parse brain config: %v", err)
		}
	} else {
		log.Printf("INFO: No local brain config found at %s. Relying on DB/Defaults.", configPath)
		config.Providers = make(map[string]ProviderConfig)
		config.Profiles = make(map[string]string)
	}

	// 2. Load from DB (Overlay)
	if db != nil {
		if err := loadFromDB(db, &config); err != nil {
			log.Printf("ERROR: Failed to load Cognitive Registry from DB: %v", err)
		} else {
			log.Println("✅ Cognitive Registry Loaded from DB.")
		}
	}

	// 3. Deployment-friendly env overrides
	// These support automation tooling without reviving the retired
	// team/agent env-map routing path. Overrides apply at provider/profile/media
	// config surfaces and win over YAML/DB defaults.
	applyEnvOverrides(&config)
	for id, provider := range config.Providers {
		config.Providers[id] = NormalizeProviderTokenDefaults(provider)
	}

	r := &Router{
		Config:     &config,
		ConfigPath: configPath,
		Adapters:   make(map[string]LLMProvider),
	}

	// 4. Initialize Adapters
	for id, pConfig := range config.Providers {
		log.Printf("DEBUG: Initializing provider %s with endpoint %s", id, pConfig.Endpoint)
		var adapter LLMProvider
		var err error

		// Map SQL 'driver' to internal 'type' if needed, or unify.
		// Migration uses 'driver', Config uses 'type'. Let's fallback.
		inputType := pConfig.Type
		if inputType == "" {
			inputType = pConfig.Driver
		}

		switch inputType {
		case "openai", "openai_compatible":
			adapter, err = NewOpenAIAdapter(pConfig)
		case "anthropic":
			adapter, err = NewAnthropicAdapter(pConfig)
		case "google":
			adapter, err = NewGoogleAdapter(pConfig)
		case "ollama":
			// Ollama matches OpenAI Compatible in our adapter, but let's be explicit
			// Check if we have a dedicated Ollama adapter or reuse OpenAI
			// Reuse OpenAI for now as it supports /v1
			pConfig.Type = "openai_compatible" // Force type for adapter logic
			adapter, err = NewOpenAIAdapter(pConfig)
		default:
			err = fmt.Errorf("unknown provider type: %s", inputType)
		}

		if err != nil {
			// Log but don't crash? For now, we allow partial failures except for critical ones.
			fmt.Printf("⚠️ Failed to init provider %s: %v\n", id, err)
			continue
		}
		r.Adapters[id] = adapter
	}

	// 5. Degraded startup posture
	// Fail closed when no provider is configured instead of silently probing
	// desktop-local loopback addresses that do not exist in deployed runtimes.
	if len(r.Adapters) == 0 {
		log.Println("WARN: Zero cognitive adapters initialized after YAML/DB/env resolution. Cognitive Engine will operate in DEGRADED mode until an explicit provider endpoint is configured.")
	}

	if rebound := r.EnsureDefaultProfileBindings(); len(rebound) > 0 {
		log.Printf("INFO: rebound default cognitive profiles to fallback provider: %v", rebound)
	}

	// 6. Discovery & Grading (startup scope)
	// Only auto-configure if we have providers.
	// Startup intentionally probes only default Ollama and profile-routed providers,
	// so we don't try connecting to every declared backend unless explicitly configured.
	if len(r.Adapters) > 0 {
		r.AutoConfigureStartup(context.Background())
	}

	return r, nil
}

// InferWithContract executes the request against the configured profile/provider
func (r *Router) InferWithContract(ctx context.Context, req InferRequest) (*InferResponse, error) {
	resolution := r.resolveExecutionProvider(req.Profile, req.Provider)
	if !resolution.Available {
		return nil, fmt.Errorf("%s", resolution.Summary)
	}
	providerID := resolution.ProviderID

	// 2. Get Adapter
	adapter, ok := r.Adapters[providerID]
	if !ok {
		return nil, fmt.Errorf("provider '%s' is not initialized at runtime", providerID)
	}

	// 3. Execute
	// Defaults for options
	providerCfg := NormalizeProviderTokenDefaults(r.Config.Providers[providerID])
	opts := InferOptions{
		Temperature: 0.7, // TODO: Load from Profile config
		MaxTokens:   providerCfg.MaxOutputTokens,
		Messages:    req.Messages,
	}

	resp, err := adapter.Infer(ctx, req.Prompt, opts)
	if err == nil && resp != nil {
		// Record tokens for telemetry. If the adapter didn't report token
		// count, estimate at ~4 chars per token (conservative approximation).
		tokens := resp.TokensUsed
		if tokens == 0 && resp.Text != "" {
			tokens = len(resp.Text) / 4
			if tokens < 1 {
				tokens = 1
			}
		}
		r.RecordTokens(tokens)
	}
	if err != nil {
		// --- Runtime Self-Recovery ---
		// If inference fails, we should check if the provider is still healthy.
		// If dead, we trigger AutoConfigure and retry ONCE.
		fmt.Printf("⚠️ Inference failed on '%s': %v. Attempting Self-Recovery...\n", providerID, err)

		// 1. Probe specific provider to confirm death (avoid jitter)
		healthy, probeErr := adapter.Probe(ctx)
		if !healthy {
			fmt.Printf("❌ Provider '%s' confirmed DEAD (%v). Re-calibrating Matrix...\n", providerID, probeErr)

			// 2. Trigger Auto-Config (Heal)
			r.AutoConfigure(ctx)

			// 3. Retry on NEW provider
			recovered := r.resolveExecutionProvider(req.Profile, req.Provider)
			if !recovered.Available {
				return nil, fmt.Errorf("recovery failed: %s", recovered.Summary)
			}
			newProviderID := recovered.ProviderID

			if newProviderID == providerID {
				return nil, fmt.Errorf("recovery failed: no alternative provider found (stuck on %s, probe error: %v)", providerID, probeErr)
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

// Embed generates a text embedding vector using the first available EmbedProvider.
// Resolution order: "embed" profile → first Ollama-compatible adapter → first adapter.
func (r *Router) Embed(ctx context.Context, text string, model string) ([]float64, error) {
	if model == "" {
		model = DefaultEmbedModel
	}

	// 1. Try "embed" profile if configured
	if providerID, ok := r.Config.Profiles["embed"]; ok {
		if adapter, ok := r.Adapters[providerID]; ok {
			if ep, ok := adapter.(EmbedProvider); ok {
				return ep.Embed(ctx, text, model)
			}
		}
	}

	// 2. Try any adapter that implements EmbedProvider
	for _, adapter := range r.Adapters {
		if ep, ok := adapter.(EmbedProvider); ok {
			return ep.Embed(ctx, text, model)
		}
	}

	return nil, fmt.Errorf("no embedding provider available (need OpenAI-compatible adapter)")
}
