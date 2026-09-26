package cognitive

import (
	"strings"
	"testing"
)

// --- Unit tests for validateRootProvider / applyRootProviderDefaults ---

func TestValidateRootProvider_UnsetIsNoOp(t *testing.T) {
	cfg := &BrainConfig{}
	if err := validateRootProvider(cfg); err != nil {
		t.Fatalf("expected unset root_provider to be a no-op, got %v", err)
	}
}

func TestValidateRootProvider_UnknownProviderFailsClosed(t *testing.T) {
	cfg := &BrainConfig{
		RootProvider: "vllm",
		Providers:    map[string]ProviderConfig{},
	}
	err := validateRootProvider(cfg)
	if err == nil {
		t.Fatal("expected error for root_provider naming an unconfigured provider")
	}
	if !strings.Contains(err.Error(), "vllm") {
		t.Fatalf("expected error to name the unresolved provider, got %v", err)
	}
}

func TestValidateRootProvider_DisabledProviderFailsClosed(t *testing.T) {
	cfg := &BrainConfig{
		RootProvider: "vllm",
		Providers: map[string]ProviderConfig{
			"vllm": {Type: "openai_compatible", Enabled: false},
		},
	}
	err := validateRootProvider(cfg)
	if err == nil {
		t.Fatal("expected error for root_provider naming a disabled provider")
	}
	if !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected error to mention disabled, got %v", err)
	}
}

func TestValidateRootProvider_EnabledProviderPasses(t *testing.T) {
	cfg := &BrainConfig{
		RootProvider: "vllm",
		Providers: map[string]ProviderConfig{
			"vllm": {Type: "openai_compatible", Enabled: true},
		},
	}
	if err := validateRootProvider(cfg); err != nil {
		t.Fatalf("expected enabled root provider to pass validation, got %v", err)
	}
}

func TestApplyRootProviderDefaults_UnsetLeavesProfilesUntouched(t *testing.T) {
	cfg := &BrainConfig{
		Profiles: map[string]string{"chat": "ollama"},
	}
	applyRootProviderDefaults(cfg)
	if len(cfg.Profiles) != 1 || cfg.Profiles["chat"] != "ollama" {
		t.Fatalf("expected profiles unchanged when root_provider is unset, got %v", cfg.Profiles)
	}
	for _, profile := range defaultExecutionProfiles {
		if profile == "chat" {
			continue
		}
		if _, ok := cfg.Profiles[profile]; ok {
			t.Fatalf("expected profile %q to stay unbound with root_provider unset", profile)
		}
	}
}

func TestApplyRootProviderDefaults_FillsUnboundProfilesOnly(t *testing.T) {
	cfg := &BrainConfig{
		RootProvider: "vllm",
		Profiles: map[string]string{
			"chat": "ollama", // explicit binding must survive
		},
	}
	applyRootProviderDefaults(cfg)

	if cfg.Profiles["chat"] != "ollama" {
		t.Fatalf("expected explicit profile binding to win over root_provider, got %q", cfg.Profiles["chat"])
	}
	for _, profile := range defaultExecutionProfiles {
		if profile == "chat" {
			continue
		}
		if got := cfg.Profiles[profile]; got != "vllm" {
			t.Fatalf("expected profile %q bound to root provider %q, got %q", profile, "vllm", got)
		}
	}
}

// --- NewRouter-level (end-to-end) tests ---

