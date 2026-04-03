package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

type cognitiveTestProvider struct{}

func (cognitiveTestProvider) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "ok", Provider: "mock", ModelUsed: "test-model"}, nil
}

func (cognitiveTestProvider) Probe(context.Context) (bool, error) {
	return true, nil
}

func TestCognitiveMatrix(t *testing.T) {
	// 1. Setup Mock AdminServer with Cognitive Engine
	cfg := &cognitive.BrainConfig{
		Profiles: map[string]string{"test": "mock"},
		Providers: map[string]cognitive.ProviderConfig{
			"mock": {Type: "mock", ModelID: "gpt-4"},
		},
	}

	// Create a partial AdminServer just for this handler
	s := &AdminServer{
		Cognitive: &cognitive.Router{
			Config: cfg,
		},
	}

	// 2. Create Request
	req, _ := http.NewRequest("GET", "/api/v1/cognitive/matrix", nil)
	rr := httptest.NewRecorder()

	// 3. Invoke Handler
	handler := http.HandlerFunc(s.HandleCognitiveConfig)
	handler.ServeHTTP(rr, req)

	// 4. Assertions
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var resp cognitive.BrainConfig
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if val, ok := resp.Profiles["test"]; !ok || val != "mock" {
		t.Errorf("unexpected profile config: %v", resp.Profiles)
	}
}

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

func TestHandleChat_UsesProposalModeForMutationIntentWithoutReadableText(t *testing.T) {
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
			"text":       "",
			"tools_used": []string{"write_file"},
			"availability": map[string]any{
				"available": false,
				"code":      "empty_provider_output",
				"summary":   "provider returned no readable content",
			},
		})
		msg.Respond(resp)
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"create workspace/test.txt"}]}`)
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
	if !resp.OK {
		t.Fatalf("expected ok=true, got error=%q", resp.Error)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("marshal response data: %v", err)
	}
	var envelope protocol.CTSEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Mode != protocol.ModeProposal {
		t.Fatalf("mode = %q, want %q", envelope.Mode, protocol.ModeProposal)
	}
	if envelope.TemplateID != protocol.TemplateChatToProposal {
		t.Fatalf("template = %q, want %q", envelope.TemplateID, protocol.TemplateChatToProposal)
	}

	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Proposal == nil {
		t.Fatal("expected proposal payload")
	}
	if payload.Text == "" {
		t.Fatal("expected non-empty proposal text fallback")
	}
}

func TestHandleChat_RetriesUnexpectedMutationForReadOnlyPrompt(t *testing.T) {
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
		var resp []byte
		if replyCount == 1 {
			resp, _ = json.Marshal(map[string]any{
				"text":        `{"tool_call":{"name":"broadcast","arguments":{"message":"ask everyone"}}}`,
				"tools_used":  []string{"broadcast"},
				"provider_id": "mock",
				"model_used":  "test-model",
			})
		} else {
			resp, _ = json.Marshal(map[string]any{
				"text":        "Workspace V8 centers Soma, governed execution, and visible review loops.",
				"provider_id": "mock",
				"model_used":  "test-model",
			})
		}
		msg.Respond(resp)
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"Summarize the current Workspace V8 design objectives."}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if replyCount != 2 {
		t.Fatalf("replyCount = %d, want 2", replyCount)
	}

	var resp protocol.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got error=%q", resp.Error)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("marshal response data: %v", err)
	}
	var envelope protocol.CTSEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Mode != protocol.ModeAnswer {
		t.Fatalf("mode = %q, want %q", envelope.Mode, protocol.ModeAnswer)
	}

	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Proposal != nil {
		t.Fatalf("did not expect proposal payload: %+v", payload.Proposal)
	}
	if !strings.Contains(payload.Text, "governed execution") {
		t.Fatalf("text = %q, want readable answer after retry", payload.Text)
	}
}

func TestHandleChat_ClassifiesArtifactAnswerAskClass(t *testing.T) {
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
			"text": "I created a deliverable for review.",
			"artifacts": []map[string]any{
				{"type": "document", "title": "Creative Brief", "content": "# Brief"},
			},
		})
		msg.Respond(resp)
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"Create a brief for this launch."}]}`)
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
	data, _ := json.Marshal(resp.Data)
	var envelope protocol.CTSEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.AskClass != protocol.AskClassGovernedArtifact {
		t.Fatalf("payload ask class = %q, want %q", payload.AskClass, protocol.AskClassGovernedArtifact)
	}
}

