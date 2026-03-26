package cognitive_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

// TestInfer_Mock ensures that we can inject a mock adapter
func TestInfer_Mock(t *testing.T) {
	// 1. Setup
	r, err := cognitive.NewRouter("../../config/cognitive.yaml", nil)
	if err != nil {
		t.Fatalf("Failed to load generic config: %v", err)
	}

	// 2. Inject deterministic mock adapter and pin the profile mapping explicitly.
	// This prevents test drift when local config remaps "sentry" to a different provider.
	const mockProviderID = "mock-test"
	r.Adapters[mockProviderID] = &cognitive.MockAdapter{FixedResponse: "Explicit Mock"}
	if r.Config.Providers == nil {
		r.Config.Providers = make(map[string]cognitive.ProviderConfig)
	}
	r.Config.Providers[mockProviderID] = cognitive.ProviderConfig{
		Type:    "mock",
		ModelID: "mock-test",
		Enabled: true,
	}
	if r.Config.Profiles == nil {
		r.Config.Profiles = make(map[string]string)
	}
	r.Config.Profiles["sentry"] = mockProviderID

	// 3. Execute
	req := cognitive.InferRequest{
		Profile: "sentry",
		Prompt:  "Status Check",
	}

	resp, err := r.InferWithContract(context.Background(), req)
	if err != nil {
		t.Fatalf("Mock inference failed: %v", err)
	}

	// 4. Verify
	// MockAdapter returns Provider="mock"
	if resp.Provider != "mock" {
		t.Errorf("Expected provider 'mock', got '%s'", resp.Provider)
	}
	if resp.Text != "Explicit Mock" {
		t.Errorf("Expected 'Explicit Mock', got '%s'", resp.Text)
	}
}

// TestInfer_Live runs against the actual configured provider.
// REQUIREMENT: Env var TEST_LIVE_LLM=true
func TestInfer_Live(t *testing.T) {
	if os.Getenv("TEST_LIVE_LLM") != "true" {
		t.Skip("Skipping Live Integration Test. Set TEST_LIVE_LLM=true to execute.")
	}

	// 1. Setup
	r, err := cognitive.NewRouter("../../config/cognitive.yaml", nil)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 2. Execute against 'sentry' (which points to ollama usually)
	req := cognitive.InferRequest{
		Profile: "sentry",
		Prompt:  "Reply with 'PONG'",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := r.InferWithContract(ctx, req)
	if err != nil {
		t.Fatalf("Live inference failed: %v", err)
	}

	t.Logf("Live Response: %s (Provider: %s)", resp.Text, resp.Provider)
}
