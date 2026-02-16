package cognitive

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
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
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		if !strings.HasPrefix(host, "http") {
			host = "http://" + host
		}
		// Patch all ollama-compatible endpoints (any openai_compatible provider)
		for k, v := range config.Providers {
			if k == "ollama" || v.Type == "openai_compatible" || v.Driver == "ollama" {
				v.Endpoint = host + "/v1" // Standardize on /v1 for adapter
				config.Providers[k] = v
			}
		}
	}

	r := &Router{
		Config:     &config,
		ConfigPath: configPath,
		Adapters:   make(map[string]LLMProvider),
	}

	// 4. Initialize Adapters
	for id, pConfig := range config.Providers {
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

	// 5. Emergency Sovereign Fallback
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

	// 6. Discovery & Grading
	// Only auto-configure if we have providers
	if len(r.Adapters) > 0 {
		r.AutoConfigure(context.Background())
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

		// Parse JSON Config for ModelID etc
		var extra struct {
			ModelID string `json:"model_id"`
		}
		if len(configJSON) > 0 {
			_ = json.Unmarshal(configJSON, &extra)
		}

		pConfig := ProviderConfig{
			Type:       driver, // Mapping 'driver' to 'type'
			Driver:     driver, // Keep original
			Endpoint:   baseURL,
			ModelID:    extra.ModelID,
			AuthKeyEnv: envVar.String,
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

	// 1. Resolve Profile -> ProviderID (3-tier fallback)
	providerID, ok := r.Config.Profiles[req.Profile]
	if !ok {
		log.Printf("WARN: Profile '%s' not found. Attempting fallback chain...", req.Profile)

		// Tier 1: Try sentry (safe default)
		if sentryID, exists := r.Config.Profiles["sentry"]; exists {
			log.Printf("Fallback: '%s' -> sentry ('%s')", req.Profile, sentryID)
			providerID = sentryID
		} else if len(r.Adapters) > 0 {
			// Tier 2: First available adapter
			for k := range r.Adapters {
				log.Printf("Fallback: '%s' -> first available adapter '%s'", req.Profile, k)
				providerID = k
				break
			}
		} else {
			// Tier 3: No profiles, no adapters — total cognitive blackout
			return nil, fmt.Errorf("profile '%s' not found — no providers available (check config/cognitive.yaml and DB connectivity)", req.Profile)
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
			newProviderID, ok := r.Config.Profiles[req.Profile]
			if !ok {
				return nil, fmt.Errorf("profile lost during recovery")
			}

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
