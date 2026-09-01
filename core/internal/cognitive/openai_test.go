package cognitive

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestOpenAIAdapter_InferLiteLLMCompatibleResponse(t *testing.T) {
	t.Setenv("LITELLM_PROXY_API_KEY_TEST", "proxy-secret")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("request path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("request method = %q, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer proxy-secret" {
			t.Errorf("Authorization = %q, want Bearer proxy-secret", got)
		}

		var request struct {
			Model       string        `json:"model"`
			Messages    []ChatMessage `json:"messages"`
			Temperature float64       `json:"temperature"`
			MaxTokens   int           `json:"max_tokens"`
			Stop        []string      `json:"stop"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if request.Model != "team-coder" {
			t.Errorf("request model = %q, want team-coder", request.Model)
		}
		if len(request.Messages) != 1 || request.Messages[0].Role != "system" || request.Messages[0].Content != "stay bounded" {
			t.Errorf("request messages = %#v, want structured message", request.Messages)
		}
		if request.Temperature != 0.2 || request.MaxTokens != 321 || len(request.Stop) != 1 || request.Stop[0] != "DONE" {
			t.Errorf("request options = temperature %v, max_tokens %d, stop %#v", request.Temperature, request.MaxTokens, request.Stop)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"chatcmpl-litellm",
			"object":"chat.completion",
			"created":1,
			"model":"openai/gpt-4.1-mini-2025-04-14",
			"choices":[{"index":0,"message":{"role":"assistant","content":"gateway response"},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":7,"completion_tokens":11,"total_tokens":18}
		}`))
	}))
	t.Cleanup(server.Close)

	adapter, err := NewOpenAIAdapter(ProviderConfig{
		Type:       "openai_compatible",
		Endpoint:   server.URL + "/v1",
		ModelID:    "team-coder",
		AuthKeyEnv: "LITELLM_PROXY_API_KEY_TEST",
	})
	if err != nil {
		t.Fatalf("NewOpenAIAdapter: %v", err)
	}

	resp, err := adapter.Infer(context.Background(), "ignored legacy prompt", InferOptions{
		Temperature: 0.2,
		MaxTokens:   321,
		Stop:        []string{"DONE"},
		Messages:    []ChatMessage{{Role: "system", Content: "stay bounded"}},
	})
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	if resp.Text != "gateway response" {
		t.Fatalf("Text = %q, want gateway response", resp.Text)
	}
	if resp.Provider != "openai_compatible" {
		t.Fatalf("Provider = %q, want openai_compatible", resp.Provider)
	}
	if resp.ModelUsed != "openai/gpt-4.1-mini-2025-04-14" {
		t.Fatalf("ModelUsed = %q, want response model", resp.ModelUsed)
	}
	if resp.TokensUsed != 18 {
		t.Fatalf("TokensUsed = %d, want reported total 18", resp.TokensUsed)
	}
}

func TestOpenAIAdapter_InferLiteLLMToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"chatcmpl-tool",
			"object":"chat.completion",
			"model":"anthropic/claude-sonnet-4",
			"choices":[{"index":0,"message":{"role":"assistant","content":null,"tool_calls":[{"id":"call_1","type":"function","function":{"name":"write_file","arguments":"{\"path\":\"workspace/result.txt\",\"content\":\"ready\"}"}}]},"finish_reason":"tool_calls"}],
			"usage":{"prompt_tokens":9,"completion_tokens":6,"total_tokens":15}
		}`))
	}))
	t.Cleanup(server.Close)

	adapter, err := NewOpenAIAdapter(ProviderConfig{
		Type:     "openai_compatible",
		Endpoint: server.URL + "/v1",
		ModelID:  "team-coder",
		AuthKey:  "proxy-secret",
	})
	if err != nil {
		t.Fatalf("NewOpenAIAdapter: %v", err)
	}

	resp, err := adapter.Infer(context.Background(), "write the result", InferOptions{})
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	wantText := `{"tool_call":{"arguments":{"content":"ready","path":"workspace/result.txt"},"name":"write_file"}}`
	if resp.Text != wantText {
		t.Fatalf("Text = %q, want %q", resp.Text, wantText)
	}
	if resp.ModelUsed != "anthropic/claude-sonnet-4" || resp.TokensUsed != 15 {
		t.Fatalf("response metadata = model %q, tokens %d", resp.ModelUsed, resp.TokensUsed)
	}
}

func TestOpenAIAdapter_InferFailsClosedWithoutChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"openai/gpt-4.1-mini","choices":[],"usage":{"total_tokens":4}}`))
	}))
	t.Cleanup(server.Close)

	adapter, err := NewOpenAIAdapter(ProviderConfig{
		Type:     "openai_compatible",
		Endpoint: server.URL + "/v1",
		ModelID:  "team-coder",
		AuthKey:  "proxy-secret",
	})
	if err != nil {
		t.Fatalf("NewOpenAIAdapter: %v", err)
	}

	_, err = adapter.Infer(context.Background(), "hello", InferOptions{})
	if err == nil || !strings.Contains(err.Error(), "no choices returned") {
		t.Fatalf("Infer error = %v, want no choices returned", err)
	}
}

