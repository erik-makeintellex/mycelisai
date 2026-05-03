package server

import (
	"context"

	"github.com/mycelis/core/internal/cognitive"
)

type cognitiveTestProvider struct{}

func (cognitiveTestProvider) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "ok", Provider: "mock", ModelUsed: "test-model"}, nil
}

func (cognitiveTestProvider) Probe(context.Context) (bool, error) {
	return true, nil
}
