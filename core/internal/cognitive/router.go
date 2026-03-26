package cognitive

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
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

	// 3. Dynamic Overrides (Docker Support)
	// OLLAMA_HOST can be a bind address (e.g. "0.0.0.0") used by Ollama itself
	// to listen on all interfaces, or a reachable endpoint (e.g. "192.168.50.156:11434").
	// Only patch provider endpoints when the value is a routable address.
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		log.Printf("DEBUG: Found OLLAMA_HOST env var: %s", host)
		stripped := strings.TrimPrefix(strings.TrimPrefix(host, "http://"), "https://")
		stripped = strings.Split(stripped, ":")[0] // extract just the host portion
		if stripped == "0.0.0.0" || stripped == "" {
			log.Println("DEBUG: OLLAMA_HOST is a bind address (0.0.0.0), skipping provider patching.")
		} else {
			if !strings.HasPrefix(host, "http") {
				host = "http://" + host
			}
			// Patch all ollama-compatible endpoints (any openai_compatible provider)
			for k, v := range config.Providers {
				if k == "ollama" || v.Type == "openai_compatible" || v.Driver == "ollama" {
					v.Endpoint = host + "/v1" // Standardize on /v1 for adapter
					config.Providers[k] = v
					log.Printf("DEBUG: Patched provider %s endpoint to %s", k, v.Endpoint)
				}
			}
		}
	} else {
		log.Println("DEBUG: No OLLAMA_HOST env var found.")
	}

	// 4. Deployment-friendly env overrides
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

	// 5. Initialize Adapters
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

	// 6. Emergency Sovereign Fallback
	// If zero adapters were initialized (YAML missing + DB down), attempt to
	// discover a local Ollama instance at well-known endpoints. This implements
	// the Universal Sovereignty principle: the organism must survive in isolation.
	if len(r.Adapters) == 0 {
		log.Println("WARN: Zero cognitive adapters initialized. Attempting emergency Ollama discovery...")
		emergencyEndpoints := []string{
			"http://localhost:11434/v1",
			"http://127.0.0.1:11434/v1",
		}
		// Also check OLLAMA_HOST env
		if host := os.Getenv("OLLAMA_HOST"); host != "" {
			if !strings.HasPrefix(host, "http") {
				host = "http://" + host
			}
			emergencyEndpoints = append([]string{host + "/v1"}, emergencyEndpoints...)
		}

		for _, ep := range emergencyEndpoints {
			emergencyConfig := ProviderConfig{
				Type:     "openai_compatible",
				Endpoint: ep,
				ModelID:  "qwen2.5-coder:7b",
				AuthKey:  "ollama",
			}
			adapter, err := NewOpenAIAdapter(emergencyConfig)
			if err != nil {
				continue
			}
			// Probe with short timeout
			probeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			healthy, _ := adapter.Probe(probeCtx)
			cancel()
			if healthy {
				log.Printf("Emergency Ollama discovered at %s", ep)
				r.Adapters["emergency-ollama"] = adapter
				r.Config.Providers["emergency-ollama"] = emergencyConfig
				// Wire all missing profiles to emergency adapter
				for _, profile := range []string{"architect", "sentry", "coder", "chat", "creative", "overseer"} {
					if _, exists := r.Config.Profiles[profile]; !exists {
						r.Config.Profiles[profile] = "emergency-ollama"
					}
				}
				break
			}
		}
		if len(r.Adapters) == 0 {
			log.Println("WARN: Emergency Ollama discovery failed. Cognitive Engine will operate in DEGRADED mode.")
		}
	}

	if rebound := r.EnsureDefaultProfileBindings(); len(rebound) > 0 {
		log.Printf("INFO: rebound default cognitive profiles to fallback provider: %v", rebound)
	}

	// 7. Discovery & Grading (startup scope)
	// Only auto-configure if we have providers.
	// Startup intentionally probes only default Ollama and profile-routed providers,
	// so we don't try connecting to every declared backend unless explicitly configured.
	if len(r.Adapters) > 0 {
		r.AutoConfigureStartup(context.Background())
	}

	return r, nil
}

func loadFromDB(db *sql.DB, config *BrainConfig) error {
	// A. Load Providers
	rows, err := db.Query("SELECT id, driver, base_url, api_key_env_var, config FROM llm_providers")
	if err != nil {
		return err
	}
	defer rows.Close()

	if config.Providers == nil {
		config.Providers = make(map[string]ProviderConfig)
	}

	for rows.Next() {
		var id, driver, baseURL string
		var envVar sql.NullString
		var configJSON []byte

		if err := rows.Scan(&id, &driver, &baseURL, &envVar, &configJSON); err != nil {
			log.Printf("WARN: Skipping bad provider row: %v", err)
			continue
		}

		pConfig := config.Providers[id]
		pConfig.Driver = driver
		if strings.TrimSpace(pConfig.Type) == "" {
			pConfig.Type = driver
		}
		if strings.TrimSpace(baseURL) != "" {
			pConfig.Endpoint = baseURL
		}
		if envVar.Valid && strings.TrimSpace(envVar.String) != "" {
			pConfig.AuthKeyEnv = envVar.String
		}
		if len(configJSON) > 0 {
			var extra struct {
				ModelID string `json:"model_id"`
			}
			if err := json.Unmarshal(configJSON, &extra); err == nil && strings.TrimSpace(extra.ModelID) != "" {
				pConfig.ModelID = extra.ModelID
			}
		}
		config.Providers[id] = pConfig
	}

	// B. Load Profiles (System Config)
	// Map "role.curator" -> profile "curator"
	rows2, err := db.Query("SELECT key, value FROM system_config WHERE key LIKE 'role.%'")
	if err != nil {
		return err
	}
	defer rows2.Close()

	if config.Profiles == nil {
		config.Profiles = make(map[string]string)
	}

	for rows2.Next() {
		var key, providerID string
		if err := rows2.Scan(&key, &providerID); err != nil {
			continue
		}
		// "role.architect" -> "architect"
		profileName := strings.TrimPrefix(key, "role.")
		config.Profiles[profileName] = providerID
	}

	return nil
}

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
	// For each profile, check if its provider is in the probe set and unhealthy.
	for profileName, providerID := range r.Config.Profiles {
		res, exists := discoveryResults[providerID]
		if !exists {
			// Provider wasn't part of this probe set; skip without treating as unknown.
			continue
		}

		if !res.Healthy {
			fmt.Printf("⚠️ Profile '%s' provider '%s' is DOWN. Attempting fallback...\n", profileName, providerID)

			// Simple Fallback: Find FIRST healthy provider in the probe set.
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

// AutoConfigure probes providers and re-routes profiles if necessary
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
	// Preserve auth key if caller omitted it (edit form leaves key blank)
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

	data, err := yaml.Marshal(r.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(r.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	log.Printf("Cognitive config saved to %s", r.ConfigPath)
	return nil
}
