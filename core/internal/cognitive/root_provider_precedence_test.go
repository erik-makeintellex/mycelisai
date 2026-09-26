package cognitive

import (
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// shippedStyleConfig mirrors the committed cognitive.yaml shape: every
// execution profile is bound explicitly to a local Ollama provider, and a
// vllm provider is configured and enabled.
const shippedStyleConfig = `
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
  admin: ollama
  architect: ollama
  chat: ollama
  coder: ollama
  creative: ollama
  overseer: ollama
  sentry: ollama
`

func assertBinding(t *testing.T, cfg *BrainConfig, profile, provider, source string) {
	t.Helper()
	if got := cfg.Profiles[profile]; got != provider {
		t.Fatalf("profile %q provider = %q, want %q", profile, got, provider)
	}
	if got := cfg.ProfileSource(profile); got != source {
		t.Fatalf("profile %q source = %q, want %q", profile, got, source)
	}
}

func TestPrecedence_RootOverridesYAMLDefaults(t *testing.T) {
	clearProviderRoutingEnv(t)
	t.Setenv("MYCELIS_ROOT_PROVIDER", "vllm")

	router, err := NewRouter(writeTestCognitiveConfig(t, shippedStyleConfig), nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	for _, profile := range defaultExecutionProfiles {
		assertBinding(t, router.Config, profile, "vllm", ProfileSourceRoot)
	}
	bindings := router.Config.EffectiveProfileBindings()
	if got := bindings["coder"]; got.ModelID != "Qwen/Qwen2.5-Coder-14B-Instruct-AWQ" || got.Source != ProfileSourceRoot {
		t.Fatalf("coder binding = %+v", got)
	}
}

func TestPrecedence_EnvOverrideBeatsRoot(t *testing.T) {
	clearProviderRoutingEnv(t)
	t.Setenv("MYCELIS_ROOT_PROVIDER", "vllm")
	t.Setenv("MYCELIS_PROFILE_CODER_PROVIDER", "ollama")

	router, err := NewRouter(writeTestCognitiveConfig(t, shippedStyleConfig), nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	assertBinding(t, router.Config, "coder", "ollama", ProfileSourceOverride)
	assertBinding(t, router.Config, "chat", "vllm", ProfileSourceRoot)
}

func TestPrecedence_EmptyEnvOverrideIsUnset(t *testing.T) {
	clearProviderRoutingEnv(t)
	t.Setenv("MYCELIS_ROOT_PROVIDER", "vllm")
	// Compose passes ${MYCELIS_PROFILE_CHAT_PROVIDER:-} as an empty string
	// when the operator sets nothing; that must not count as an override.
	t.Setenv("MYCELIS_PROFILE_CHAT_PROVIDER", "")
	t.Setenv("MYCELIS_PROFILE_ADMIN_PROVIDER", "   ")

	router, err := NewRouter(writeTestCognitiveConfig(t, shippedStyleConfig), nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	assertBinding(t, router.Config, "chat", "vllm", ProfileSourceRoot)
	assertBinding(t, router.Config, "admin", "vllm", ProfileSourceRoot)
}

func TestPrecedence_DBOverlayBeatsRoot(t *testing.T) {
	clearProviderRoutingEnv(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id, driver, base_url, api_key_env_var, config FROM llm_providers").
		WillReturnRows(sqlmock.NewRows([]string{"id", "driver", "base_url", "api_key_env_var", "config"}))
	mock.ExpectQuery("SELECT key, value FROM system_config WHERE key LIKE 'role\\.%'").
		WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
			AddRow("role.architect", "ollama").
			AddRow("role.sentry", ""))

	// Same order NewRouter uses: YAML defaults, DB overlay, env, root.
	cfg := &BrainConfig{
		RootProvider: "vllm",
		Providers: map[string]ProviderConfig{
			"vllm":   {Type: "openai_compatible", Enabled: true, ModelID: "m"},
			"ollama": {Type: "ollama", Enabled: true, ModelID: "o"},
		},
		Profiles: map[string]string{"architect": "ollama", "chat": "ollama"},
	}
	markLoadedProfilesAsDefault(cfg)
	if err := loadFromDB(db, cfg); err != nil {
		t.Fatalf("loadFromDB: %v", err)
	}
	applyEnvOverrides(cfg)
	if err := validateRootProvider(cfg); err != nil {
		t.Fatalf("validateRootProvider: %v", err)
	}
	applyRootProviderDefaults(cfg)

	assertBinding(t, cfg, "architect", "ollama", ProfileSourceOverride)
	assertBinding(t, cfg, "chat", "vllm", ProfileSourceRoot)
	// An empty DB row is not an override: root still applies.
	assertBinding(t, cfg, "sentry", "vllm", ProfileSourceRoot)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestPrecedence_UnsetRootKeepsShippedDefaults(t *testing.T) {
	clearProviderRoutingEnv(t)

	router, err := NewRouter(writeTestCognitiveConfig(t, shippedStyleConfig), nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	if router.Config.RootProvider != "" {
		t.Fatalf("RootProvider = %q, want empty", router.Config.RootProvider)
	}
	for _, profile := range defaultExecutionProfiles {
		assertBinding(t, router.Config, profile, "ollama", ProfileSourceDefault)
	}
}

func TestPrecedence_BadRootFailsClosedEvenWhenYAMLBindsEverything(t *testing.T) {
	for _, tc := range []struct{ name, root, want string }{
		{"unknown", "does-not-exist", "does not match"},
		{"disabled", "lmstudio", "disabled"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clearProviderRoutingEnv(t)
			t.Setenv("MYCELIS_ROOT_PROVIDER", tc.root)
			config := strings.Replace(shippedStyleConfig, "providers:\n", "providers:\n  lmstudio:\n    type: openai_compatible\n    endpoint: http://127.0.0.1:1234/v1\n    enabled: false\n", 1)
			router, err := NewRouter(writeTestCognitiveConfig(t, config), nil)
			if err == nil || router != nil {
				t.Fatalf("expected fail-closed for %s root, got router=%v err=%v", tc.name, router, err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

func TestPrecedence_RuntimeOverrideAndPersistenceKeepDefaults(t *testing.T) {
	clearProviderRoutingEnv(t)
	t.Setenv("MYCELIS_ROOT_PROVIDER", "vllm")
	path := writeTestCognitiveConfig(t, shippedStyleConfig)

	router, err := NewRouter(path, nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	router.Config.SetProfileOverride("creative", "ollama")
	assertBinding(t, router.Config, "creative", "ollama", ProfileSourceOverride)

	if err := router.SaveConfig(); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}
	// Root-derived bindings are persisted as their shipped defaults, so
	// unsetting MYCELIS_ROOT_PROVIDER later restores today's routing.
	if strings.Contains(string(saved), "chat: vllm") || !strings.Contains(string(saved), "chat: ollama") {
		t.Fatalf("root binding leaked into persisted profiles:\n%s", saved)
	}
	if strings.Contains(string(saved), "root_provider") {
		t.Fatalf("env root_provider leaked into persisted config:\n%s", saved)
	}
}
