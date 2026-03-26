package cognitive

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type startupProbeStub struct {
	probeCalls int
	healthy    bool
}

func (s *startupProbeStub) Infer(ctx context.Context, prompt string, opts InferOptions) (*InferResponse, error) {
	return &InferResponse{Text: "ok", Provider: "stub", ModelUsed: "stub"}, nil
}

func (s *startupProbeStub) Probe(ctx context.Context) (bool, error) {
	s.probeCalls++
	return s.healthy, nil
}

func TestStartupProbeProviderIDs_DefaultOllamaAndProfileRouted(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Profiles: map[string]string{
				"chat": "ollama",
			},
		},
		Adapters: map[string]LLMProvider{
			"ollama":   &startupProbeStub{healthy: true},
			"vllm":     &startupProbeStub{healthy: true},
			"lmstudio": &startupProbeStub{healthy: true},
		},
	}

	ids := r.startupProbeProviderIDs()
	if _, ok := ids["ollama"]; !ok {
		t.Fatal("expected ollama in startup probe set")
	}
	if _, ok := ids["vllm"]; ok {
		t.Fatal("did not expect vllm in startup probe set without profile routing")
	}
	if _, ok := ids["lmstudio"]; ok {
		t.Fatal("did not expect lmstudio in startup probe set without profile routing")
	}
}

func TestAutoConfigureStartup_ProbesOnlyScopedProviders(t *testing.T) {
	ollama := &startupProbeStub{healthy: true}
	vllm := &startupProbeStub{healthy: true}
	lmstudio := &startupProbeStub{healthy: true}

	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"ollama":   {ModelID: "qwen2.5-coder:7b"},
				"vllm":     {ModelID: "qwen2.5-coder"},
				"lmstudio": {ModelID: "default"},
			},
			Profiles: map[string]string{
				"chat": "ollama",
			},
		},
		Adapters: map[string]LLMProvider{
			"ollama":   ollama,
			"vllm":     vllm,
			"lmstudio": lmstudio,
		},
	}

	r.AutoConfigureStartup(context.Background())

	if ollama.probeCalls == 0 {
		t.Fatal("expected ollama to be probed during startup")
	}
	if vllm.probeCalls != 0 {
		t.Fatalf("expected vllm probeCalls=0, got %d", vllm.probeCalls)
	}
	if lmstudio.probeCalls != 0 {
		t.Fatalf("expected lmstudio probeCalls=0, got %d", lmstudio.probeCalls)
	}
}

func TestAutoConfigureStartup_ProbesAdditionalProfileProvider(t *testing.T) {
	ollama := &startupProbeStub{healthy: true}
	vllm := &startupProbeStub{healthy: true}

	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"ollama": {ModelID: "qwen2.5-coder:7b"},
				"vllm":   {ModelID: "qwen2.5-coder"},
			},
			Profiles: map[string]string{
				"chat":  "ollama",
				"coder": "vllm",
			},
		},
		Adapters: map[string]LLMProvider{
			"ollama": ollama,
			"vllm":   vllm,
		},
	}

	r.AutoConfigureStartup(context.Background())

	if ollama.probeCalls == 0 {
		t.Fatal("expected ollama to be probed")
	}
	if vllm.probeCalls == 0 {
		t.Fatal("expected vllm to be probed when explicitly profile-routed")
	}
}

func TestLoadFromDB_PreservesYAMLExecutionFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	providerRows := sqlmock.NewRows([]string{"id", "driver", "base_url", "api_key_env_var", "config"}).
		AddRow("ollama", "openai_compatible", "http://127.0.0.1:11434/v1", sql.NullString{String: "OLLAMA_API_KEY", Valid: true}, []byte(`{"model_id":"qwen2.5-coder:7b"}`))
	mock.ExpectQuery("SELECT id, driver, base_url, api_key_env_var, config FROM llm_providers").WillReturnRows(providerRows)
	profileRows := sqlmock.NewRows([]string{"key", "value"}).
		AddRow("role.chat", "ollama")
	mock.ExpectQuery("SELECT key, value FROM system_config WHERE key LIKE 'role\\.%'").WillReturnRows(profileRows)

	config := &BrainConfig{
		Providers: map[string]ProviderConfig{
			"ollama": {
				Enabled:      true,
				Location:     "local",
				DataBoundary: "local_only",
			},
		},
	}

	if err := loadFromDB(db, config); err != nil {
		t.Fatalf("loadFromDB: %v", err)
	}
	if !config.Providers["ollama"].Enabled {
		t.Fatal("expected enabled flag from YAML to survive DB overlay")
	}
	if config.Providers["ollama"].Location != "local" {
		t.Fatalf("location = %q, want local", config.Providers["ollama"].Location)
	}
	if config.Providers["ollama"].ModelID != "qwen2.5-coder:7b" {
		t.Fatalf("model_id = %q", config.Providers["ollama"].ModelID)
	}
	if got := config.Profiles["chat"]; got != "ollama" {
		t.Fatalf("chat profile = %q, want ollama", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
