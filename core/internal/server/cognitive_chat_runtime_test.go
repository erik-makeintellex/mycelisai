package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestHandleChat_ReturnsDeterministicRuntimeStateSummaryWithoutNATS(t *testing.T) {
	s := newTestServer()
	s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:   "org-1",
			Name: "Acme Org",
		},
		Departments: []OrganizationDepartmentSummary{
			{ID: "marketing-team", Name: "Marketing"},
		},
	})

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"what is your current state"}],"organization_id":"org-1","team_id":"marketing-team","team_name":"Marketing"}`)
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
	if payload.AskClass != protocol.AskClassDirectAnswer {
		t.Fatalf("ask class = %q, want %q", payload.AskClass, protocol.AskClassDirectAnswer)
	}
	if !strings.Contains(payload.Text, "Current Mycelis runtime state:") {
		t.Fatalf("text = %q, want runtime state summary", payload.Text)
	}
	if !strings.Contains(payload.Text, "Current organization focus: Acme Org.") {
		t.Fatalf("text = %q, want organization focus", payload.Text)
	}
	if !strings.Contains(payload.Text, "Current team focus: Marketing.") {
		t.Fatalf("text = %q, want team focus", payload.Text)
	}
}

func TestHandleChat_ReturnsDeterministicTeamRosterSummary(t *testing.T) {
	wireNATS := withNATS(t)
	s := newTestServer(wireNATS)
	s.Soma = swarm.NewTestSoma([]*swarm.TeamManifest{
		{
			ID:   "alpha-team",
			Name: "Alpha Team",
			Members: []protocol.AgentManifest{
				{ID: "alpha-lead"},
				{ID: "alpha-specialist"},
			},
		},
		{
			ID:   "beta-team",
			Name: "Beta Team",
			Members: []protocol.AgentManifest{
				{ID: "beta-lead"},
			},
		},
	})

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"what teams currently exist"}]}`)
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
	if !strings.Contains(payload.Text, "Active teams:") {
		t.Fatalf("text = %q, want team roster summary", payload.Text)
	}
	if !strings.Contains(payload.Text, "Alpha Team (2 members)") {
		t.Fatalf("text = %q, want alpha team summary", payload.Text)
	}
	if !strings.Contains(payload.Text, "Beta Team (1 member)") {
		t.Fatalf("text = %q, want beta team summary", payload.Text)
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
