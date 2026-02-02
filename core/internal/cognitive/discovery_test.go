package cognitive_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

// MockProvider implementation for testing
type MockProvider struct {
	IsHealthy bool
	Err       error
}

func (m *MockProvider) Infer(ctx context.Context, prompt string, opts cognitive.InferOptions) (*cognitive.InferResponse, error) {
	if !m.IsHealthy {
		return nil, m.Err
	}
	return &cognitive.InferResponse{
		Text:      "Mock Response",
		ModelUsed: "mock-model",
		Provider:  "mock",
	}, nil
}

func (m *MockProvider) Probe(ctx context.Context) (bool, error) {
	if !m.IsHealthy {
		return false, m.Err
	}
	return true, nil
}

func TestGradeModel(t *testing.T) {
	tests := []struct {
		modelID string
		want    cognitive.Tier
	}{
		{"gpt-4-turbo", cognitive.TierS},
		{"claude-3-opus", cognitive.TierS},
		{"claude-3-5-sonnet", cognitive.TierA},
		{"qwen2.5-32b", cognitive.TierA},
		{"qwen2.5:7b", cognitive.TierB},
		{"mistral-nemo", cognitive.TierB},
		{"phi-3-mini", cognitive.TierC},
		{"unknown-model", cognitive.TierB}, // Default
	}

	for _, tt := range tests {
		got := cognitive.GradeModel(tt.modelID)
		if got != tt.want {
			t.Errorf("GradeModel(%q) = %v, want %v", tt.modelID, got, tt.want)
		}
	}
}

func TestServiceDiscovery(t *testing.T) {
	providers := map[string]cognitive.LLMProvider{
		"healthy": &MockProvider{IsHealthy: true},
		"sick":    &MockProvider{IsHealthy: false, Err: errors.New("connection refused")},
	}

	sd := cognitive.NewServiceDiscovery(providers)
	results := sd.DiscoverAll(context.Background())

	if !results["healthy"].Healthy {
		t.Error("Expected 'healthy' provider to be healthy")
	}
	if results["sick"].Healthy {
		t.Error("Expected 'sick' provider to be unhealthy")
	}
}

// TestRouter_AutoConfig validates that the Router automatically re-routes traffic
// from a sick provider to a healthy one during initialization.
func TestRouter_AutoConfig(t *testing.T) {
	// 1. Setup Config
	config := &cognitive.BrainConfig{
		Providers: map[string]cognitive.ProviderConfig{
			"primary_dead": {Type: "mock", ModelID: "gpt-4"},
			"backup_live":  {Type: "mock", ModelID: "qwen2.5"},
		},
		Profiles: map[string]string{
			"architect": "primary_dead", // Points to dead provider
		},
	}

	// 2. Setup Adapters (Manually injected as we are unit testing logic, not loading yaml)
	adapters := map[string]cognitive.LLMProvider{
		"primary_dead": &MockProvider{IsHealthy: false, Err: errors.New("dead")},
		"backup_live":  &MockProvider{IsHealthy: true},
	}

	// 3. Create Router Manually (simulating NewRouter logic)
	r := &cognitive.Router{
		Config:   config,
		Adapters: adapters,
	}

	// 4. Run Discovery & AutoConfig Logic (Copied/Refactored from NewRouter)
	// Ideally NewRouter logic should be extracted to a method like `r.AutoConfig()`
	// For now, we call the newly created public method `AutoConfigure()` if we create it,
	// OR we refactor Router to expose it.
	// Action: I will refactor Router to expose `AutoConfigure()` to make it testable.

	// Assuming r.AutoConfigure() exists (Implementing it in next step)
	r.AutoConfigure(context.Background())

	// 5. Assert Switch
	currentProviderID := r.Config.Profiles["architect"]
	if currentProviderID != "backup_live" {
		t.Errorf("AutoConfig failed. Expected 'backup_live', got '%s'", currentProviderID)
	}
}
