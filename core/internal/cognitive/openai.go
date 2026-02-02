package cognitive

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIAdapter struct {
	client *openai.Client
	model  string
}

func NewOpenAIAdapter(config ProviderConfig) (*OpenAIAdapter, error) {
	// 1. Resolve Auth Key
	apiKey := config.AuthKey
	if config.AuthKeyEnv != "" {
		if envVal := os.Getenv(config.AuthKeyEnv); envVal != "" {
			apiKey = envVal
		}
	}
	// Fallback for local providers (Ollama needs dummy key)
	if apiKey == "" && config.Type == "openai_compatible" {
		apiKey = "dummy"
	}
	if apiKey == "" {
		return nil, fmt.Errorf("missing api key")
	}

	// 2. Configure Client
	clientConfig := openai.DefaultConfig(apiKey)
	if config.Endpoint != "" {
		clientConfig.BaseURL = config.Endpoint
	}

	return &OpenAIAdapter{
		client: openai.NewClientWithConfig(clientConfig),
		model:  config.ModelID,
	}, nil
}

func (a *OpenAIAdapter) Infer(ctx context.Context, prompt string, opts InferOptions) (*InferResponse, error) {
	req := openai.ChatCompletionRequest{
		Model: a.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: float32(opts.Temperature),
		MaxTokens:   opts.MaxTokens,
		Stop:        opts.Stop,
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai inference failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	return &InferResponse{
		Text:      resp.Choices[0].Message.Content,
		ModelUsed: a.model,
		Provider:  "openai",
	}, nil
}

func (a *OpenAIAdapter) Probe(ctx context.Context) (bool, error) {
	// Simple connectivity check: List Models or empty chat?
	// Listing models is safer/cheaper usually, but might return huge list.
	// Let's try to send an empty/hello message to check Auth?
	// Or use ListModels if available. sashabaranov has ListModels.

	_, err := a.client.ListModels(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}
