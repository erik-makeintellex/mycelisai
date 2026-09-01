package cognitive

import (
	"context"
	"crypto/sha256"
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
			User        string        `json:"user"`
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
		if !strings.HasPrefix(request.User, "mycelis-v1-") {
			t.Errorf("request user = %q, want opaque Mycelis correlation", request.User)
		}
		for _, rawID := range []string{"run-123", "delivery-team", "coder-1"} {
			if strings.Contains(request.User, rawID) {
				t.Errorf("request user exposed raw identifier %q: %q", rawID, request.User)
			}
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
		Type:         "openai_compatible",
		Endpoint:     server.URL + "/v1",
		ModelID:      "team-coder",
		AuthKeyEnv:   "LITELLM_PROXY_API_KEY_TEST",
		ModelGateway: true,
		DataBoundary: "leaves_org",
	})
	if err != nil {
		t.Fatalf("NewOpenAIAdapter: %v", err)
	}

	resp, err := adapter.Infer(context.Background(), "ignored legacy prompt", InferOptions{
		Temperature: 0.2,
		MaxTokens:   321,
		Stop:        []string{"DONE"},
		Messages:    []ChatMessage{{Role: "system", Content: "stay bounded"}},
		Correlation: InferenceCorrelation{RunID: "run-123", TeamID: "delivery-team", AgentID: "coder-1"},
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
	if resp.UpstreamResponseID != "chatcmpl-litellm" {
		t.Fatalf("UpstreamResponseID = %q, want chatcmpl-litellm", resp.UpstreamResponseID)
	}
	if resp.PromptTokens != 7 || resp.CompletionTokens != 11 || resp.TokensUsed != 18 {
		t.Fatalf("usage = prompt %d, completion %d, total %d", resp.PromptTokens, resp.CompletionTokens, resp.TokensUsed)
	}
}

func TestOpenAIAdapter_InferLiteLLMToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			User string `json:"user"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if request.User != "" {
			t.Errorf("ordinary provider sent gateway correlation user %q", request.User)
		}
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

	resp, err := adapter.Infer(context.Background(), "write the result", InferOptions{
		Correlation: InferenceCorrelation{RunID: "run-ordinary", TeamID: "ordinary-team", AgentID: "ordinary-agent"},
	})
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	wantText := `{"tool_call":{"arguments":{"content":"ready","path":"workspace/result.txt"},"name":"write_file"}}`
	if resp.Text != wantText {
		t.Fatalf("Text = %q, want %q", resp.Text, wantText)
	}
	if resp.ModelUsed != "anthropic/claude-sonnet-4" || resp.UpstreamResponseID != "chatcmpl-tool" {
		t.Fatalf("response identity = model %q, upstream response %q", resp.ModelUsed, resp.UpstreamResponseID)
	}
	if resp.PromptTokens != 9 || resp.CompletionTokens != 6 || resp.TokensUsed != 15 {
		t.Fatalf("usage = prompt %d, completion %d, total %d", resp.PromptTokens, resp.CompletionTokens, resp.TokensUsed)
	}
}

func TestInferResponse_AccountingJSONCompatibility(t *testing.T) {
	raw, err := json.Marshal(InferResponse{
		Text:               "ready",
		ModelUsed:          "upstream/model",
		Provider:           "gateway",
		UpstreamResponseID: "chatcmpl-123",
		PromptTokens:       3,
		CompletionTokens:   5,
		TokensUsed:         8,
	})
	if err != nil {
		t.Fatalf("Marshal InferResponse: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("Unmarshal InferResponse: %v", err)
	}
	wantNumbers := map[string]float64{
		"prompt_tokens": 3, "completion_tokens": 5, "tokens_used": 8,
	}
	for key, want := range wantNumbers {
		if payload[key] != want {
			t.Fatalf("%s = %#v, want %v", key, payload[key], want)
		}
	}
	if payload["upstream_response_id"] != "chatcmpl-123" {
		t.Fatalf("upstream_response_id = %#v, want chatcmpl-123", payload["upstream_response_id"])
	}

	minimal, err := json.Marshal(InferResponse{Text: "ready", ModelUsed: "model", Provider: "provider"})
	if err != nil {
		t.Fatalf("Marshal minimal InferResponse: %v", err)
	}
	for _, field := range []string{"upstream_response_id", "prompt_tokens", "completion_tokens", "tokens_used"} {
		if strings.Contains(string(minimal), `"`+field+`"`) {
			t.Fatalf("missing upstream field %q was synthesized in %s", field, minimal)
		}
	}
}

func TestOpenAIAdapter_ModelGatewayRequiresExplicitDataBoundary(t *testing.T) {
	_, err := NewOpenAIAdapter(ProviderConfig{
		Type:         "openai_compatible",
		Endpoint:     "http://127.0.0.1:4000/v1",
		ModelID:      "mycelis-default",
		AuthKey:      "test-key",
		ModelGateway: true,
	})
	if err == nil || !strings.Contains(err.Error(), "explicit local_only or leaves_org") {
		t.Fatalf("NewOpenAIAdapter error = %v, want explicit gateway boundary failure", err)
	}
}

func TestOpaqueGatewayCorrelation_IsStableBoundedAndOpaque(t *testing.T) {
	correlation := InferenceCorrelation{RunID: "run-123", TeamID: "delivery-team", AgentID: "coder-1"}
	key := deriveGatewayCorrelationKey("scoped-proxy-client-key")
	first, err := opaqueGatewayCorrelation(correlation, key)
	if err != nil {
		t.Fatalf("opaqueGatewayCorrelation: %v", err)
	}
	second, err := opaqueGatewayCorrelation(correlation, key)
	if err != nil {
		t.Fatalf("opaqueGatewayCorrelation repeat: %v", err)
	}
	changed, err := opaqueGatewayCorrelation(InferenceCorrelation{RunID: "run-124", TeamID: "delivery-team", AgentID: "coder-1"}, key)
	if err != nil {
		t.Fatalf("opaqueGatewayCorrelation changed: %v", err)
	}
	if first == "" || first != second || first == changed {
		t.Fatalf("correlation tokens first=%q second=%q changed=%q", first, second, changed)
	}
	otherDeployment, err := opaqueGatewayCorrelation(correlation, deriveGatewayCorrelationKey("different-scoped-key"))
	if err != nil {
		t.Fatalf("opaqueGatewayCorrelation other deployment: %v", err)
	}
	if first == otherDeployment {
		t.Fatal("correlation token must be deployment-key scoped")
	}
	if len(first) != len("mycelis-v1-")+sha256.Size*2 {
		t.Fatalf("correlation token length = %d, want bounded digest", len(first))
	}
	for _, rawID := range []string{correlation.RunID, correlation.TeamID, correlation.AgentID} {
		if strings.Contains(first, rawID) {
			t.Fatalf("correlation token exposed raw identifier %q", rawID)
		}
	}

	_, err = opaqueGatewayCorrelation(InferenceCorrelation{TeamID: "unsafe team/id"}, key)
	if err == nil || strings.Contains(err.Error(), "unsafe team/id") {
		t.Fatalf("unsafe correlation error = %v, want redacted validation failure", err)
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
