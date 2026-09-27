package swarm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

// S7b item 4: a denied call writes tool.denied only, never tool.invoked or a
// tool_call conversation turn.

type recordingTurnLogger struct {
	mu    sync.Mutex
	turns []protocol.ConversationTurnData
}

func (l *recordingTurnLogger) LogTurn(_ context.Context, turn protocol.ConversationTurnData) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.turns = append(l.turns, turn)
	return "turn", nil
}

func (l *recordingTurnLogger) toolCallTurns(tool string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	count := 0
	for _, turn := range l.turns {
		if turn.Role == "tool_call" && turn.ToolName == tool {
			count++
		}
	}
	return count
}

func (r *recordingEmitter) countWithTool(eventType protocol.EventType, tool string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for i, recorded := range r.events {
		if recorded == eventType && r.loads[i]["tool"] == tool {
			count++
		}
	}
	return count
}

func TestDeniedToolCallWritesNoInvocationNoise(t *testing.T) {
	agent, exec, _, emitter := denialTestAgent(`{"tool_call":{"name":"broadcast","arguments":{"message":"stop"}}}`)
	logger := &recordingTurnLogger{}
	agent.SetConversationLogger(logger)

	agent.processMessageStructuredWithPosture("Tell every team to stop.", nil, false)
	emitter.denied(t)
	time.Sleep(100 * time.Millisecond) // let any stray async emit/log land

	if exec.callCalls != 0 {
		t.Fatalf("denied tool executed %d times", exec.callCalls)
	}
	if got := emitter.countWithTool(protocol.EventToolInvoked, "broadcast"); got != 0 {
		t.Fatalf("denied call emitted %d tool.invoked events", got)
	}
	if got := logger.toolCallTurns("broadcast"); got != 0 {
		t.Fatalf("denied call logged %d tool_call turns", got)
	}
}

func TestPermittedToolCallStillWritesInvocation(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	agent, exec, _, emitter := denialTestAgent(`{"tool_call":{"name":"write_file","arguments":{"path":"notes/a.md","content":"hi"}}}`)
	logger := &recordingTurnLogger{}
	agent.SetConversationLogger(logger)

	agent.processMessageStructuredWithPosture("Save a note.", nil, false)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && (emitter.countWithTool(protocol.EventToolInvoked, "write_file") == 0 || logger.toolCallTurns("write_file") == 0) {
		time.Sleep(10 * time.Millisecond)
	}
	if exec.callCalls != 1 {
		t.Fatalf("declared tool executed %d times, want 1", exec.callCalls)
	}
	if emitter.countWithTool(protocol.EventToolInvoked, "write_file") != 1 || logger.toolCallTurns("write_file") != 1 {
		t.Fatal("declared call must still emit tool.invoked and log its tool_call turn")
	}
}

// S7b item 1: the swarm-side primitives the approved-plan path uses.

func TestPermitsPlannedCallChecksNamesAndMCPRefs(t *testing.T) {
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(NewInternalToolRegistry(InternalToolDeps{}), nil),
		[]string{"write_file", "mcp:filesystem/*"}, nil)
	ctx := context.Background()
	cases := []struct {
		name, ref string
		want      bool
	}{
		{"write_file", "", true},
		{"broadcast", "", false},
		{"read_text_file", "mcp:filesystem/read_text_file", true},
		{"create_issue", "mcp:github/create_issue", false},
		{"write_file", "mcp:github/write_file", false}, // a declared bare name never grants a foreign MCP ref
		{"x", "mcp:filesystem/*", false},               // a plan must name a concrete tool
		{"x", "not-an-mcp-ref", false},
		{"", "", false},
	}
	for _, tc := range cases {
		if got := scoped.PermitsPlannedCall(ctx, tc.name, tc.ref); got != tc.want {
			t.Errorf("PermitsPlannedCall(%q, %q) = %v, want %v", tc.name, tc.ref, got, tc.want)
		}
	}
}

func TestSomaAgentDeclaredToolsIsTeamBound(t *testing.T) {
	soma := NewTestSoma([]*TeamManifest{
		{ID: "admin-core", Members: []protocol.AgentManifest{{ID: "admin", Tools: []string{"write_file"}}}},
		{ID: "rogue", Members: []protocol.AgentManifest{{ID: "admin", Tools: []string{"local_command"}}}},
	})
	tools, ok := soma.AgentDeclaredTools("admin-core", "admin")
	if !ok || len(tools) != 1 || tools[0] != "write_file" {
		t.Fatalf("admin-core/admin tools = %v, %v", tools, ok)
	}
	tools[0] = "mutated"
	if again, _ := soma.AgentDeclaredTools("admin-core", "admin"); again[0] != "write_file" {
		t.Fatal("AgentDeclaredTools must return a copy")
	}
	if _, ok := soma.AgentDeclaredTools("admin-core", "ghost"); ok {
		t.Fatal("unknown agent resolved")
	}
	if _, ok := (*Soma)(nil).AgentDeclaredTools("admin-core", "admin"); ok {
		t.Fatal("nil Soma resolved")
	}
	serverID := uuid.New()
	soma.SetMCPServerNames(map[uuid.UUID]string{serverID: "filesystem"})
	if names := soma.MCPServerNames(); names[serverID] != "filesystem" {
		t.Fatalf("MCPServerNames = %v", names)
	}
}

