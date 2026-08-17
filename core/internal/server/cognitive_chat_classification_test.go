package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/somacommands"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

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

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"Summarize the delivery strategy for this workspace."}]}`)
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

func TestInferMutationToolsTreatsNeedResearchTeamAsBlueprintDelegation(t *testing.T) {
	tools := inferMutationToolsFromText("i need an indepth ai research team that can take on various aspects of current research to understand optimal agentry architecture")

	if !containsString(tools, "generate_blueprint") {
		t.Fatalf("tools = %v, want generate_blueprint", tools)
	}
	if !containsString(tools, "delegate") {
		t.Fatalf("tools = %v, want delegate", tools)
	}
}

func TestInferMutationToolsTreatsToolPostureGuidanceAsReadOnly(t *testing.T) {
	tests := []string{
		"Check available tools and walk me through enabling what is missing.",
		"show me currently configured tools",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if tools := inferMutationToolsFromText(input); len(tools) != 0 {
				t.Fatalf("tools = %#v, want no mutation tools", tools)
			}
		})
	}
}

func TestOutcomeTemplateManifestUserQuotesRouteConfigurationCommands(t *testing.T) {
	registry, err := somacommands.LoadDefault()
	if err != nil {
		t.Fatalf("load Soma command manifests: %v", err)
	}
	quotes := map[string]string{}
	for _, command := range registry.Commands {
		quotes[command.Handler] = command.UserQuote
	}

	tests := []struct {
		handler      string
		wantMutation []string
		wantReadOnly []string
	}{
		{handler: "preview_config_document", wantReadOnly: []string{"preview_config_document"}},
		{handler: "store_config_document", wantMutation: []string{"store_config_document"}},
		{handler: "activate_config_document", wantMutation: []string{"activate_config_document"}},
	}
	for _, tc := range tests {
		t.Run(tc.handler, func(t *testing.T) {
			quote := strings.TrimSpace(quotes[tc.handler])
			if quote == "" {
				t.Fatalf("manifest user_quote missing for %s", tc.handler)
			}
			if got := inferMutationToolsFromText(quote); !slices.Equal(got, tc.wantMutation) {
				t.Fatalf("inferMutationToolsFromText(%q) = %v, want %v", quote, got, tc.wantMutation)
			}
			if got := inferReadOnlyConfigToolsFromText(quote); !slices.Equal(got, tc.wantReadOnly) {
				t.Fatalf("inferReadOnlyConfigToolsFromText(%q) = %v, want %v", quote, got, tc.wantReadOnly)
			}
			messages, mutationTools := normalizeChatRequestMessages([]chatRequestMessage{{Role: "user", Content: quote}})
			if !slices.Equal(mutationTools, tc.wantMutation) {
				t.Fatalf("normalized mutation tools = %v, want %v", mutationTools, tc.wantMutation)
			}
			latest := latestUserMessageContent(messages)
			if len(tc.wantMutation) > 0 && !strings.Contains(latest, governedMutationRoutePrefix) {
				t.Fatalf("mutation quote did not use governed route: %q", latest)
			}
			if len(tc.wantReadOnly) > 0 && (!strings.Contains(latest, directAnswerRoutePrefix) || !strings.Contains(latest, "preview_config_document")) {
				t.Fatalf("preview quote did not retain read-only preview guidance: %q", latest)
			}
		})
	}
}

func TestOutcomeTemplateRequestsDoNotRouteToGenericFileOrRecurringTemplateTools(t *testing.T) {
	tests := map[string][]string{
		"Draft an Outcome Template as YAML for this request.": nil,
		"Save this reusable Outcome Template.":                {"store_config_document"},
		"Apply this recurring Outcome Template from now on.":  {"activate_config_document"},
	}
	for input, want := range tests {
		if got := inferMutationToolsFromText(input); !slices.Equal(got, want) {
			t.Fatalf("inferMutationToolsFromText(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestNormalizeRetryRequestUsesPriorUserIntent(t *testing.T) {
	messages := []chatRequestMessage{
		{Role: "user", Content: "create a research team and have them generate documentation"},
		{Role: "assistant", Content: "Soma hit a server-side failure while handling the request."},
		{Role: "user", Content: "try again"},
	}

	normalized := normalizeRetryRequest(messages)
	latest := latestUserMessageContent(normalized)

	if strings.Contains(latest, "try again") {
		t.Fatalf("latest retry request was not normalized: %q", latest)
	}
	if !strings.Contains(latest, "create a research team") {
		t.Fatalf("latest retry request = %q, want prior user intent", latest)
	}
	tools := inferMutationToolsFromText(latest)
	if !containsString(tools, "generate_blueprint") || !containsString(tools, "delegate") || containsString(tools, "write_file") {
		t.Fatalf("retry mutation tools = %v, want team proposal tools without write_file", tools)
	}
}