func TestNewRouter_RootProviderFillsUnboundProfiles(t *testing.T) {
	clearProviderRoutingEnv(t)

	configPath := writeTestCognitiveConfig(t, `
root_provider: vllm
providers:
  vllm:
    type: openai_compatible
    endpoint: http://host.docker.internal:8000/v1
    model_id: Qwen/Qwen2.5-Coder-14B-Instruct-AWQ
    enabled: true
    data_boundary: local_only
profiles:
  chat: vllm
`)

	router, err := NewRouter(configPath, nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	for _, profile := range defaultExecutionProfiles {
		if got := router.Config.Profiles[profile]; got != "vllm" {
			t.Fatalf("expected profile %q bound to root provider vllm, got %q", profile, got)
		}
	}
}

func TestNewRouter_ExplicitProfileOverrideWinsOverRootProvider(t *testing.T) {
	clearProviderRoutingEnv(t)

	configPath := writeTestCognitiveConfig(t, `
root_provider: vllm
providers:
  vllm:
    type: openai_compatible
    endpoint: http://host.docker.internal:8000/v1
    model_id: Qwen/Qwen2.5-Coder-14B-Instruct-AWQ
    enabled: true
    data_boundary: local_only
  ollama:
    type: ollama
    endpoint: http://127.0.0.1:11434/v1
    model_id: qwen2.5-coder:7b
    enabled: true
    data_boundary: local_only
profiles:
  chat: ollama
`)
	t.Setenv("MYCELIS_PROFILE_ADMIN_PROVIDER", "ollama")

	router, err := NewRouter(configPath, nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	if got := router.Config.Profiles["chat"]; got != "ollama" {
		t.Fatalf("expected YAML-explicit chat profile to win over root_provider, got %q", got)
	}
	if got := router.Config.Profiles["admin"]; got != "ollama" {
		t.Fatalf("expected env-explicit admin profile to win over root_provider, got %q", got)
	}
	if got := router.Config.Profiles["coder"]; got != "vllm" {
		t.Fatalf("expected unbound coder profile to fall back to root_provider, got %q", got)
	}
}

func TestNewRouter_UnknownRootProviderFailsClosed(t *testing.T) {
	clearProviderRoutingEnv(t)

	configPath := writeTestCognitiveConfig(t, `
root_provider: does-not-exist
providers:
  ollama:
    type: ollama
    endpoint: http://127.0.0.1:11434/v1
    model_id: qwen2.5-coder:7b
    enabled: true
profiles:
  chat: ollama
`)

	router, err := NewRouter(configPath, nil)
	if err == nil {
		t.Fatalf("expected NewRouter to reject an unknown root_provider, got router=%+v", router)
	}
	if router != nil {
		t.Fatalf("expected nil router on root_provider validation failure, got %+v", router)
	}
}

func TestNewRouter_DisabledRootProviderFailsClosed(t *testing.T) {
	clearProviderRoutingEnv(t)

	configPath := writeTestCognitiveConfig(t, `
root_provider: vllm
providers:
  vllm:
    type: openai_compatible
    endpoint: http://host.docker.internal:8000/v1
    model_id: Qwen/Qwen2.5-Coder-14B-Instruct-AWQ
    enabled: false
profiles: {}
`)

	router, err := NewRouter(configPath, nil)
	if err == nil {
		t.Fatalf("expected NewRouter to reject a disabled root_provider, got router=%+v", router)
	}
	if router != nil {
		t.Fatalf("expected nil router on disabled root_provider, got %+v", router)
	}
}

func TestNewRouter_UnsetRootProviderKeepsTodaysBehavior(t *testing.T) {
	clearProviderRoutingEnv(t)

	configPath := writeTestCognitiveConfig(t, `
providers:
  ollama:
    type: ollama
    endpoint: http://127.0.0.1:11434/v1
    model_id: qwen2.5-coder:7b
    enabled: true
profiles:
  chat: ollama
`)

	router, err := NewRouter(configPath, nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	if router.Config.RootProvider != "" {
		t.Fatalf("expected empty RootProvider, got %q", router.Config.RootProvider)
	}
	for _, profile := range defaultExecutionProfiles {
		if profile == "chat" {
			continue
		}
		if got := router.Config.Profiles[profile]; got != "" {
			t.Fatalf("expected profile %q to stay unbound without root_provider, got %q", profile, got)
		}
	}
}

func TestNewRouter_RootProviderEnvOverrideAppliesRootAlias(t *testing.T) {
	clearProviderRoutingEnv(t)
	t.Setenv("MYCELIS_ROOT_PROVIDER", "vllm")
	t.Setenv("MYCELIS_PROVIDER_VLLM_TYPE", "openai_compatible")
	t.Setenv("MYCELIS_PROVIDER_VLLM_ENABLED", "true")
	t.Setenv("MYCELIS_PROVIDER_VLLM_ENDPOINT", "http://host.docker.internal:8000/v1")
	t.Setenv("MYCELIS_PROVIDER_VLLM_MODEL_ID", "Qwen/Qwen2.5-Coder-14B-Instruct-AWQ")

	configPath := writeTestCognitiveConfig(t, `
providers: {}
profiles: {}
`)

	router, err := NewRouter(configPath, nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	if router.Config.RootProvider != "vllm" {
		t.Fatalf("expected MYCELIS_ROOT_PROVIDER to set RootProvider, got %q", router.Config.RootProvider)
	}
	for _, profile := range defaultExecutionProfiles {
		if got := router.Config.Profiles[profile]; got != "vllm" {
			t.Fatalf("expected profile %q bound to env root provider vllm, got %q", profile, got)
		}
	}
}
