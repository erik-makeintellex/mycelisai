package cognitive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	AnthropicDefaultEndpoint = "https://api.anthropic.com/v1/messages"
	AnthropicVersion         = "2023-06-01"
)

type AnthropicAdapter struct {
	apiKey   string
	model    string
	endpoint string
}

func NewAnthropicAdapter(config ProviderConfig) (*AnthropicAdapter, error) {
	// 1. Resolve Auth Key
	apiKey := config.AuthKey
	if config.AuthKeyEnv != "" {
		if envVal := os.Getenv(config.AuthKeyEnv); envVal != "" {
			apiKey = envVal
		}
	}
	if apiKey == "" {
		return nil, fmt.Errorf("missing api key for anthropic")
	}

	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = AnthropicDefaultEndpoint
	}

	return &AnthropicAdapter{
		apiKey:   apiKey,
		model:    config.ModelID,
		endpoint: endpoint,
	}, nil
}

// --- Request/Response Structs ---

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
	MaxTokens int                `json:"max_tokens,omitempty"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (a *AnthropicAdapter) Infer(ctx context.Context, prompt string, opts InferOptions) (*InferResponse, error) {
	// 1. Prepare Request Payload
	payload := anthropicRequest{
		Model:     a.model,
		Messages:  []anthropicMessage{{Role: "user", Content: prompt}},
		MaxTokens: opts.MaxTokens,
	}

	// Default MaxTokens if not set (Anthropic requires it)
	if payload.MaxTokens == 0 {
		payload.MaxTokens = 1024
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal anthropic request: %w", err)
	}

	// 2. Build HTTP Request
	req, err := http.NewRequestWithContext(ctx, "POST", a.endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", AnthropicVersion)
	req.Header.Set("content-type", "application/json")

	// 3. Execute
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic connection failed: %w", err)
	}
	defer resp.Body.Close()

	// 4. Decode Response
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("anthropic error %d: %s", resp.StatusCode, string(body))
	}

	var result anthropicResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode anthropic response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("anthropic api error: %s - %s", result.Error.Type, result.Error.Message)
	}

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("anthropic returned empty content")
	}

	return &InferResponse{
		Text:      result.Content[0].Text,
		ModelUsed: a.model,
		Provider:  "anthropic",
	}, nil
}
