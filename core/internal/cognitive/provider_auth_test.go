package cognitive

import (
	"context"
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
