package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestHandleChat_ConfigMutationsRemainGovernedProposals(t *testing.T) {
	tests := []struct {
		name      string
		quote     string
		tool      string
		arguments map[string]any
	}{
		{name: "store", quote: "Save this outcome template.", tool: "store_config_document", arguments: map[string]any{
			"format": "yaml", "content": "model-authored-config",
		}},
		{name: "activate", quote: "Use this outcome template.", tool: "activate_config_document", arguments: map[string]any{
			"record_id": "11111111-1111-1111-1111-111111111111", "action": "activate",
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer(withNATS(t))
			s.Cognitive = &cognitive.Router{
				Config: &cognitive.BrainConfig{
					Profiles: map[string]string{"chat": "mock"},
					Providers: map[string]cognitive.ProviderConfig{
						"mock": {Type: "mock", Enabled: true, ModelID: "test-model"},
					},
				},
				Adapters: map[string]cognitive.LLMProvider{"mock": cognitiveTestProvider{}},
			}

			_, err := s.NC.Subscribe("swarm.council.admin.request", func(msg *nats.Msg) {
				response, _ := json.Marshal(chatAgentResult{
					Text:      "Soma prepared the requested configuration action.",
					ToolsUsed: []string{tc.tool},
					PlannedToolCalls: []protocol.PlannedToolCall{{
						Name: tc.tool, Arguments: tc.arguments,
					}},
				})
				_ = msg.Respond(response)
			})
			if err != nil {
				t.Fatalf("subscribe: %v", err)
			}
			if err := s.NC.Flush(); err != nil {
				t.Fatalf("flush: %v", err)
			}

			reqBody, _ := json.Marshal(map[string]any{
				"messages": []chatRequestMessage{{Role: "user", Content: tc.quote}},
			})
			req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(reqBody))
			rr := httptest.NewRecorder()
			http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
			}
			var response protocol.APIResponse
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			data, _ := json.Marshal(response.Data)
			var envelope protocol.CTSEnvelope
			if err := json.Unmarshal(data, &envelope); err != nil {
				t.Fatalf("decode envelope: %v", err)
			}
			if envelope.Mode != protocol.ModeProposal {
				t.Fatalf("mode = %q, want proposal", envelope.Mode)
			}
			var payload protocol.ChatResponsePayload
			if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload.Proposal == nil || !containsString(payload.Proposal.Tools, tc.tool) {
				t.Fatalf("proposal = %#v, want tool %s", payload.Proposal, tc.tool)
			}
			if payload.ExecutionSummary == nil || payload.ExecutionSummary.NextStep == nil || payload.ExecutionSummary.NextStep.Action != "confirm_action" {
				t.Fatalf("execution summary = %#v, want confirm_action", payload.ExecutionSummary)
			}
			if payload.Proposal.RiskLevel != "medium" {
				t.Fatalf("proposal risk = %q, want medium", payload.Proposal.RiskLevel)
			}
		})
	}
}
