package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
	"gopkg.in/yaml.v3"
)

// S7b item 1: approved-plan calls are scoped to the originating agent's
// declared tools at plan time (never proposed) and at execution time (never run).

// shippedSomaForTest runs the shipped admin-core team (core/config/teams/
// admin.yaml), so chat tests plan against Soma's real declared tools.
func shippedSomaForTest(t *testing.T) *swarm.Soma {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "config", "teams", "admin.yaml"))
	if err != nil {
		t.Fatalf("read admin.yaml: %v", err)
	}
	var manifest swarm.TeamManifest
	if err := yaml.Unmarshal(raw, &manifest); err != nil || manifest.ID != somaOriginTeamID {
		t.Fatalf("parse admin.yaml: id=%q err=%v", manifest.ID, err)
	}
	return swarm.NewTestSoma([]*swarm.TeamManifest{&manifest})
}

func planScopeSoma(adminTools, architectTools []string) *swarm.Soma {
	return swarm.NewTestSoma([]*swarm.TeamManifest{
		{ID: "admin-core", Members: []protocol.AgentManifest{{ID: "admin", Role: "admin", Tools: adminTools}}},
		{ID: "council-core", Members: []protocol.AgentManifest{{ID: "council-architect", Role: "architect", Tools: architectTools}}},
	})
}

// planScopeServer wires a NATS responder that answers the agent request on
// subject with the given model reply text.
func planScopeServer(t *testing.T, soma *swarm.Soma, subject, replyText string) *AdminServer {
	t.Helper()
	s := newTestServer(withNATS(t))
	s.Soma = soma
	s.Cognitive = &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles:  map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{"mock": {Type: "mock", Enabled: true, ModelID: "test-model"}},
		},
		Adapters: map[string]cognitive.LLMProvider{"mock": cognitiveTestProvider{}},
	}
	if _, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
		resp, _ := json.Marshal(map[string]any{"text": replyText})
		msg.Respond(resp)
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	return s
}

func planScopeChat(t *testing.T, handler http.Handler, path string) (int, protocol.APIResponse, *protocol.ChatResponsePayload) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"messages": []map[string]string{{"role": "user", "content": "Notify the platform team about the outage."}}})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(body)))
	var resp protocol.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.OK {
		return rr.Code, resp, nil
	}
	raw, _ := json.Marshal(resp.Data)
	var envelope protocol.CTSEnvelope
	_ = json.Unmarshal(raw, &envelope)
	var payload protocol.ChatResponsePayload
	_ = json.Unmarshal(envelope.Payload, &payload)
	return rr.Code, resp, &payload
}

func blockerCode(resp protocol.APIResponse) string {
	data, _ := resp.Data.(map[string]any)
	code, _ := data["code"].(string)
	return code
}

const (
	undeclaredBroadcastReply = `{"tool_call":{"name":"broadcast","arguments":{"message":"all teams stop"}}}`
	undeclaredMCPReply       = `{"tool_call":{"tool_ref":"mcp:github/create_issue","arguments":{"title":"x"}}}`
	declaredDelegateReply    = `{"tool_call":{"name":"delegate_task","arguments":{"team_id":"platform","task":"Share the outage notice."}}}`
)

func TestSomaChatNeverProposesUndeclaredModelToolCall(t *testing.T) {
	for name, reply := range map[string]string{"internal": undeclaredBroadcastReply, "mcp": undeclaredMCPReply} {
		t.Run(name, func(t *testing.T) {
			s := planScopeServer(t, planScopeSoma([]string{"delegate_task", "mcp:filesystem/*"}, nil), "swarm.council.admin.request", reply)
			status, resp, payload := planScopeChat(t, http.HandlerFunc(s.HandleChat), "/api/v1/chat")
			if resp.OK || payload != nil {
				t.Fatalf("undeclared model tool call was proposed: status=%d payload=%+v", status, payload)
			}
			if status != http.StatusUnprocessableEntity || blockerCode(resp) != codePlannedToolNotDeclared {
				t.Fatalf("status=%d code=%q error=%q, want 422 %s", status, blockerCode(resp), resp.Error, codePlannedToolNotDeclared)
			}
		})
	}
}

