package memory

import (
	"context"
	"fmt"

	"github.com/mycelis/core/internal/cognitive"
)

// pheMockLLMAdapter is a local test double for cognitive.LLMProvider.
//
// F21: cognitive.MockAdapter used to live in a non-test file
// (core/internal/cognitive/mock.go) so it could be imported here across
// packages. It has moved to a _test.go file in the cognitive package, which
// Go does not expose to other packages' tests, so the memory package's own
// tests carry their own minimal mock adapter instead of importing one.
type pheMockLLMAdapter struct {
	FixedResponse string
}

func (m *pheMockLLMAdapter) Infer(_ context.Context, prompt string, _ cognitive.InferOptions) (*cognitive.InferResponse, error) {
	resp := m.FixedResponse
	if resp == "" {
		resp = fmt.Sprintf("Mock Response to: %s", prompt)
	}
	return &cognitive.InferResponse{
		Text:      resp,
		ModelUsed: "mock-model",
		Provider:  "mock",
	}, nil
}

func (m *pheMockLLMAdapter) Probe(_ context.Context) (bool, error) {
	return true, nil
}
