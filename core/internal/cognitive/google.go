package cognitive

import (
	"context"
	"fmt"
)

type GoogleAdapter struct {
	config ProviderConfig
}

func NewGoogleAdapter(config ProviderConfig) *GoogleAdapter {
	return &GoogleAdapter{config: config}
}

func (a *GoogleAdapter) Infer(ctx context.Context, prompt string, opts InferOptions) (*InferResponse, error) {
	return nil, fmt.Errorf("google provider not yet implemented")
}
