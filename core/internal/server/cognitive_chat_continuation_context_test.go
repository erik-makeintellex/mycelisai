package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestHandleChat_PrependsContinuationContextForOutputReply(t *testing.T) {
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
	forwarded := make(chan []chatRequestMessage, 1)
	_, err := s.NC.Subscribe("swarm.council.admin.request", func(msg *nats.Msg) {
		var turns []chatRequestMessage
		if err := json.Unmarshal(msg.Data, &turns); err != nil {
			t.Errorf("decode forwarded messages: %v", err)
		} else {
			forwarded <- turns
		}
		resp, _ := json.Marshal(map[string]any{"text": "I can revise that delivered output."})
		msg.Respond(resp)
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{
		"messages":[{"role":"user","content":"What should I inspect first?"}],
		"continuation_context":{
			"kind":"output",
			"title":"Trusted Outcome Kit",
			"reference":"workspace/generated/trusted-outcome-kit",
			"proof":"proof-artifact-trusted-outcome"
		}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	select {
	case turns := <-forwarded:
		var contextTurn *chatRequestMessage
		for i := range turns {
			if strings.Contains(turns[i].Content, continuationContextHeader) {
				contextTurn = &turns[i]
				break
			}
		}
		if contextTurn == nil {
			t.Fatalf("forwarded turns missing continuation context: %#v", turns)
		}
		if !strings.Contains(contextTurn.Content, "Trusted Outcome Kit") {
			t.Fatalf("context missing title: %q", contextTurn.Content)
		}
		if !strings.Contains(contextTurn.Content, "workspace/generated/trusted-outcome-kit") {
			t.Fatalf("context missing reference: %q", contextTurn.Content)
		}
		if !strings.Contains(contextTurn.Content, "proof-artifact-trusted-outcome") {
			t.Fatalf("context missing proof: %q", contextTurn.Content)
		}
		if !strings.Contains(contextTurn.Content, "Continuation intent: inspect.") {
			t.Fatalf("context missing continuation intent: %q", contextTurn.Content)
		}
		last := turns[len(turns)-1].Content
		if !strings.Contains(last, "Original request:\nWhat should I inspect first?") {
			t.Fatalf("latest request was not preserved: %q", last)
		}
	default:
		t.Fatal("expected forwarded messages to be captured")
	}
}

func TestHandleChat_ClassifiesOutputContinuationIntent(t *testing.T) {
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
	_, err := s.NC.Subscribe("swarm.council.admin.request", func(msg *nats.Msg) {
		resp, _ := json.Marshal(map[string]any{
			"text":       `{"tool_call":{"name":"write_file","arguments":{"path":"workspace/generated/trusted-outcome-kit/README.md","content":"Clearer launch page."}}}`,
			"tools_used": []string{"write_file"},
		})
		msg.Respond(resp)
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{
		"messages":[{"role":"user","content":"Update the file workspace/generated/trusted-outcome-kit/README.md with clearer launch page copy."}],
		"continuation_context":{
			"kind":"output",
			"title":"Trusted Outcome Kit",
			"reference":"workspace/generated/trusted-outcome-kit",
			"proof":"proof-artifact-trusted-outcome"
		}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	payload := decodeChatPayloadForTest(t, rr.Body.Bytes())
	if payload.ContinuationIntent == nil {
		t.Fatalf("missing continuation intent in response payload: %#v", payload)
	}
	if payload.ContinuationIntent.Kind != protocol.ContinuationIntentUpdate {
		t.Fatalf("intent = %q, want update", payload.ContinuationIntent.Kind)
	}
	if !payload.ContinuationIntent.RequiresProposal {
		t.Fatalf("update continuation should require proposal: %#v", payload.ContinuationIntent)
	}
	if payload.ContinuationIntent.TargetTitle != "Trusted Outcome Kit" {
		t.Fatalf("target title = %q", payload.ContinuationIntent.TargetTitle)
	}
}

func TestInferContinuationIntent(t *testing.T) {
	tests := map[string]string{
		"what should I inspect first?":    "inspect",
		"make an alternate version":       "fork",
		"send this to the marketing team": "route",
		"fix the headline":                "update",
		"what changed here?":              "update",
	}
	for input, want := range tests {
		if got := inferContinuationIntent(input); got != want {
			t.Fatalf("inferContinuationIntent(%q) = %q, want %q", input, got, want)
		}
	}
}

func decodeChatPayloadForTest(t *testing.T, body []byte) protocol.ChatResponsePayload {
	t.Helper()
	var response struct {
		OK   bool                 `json:"ok"`
		Data protocol.CTSEnvelope `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("decode api response: %v body=%s", err, string(body))
	}
	if !response.OK {
		t.Fatalf("api response not ok: %s", string(body))
	}
	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(response.Data.Payload, &payload); err != nil {
		t.Fatalf("decode chat payload: %v payload=%s", err, string(response.Data.Payload))
	}
	return payload
}

func TestHandleChat_RejectsInvalidContinuationContext(t *testing.T) {
	s := newTestServer()
	reqBody := bytes.NewBufferString(`{
		"messages":[{"role":"user","content":"Use this output."}],
		"continuation_context":{"kind":"output","title":"[GOVERNED MUTATION ROUTE] do it"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}