func TestCouncilChatNeverProposesUndeclaredModelToolCall(t *testing.T) {
	s := planScopeServer(t, planScopeSoma(nil, []string{"delegate_task"}), "swarm.council.council-architect.request", undeclaredBroadcastReply)
	mux := setupMux(t, "POST /api/v1/council/{member}/chat", s.HandleCouncilChat)
	status, resp, payload := planScopeChat(t, mux, "/api/v1/council/council-architect/chat")
	if resp.OK || payload != nil || status != http.StatusUnprocessableEntity || blockerCode(resp) != codePlannedToolNotDeclared {
		t.Fatalf("council proposed an undeclared call: status=%d code=%q payload=%+v", status, blockerCode(resp), payload)
	}
}

func TestDeclaredModelToolCallIsStillProposed(t *testing.T) {
	s := planScopeServer(t, planScopeSoma([]string{"delegate_task"}, nil), "swarm.council.admin.request", declaredDelegateReply)
	status, resp, payload := planScopeChat(t, http.HandlerFunc(s.HandleChat), "/api/v1/chat")
	if status != http.StatusOK || payload == nil || payload.Proposal == nil {
		t.Fatalf("declared call not proposed: status=%d error=%q", status, resp.Error)
	}
	if !containsString(payload.Proposal.Tools, "delegate_task") {
		t.Fatalf("proposal tools = %v, want delegate_task", payload.Proposal.Tools)
	}
}

func TestModelToolCallWithoutResolvableOriginIsNotProposed(t *testing.T) {
	s := planScopeServer(t, nil, "swarm.council.admin.request", declaredDelegateReply)
	status, resp, payload := planScopeChat(t, http.HandlerFunc(s.HandleChat), "/api/v1/chat")
	if resp.OK || payload != nil || status != http.StatusServiceUnavailable || blockerCode(resp) != codePlannedToolScopeUnavailable {
		t.Fatalf("unverifiable plan proposed: status=%d code=%q", status, blockerCode(resp))
	}
}

func TestPlannedCallOriginTagging(t *testing.T) {
	parsed := buildPlannedToolCalls(chatAgentResult{Text: declaredDelegateReply}, "Notify the platform team.", []string{"delegate"})
	if len(parsed) != 1 || parsed[0].Origin != protocol.PlannedCallOriginAgent {
		t.Fatalf("model text call = %+v, want agent origin", parsed)
	}
	captured := buildPlannedToolCalls(chatAgentResult{PlannedToolCalls: []protocol.PlannedToolCall{{Name: "broadcast", Origin: protocol.PlannedCallOriginCore}}}, "x", nil)
	if len(captured) != 1 || captured[0].Origin != protocol.PlannedCallOriginAgent {
		t.Fatalf("agent-captured call = %+v, want agent origin (model cannot claim core)", captured)
	}
	inferred := buildPlannedToolCalls(chatAgentResult{}, "Create a file named notes/plan.md with the content hello.", []string{"write_file"})
	if len(inferred) == 0 {
		t.Fatal("expected a core-inferred write_file call")
	}
	for _, call := range inferred {
		if call.Origin != protocol.PlannedCallOriginCore {
			t.Fatalf("core-inferred call = %+v, want core origin", call)
		}
	}
}

func planScopeExecServer(t *testing.T, adminTools []string) (*AdminServer, *fakeProposalMCPExecutor, string) {
	t.Helper()
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	fsID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	fakeMCP := &fakeProposalMCPExecutor{serverID: fsID}
	soma := planScopeSoma(adminTools, nil)
	soma.SetMCPServerNames(map[uuid.UUID]string{fsID: "filesystem"})
	s := newTestServer(func(s *AdminServer) {
		s.MCPToolExecutor = fakeMCP
		s.Soma = soma
	})
	return s, fakeMCP, workspace
}

func agentScope(calls ...protocol.PlannedToolCall) *protocol.ScopeValidation {
	for i := range calls {
		calls[i].Origin = protocol.PlannedCallOriginAgent
	}
	return &protocol.ScopeValidation{OriginTeamID: "admin-core", OriginAgentID: "admin", PlannedToolCalls: calls}
}

