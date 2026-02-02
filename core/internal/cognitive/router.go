package cognitive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

// Router manages model selection and inference
type Router struct {
	Config *BrainConfig
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
		for i, m := range config.Models {
			if m.Provider == "ollama" {
				config.Models[i].Endpoint = host
			}
		}
	}

	return &Router{Config: &config}, nil
}

// InferRequest represents the API payload
type InferRequest struct {
	Profile string `json:"profile"`
	Prompt  string `json:"prompt"`
}

// InferResponse represents the API response
type InferResponse struct {
	Text      string `json:"text"`
	ModelUsed string `json:"model_used"`
	Provider  string `json:"provider"`
}

// Infer determines which model to use and executes the CQA loop
func (r *Router) Infer(req InferRequest) (*InferResponse, error) {
	// Root context for the whole operation
	// We might want a global timeout or use Background
	return r.InferWithContract(context.Background(), req)
}

// executeRequest performs the actual provider call (Low Level)
func (r *Router) executeRequest(ctx context.Context, targetModel *Model, prompt string, temp float64) (*InferResponse, error) {
	// 3. (Real Logic) Call Provider
	switch targetModel.Provider {
	case "ollama":
		return r.callOllama(ctx, targetModel, prompt, temp)
	case "openai":
		return r.callOpenAI(ctx, targetModel, prompt, temp)
	}

	// Fallback Stub for unknown providers
	response := &InferResponse{
		Text:      fmt.Sprintf("[%s via %s] %s", targetModel.Name, targetModel.Provider, prompt),
		ModelUsed: targetModel.ID,
		Provider:  targetModel.Provider,
	}

	return response, nil
}

func (r *Router) callOllama(ctx context.Context, model *Model, prompt string, temp float64) (*InferResponse, error) {
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

func (r *Router) callOpenAI(ctx context.Context, model *Model, prompt string, temp float64) (*InferResponse, error) {
	// 1. Get API Key
	apiKey := os.Getenv(model.AuthKeyEnv)
	if apiKey == "" {
		return nil, fmt.Errorf("missing api key for %s (env: %s)", model.ID, model.AuthKeyEnv)
	}

	// 2. Construct URL (Assume Endpoint is base, e.g., https://api.openai.com/v1)
	// We append /chat/completions standard
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
