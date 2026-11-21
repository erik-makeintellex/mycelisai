package cognitive

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIAdapter_UsesBearerAuthHeader(t *testing.T) {
	t.Parallel()

	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"chatcmpl-test","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`)
	}))
	defer server.Close()

	adapter, err := NewOpenAIAdapter(ProviderConfig{
		Type:     "openai",
		Endpoint: server.URL + "/v1",
		ModelID:  "gpt-test",
		AuthKey:  "openai-test-key",
	})
	if err != nil {
		t.Fatalf("NewOpenAIAdapter() error = %v", err)
	}

	if _, err := adapter.Infer(context.Background(), "hello", InferOptions{MaxTokens: 8}); err != nil {
		t.Fatalf("Infer() error = %v", err)
	}

	if authHeader != "Bearer openai-test-key" {
		t.Fatalf("expected Authorization bearer header, got %q", authHeader)
	}
}

func TestAnthropicAdapter_UsesRequiredHeaders(t *testing.T) {
	t.Parallel()

	var apiKeyHeader string
	var versionHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader = r.Header.Get("x-api-key")
		versionHeader = r.Header.Get("anthropic-version")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"msg-test","content":[{"text":"ok"}]}`)
	}))
	defer server.Close()

	adapter, err := NewAnthropicAdapter(ProviderConfig{
		Type:     "anthropic",
		Endpoint: server.URL,
		ModelID:  "claude-test",
		AuthKey:  "anthropic-test-key",
	})
	if err != nil {
		t.Fatalf("NewAnthropicAdapter() error = %v", err)
	}

	if _, err := adapter.Infer(context.Background(), "hello", InferOptions{MaxTokens: 8}); err != nil {
		t.Fatalf("Infer() error = %v", err)
	}

	if apiKeyHeader != "anthropic-test-key" {
		t.Fatalf("expected x-api-key header, got %q", apiKeyHeader)
	}
	if versionHeader != AnthropicVersion {
		t.Fatalf("expected anthropic-version %q, got %q", AnthropicVersion, versionHeader)
	}
}

func TestGoogleAdapter_UsesXGoogAPIKeyHeader(t *testing.T) {
	t.Parallel()

	var apiKeyHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader = r.Header.Get("x-goog-api-key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"candidates":[{"content":{"parts":[{"text":"ok"}]}}]}`)
	}))
	defer server.Close()

	adapter, err := NewGoogleAdapter(ProviderConfig{
		Type:     "google",
		Endpoint: server.URL + "/v1beta/models",
		ModelID:  "gemini-test",
		AuthKey:  "gemini-test-key",
	})
	if err != nil {
		t.Fatalf("NewGoogleAdapter() error = %v", err)
	}

	if _, err := adapter.Infer(context.Background(), "hello", InferOptions{MaxTokens: 8}); err != nil {
		t.Fatalf("Infer() error = %v", err)
	}

	if apiKeyHeader != "gemini-test-key" {
		t.Fatalf("expected x-goog-api-key header, got %q", apiKeyHeader)
	}
}

func TestAnthropicAdapter_ForwardsMultiTurnMessagesAndSystemPrompt(t *testing.T) {
	t.Parallel()

	var received anthropicRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"msg-multi","content":[{"text":"claude-reply"}]}`)
	}))
	defer server.Close()

	adapter, err := NewAnthropicAdapter(ProviderConfig{
		Type:     "anthropic",
		Endpoint: server.URL,
		ModelID:  "claude-3-7-sonnet-latest",
		AuthKey:  "anthropic-test-key",
	})
	if err != nil {
		t.Fatalf("NewAnthropicAdapter() error = %v", err)
	}

	resp, err := adapter.Infer(context.Background(), "fallback", InferOptions{
		MaxTokens: 256,
		Messages: []ChatMessage{
			{Role: "system", Content: "You are Soma, a governed orchestrator."},
			{Role: "user", Content: "Plan the release."},
			{Role: "assistant", Content: "Here is the draft."},
			{Role: "user", Content: "Execute step 1."},
		},
	})
	if err != nil {
		t.Fatalf("Infer() error = %v", err)
	}
	if resp.Text != "claude-reply" {
		t.Fatalf("expected claude-reply, got %q", resp.Text)
	}
	if received.System != "You are Soma, a governed orchestrator." {
		t.Fatalf("expected extracted system prompt, got %q", received.System)
	}
	if len(received.Messages) != 3 || received.Messages[1].Role != "assistant" || received.Messages[2].Content != "Execute step 1." {
		t.Fatalf("unexpected anthropic messages payload: %+v", received.Messages)
	}
}

func TestGoogleAdapter_ForwardsMultiTurnMessagesAndSystemInstruction(t *testing.T) {
	t.Parallel()

	var received googleRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"candidates":[{"content":{"parts":[{"text":"gemini-reply"}]}}]}`)
	}))
	defer server.Close()

	adapter, err := NewGoogleAdapter(ProviderConfig{
		Type:     "google",
		Endpoint: server.URL + "/v1beta/models",
		ModelID:  "gemini-2.5-pro",
		AuthKey:  "gemini-test-key",
	})
	if err != nil {
		t.Fatalf("NewGoogleAdapter() error = %v", err)
	}

	resp, err := adapter.Infer(context.Background(), "fallback", InferOptions{
		MaxTokens: 256,
		Messages: []ChatMessage{
			{Role: "system", Content: "Governed Gemini instruction."},
			{Role: "user", Content: "Review architecture."},
			{Role: "assistant", Content: "Reviewed."},
			{Role: "user", Content: "Approve gate."},
		},
	})
	if err != nil {
		t.Fatalf("Infer() error = %v", err)
	}
	if resp.Text != "gemini-reply" {
		t.Fatalf("expected gemini-reply, got %q", resp.Text)
	}
	if received.SystemInstruction == nil || len(received.SystemInstruction.Parts) != 1 || received.SystemInstruction.Parts[0].Text != "Governed Gemini instruction." {
		t.Fatalf("expected systemInstruction, got %+v", received.SystemInstruction)
	}
	if len(received.Contents) != 3 || received.Contents[1].Role != "model" || received.Contents[2].Parts[0].Text != "Approve gate." {
		t.Fatalf("unexpected google contents payload: %+v", received.Contents)
	}
}
