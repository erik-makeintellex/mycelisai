package cognitive

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIAdapter struct {
	client       *openai.Client
	model        string
	provider     string
	modelGateway bool
	// gatewayCorrelationKey is a one-way derivative of the resolved scoped
	// gateway client key. The raw credential is never retained in adapter-owned
	// correlation state.
	gatewayCorrelationKey []byte
}

var safeInferenceCorrelationID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

func NewOpenAIAdapter(config ProviderConfig) (*OpenAIAdapter, error) {
	// 1. Resolve Auth Key
	apiKey := config.AuthKey
	if config.AuthKeyEnv != "" {
		if envVal := os.Getenv(config.AuthKeyEnv); envVal != "" {
			apiKey = envVal
		}
	}
	// Fallback for local providers (Ollama needs dummy key)
	if apiKey == "" && config.Type == "openai_compatible" && !config.ModelGateway {
		apiKey = "dummy"
	}
	if apiKey == "" {
		return nil, fmt.Errorf("missing api key")
	}
	if config.ModelGateway {
		switch strings.TrimSpace(config.DataBoundary) {
		case "local_only", "leaves_org":
		default:
			return nil, fmt.Errorf("model gateway requires an explicit local_only or leaves_org data boundary")
		}
	}

	// 2. Configure Client
	clientConfig := openai.DefaultConfig(apiKey)
	if config.Endpoint != "" {
		clientConfig.BaseURL = config.Endpoint
	}

	provider := strings.TrimSpace(config.Type)
	if provider == "" {
		provider = "openai"
	}

	var gatewayCorrelationKey []byte
	if config.ModelGateway {
		gatewayCorrelationKey = deriveGatewayCorrelationKey(apiKey)
	}

	return &OpenAIAdapter{
		client:                openai.NewClientWithConfig(clientConfig),
		model:                 config.ModelID,
		provider:              provider,
		modelGateway:          config.ModelGateway,
		gatewayCorrelationKey: gatewayCorrelationKey,
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
	if a.modelGateway {
		correlationUser, err := opaqueGatewayCorrelation(opts.Correlation, a.gatewayCorrelationKey)
		if err != nil {
			return nil, fmt.Errorf("invalid model gateway correlation: %w", err)
		}
		reqBody.User = correlationUser
	}

	resp, err := a.client.CreateChatCompletion(ctx, reqBody)
	if err != nil {
		return nil, normalizedOpenAIError("openai inference", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	text := normalizeOpenAIMessage(resp.Choices[0].Message)
	modelUsed := resp.Model
	if strings.TrimSpace(modelUsed) == "" {
		modelUsed = a.model
	}

	return &InferResponse{
		Text:               text,
		ModelUsed:          modelUsed,
		Provider:           a.provider,
		UpstreamResponseID: resp.ID,
		PromptTokens:       resp.Usage.PromptTokens,
		CompletionTokens:   resp.Usage.CompletionTokens,
		TokensUsed:         resp.Usage.TotalTokens,
	}, nil
}

func deriveGatewayCorrelationKey(apiKey string) []byte {
	derived := sha256.Sum256([]byte("mycelis:model-gateway-correlation-key:v1\x00" + apiKey))
	key := make([]byte, len(derived))
	copy(key, derived[:])
	return key
}

func opaqueGatewayCorrelation(correlation InferenceCorrelation, key []byte) (string, error) {
	values := []struct {
		name  string
		value string
	}{
		{name: "run_id", value: strings.TrimSpace(correlation.RunID)},
		{name: "team_id", value: strings.TrimSpace(correlation.TeamID)},
		{name: "agent_id", value: strings.TrimSpace(correlation.AgentID)},
	}

	hasScope := false
	if len(key) != sha256.Size {
		return "", fmt.Errorf("gateway correlation key is unavailable")
	}
	hash := hmac.New(sha256.New, key)
	_, _ = hash.Write([]byte("mycelis:model-gateway-correlation:v1\x00"))
	for _, item := range values {
		if item.value != "" {
			hasScope = true
			if !safeInferenceCorrelationID.MatchString(item.value) {
				return "", fmt.Errorf("%s is not a bounded identifier", item.name)
			}
		}
		_, _ = fmt.Fprintf(hash, "%s:%d:%s\x00", item.name, len(item.value), item.value)
	}
	if !hasScope {
		return "", nil
	}
	return fmt.Sprintf("mycelis-v1-%x", hash.Sum(nil)), nil
}

func normalizeOpenAIMessage(msg openai.ChatCompletionMessage) string {
	if len(msg.ToolCalls) > 0 {
		if payload, ok := synthesizeToolCallPayload(msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments); ok {
			return payload
		}
	}
	if msg.FunctionCall != nil {
		if payload, ok := synthesizeToolCallPayload(msg.FunctionCall.Name, msg.FunctionCall.Arguments); ok {
			return payload
		}
	}
	if strings.TrimSpace(msg.Content) != "" {
		return msg.Content
	}
	if strings.TrimSpace(msg.Refusal) != "" {
		return msg.Refusal
	}
	return msg.Content
}

func synthesizeToolCallPayload(name, arguments string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false
	}

	args := map[string]any{}
	if rawArgs := strings.TrimSpace(arguments); rawArgs != "" {
		if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
			args["raw_arguments"] = rawArgs
		}
	}

	payload, err := json.Marshal(map[string]any{
		"tool_call": map[string]any{
			"name":      name,
			"arguments": args,
		},
	})
	if err != nil {
		return "", false
	}
	return string(payload), true
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
		return nil, normalizedOpenAIError("embedding", err)
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
		return false, normalizedOpenAIError("provider probe", err)
	}
	return true, nil
}

// normalizedOpenAIError retains only safe request category/status information.
// Upstream bodies and provider messages can contain prompts, credentials, or
// deployment details and must not cross the adapter boundary.
func normalizedOpenAIError(operation string, err error) error {
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%s canceled: %w", operation, context.Canceled)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s timed out: %w", operation, context.DeadlineExceeded)
	}

	var apiErr *openai.APIError
	if errors.As(err, &apiErr) && apiErr.HTTPStatusCode > 0 {
		return fmt.Errorf("%s failed (status %d)", operation, apiErr.HTTPStatusCode)
	}
	var requestErr *openai.RequestError
	if errors.As(err, &requestErr) && requestErr.HTTPStatusCode > 0 {
		return fmt.Errorf("%s failed (status %d)", operation, requestErr.HTTPStatusCode)
	}
	return fmt.Errorf("%s failed", operation)
}
