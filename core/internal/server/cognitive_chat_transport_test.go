package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestHandleChat_UnwrapsReadableJSONEnvelopeFromAgent(t *testing.T) {
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
		msg.Respond([]byte(`{"message":"Council-Sentry protects the runtime and reviews operational risk."}`))
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"Introduce sentry."}]}`)
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
	if payload.Text != "Council-Sentry protects the runtime and reviews operational risk." {
		t.Fatalf("payload.text = %q", payload.Text)
	}
}

func TestHandleChat_ReturnsStructuredTransportBlockerWhenAdminHasNoResponder(t *testing.T) {
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

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"hello"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusServiceUnavailable, rr.Body.String())
	}

	var resp protocol.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != "Soma is currently unreachable from the workspace runtime." {
		t.Fatalf("error = %q", resp.Error)
	}
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["code"] != "transport_unavailable" {
		t.Fatalf("code = %v", data["code"])
	}
}

func TestBuildTransportChatBlocker_ClassifiesTimeout(t *testing.T) {
	status, blocker := buildTransportChatBlocker("Soma", context.DeadlineExceeded)

	if status != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d", status, http.StatusGatewayTimeout)
	}
	if blocker.Code != "transport_timeout" {
		t.Fatalf("code = %q, want transport_timeout", blocker.Code)
	}
	if blocker.Summary != "Soma did not respond before the request deadline." {
		t.Fatalf("summary = %q", blocker.Summary)
	}
}

func TestBuildTransportChatBlocker_ClassifiesBackpressure(t *testing.T) {
	status, blocker := buildTransportChatBlocker("Soma", errors.New("nats: outbound buffer limit exceeded"))

	if status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", status, http.StatusServiceUnavailable)
	}
	if blocker.Code != "transport_backpressure" {
		t.Fatalf("code = %q, want transport_backpressure", blocker.Code)
	}
	if blocker.Summary != "Soma is overloaded right now and could not process the request." {
		t.Fatalf("summary = %q", blocker.Summary)
	}
}

func TestBuildTransportChatBlocker_ClassifiesNoResponders(t *testing.T) {
	status, blocker := buildTransportChatBlocker("Soma", nats.ErrNoResponders)

	if status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", status, http.StatusServiceUnavailable)
	}
	if blocker.Code != "transport_unavailable" {
		t.Fatalf("code = %q, want transport_unavailable", blocker.Code)
	}
	if blocker.Summary != "Soma is currently unreachable from the workspace runtime." {
		t.Fatalf("summary = %q", blocker.Summary)
	}
}

func TestHandleChat_BlocksUnreadableStructuredReplyAfterRetry(t *testing.T) {
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
	replyCount := 0
	_, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
		replyCount++
		msg.Respond([]byte(`{"status":"ok","details":{"member":"council-sentry"}}`))
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"tell sentry to show me their context"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusBadGateway, rr.Body.String())
	}
	if replyCount != 2 {
		t.Fatalf("replyCount = %d, want 2", replyCount)
	}

	var resp protocol.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != "Soma drifted into a governed action while answering a read-only request. Retry the request or restate it more directly." {
		t.Fatalf("error = %q", resp.Error)
	}
}
