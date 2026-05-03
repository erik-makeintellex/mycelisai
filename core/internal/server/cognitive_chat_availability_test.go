package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/searchcap"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestHandleChat_RequiresAvailableCognitiveEngine(t *testing.T) {
	s := &AdminServer{
		Cognitive: &cognitive.Router{
			Config: &cognitive.BrainConfig{
				Providers: map[string]cognitive.ProviderConfig{
					"ollama": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b"},
				},
				Profiles: map[string]string{
					"chat": "ollama",
				},
			},
		},
	}

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"hello"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusServiceUnavailable)
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["ok"] != false {
		t.Fatalf("ok = %v, want false", resp["ok"])
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["code"] != cognitive.ExecutionProviderDisabled {
		t.Fatalf("code = %v, want %s", data["code"], cognitive.ExecutionProviderDisabled)
	}
	if data["setup_required"] != true {
		t.Fatalf("setup_required = %v, want true", data["setup_required"])
	}
}

func TestHandleChat_AnswersSearchCapabilityWithoutNATS(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderSearXNG, SearXNGEndpoint: "http://searxng.local"}, nil, nil)
	})

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"are you able to make web requests without brave tokens?"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var resp protocol.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("marshal response data: %v", err)
	}
	var envelope protocol.CTSEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if !strings.Contains(payload.Text, "SearXNG") && !strings.Contains(payload.Text, "searxng") {
		t.Fatalf("payload.text = %q, want searxng capability", payload.Text)
	}
	if !strings.Contains(payload.Text, "Hosted Brave tokens are not required") {
		t.Fatalf("payload.text = %q, want token-free path", payload.Text)
	}
}

func TestHandleChat_ReturnsStructuredBlockerForEmptyAgentAnswer(t *testing.T) {
	wireNATS := withNATS(t)
	s := newTestServer(wireNATS)
	s.Cognitive = &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{
				"mock": {Type: "mock", Enabled: true, ModelID: "test-model"},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"mock": cognitiveTestProvider{},
		},
	}

	subject := "swarm.council.admin.request"
	_, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
		resp, _ := json.Marshal(map[string]any{
			"text": "",
			"availability": map[string]any{
				"available": false,
				"code":      "empty_provider_output",
				"summary":   "provider returned no readable content",
			},
			"provider_id": "mock",
			"model_used":  "test-model",
		})
		msg.Respond(resp)
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"hello"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusBadGateway, rr.Body.String())
	}

	var resp protocol.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.OK {
		t.Fatal("expected ok=false")
	}
	if resp.Error != "provider returned no readable content" {
		t.Fatalf("error = %q, want provider blocker summary", resp.Error)
	}
}

func TestHandleChat_ReturnsStructuredBlockerForEmptyAgentEnvelope(t *testing.T) {
	wireNATS := withNATS(t)
	s := newTestServer(wireNATS)
	s.Cognitive = &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{
				"mock": {Type: "mock", Enabled: true, ModelID: "test-model"},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"mock": cognitiveTestProvider{},
		},
	}

	subject := "swarm.council.admin.request"
	_, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
		msg.Respond([]byte(`{"provider_id":"mock","model_used":"test-model"}`))
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"hello"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusBadGateway, rr.Body.String())
	}

	var resp protocol.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.OK {
		t.Fatal("expected ok=false")
	}
	if resp.Error != "Soma could not produce a readable reply for that request." {
		t.Fatalf("error = %q, want structured blocker summary", resp.Error)
	}
}
