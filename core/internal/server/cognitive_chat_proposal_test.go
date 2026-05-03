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
