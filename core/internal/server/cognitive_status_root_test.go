package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

type statusProfile struct {
	ProviderID string `json:"provider_id"`
	ModelID    string `json:"model_id"`
	Source     string `json:"source"`
}

type statusRootResponse struct {
	RootProvider      *string                  `json:"root_provider"`
	RootProviderModel string                   `json:"root_provider_model"`
	Profiles          map[string]statusProfile `json:"profiles"`
}

func getCognitiveStatus(t *testing.T, s *AdminServer) statusRootResponse {
	t.Helper()
	rr := httptest.NewRecorder()
	http.HandlerFunc(s.HandleCognitiveStatus).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/cognitive/status", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	var resp statusRootResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

func rootStatusConfig() *cognitive.BrainConfig {
	cfg := &cognitive.BrainConfig{
		RootProvider: "vllm",
		Providers: map[string]cognitive.ProviderConfig{
			"vllm":   {Type: "openai_compatible", ModelID: "Qwen/Qwen2.5-Coder-14B-Instruct-AWQ", Enabled: false},
			"ollama": {Type: "ollama", ModelID: "qwen2.5-coder:7b", Enabled: false},
		},
		Profiles:       map[string]string{"chat": "vllm"},
		ProfileSources: map[string]string{"chat": cognitive.ProfileSourceRoot},
	}
	cfg.SetProfileOverride("coder", "ollama")
	return cfg
}

func TestHandleCognitiveStatus_ReportsRootAndProfileSources(t *testing.T) {
	s := &AdminServer{Cognitive: &cognitive.Router{Config: rootStatusConfig(), Adapters: map[string]cognitive.LLMProvider{}}}
	resp := getCognitiveStatus(t, s)

	if resp.RootProvider == nil || *resp.RootProvider != "vllm" {
		t.Fatalf("root_provider = %v, want vllm", resp.RootProvider)
	}
	if resp.RootProviderModel != "Qwen/Qwen2.5-Coder-14B-Instruct-AWQ" {
		t.Fatalf("root_provider_model = %q", resp.RootProviderModel)
	}
	if got := resp.Profiles["chat"]; got.ProviderID != "vllm" || got.Source != "root" || got.ModelID == "" {
		t.Fatalf("chat = %+v", got)
	}
	if got := resp.Profiles["coder"]; got.ProviderID != "ollama" || got.Source != "override" {
		t.Fatalf("coder = %+v", got)
	}
	if got := resp.Profiles["sentry"]; got.ProviderID != "" || got.Source != "unbound" {
		t.Fatalf("sentry = %+v", got)
	}
}

func TestHandleCognitiveStatus_OmitsRootWhenUnset(t *testing.T) {
	cfg := &cognitive.BrainConfig{
		Providers: map[string]cognitive.ProviderConfig{"ollama": {Type: "ollama", ModelID: "m"}},
		Profiles:  map[string]string{"chat": "ollama"},
	}
	resp := getCognitiveStatus(t, &AdminServer{Cognitive: &cognitive.Router{Config: cfg, Adapters: map[string]cognitive.LLMProvider{}}})
	if resp.RootProvider != nil {
		t.Fatalf("root_provider = %q, want absent", *resp.RootProvider)
	}
	if got := resp.Profiles["chat"]; got.Source != "default" || got.ProviderID != "ollama" {
		t.Fatalf("chat = %+v", got)
	}
}

func TestHandleUpdateProfiles_MarksOperatorOverride(t *testing.T) {
	cfg := rootStatusConfig()
	router := &cognitive.Router{Config: cfg, ConfigPath: filepath.Join(t.TempDir(), "cognitive.yaml"), Adapters: map[string]cognitive.LLMProvider{}}
	s := &AdminServer{Cognitive: router}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/cognitive/profiles", strings.NewReader(`{"profiles":{"chat":"ollama"}}`))
	http.HandlerFunc(s.HandleUpdateProfiles).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	if cfg.Profiles["chat"] != "ollama" || cfg.ProfileSource("chat") != cognitive.ProfileSourceOverride {
		t.Fatalf("chat = %q (%s), want ollama override", cfg.Profiles["chat"], cfg.ProfileSource("chat"))
	}
}