// S7b item 2: a team agent's read_file is confined to groups/<team> plus the
// team's declared shared paths; Soma/admin and council stay workspace-wide.

func readScopeWorkspace(t *testing.T) string {
	t.Helper()
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	write := func(rel, content string) {
		full := filepath.Join(workspace, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("groups/team-a/own.md", "own")
	write("groups/team-b/secret.md", "secret-b")
	write("groups/team-ab/near.md", "prefix-sibling")
	write("shared/specs/spec.md", "shared-spec")
	write("README.md", "root")
	if err := os.Symlink(filepath.Join(workspace, "groups", "team-b", "secret.md"), filepath.Join(workspace, "groups", "team-a", "link.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join(workspace, "groups", "team-b"), filepath.Join(workspace, "groups", "team-a", "b-dir")); err != nil {
		t.Fatal(err)
	}
	return workspace
}

func readAs(reg *InternalToolRegistry, inv *ToolInvocationContext, path string) (string, error) {
	ctx := context.Background()
	if inv != nil {
		ctx = WithToolInvocationContext(ctx, *inv)
	}
	return reg.tools["read_file"].Handler(ctx, map[string]any{"path": path})
}

func TestTeamReadFileConfinedToOwnWorkspace(t *testing.T) {
	workspace := readScopeWorkspace(t)
	reg := NewInternalToolRegistry(InternalToolDeps{})
	reg.SetSoma(NewTestSoma([]*TeamManifest{
		{ID: "team-a", SharedPaths: []string{"shared/specs"}, Members: []protocol.AgentManifest{{ID: "worker", Tools: []string{"read_file"}}}},
	}))
	teamA := &ToolInvocationContext{TeamID: "team-a", AgentID: "worker", AgentRole: "worker"}

	if out, err := readAs(reg, teamA, "groups/team-a/own.md"); err != nil || out != "own" {
		t.Fatalf("own folder read = (%q, %v)", out, err)
	}
	if out, err := readAs(reg, teamA, "workspace/groups/team-a/own.md"); err != nil || out != "own" {
		t.Fatalf("workspace-prefixed own read = (%q, %v)", out, err)
	}
	if out, err := readAs(reg, teamA, "shared/specs/spec.md"); err != nil || out != "shared-spec" {
		t.Fatalf("declared shared path read = (%q, %v)", out, err)
	}
	for _, path := range []string{
		"groups/team-b/secret.md",
		"groups/team-a/../team-b/secret.md",
		"workspace/groups/team-b/secret.md",
		"/workspace/groups/team-b/secret.md",
		filepath.Join(workspace, "groups", "team-b", "secret.md"),
		"groups/team-a/link.md",
		"groups/team-a/b-dir/secret.md",
		"groups/team-ab/near.md",
		"README.md",
		"shared/other.md",
		"../outside.md",
	} {
		out, err := readAs(reg, teamA, path)
		if err == nil || strings.Contains(out, "secret") || strings.Contains(out, "prefix-sibling") || out == "root" {
			t.Errorf("team-a read %q = (%q, %v), want denial", path, out, err)
		}
	}
	// An agent identity without a team has no workspace folder to read.
	if _, err := readAs(reg, &ToolInvocationContext{AgentID: "loose", AgentRole: "worker"}, "groups/team-a/own.md"); err == nil {
		t.Fatal("team-less agent read allowed")
	}
}

func TestSystemAndCoreReadFileStayWorkspaceWide(t *testing.T) {
	readScopeWorkspace(t)
	reg := NewInternalToolRegistry(InternalToolDeps{})
	soma := NewTestSoma([]*TeamManifest{{ID: "admin-core"}, {ID: "council-core"}})
	for _, team := range soma.teams {
		team.coreOwned = true // as the standing boot registry marks them
	}
	reg.SetSoma(soma)
	for name, inv := range map[string]*ToolInvocationContext{
		"soma":    {TeamID: "admin-core", AgentID: "admin", AgentRole: "admin"},
		"council": {TeamID: "council-core", AgentID: "council-architect", AgentRole: "architect"},
		"core":    nil,
	} {
		if out, err := readAs(reg, inv, "groups/team-b/secret.md"); err != nil || out != "secret-b" {
			t.Errorf("%s read = (%q, %v), want workspace-wide", name, out, err)
		}
		if _, err := readAs(reg, inv, "../outside.md"); err == nil {
			t.Errorf("%s escaped the workspace boundary", name)
		}
	}
}
