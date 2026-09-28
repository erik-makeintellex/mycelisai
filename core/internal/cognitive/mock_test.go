package cognitive

import (
	"context"
	"fmt"
)

// MockAdapter is a test double for LLMProvider. F21: it must never ship in the
// production binary as a real-looking provider, so it lives in a _test.go file;
// it is compiled only into the cognitive package's own test builds (both
// internal `package cognitive` tests and the external `package cognitive_test`
// tests in this directory can see it).
type MockAdapter struct {
	FixedResponse string
}

func (m *MockAdapter) Infer(ctx context.Context, prompt string, opts InferOptions) (*InferResponse, error) {
	resp := m.FixedResponse
	if resp == "" {
		resp = fmt.Sprintf("Mock Response to: %s", prompt)
	}
	return &InferResponse{
		Text:      resp,
		ModelUsed: "mock-model",
		Provider:  "mock",
	}, nil
}

func (m *MockAdapter) Probe(ctx context.Context) (bool, error) {
	return true, nil
}