func TestHandleChat_ClassifiesConsultedAnswerAskClass(t *testing.T) {
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
			"text": "The architect reviewed the plan and here is the recommendation.",
			"consultations": []map[string]any{
				{"member": "council-architect", "summary": "Recommend the safer route."},
			},
		})
		msg.Respond(resp)
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"Review the architecture tradeoffs."}]}`)
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
	data, _ := json.Marshal(resp.Data)
	var envelope protocol.CTSEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.AskClass != protocol.AskClassSpecialist {
		t.Fatalf("payload ask class = %q, want %q", payload.AskClass, protocol.AskClassSpecialist)
	}
}

func TestHandleChat_BlocksUnexpectedMutationForReadOnlyPromptAfterRetry(t *testing.T) {
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
		resp, _ := json.Marshal(map[string]any{
			"text":        `{"tool_call":{"name":"broadcast","arguments":{"message":"ask everyone"}}}`,
			"tools_used":  []string{"broadcast"},
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

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"Summarize the current Workspace V8 design objectives."}]}`)
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
	if resp.OK {
		t.Fatal("expected ok=false")
	}
	if !strings.Contains(resp.Error, "read-only request") {
		t.Fatalf("error = %q, want drift blocker summary", resp.Error)
	}
}

func TestHandleChat_RoutesLatestMutationTurnToProposalAcrossThreadHistory(t *testing.T) {
	cases := []struct {
		name                 string
		messages             []chatRequestMessage
		wantMode             protocol.ExecutionMode
		wantProposalTool     string
		wantRouteHintApplied bool
	}{
		{
			name: "direct answer stays answer mode",
			messages: []chatRequestMessage{
				{Role: "user", Content: "Summarize the current Workspace V8 design objectives."},
			},
			wantMode:             protocol.ModeAnswer,
			wantRouteHintApplied: false,
		},
		{
			name: "mixed thread mutation still routes to proposal",
			messages: []chatRequestMessage{
				{Role: "user", Content: "Summarize the current Workspace V8 design objectives."},
				{Role: "assistant", Content: "Here is a readable answer."},
				{Role: "user", Content: "Create a simple python file named workspace/logs/qa_browser_mutation_test.py that prints hello world."},
			},
			wantMode:             protocol.ModeProposal,
			wantProposalTool:     "write_file",
			wantRouteHintApplied: true,
		},
		{
			name: "rephrased mutation after answer still routes to proposal",
			messages: []chatRequestMessage{
				{Role: "user", Content: "Summarize the current Workspace V8 design objectives."},
				{Role: "assistant", Content: "Here is a readable answer."},
				{Role: "user", Content: "Please write a new python file named workspace/logs/rephrased_mutation_test.py that prints hello world."},
			},
			wantMode:             protocol.ModeProposal,
			wantProposalTool:     "write_file",
			wantRouteHintApplied: true,
		},
		{
			name: "clean first turn mutation routes to proposal",
			messages: []chatRequestMessage{
				{Role: "user", Content: "Create a simple python file named workspace/logs/first_turn_mutation_test.py that prints hello world."},
			},
			wantMode:             protocol.ModeProposal,
			wantProposalTool:     "write_file",
			wantRouteHintApplied: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
			forwarded := make(chan []chatRequestMessage, 1)
			_, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
				var turns []chatRequestMessage
				if err := json.Unmarshal(msg.Data, &turns); err != nil {
					t.Errorf("decode forwarded messages: %v", err)
				} else {
					forwarded <- turns
				}
				resp, _ := json.Marshal(map[string]any{
					"text": "```python\nprint('hello world')\n```",
				})
				msg.Respond(resp)
			})
			if err != nil {
				t.Fatalf("subscribe: %v", err)
			}
			if err := s.NC.Flush(); err != nil {
				t.Fatalf("flush: %v", err)
			}

			reqBody, err := json.Marshal(map[string]any{"messages": tc.messages})
			if err != nil {
				t.Fatalf("marshal request: %v", err)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBuffer(reqBody))
			rr := httptest.NewRecorder()

			http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
			}

			var resp protocol.APIResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if !resp.OK {
				t.Fatalf("expected ok=true, got error=%q", resp.Error)
			}

			data, err := json.Marshal(resp.Data)
			if err != nil {
				t.Fatalf("marshal response data: %v", err)
			}
			var envelope protocol.CTSEnvelope
			if err := json.Unmarshal(data, &envelope); err != nil {
				t.Fatalf("decode envelope: %v", err)
			}
			if envelope.Mode != tc.wantMode {
				t.Fatalf("mode = %q, want %q", envelope.Mode, tc.wantMode)
			}

			var payload protocol.ChatResponsePayload
			if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if tc.wantMode == protocol.ModeProposal {
				if payload.Proposal == nil {
					t.Fatal("expected proposal payload")
				}
				if tc.wantProposalTool != "" {
					if len(payload.Proposal.Tools) == 0 {
						t.Fatal("expected proposed tools")
					}
					found := false
					for _, tool := range payload.Proposal.Tools {
						if tool == tc.wantProposalTool {
							found = true
							break
						}
					}
					if !found {
						t.Fatalf("proposal tools = %v, want to include %q", payload.Proposal.Tools, tc.wantProposalTool)
					}
					if !containsString(payload.ToolsUsed, tc.wantProposalTool) {
						t.Fatalf("chat payload tools_used = %v, want to include %q", payload.ToolsUsed, tc.wantProposalTool)
					}
				}
			} else {
				if payload.Proposal != nil {
					t.Fatalf("did not expect proposal payload: %+v", payload.Proposal)
				}
			}

			select {
			case turns := <-forwarded:
				if tc.wantRouteHintApplied {
					if len(turns) == 0 {
						t.Fatal("expected forwarded messages")
					}
					last := turns[len(turns)-1]
					if !strings.Contains(last.Content, governedMutationRoutePrefix) {
						t.Fatalf("last user content missing route hint: %q", last.Content)
					}
					if !strings.Contains(last.Content, tc.messages[len(tc.messages)-1].Content) {
						t.Fatalf("last user content lost original request: %q", last.Content)
					}
				}
			default:
				if tc.wantRouteHintApplied {
					t.Fatal("expected forwarded messages to be captured")
				}
			}
		})
	}
}

