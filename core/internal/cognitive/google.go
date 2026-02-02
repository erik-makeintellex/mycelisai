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
	GoogleDefaultEndpoint = "https://generativelanguage.googleapis.com/v1beta/models"
)

type GoogleAdapter struct {
	apiKey   string
	model    string
	endpoint string
}

func NewGoogleAdapter(config ProviderConfig) (*GoogleAdapter, error) {
	// 1. Resolve Auth Key
	apiKey := config.AuthKey
	if config.AuthKeyEnv != "" {
		if envVal := os.Getenv(config.AuthKeyEnv); envVal != "" {
			apiKey = envVal
		}
	}
	if apiKey == "" {
		return nil, fmt.Errorf("missing api key for google")
	}

	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = GoogleDefaultEndpoint
	}

	return &GoogleAdapter{
		apiKey:   apiKey,
		model:    config.ModelID,
		endpoint: endpoint,
	}, nil
}

// --- Request/Response Structs ---

type googlePart struct {
	Text string `json:"text"`
}

type googleContent struct {
	Role  string       `json:"role,omitempty"` // "user" or "model"
	Parts []googlePart `json:"parts"`
}

type googleRequest struct {
	Contents []googleContent `json:"contents"`

	// Safety, Generation Config could go here
	GenerationConfig struct {
		Temperature     float64 `json:"temperature,omitempty"`
		MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	} `json:"generationConfig,omitempty"`
}

type googleResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`

	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (g *GoogleAdapter) Infer(ctx context.Context, prompt string, opts InferOptions) (*InferResponse, error) {
	// 1. Prepare Payload
	payload := googleRequest{
		Contents: []googleContent{
			{
				Role:  "user",
				Parts: []googlePart{{Text: prompt}},
			},
		},
	}

	// Apply Options
	payload.GenerationConfig.Temperature = opts.Temperature
	payload.GenerationConfig.MaxOutputTokens = opts.MaxTokens

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal google request: %w", err)
	}

	// 2. Build URL (API Key is query param)
	// https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-pro:generateContent?key=...
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", g.endpoint, g.model, g.apiKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 3. Execute
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google connection failed: %w", err)
	}
	defer resp.Body.Close()

	// 4. Decode
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("google error %d: %s", resp.StatusCode, string(body))
	}

	var result googleResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode google response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("google api error %d: %s", result.Error.Code, result.Error.Message)
	}

	if len(result.Candidates) == 0 {
		return nil, fmt.Errorf("google returned no candidates")
	}

	if len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("google returned empty content parts")
	}

	return &InferResponse{
		Text:      result.Candidates[0].Content.Parts[0].Text,
		ModelUsed: g.model,
		Provider:  "google",
	}, nil
}

func (g *GoogleAdapter) Probe(ctx context.Context) (bool, error) {
	opts := InferOptions{MaxTokens: 1}
	_, err := g.Infer(ctx, "ping", opts)
	if err != nil {
		return false, err
	}
	return true, nil
}
