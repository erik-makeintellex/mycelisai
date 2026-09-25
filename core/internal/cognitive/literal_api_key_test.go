package cognitive

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

// --- Security QA follow-up: a literal api_key in a tracked config source
// --- must never be silently dropped without explanation.

func TestDetectLiteralProviderAPIKeys(t *testing.T) {
	data := []byte(`
providers:
  has-literal:
    type: anthropic
    api_key: sk-should-never-be-logged
  no-key:
    type: openai_compatible
  empty-key:
    type: openai_compatible
    api_key: ""
`)
	found := detectLiteralProviderAPIKeys(data)
	if len(found) != 1 || found[0] != "has-literal" {
		t.Fatalf("detectLiteralProviderAPIKeys = %v, want [has-literal]", found)
	}
}

func TestDetectLiteralProviderAPIKeys_EmptyProducesNoDetection(t *testing.T) {
	data := []byte(`
providers:
  ollama:
    type: openai_compatible
    api_key: ""
`)
	if found := detectLiteralProviderAPIKeys(data); len(found) != 0 {
		t.Fatalf("expected no detected providers for empty api_key, got %v", found)
	}
}

// TestNewRouter_WarnsOnLiteralAPIKeyWithoutProviderValue proves: (1) a WARN
// is logged naming the provider id, (2) the secret value is never logged,
// and (3) the provider is flagged so adapter init can explain the failure.
func TestNewRouter_WarnsOnLiteralAPIKeyWithoutProviderValue(t *testing.T) {
	clearProviderRoutingEnv(t)

	secret := "sk-super-secret-should-never-be-logged"
	configPath := writeTestCognitiveConfig(t, `
providers:
  production_claude:
    type: anthropic
    model_id: claude-3-opus
    api_key: `+secret+`
    enabled: true
    data_boundary: leaves_org
profiles:
  chat: production_claude
`)

	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })

	router, err := NewRouter(configPath, nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	logged := output.String()
	if !strings.Contains(logged, `provider "production_claude"`) {
		t.Fatalf("expected WARN naming the provider id, got log: %q", logged)
	}
	if strings.Contains(logged, secret) {
		t.Fatalf("literal api_key value leaked into logs: %q", logged)
	}

	provider, ok := router.Config.Providers["production_claude"]
	if !ok {
		t.Fatal("expected production_claude provider to be loaded")
	}
	if !provider.LiteralAPIKeyIgnored {
		t.Fatal("expected LiteralAPIKeyIgnored to be set for a provider with a literal api_key")
	}
}

// TestNewRouter_EmptyAPIKeyProducesNoWarning proves the negative: an empty
// (or absent) api_key must not trigger the literal-key warning/flag.
func TestNewRouter_EmptyAPIKeyProducesNoWarning(t *testing.T) {
	clearProviderRoutingEnv(t)

	configPath := writeTestCognitiveConfig(t, `
providers:
  ollama:
    type: openai_compatible
    endpoint: http://127.0.0.1:11434/v1
    model_id: qwen2.5-coder:7b
    api_key: ""
    enabled: true
profiles:
  chat: ollama
`)

	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })

	router, err := NewRouter(configPath, nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	if strings.Contains(output.String(), "literal api_key") {
		t.Fatalf("expected no literal-api_key warning for an empty key, got log: %q", output.String())
	}
	if router.Config.Providers["ollama"].LiteralAPIKeyIgnored {
		t.Fatal("expected LiteralAPIKeyIgnored to remain false for an empty api_key")
	}
}

// TestNewRouter_LiteralAPIKeyPathIsNotFatal proves startup does not fail for
// this condition: the affected provider fails closed on its own (no
// adapter), while unrelated providers still load normally.
func TestNewRouter_LiteralAPIKeyPathIsNotFatal(t *testing.T) {
	clearProviderRoutingEnv(t)

	configPath := writeTestCognitiveConfig(t, `
providers:
  production_claude:
    type: anthropic
    model_id: claude-3-opus
    api_key: sk-ignored
    enabled: true
    data_boundary: leaves_org
  ollama:
    type: openai_compatible
    endpoint: http://127.0.0.1:11434/v1
    model_id: qwen2.5-coder:7b
    enabled: true
profiles:
  chat: ollama
`)

	router, err := NewRouter(configPath, nil)
	if err != nil {
		t.Fatalf("NewRouter must not fail startup for a literal api_key: %v", err)
	}
	if _, ok := router.Adapters["production_claude"]; ok {
		t.Fatal("expected production_claude adapter to fail closed (no usable key), not initialize")
	}
	if _, ok := router.Adapters["ollama"]; !ok {
		t.Fatal("expected ollama to still initialize despite the unrelated literal-key provider")
	}
}

func TestNewAnthropicAdapter_LiteralAPIKeyIgnoredReturnsGuidance(t *testing.T) {
	_, err := NewAnthropicAdapter(ProviderConfig{
		Type:                 "anthropic",
		ModelID:              "claude-3-opus",
		LiteralAPIKeyIgnored: true,
	})
	if err == nil {
		t.Fatal("expected an error for a provider with no usable key")
	}
	if !strings.Contains(err.Error(), LiteralAPIKeyGuidanceError) {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), LiteralAPIKeyGuidanceError)
	}
}

func TestNewGoogleAdapter_LiteralAPIKeyIgnoredReturnsGuidance(t *testing.T) {
	_, err := NewGoogleAdapter(ProviderConfig{
		Type:                 "google",
		ModelID:              "gemini-1.5-pro",
		LiteralAPIKeyIgnored: true,
	})
	if err == nil {
		t.Fatal("expected an error for a provider with no usable key")
	}
	if !strings.Contains(err.Error(), LiteralAPIKeyGuidanceError) {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), LiteralAPIKeyGuidanceError)
	}
}

func TestNewOpenAIAdapter_ModelGatewayLiteralAPIKeyIgnoredReturnsGuidance(t *testing.T) {
	// A model gateway is never eligible for the local "dummy" key fallback,
	// so a literal-key-ignored gateway with no api_key_env must surface the
	// explicit guidance rather than the generic "missing api key".
	_, err := NewOpenAIAdapter(ProviderConfig{
		Type:                 "openai_compatible",
		ModelID:              "mycelis-default",
		ModelGateway:         true,
		DataBoundary:         "leaves_org",
		LiteralAPIKeyIgnored: true,
	})
	if err == nil {
		t.Fatal("expected an error for a gateway provider with no usable key")
	}
	if !strings.Contains(err.Error(), LiteralAPIKeyGuidanceError) {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), LiteralAPIKeyGuidanceError)
	}
}

// TestNewAnthropicAdapter_APIKeyEnvStillWorksWhenLiteralWasIgnored proves the
// positive case: api_key_env remains a fully supported path even for a
// provider whose source config also had a (now-ignored) literal api_key.
func TestNewAnthropicAdapter_APIKeyEnvStillWorksWhenLiteralWasIgnored(t *testing.T) {
	t.Setenv("TEST_ANTHROPIC_KEY", "anthropic-real-key")

	adapter, err := NewAnthropicAdapter(ProviderConfig{
		Type:                 "anthropic",
		ModelID:              "claude-3-opus",
		AuthKeyEnv:           "TEST_ANTHROPIC_KEY",
		LiteralAPIKeyIgnored: true,
	})
	if err != nil {
		t.Fatalf("expected api_key_env to still work, got error: %v", err)
	}
	if adapter == nil {
		t.Fatal("expected a constructed adapter")
	}
}
