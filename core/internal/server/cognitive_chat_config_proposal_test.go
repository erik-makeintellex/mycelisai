package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func TestHandleChat_ConfigStoreRemainsGovernedProposal(t *testing.T) {
	tests := []struct {
		name      string
		quote     string
		tool      string
		arguments map[string]any
	}{
		{name: "store", quote: "Save this outcome template.", tool: "store_config_document", arguments: map[string]any{
			"format": "yaml", "content": "model-authored-config",
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer(withNATS(t))
			var agentRequests atomic.Int32
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
				agentRequests.Add(1)
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
			if agentRequests.Load() != 1 {
				t.Fatalf("agent requests = %d, want 1 for source-less save", agentRequests.Load())
			}
		})
	}
}

func TestHandleChat_ConfigInlineStoreSkipsProviderInference(t *testing.T) {
	s := newTestServer(withNATS(t))
	var agentRequests atomic.Int32
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
		agentRequests.Add(1)
		_ = msg.Respond([]byte(`{"text":"unexpected provider request"}`))
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody, _ := json.Marshal(map[string]any{
		"messages": []chatRequestMessage{{Role: "user", Content: "Save this Worker Profile:\n" + retainedWorkerProfileYAML}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()
	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if agentRequests.Load() != 0 {
		t.Fatalf("agent requests = %d, want 0 for deterministic inline save", agentRequests.Load())
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
}

func TestConfigDocumentRequestBoundarySurvivesApprovedScopeRoundTrip(t *testing.T) {
	scope := protocol.ScopeValidation{ConfigRequestBoundary: configDocumentRequestBoundary(
		" org-1 ", " team-1 ", " operator-1 ",
	)}
	raw, err := json.Marshal(scope)
	if err != nil {
		t.Fatal(err)
	}
	var restored protocol.ScopeValidation
	if err := json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	boundary := restored.ConfigRequestBoundary
	if boundary == nil || boundary.OrganizationID != "org-1" || boundary.WorkspaceID != "org-1" ||
		boundary.TeamID != "team-1" || boundary.OperatorID != "operator-1" {
		t.Fatalf("restored request boundary = %#v", boundary)
	}
}
