package swarm

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// preflightTestAgent runs an agent over a fake registry whose consult_council
// handler stands in for the council bus.
func preflightTestAgent(t *testing.T, declared []string) (*Agent, map[string]int32Counter, *scriptedDenialProvider, *recordingEmitter) {
	t.Helper()
	reg, counts := scopeTestRegistry(t, "consult_council", "local_command", "write_file")
	provider := &scriptedDenialProvider{replies: []string{`{"tool_call":{"name":"local_command","arguments":{"command":"cat /etc/hosts"}}}`}}
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles:  map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{"mock": {Type: "mock", Enabled: true, ModelID: "test-model"}},
		},
		Adapters: map[string]cognitive.LLMProvider{"mock": provider},
	}
	agent := NewAgent(context.Background(), agentManifestWithTools("worker", declared), "team-a", nil, router, NewCompositeToolExecutor(reg, nil))
	emitter := &recordingEmitter{}
	agent.SetEventEmitter(emitter, "run-1")
	out := map[string]int32Counter{}
	for name, count := range counts {
		out[name] = count
	}
	return agent, out, provider, emitter
}

type int32Counter interface{ Load() int32 }

func TestUndeclaredPreflightToolNeverReachesCouncil(t *testing.T) {
	agent, counts, provider, emitter := preflightTestAgent(t, []string{"write_file"})

	agent.processMessageStructuredWithPosture("Show the hosts file.", nil, false)

	if got := counts["consult_council"].Load(); got != 0 {
		t.Fatalf("council contacted %d times for an undeclared local_command", got)
	}
	if got := counts["local_command"].Load(); got != 0 {
		t.Fatalf("undeclared local_command executed %d times", got)
	}
	if payload := emitter.denied(t); payload["tool"] != "local_command" || payload["phase"] != "lookup" {
		t.Fatalf("tool.denied payload = %#v", payload)
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	last := provider.prompts[len(provider.prompts)-1]
	if !strings.Contains(last, "is not permitted for this agent and was not executed") || strings.Contains(last, "Preflight (") {
		t.Fatal("model did not get the plain denial, or saw council preflight output")
	}
}

func TestDeclaredPreflightToolStillConsultsCouncil(t *testing.T) {
	agent, counts, _, _ := preflightTestAgent(t, []string{"local_command"})

	agent.processMessageStructuredWithPosture("Show the hosts file.", nil, false)

	if got := counts["consult_council"].Load(); got != 1 {
		t.Fatalf("declared local_command preflight consulted council %d times, want 1", got)
	}
}

func childManifest(tools ...string) *TeamManifest {
	return &TeamManifest{ID: "child", Members: []protocol.AgentManifest{{ID: "child-agent", Tools: tools}}}
}

func TestCreateTeamChildToolsMustStayWithinCaller(t *testing.T) {
	cases := []struct {
		parent, child []string
		denied        string
	}{
		{parent: []string{"store_artifact"}, child: []string{"store_artifact"}},
		{parent: []string{"store_artifact"}, child: []string{"local_command"}, denied: "local_command"},
		{parent: []string{"store_artifact"}, child: []string{"mcp:filesystem/*"}, denied: "mcp:filesystem/*"},
		{parent: []string{"store_artifact"}, child: []string{"mcp:filesystem/read_file"}, denied: "mcp:filesystem/read_file"},
		{parent: []string{"store_artifact"}, child: []string{"toolset:workspace"}, denied: "toolset:workspace"},
		{parent: []string{"mcp:filesystem/read_file"}, child: []string{"mcp:filesystem/*"}, denied: "mcp:filesystem/*"},
		{parent: []string{"mcp:filesystem/*"}, child: []string{"mcp:filesystem/read_file", "mcp:filesystem/*"}},
		{parent: []string{"toolset:workspace"}, child: []string{"toolset:workspace"}},
	}
	for _, tc := range cases {
		ctx := withCallerToolScope(context.Background(), newAgentToolScope(tc.parent, nil))
		err := childToolsWithinCaller(ctx, childManifest(tc.child...))
		if tc.denied == "" && err != nil {
			t.Fatalf("parent %v child %v: unexpected refusal %v", tc.parent, tc.child, err)
		}
		if tc.denied != "" && (err == nil || !strings.Contains(err.Error(), "not permitted: "+tc.denied)) {
			t.Fatalf("parent %v child %v: err = %v, want refusal naming %s", tc.parent, tc.child, err, tc.denied)
		}
	}
	// No agent scope (Core restore, operator-approved plan): not narrowed here.
	if err := childToolsWithinCaller(context.Background(), childManifest("local_command")); err != nil {
		t.Fatalf("unscoped create_team refused: %v", err)
	}
}

func TestScopedCallToolHandsCallerScopeToCreateTeam(t *testing.T) {
	reg := &InternalToolRegistry{tools: map[string]*InternalTool{}}
	reg.tools["create_team"] = &InternalTool{Name: "create_team", Handler: func(ctx context.Context, args map[string]any) (string, error) {
		if err := childToolsWithinCaller(ctx, buildRuntimeTeamManifest(args)); err != nil {
			return "", err
		}
		return "created", nil
	}}
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(reg, nil), []string{"create_team", "store_artifact"}, nil)

	_, err := scoped.CallTool(context.Background(), InternalServerID, "create_team", map[string]any{"team_id": "kids", "tools": []any{"local_command", "mcp:fetch/*"}})
	if err == nil || !strings.Contains(err.Error(), "not permitted: local_command, mcp:fetch/*") {
		t.Fatalf("widening create_team err = %v", err)
	}
	if out, err := scoped.CallTool(context.Background(), InternalServerID, "create_team", map[string]any{"team_id": "kids", "tools": []any{"store_artifact"}}); err != nil || out != "created" {
		t.Fatalf("in-scope create_team = (%q, %v)", out, err)
	}
}

func TestDeniedToolNameIsBounded(t *testing.T) {
	long := strings.Repeat("x", 300)
	for _, text := range []string{toolDeniedFeedback(long), (&ToolNotPermittedError{Tool: long}).Error()} {
		if strings.Contains(text, strings.Repeat("x", 129)) || !strings.Contains(text, strings.Repeat("x", 128)+"…") {
			t.Fatalf("tool name not bounded to 128 runes: %d runes", utf8.RuneCountInString(text))
		}
	}
	if got := boundedToolName(" broadcast "); got != "broadcast" {
		t.Fatalf("short name changed: %q", got)
	}
}
