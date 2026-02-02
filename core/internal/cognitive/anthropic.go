package cognitive

import (
	"context"
	"fmt"
)

type AnthropicAdapter struct {
	config ProviderConfig
}

func NewAnthropicAdapter(config ProviderConfig) *AnthropicAdapter {
	return &AnthropicAdapter{config: config}
}

func (a *AnthropicAdapter) Infer(ctx context.Context, prompt string, opts InferOptions) (*InferResponse, error) {
	return nil, fmt.Errorf("anthropic provider not yet implemented")
}