func TestOpenAIAdapter_ProbeLiteLLMFailsClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("request path = %q, want /v1/models", r.URL.Path)
		}
		http.Error(w, `{"error":{"message":"gateway unavailable"}}`, http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	adapter, err := NewOpenAIAdapter(ProviderConfig{
		Type:     "openai_compatible",
		Endpoint: server.URL + "/v1",
		ModelID:  "team-coder",
		AuthKey:  "proxy-secret",
	})
	if err != nil {
		t.Fatalf("NewOpenAIAdapter: %v", err)
	}

	healthy, err := adapter.Probe(context.Background())
	if healthy || err == nil {
		t.Fatalf("Probe = (%v, %v), want unhealthy error", healthy, err)
	}
	if got := err.Error(); got != "provider probe failed (status 503)" {
		t.Fatalf("Probe error = %q, want normalized status", got)
	}
}

func TestOpenAIAdapter_DoesNotExposeGatewayErrorBody(t *testing.T) {
	const sentinel = "secret-upstream-detail-do-not-expose"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"` + sentinel + `"}}`))
	}))
	t.Cleanup(server.Close)

	adapter, err := NewOpenAIAdapter(ProviderConfig{
		Type:     "openai_compatible",
		Endpoint: server.URL + "/v1",
		ModelID:  "team-coder",
		AuthKey:  "proxy-secret",
	})
	if err != nil {
		t.Fatalf("NewOpenAIAdapter: %v", err)
	}

	_, err = adapter.Infer(context.Background(), "hello", InferOptions{})
	if err == nil {
		t.Fatal("Infer error = nil, want normalized gateway failure")
	}
	if strings.Contains(err.Error(), sentinel) || err.Error() != "openai inference failed (status 429)" {
		t.Fatalf("Infer error = %q, want normalized status without upstream body", err)
	}
}

func TestNormalizeOpenAIMessage_PrefersStructuredToolCallsWhenContentEmpty(t *testing.T) {
	msg := openai.ChatCompletionMessage{
		ToolCalls: []openai.ToolCall{
			{
				Function: openai.FunctionCall{
					Name:      "write_file",
					Arguments: `{"path":"workspace/test.txt","content":"hello"}`,
				},
			},
		},
	}

	got := normalizeOpenAIMessage(msg)
	want := `{"tool_call":{"arguments":{"content":"hello","path":"workspace/test.txt"},"name":"write_file"}}`
	if got != want {
		t.Fatalf("normalizeOpenAIMessage() = %q, want %q", got, want)
	}
}

func TestNormalizeOpenAIMessage_FallsBackToFunctionCall(t *testing.T) {
	msg := openai.ChatCompletionMessage{
		FunctionCall: &openai.FunctionCall{
			Name:      "delegate",
			Arguments: `{"team_id":"admin-core"}`,
		},
	}

	got := normalizeOpenAIMessage(msg)
	want := `{"tool_call":{"arguments":{"team_id":"admin-core"},"name":"delegate"}}`
	if got != want {
		t.Fatalf("normalizeOpenAIMessage() = %q, want %q", got, want)
	}
}

func TestNormalizeOpenAIMessage_PrefersStructuredToolCallsOverProse(t *testing.T) {
	msg := openai.ChatCompletionMessage{
		Content: "Here is the draft answer, but I also need to call a tool.",
		ToolCalls: []openai.ToolCall{
			{
				Function: openai.FunctionCall{
					Name:      "write_file",
					Arguments: `{"path":"workspace/test.txt","content":"hello"}`,
				},
			},
		},
	}

	got := normalizeOpenAIMessage(msg)
	want := `{"tool_call":{"arguments":{"content":"hello","path":"workspace/test.txt"},"name":"write_file"}}`
	if got != want {
		t.Fatalf("normalizeOpenAIMessage() = %q, want %q", got, want)
	}
}
