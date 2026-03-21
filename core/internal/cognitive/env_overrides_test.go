package cognitive

import "testing"

func TestApplyEnvOverrides_ExistingProviderAndProfile(t *testing.T) {
	t.Setenv("MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_MODEL_ID", "qwen3:8b")
	t.Setenv("MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT", "http://192.168.50.156:11434/v1")
	t.Setenv("MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENABLED", "true")
	t.Setenv("MYCELIS_PROFILE_CHAT_PROVIDER", "local_ollama_dev")

	cfg := &BrainConfig{
		Providers: map[string]ProviderConfig{
			"local-ollama-dev": {
				Type:     "ollama",
				Endpoint: "http://127.0.0.1:11434/v1",
				ModelID:  "qwen2.5-coder:7b",
				Enabled:  false,
			},
		},
		Profiles: map[string]string{
			"chat": "ollama",
		},
	}

	applyEnvOverrides(cfg)

	provider := cfg.Providers["local-ollama-dev"]
	if provider.ModelID != "qwen3:8b" {
		t.Fatalf("expected provider model override, got %q", provider.ModelID)
	}
	if provider.Endpoint != "http://192.168.50.156:11434/v1" {
		t.Fatalf("expected provider endpoint override, got %q", provider.Endpoint)
	}
	if !provider.Enabled {
		t.Fatal("expected provider enabled override to be true")
	}
	if cfg.Profiles["chat"] != "local-ollama-dev" {
		t.Fatalf("expected chat profile override to local-ollama-dev, got %q", cfg.Profiles["chat"])
	}
}

func TestApplyEnvOverrides_CreatesProviderAndMedia(t *testing.T) {
	t.Setenv("MYCELIS_PROVIDER_TEAMLEAD_OLLAMA_TYPE", "ollama")
	t.Setenv("MYCELIS_PROVIDER_TEAMLEAD_OLLAMA_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("MYCELIS_PROVIDER_TEAMLEAD_OLLAMA_MODEL_ID", "qwen3:8b")
	t.Setenv("MYCELIS_PROVIDER_TEAMLEAD_OLLAMA_ENABLED", "true")
	t.Setenv("MYCELIS_PROFILE_ADMIN_PROVIDER", "teamlead_ollama")
	t.Setenv("MYCELIS_MEDIA_ENDPOINT", "http://127.0.0.1:8001/v1")
	t.Setenv("MYCELIS_MEDIA_MODEL_ID", "flux-schnell")

	cfg := &BrainConfig{}
	applyEnvOverrides(cfg)

	provider, ok := cfg.Providers["teamlead-ollama"]
	if !ok {
		t.Fatal("expected env-defined provider to be created")
	}
	if provider.Type != "ollama" {
		t.Fatalf("expected provider type override, got %q", provider.Type)
	}
	if provider.ModelID != "qwen3:8b" {
		t.Fatalf("expected provider model override, got %q", provider.ModelID)
	}
	if !provider.Enabled {
		t.Fatal("expected provider to be enabled")
	}
	if cfg.Profiles["admin"] != "teamlead-ollama" {
		t.Fatalf("expected admin profile override, got %q", cfg.Profiles["admin"])
	}
	if cfg.Media == nil {
		t.Fatal("expected media config to be created")
	}
	if cfg.Media.Endpoint != "http://127.0.0.1:8001/v1" {
		t.Fatalf("expected media endpoint override, got %q", cfg.Media.Endpoint)
	}
	if cfg.Media.ModelID != "flux-schnell" {
		t.Fatalf("expected media model override, got %q", cfg.Media.ModelID)
	}
}

func TestApplyEnvOverrides_InvalidBoolIgnored(t *testing.T) {
	t.Setenv("MYCELIS_PROVIDER_OLLAMA_ENABLED", "definitely")

	cfg := &BrainConfig{
		Providers: map[string]ProviderConfig{
			"ollama": {Enabled: true},
		},
	}

	applyEnvOverrides(cfg)

	if !cfg.Providers["ollama"].Enabled {
		t.Fatal("expected invalid bool override to leave enabled state unchanged")
	}
}
