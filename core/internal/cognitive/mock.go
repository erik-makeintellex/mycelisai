package cognitive

import (
	"context"
	"fmt"
)

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