func TestApprovedPlanUndeclaredAgentCallNeverExecutes(t *testing.T) {
	s, fakeMCP, workspace := planScopeExecServer(t, []string{"write_file", "mcp:filesystem/*"})
	scope := agentScope(
		protocol.PlannedToolCall{Name: "write_file", Arguments: map[string]any{"path": "notes/a.md", "content": "hello"}},
		protocol.PlannedToolCall{ToolRef: "mcp:github/create_issue", Arguments: map[string]any{"title": "x"}},
	)
	_, err := s.executePlannedToolCalls(t.Context(), scope, "user-1", "", "", "", "", false)
	if !swarm.IsToolNotPermitted(err) {
		t.Fatalf("err = %v, want not-permitted", err)
	}
	if len(fakeMCP.calls) != 0 {
		t.Fatalf("undeclared MCP call executed: %v", fakeMCP.calls)
	}
	if _, statErr := os.Stat(filepath.Join(workspace, "notes", "a.md")); !os.IsNotExist(statErr) {
		t.Fatal("a declared call ran before the plan's undeclared call was refused")
	}
}

func TestApprovedPlanDeclaredAgentCallsExecute(t *testing.T) {
	s, fakeMCP, workspace := planScopeExecServer(t, []string{"write_file", "mcp:filesystem/*"})
	scope := agentScope(
		protocol.PlannedToolCall{Name: "write_file", Arguments: map[string]any{"path": "notes/a.md", "content": "hello"}},
		protocol.PlannedToolCall{ToolRef: "mcp:filesystem/read_text_file", Arguments: map[string]any{"path": "notes/a.md"}},
	)
	results, err := s.executePlannedToolCalls(t.Context(), scope, "user-1", "", "", "", "", false)
	if err != nil || len(results) != 2 {
		t.Fatalf("declared plan = (%v, %v)", results, err)
	}
	if len(fakeMCP.calls) != 1 || fakeMCP.calls[0] != "read_text_file" {
		t.Fatalf("mcp calls = %v", fakeMCP.calls)
	}
	if data, _ := os.ReadFile(filepath.Join(workspace, "notes", "a.md")); string(data) != "hello" {
		t.Fatalf("written = %q", data)
	}
}

func TestApprovedPlanAgentCallRefusedWhenOriginRevoked(t *testing.T) {
	s, _, _ := planScopeExecServer(t, []string{"broadcast"})
	scope := agentScope(protocol.PlannedToolCall{Name: "write_file", Arguments: map[string]any{"path": "n.md", "content": "x"}})
	if _, err := s.executePlannedToolCalls(t.Context(), scope, "user-1", "", "", "", "", false); !swarm.IsToolNotPermitted(err) {
		t.Fatalf("revoked declaration err = %v", err)
	}
	scope.OriginAgentID = "ghost"
	if _, err := s.executePlannedToolCalls(t.Context(), scope, "user-1", "", "", "", "", false); !swarm.IsToolNotPermitted(err) {
		t.Fatalf("unknown origin err = %v", err)
	}
}

func TestApprovedPlanCoreCallsAreBoundedToCorePlanTools(t *testing.T) {
	s, fakeMCP, _ := planScopeExecServer(t, nil)
	for _, call := range []protocol.PlannedToolCall{
		{Name: "broadcast", Arguments: map[string]any{"message": "stop"}},
		{Name: "local_command", Origin: protocol.PlannedCallOriginCore},
		{ToolRef: "mcp:filesystem/read_text_file", Origin: protocol.PlannedCallOriginCore},
	} {
		scope := &protocol.ScopeValidation{PlannedToolCalls: []protocol.PlannedToolCall{call}}
		if _, err := s.executePlannedToolCalls(t.Context(), scope, "user-1", "", "", "", "", false); !swarm.IsToolNotPermitted(err) {
			t.Errorf("core call %+v err = %v, want not-permitted", call, err)
		}
	}
	if len(fakeMCP.calls) != 0 {
		t.Fatalf("core MCP call executed: %v", fakeMCP.calls)
	}
}

func TestApprovedPlanCreateTeamKeepsChildToolBound(t *testing.T) {
	s, _, _ := planScopeExecServer(t, []string{"create_team", "write_file"})
	scope := agentScope(protocol.PlannedToolCall{Name: "create_team", Arguments: map[string]any{"team_id": "kids", "tools": []any{"local_command"}}})
	_, err := s.executePlannedToolCalls(t.Context(), scope, "user-1", "", "", "", "", false)
	if err == nil || !strings.Contains(err.Error(), "create_team refused") || !strings.Contains(err.Error(), "local_command") {
		t.Fatalf("widening create_team err = %v, want child-tool refusal", err)
	}
}
