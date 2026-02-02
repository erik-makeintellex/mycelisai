package cognitive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Router manages model selection and inference
type Router struct {
	Config    *BrainConfig
	Providers map[string]LLMProvider
}

// NewRouter loads configuration from the given path
func NewRouter(configPath string) (*Router, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read brain config: %w", err)
	}

	var config BrainConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse brain config: %w", err)
	}

	// Dynamic Override: OLLAMA_HOST
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		if !strings.HasPrefix(host, "http") {
			host = "http://" + host
		}
		for i, m := range config.Models {
			if m.Provider == "ollama" {
				config.Models[i].Endpoint = host
			}
		}
	}

	r := &Router{
		Config:    &config,
		Providers: make(map[string]LLMProvider),
	}

	// Register Default Providers
	r.Providers["ollama"] = &OllamaProvider{}
	r.Providers["openai"] = &OpenAIProvider{}

	return r, nil
}

// Infer determines which model to use and executes the CQA loop
func (r *Router) Infer(req InferRequest) (*InferResponse, error) {
	// Root context for the whole operation
	return r.InferWithContract(context.Background(), req)
}

// executeRequest performs the actual provider call via the Registry
func (r *Router) executeRequest(ctx context.Context, targetModel *Model, prompt string, temp float64) (*InferResponse, error) {
	provider, ok := r.Providers[targetModel.Provider]
	if !ok {
		// Fallback Stub for unknown providers
		return &InferResponse{
			Text:      fmt.Sprintf("[%s via %s] %s", targetModel.Name, targetModel.Provider, prompt),
			ModelUsed: targetModel.ID,
			Provider:  targetModel.Provider,
		}, nil
	}

	return provider.Call(ctx, targetModel, prompt, temp)
}

// --- Provider Implementations ---

type OllamaProvider struct{}

func (p *OllamaProvider) Call(ctx context.Context, model *Model, prompt string, temp float64) (*InferResponse, error) {
	url := fmt.Sprintf("%s/api/generate", model.Endpoint)

	payload := map[string]interface{}{
		"model":  model.Name,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": temp,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ollama payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return &InferResponse{
		Text:      result.Response,
		ModelUsed: model.ID,
		Provider:  "ollama",
	}, nil
}

type OpenAIProvider struct{}

func (p *OpenAIProvider) Call(ctx context.Context, model *Model, prompt string, temp float64) (*InferResponse, error) {
	if os.Getenv("OPENAI_MOCK") == "true" {
		return &InferResponse{Text: "mock-openai", ModelUsed: model.ID, Provider: "openai"}, nil
	}

	// 1. Get API Key (Support both Legacy and New Auth)
	apiKey := ""
	if model.Auth.Value != "" {
		apiKey = os.Getenv(model.Auth.Value) // Treat as Env Var Name for now
		if apiKey == "" {
			apiKey = model.Auth.Value // Fallback to raw value if not an env var (unsafe but supported)
		}
	} else if model.AuthKeyEnv != "" {
		apiKey = os.Getenv(model.AuthKeyEnv)
	}

	// Allow empty logic only if AuthType is none or not set, but OpenAI needs key
	if apiKey == "" {
		return nil, fmt.Errorf("missing api key for %s", model.ID)
	}

	// 2. Construct URL
	url := fmt.Sprintf("%s/chat/completions", model.Endpoint)

	// 3. Payload
	payload := map[string]interface{}{
		"model": model.Name,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": temp,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openai payload: %w", err)
	}

	// 4. Request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai error %d: %s", resp.StatusCode, string(body))
	}

	// 5. Decode
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode openai response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	return &InferResponse{
		Text:      result.Choices[0].Message.Content,
		ModelUsed: model.ID,
		Provider:  "openai",
	}, nil
}
