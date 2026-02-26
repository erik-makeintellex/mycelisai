package swarm

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/mcp"
)

// mockMCPExecutor simulates an MCP tool executor for testing.
type mockMCPExecutor struct {
	tools map[string]uuid.UUID // tool name → server ID
}

func (m *mockMCPExecutor) FindToolByName(_ context.Context, name string) (uuid.UUID, string, error) {
	if id, ok := m.tools[name]; ok {
		return id, name, nil
	}
	return uuid.Nil, "", nil
}

func (m *mockMCPExecutor) CallTool(_ context.Context, _ uuid.UUID, toolName string, _ map[string]any) (string, error) {
	return "result:" + toolName, nil
}

func TestScopedToolExecutor_AllowAll(t *testing.T) {
	// No mcp: refs → allowAll = true, all MCP tools permitted
	fsServerID := uuid.New()
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{"read_file": fsServerID}}
	composite := NewCompositeToolExecutor(nil, mock)
	scoped := NewScopedToolExecutor(composite, nil, map[uuid.UUID]string{fsServerID: "filesystem"})

	if !scoped.allowAll {
		t.Fatal("expected allowAll=true when no mcp refs")
	}

	serverID, toolName, err := scoped.FindToolByName(context.Background(), "read_file")
	if err != nil {
		t.Fatalf("FindToolByName: %v", err)
	}
	if serverID != fsServerID {
		t.Errorf("got serverID=%v, want %v", serverID, fsServerID)
	}
	if toolName != "read_file" {
		t.Errorf("got toolName=%q, want read_file", toolName)
	}
}

func TestScopedToolExecutor_FilteredAllow(t *testing.T) {
	fsServerID := uuid.New()
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{
		"read_file":  fsServerID,
		"write_file": fsServerID,
	}}
	composite := NewCompositeToolExecutor(nil, mock)

	// Allow only read_file from filesystem
	refs := []mcp.ToolRef{{ServerName: "filesystem", ToolName: "read_file"}}
	scoped := NewScopedToolExecutor(composite, refs, map[uuid.UUID]string{fsServerID: "filesystem"})

	// read_file should be allowed
	_, _, err := scoped.FindToolByName(context.Background(), "read_file")
	if err != nil {
		t.Fatalf("read_file should be allowed: %v", err)
	}

	// write_file should be denied
	_, _, err = scoped.FindToolByName(context.Background(), "write_file")
	if err == nil {
		t.Fatal("write_file should be denied but was allowed")
	}
}

func TestScopedToolExecutor_WildcardAllow(t *testing.T) {
	fsServerID := uuid.New()
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{
		"read_file":  fsServerID,
		"write_file": fsServerID,
		"list_dir":   fsServerID,
	}}
	composite := NewCompositeToolExecutor(nil, mock)

	// Allow all filesystem tools via wildcard
	refs := []mcp.ToolRef{{ServerName: "filesystem", ToolName: "*"}}
	scoped := NewScopedToolExecutor(composite, refs, map[uuid.UUID]string{fsServerID: "filesystem"})

	for _, tool := range []string{"read_file", "write_file", "list_dir"} {
		_, _, err := scoped.FindToolByName(context.Background(), tool)
		if err != nil {
			t.Errorf("%s should be allowed via wildcard: %v", tool, err)
		}
	}
}

func TestScopedToolExecutor_InternalToolsAlwaysPass(t *testing.T) {
	// Internal tools should always pass regardless of MCP refs
	internalReg := NewInternalToolRegistry(InternalToolDeps{})
	// The registry auto-registers built-in tools; use one that exists
	// or add a mock via the tools map directly.
	internalReg.tools["consult_council"] = &InternalTool{
		Name:        "consult_council",
		Description: "test",
		Handler:     func(ctx context.Context, args map[string]any) (string, error) { return "ok", nil },
	}

	composite := NewCompositeToolExecutor(internalReg, nil)

	// Scoped with specific mcp refs (not allowAll)
	refs := []mcp.ToolRef{{ServerName: "filesystem", ToolName: "read_file"}}
	scoped := NewScopedToolExecutor(composite, refs, nil)

	// Internal tool should still pass
	serverID, _, err := scoped.FindToolByName(context.Background(), "consult_council")
	if err != nil {
		t.Fatalf("internal tool should pass: %v", err)
	}
	if serverID != InternalServerID {
		t.Errorf("expected InternalServerID, got %v", serverID)
	}
}

func TestScopedToolExecutor_DeniedDifferentServer(t *testing.T) {
	fsServerID := uuid.New()
	ghServerID := uuid.New()
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{
		"read_file":    fsServerID,
		"create_issue": ghServerID,
	}}
	composite := NewCompositeToolExecutor(nil, mock)

	// Only allow filesystem tools
	refs := []mcp.ToolRef{{ServerName: "filesystem", ToolName: "*"}}
	serverNames := map[uuid.UUID]string{fsServerID: "filesystem", ghServerID: "github"}
	scoped := NewScopedToolExecutor(composite, refs, serverNames)

	// github tool should be denied
	_, _, err := scoped.FindToolByName(context.Background(), "create_issue")
	if err == nil {
		t.Fatal("create_issue (github) should be denied when only filesystem is allowed")
	}
}

func TestScopedToolExecutor_CallToolDelegates(t *testing.T) {
	fsServerID := uuid.New()
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{"read_file": fsServerID}}
	composite := NewCompositeToolExecutor(nil, mock)
	scoped := NewScopedToolExecutor(composite, nil, nil)

	result, err := scoped.CallTool(context.Background(), fsServerID, "read_file", nil)
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result != "result:read_file" {
		t.Errorf("got %q, want result:read_file", result)
	}
}
