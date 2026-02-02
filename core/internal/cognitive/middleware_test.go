package cognitive

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// --- Mock Provider ---

type MockProvider struct {
	ShouldFailCount int
	OutputSequence  []string
	CallCount       int
}

func (m *MockProvider) Call(ctx context.Context, model *Model, prompt string, temp float64) (*InferResponse, error) {
	m.CallCount++

	// Simulate Failure
	if m.CallCount <= m.ShouldFailCount {
		return nil, errors.New("simulated provider error")
	}

	// Return Output from Sequence
	idx := m.CallCount - 1 - m.ShouldFailCount
	if idx < len(m.OutputSequence) {
		return &InferResponse{
			Text:      m.OutputSequence[idx],
			ModelUsed: model.ID,
			Provider:  "mock",
		}, nil
	}

	return &InferResponse{Text: "default response", ModelUsed: model.ID, Provider: "mock"}, nil
}

// --- Tests ---

func setupRouter() *Router {
	cfg := &BrainConfig{
		Models: []Model{
			{ID: "mock-model", Provider: "mock", Name: "mock-v1", Endpoint: "http://mock"},
		},
		Profiles: map[string]Profile{
			"test-retry": {
				ActiveModel:  "mock-model",
				TimeoutMs:    1000,
				MaxRetries:   2,
				OutputSchema: "",
			},
			"test-validation": {
				ActiveModel:  "mock-model",
				TimeoutMs:    1000,
				MaxRetries:   2,
				OutputSchema: "strict_json",
			},
		},
	}

	r := &Router{
		Config:    cfg,
		Providers: make(map[string]LLMProvider),
	}
	return r
}

func TestInferWithContract_Success(t *testing.T) {
	r := setupRouter()
	mock := &MockProvider{OutputSequence: []string{"Hello World"}}
	r.Providers["mock"] = mock

	req := InferRequest{Profile: "test-retry", Prompt: "Hi"}
	resp, err := r.Infer(req)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if resp.Text != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", resp.Text)
	}
}

func TestInferWithContract_Retry(t *testing.T) {
	r := setupRouter()
	// Fail once, then succeed
	mock := &MockProvider{ShouldFailCount: 1, OutputSequence: []string{"Recovered"}}
	r.Providers["mock"] = mock

	req := InferRequest{Profile: "test-retry", Prompt: "Hi"}
	resp, err := r.Infer(req)

	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}
	if mock.CallCount != 2 {
		t.Errorf("Expected 2 calls, got %d", mock.CallCount)
	}
	if resp.Text != "Recovered" {
		t.Errorf("Expected 'Recovered', got '%s'", resp.Text)
	}
}

func TestInferWithContract_Validation(t *testing.T) {
	r := setupRouter()
	// Return bad json, then good json
	mock := &MockProvider{
		OutputSequence: []string{
			"I am not JSON",
			`{"key": "value"}`,
		},
	}
	r.Providers["mock"] = mock

	req := InferRequest{Profile: "test-validation", Prompt: "Give JSON"}
	resp, err := r.Infer(req)

	if err != nil {
		t.Fatalf("Expected success after validation retry, got error: %v", err)
	}
	if mock.CallCount != 2 {
		t.Errorf("Expected 2 calls (1 bad, 1 good), got %d", mock.CallCount)
	}
	if !strings.Contains(resp.Text, "key") {
		t.Errorf("Expected valid json, got '%s'", resp.Text)
	}
}

func TestInferWithContract_MaxRetriesExceeded(t *testing.T) {
	r := setupRouter()
	// Fail 5 times (max retries is 2)
	mock := &MockProvider{ShouldFailCount: 5}
	r.Providers["mock"] = mock

	req := InferRequest{Profile: "test-retry", Prompt: "Hi"}
	_, err := r.Infer(req)

	if err == nil {
		t.Fatal("Expected error, got success")
	}
}
