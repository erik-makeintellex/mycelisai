package cognitive

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

// ErrAIEngineUnavailable is the normalized, user-facing error Core returns
// when a bound AI Engine could not complete inference — for a direct
// provider failure or a model-gateway failure alike. The underlying
// adapter/transport error can carry endpoint URLs, dial detail, or other
// operational information; it is logged server-side only via log.Printf and
// is never included in this error's message or wrapped into it with %w.
var ErrAIEngineUnavailable = errors.New("AI engine unavailable")

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
		} else {
			// A literal api_key never loads into AuthKey (yaml:"-"), so a
			// non-empty one in the source file is otherwise silently
			// dropped. Warn once per provider (never the value) and flag
			// the provider so adapter init can explain why it still fails.
			for _, id := range detectLiteralProviderAPIKeys(data) {
				log.Printf("WARN: provider %q has a literal api_key in %s; it is not loaded (AuthKey is never read from a tracked file). Configure api_key_env with a secret reference instead.", id, configPath)
				if provider, ok := config.Providers[id]; ok {
					provider.LiteralAPIKeyIgnored = true
					config.Providers[id] = provider
				}
			}
		}
	} else {
		log.Printf("INFO: No local brain config found at %s. Relying on DB/Defaults.", configPath)
		config.Providers = make(map[string]ProviderConfig)
		config.Profiles = make(map[string]string)
	}

	// Tag the cognitive.yaml bindings as shipped defaults before any
	// override source runs; root_provider may replace defaults only.
	markLoadedProfilesAsDefault(&config)

	// 2. Load from DB (Overlay)
	if db != nil {
		if err := loadFromDB(db, &config); err != nil {
			// Startup proceeds, but never silently: status reports
			// overlay_error so operators know role.* overrides may be missing.
			config.OverlayError = true
			log.Printf("ERROR: Failed to load Cognitive Registry from DB (overlay_error; role.* overrides may be missing): %v", err)
		} else {
			log.Println("✅ Cognitive Registry Loaded from DB.")
		}
	}

	// 3. Deployment-friendly env overrides
	// These support automation tooling without reviving the retired
	// team/agent env-map routing path. Overrides apply at provider/profile/media
	// config surfaces and win over YAML/DB defaults.
	applyEnvOverrides(&config)
	recordEnvProfileOverrides(&config)
	for id, provider := range config.Providers {
		config.Providers[id] = NormalizeProviderTokenDefaults(provider)
	}

	// 3.1 Fail closed on an unusable root_provider/MYCELIS_ROOT_PROVIDER
	// before it is applied to any profile. An unset root is a no-op; a root
	// naming an unconfigured or disabled provider is a hard config error,
	// never a silent skip that leaves profiles quietly unbound.
	if err := validateRootProvider(&config); err != nil {
		return nil, fmt.Errorf("invalid cognitive config: %w", err)
	}

	// 3.2 Bind every default execution profile to the root provider unless
	// an operator override (DB overlay or MYCELIS_PROFILE_<NAME>_PROVIDER,
	// tagged above) pins it. Root replaces the cognitive.yaml defaults.
	applyRootProviderDefaults(&config)

	// 3.5 Reject a cross-data-boundary ProfileFallbacks entry before any
	// adapter is built. This is a config error, not a runtime posture: a
	// local_only profile must never even be allowed to declare a leaves_org
	// fallback, so Core refuses to start the cognitive engine on this
	// config rather than silently ignore the offending entry at request time.
	if err := validateProfileFallbackBoundaries(&config); err != nil {
		return nil, fmt.Errorf("invalid cognitive config: %w", err)
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
	r.warnMisroutedExecutionProfiles()

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
	r.mu.RLock()
	resolution := r.resolveExecutionProviderLocked(req.Profile, req.Provider)
	providerID := resolution.ProviderID
	adapter, ok := r.Adapters[providerID]
	providerCfg := NormalizeProviderTokenDefaults(r.Config.Providers[providerID])
	r.mu.RUnlock()
	if !resolution.Available {
		return nil, fmt.Errorf("%s", resolution.Summary)
	}

	// 2. Get Adapter
	if !ok {
		return nil, fmt.Errorf("provider '%s' is not initialized at runtime", providerID)
	}

	// 3. Execute
	// Defaults for options
	opts := InferOptions{
		Temperature: 0.7, // TODO: Load from Profile config
		MaxTokens:   providerCfg.MaxOutputTokens,
		Messages:    req.Messages,
		Correlation: req.Correlation,
	}

	resp, err := adapter.Infer(ctx, req.Prompt, opts)
	if err == nil && resp != nil {
		r.finalizeInferenceResponse(providerID, resp)
	}
	if err != nil {
		// Core fails closed on inference failure. It never re-routes a
		// request to a different provider mid-flight: cross-provider
		// substitution only ever happens before the call, through the
		// operator-configured, same-data-boundary ProfileFallbacks list
		// resolved in resolveExecutionProvider above.
		//
		// The adapter error (err/probeErr) can carry endpoint URLs, dial
		// detail, or other operational information, so it is logged
		// server-side only and never placed in the returned error — for a
		// direct provider and for a model gateway alike. A configured model
		// gateway is already the boundary Core selected, so it is never
		// probed (an unreachable gateway is reported the same way as any
		// other inference failure, without an extra health round-trip).
		if providerCfg.ModelGateway {
			log.Printf("WARN: inference failed on gateway provider %q: %v", providerID, err)
			return nil, fmt.Errorf("%w: provider %q", ErrAIEngineUnavailable, providerID)
		}
		healthy, probeErr := adapter.Probe(ctx)
		log.Printf("WARN: inference failed on provider %q: %v (probe healthy=%v probeErr=%v)", providerID, err, healthy, probeErr)
		return nil, fmt.Errorf("%w: provider %q", ErrAIEngineUnavailable, providerID)
	}

	return resp, nil
}

// finalizeInferenceResponse preserves the configured routing identity and only
// records usage reported by the provider. Missing usage remains unknown (zero)
// rather than being replaced with a text-length estimate.
func (r *Router) finalizeInferenceResponse(providerID string, resp *InferResponse) {
	resp.Provider = providerID
	if resp.TokensUsed > 0 {
		r.RecordTokens(resp.TokensUsed)
	}
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
	r.mu.RLock()
	providerID, ok := r.Config.Profiles["embed"]
	r.mu.RUnlock()
	if ok {
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
