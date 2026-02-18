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
	// Map abstract ChatMessage to openai.ChatCompletionMessage
	var messages []openai.ChatCompletionMessage
	if len(opts.Messages) > 0 {
		messages = make([]openai.ChatCompletionMessage, len(opts.Messages))
		for i, m := range opts.Messages {
			messages[i] = openai.ChatCompletionMessage{
				Role:    m.Role,
				Content: m.Content,
			}
		}
	} else {
		// Fallback for legacy Prompt field
		messages = []openai.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		}
	}

	reqBody := openai.ChatCompletionRequest{
		Model:       a.model,
		Messages:    messages,
		Temperature: float32(opts.Temperature),
		MaxTokens:   opts.MaxTokens,
		Stop:        opts.Stop,
	}

	resp, err := a.client.CreateChatCompletion(ctx, reqBody)
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

// Embed generates a vector embedding for the given text using the OpenAI-compatible
// /v1/embeddings endpoint. Works with Ollama (nomic-embed-text) and OpenAI (text-embedding-3-small).
func (a *OpenAIAdapter) Embed(ctx context.Context, text string, model string) ([]float64, error) {
	if model == "" {
		model = DefaultEmbedModel
	}

	resp, err := a.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.EmbeddingModel(model),
	})
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	// Convert float32 → float64 for pgvector compatibility
	raw := resp.Data[0].Embedding
	vec := make([]float64, len(raw))
	for i, v := range raw {
		vec[i] = float64(v)
	}
	return vec, nil
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