func TestResolveChatAskContract(t *testing.T) {
	direct := resolveChatAskContract("soma", false, chatAgentResult{})
	if direct.AskClass != protocol.AskClassDirectAnswer {
		t.Fatalf("direct ask class = %q, want %q", direct.AskClass, protocol.AskClassDirectAnswer)
	}
	if direct.TemplateID != protocol.TemplateChatToAnswer {
		t.Fatalf("direct template = %q, want %q", direct.TemplateID, protocol.TemplateChatToAnswer)
	}
	if direct.DefaultExecutionMode != protocol.ModeAnswer {
		t.Fatalf("direct mode = %q, want %q", direct.DefaultExecutionMode, protocol.ModeAnswer)
	}

	artifact := resolveChatAskContract("soma", false, chatAgentResult{
		Artifacts: []protocol.ChatArtifactRef{{Type: "document", Title: "Brief"}},
	})
	if artifact.AskClass != protocol.AskClassGovernedArtifact {
		t.Fatalf("artifact ask class = %q, want %q", artifact.AskClass, protocol.AskClassGovernedArtifact)
	}
	if artifact.DefaultExecutionMode != protocol.ModeAnswer {
		t.Fatalf("artifact mode = %q, want %q", artifact.DefaultExecutionMode, protocol.ModeAnswer)
	}

	specialist := resolveChatAskContract("specialist", false, chatAgentResult{})
	if specialist.AskClass != protocol.AskClassSpecialist {
		t.Fatalf("specialist ask class = %q, want %q", specialist.AskClass, protocol.AskClassSpecialist)
	}
	if specialist.DefaultAgentTarget != "specialist" {
		t.Fatalf("specialist default target = %q, want specialist", specialist.DefaultAgentTarget)
	}

	consulted := resolveChatAskContract("soma", false, chatAgentResult{
		Consultations: []protocol.ConsultationEntry{{Member: "council-architect", Summary: "Reviewed the plan."}},
	})
	if consulted.AskClass != protocol.AskClassSpecialist {
		t.Fatalf("consulted ask class = %q, want %q", consulted.AskClass, protocol.AskClassSpecialist)
	}

	mutation := resolveChatAskContract("soma", true, chatAgentResult{})
	if mutation.AskClass != protocol.AskClassGovernedMutation {
		t.Fatalf("mutation ask class = %q, want %q", mutation.AskClass, protocol.AskClassGovernedMutation)
	}
	if mutation.TemplateID != protocol.TemplateChatToProposal {
		t.Fatalf("mutation template = %q, want %q", mutation.TemplateID, protocol.TemplateChatToProposal)
	}
	if mutation.DefaultExecutionMode != protocol.ModeProposal {
		t.Fatalf("mutation mode = %q, want %q", mutation.DefaultExecutionMode, protocol.ModeProposal)
	}
	if !mutation.RequiresConfirmation {
		t.Fatal("mutation contract should require confirmation")
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
