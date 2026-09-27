package swarm

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// mockMCPExecutor simulates an MCP tool executor for testing.
type mockMCPExecutor struct {
	tools map[string]uuid.UUID // tool name → server ID
	calls []string
}

func (m *mockMCPExecutor) FindToolByName(_ context.Context, name string) (uuid.UUID, string, error) {
	if id, ok := m.tools[name]; ok {
		return id, name, nil
	}
	return uuid.Nil, "", nil
}

func (m *mockMCPExecutor) CallTool(_ context.Context, _ uuid.UUID, toolName string, _ map[string]any) (string, error) {
	m.calls = append(m.calls, toolName)
	return "result:" + toolName, nil
}

func TestScopedToolExecutor_NoMCPRefsDeniesMCP(t *testing.T) {
	// Previously zero mcp: refs meant every MCP tool was allowed.
	fsServerID := uuid.New()
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{"list_dir": fsServerID}}
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(nil, mock), []string{"remember"}, map[uuid.UUID]string{fsServerID: "filesystem"})

	if _, _, err := scoped.FindToolByName(context.Background(), "list_dir"); !IsToolNotPermitted(err) {
		t.Fatalf("list_dir without an mcp: entry: err = %v, want not permitted", err)
	}
	if _, err := scoped.CallTool(context.Background(), fsServerID, "list_dir", nil); !IsToolNotPermitted(err) {
		t.Fatalf("direct CallTool of undeclared MCP tool: err = %v, want not permitted", err)
	}
	if len(mock.calls) != 0 {
		t.Fatalf("undeclared MCP tool executed: %v", mock.calls)
	}
}

func TestScopedToolExecutor_FilteredAllow(t *testing.T) {
	fsServerID := uuid.New()
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{
		"read_file":  fsServerID,
		"write_file": fsServerID,
	}}
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(nil, mock), []string{"mcp:filesystem/read_file"}, map[uuid.UUID]string{fsServerID: "filesystem"})

	if _, _, err := scoped.FindToolByName(context.Background(), "read_file"); err != nil {
		t.Fatalf("read_file should be allowed: %v", err)
	}
	if _, _, err := scoped.FindToolByName(context.Background(), "write_file"); !IsToolNotPermitted(err) {
		t.Fatalf("write_file: err = %v, want not permitted", err)
	}
}

func TestScopedToolExecutor_WildcardAllow(t *testing.T) {
	fsServerID := uuid.New()
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{
		"read_file":  fsServerID,
		"write_file": fsServerID,
		"list_dir":   fsServerID,
	}}
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(nil, mock), []string{" mcp:filesystem/* "}, map[uuid.UUID]string{fsServerID: "filesystem"})

	for _, tool := range []string{"read_file", "write_file", "list_dir"} {
		serverID, _, err := scoped.FindToolByName(context.Background(), tool)
		if err != nil {
			t.Fatalf("%s should be allowed via wildcard: %v", tool, err)
		}
		if _, err := scoped.CallTool(context.Background(), serverID, tool, nil); err != nil {
			t.Fatalf("%s call via wildcard: %v", tool, err)
		}
	}
}

func TestScopedToolExecutor_ExplicitMCPAllowlistPrefersMCPNameCollision(t *testing.T) {
	fsServerID := uuid.New()
	internalReg := NewInternalToolRegistry(InternalToolDeps{})
	internalReg.tools["read_file"] = &InternalTool{
		Name:        "read_file",
		Description: "internal file reader",
		Handler:     func(ctx context.Context, args map[string]any) (string, error) { return "internal", nil },
	}
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{"read_file": fsServerID}}
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(internalReg, mock), []string{"mcp:filesystem/*"}, map[uuid.UUID]string{fsServerID: "filesystem"})

	serverID, toolName, err := scoped.FindToolByName(context.Background(), "read_file")
	if err != nil {
		t.Fatalf("read_file should resolve to allowed MCP tool: %v", err)
	}
	if serverID != fsServerID || toolName != "read_file" {
		t.Fatalf("got (%v, %q), want MCP filesystem read_file", serverID, toolName)
	}
}

func TestScopedToolExecutor_DeniedDifferentServer(t *testing.T) {
	fsServerID := uuid.New()
	ghServerID := uuid.New()
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{
		"read_file":    fsServerID,
		"create_issue": ghServerID,
	}}
	serverNames := map[uuid.UUID]string{fsServerID: "filesystem", ghServerID: "github"}
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(nil, mock), []string{"mcp:filesystem/*"}, serverNames)

	if _, _, err := scoped.FindToolByName(context.Background(), "create_issue"); !IsToolNotPermitted(err) {
		t.Fatalf("create_issue (github): err = %v, want not permitted", err)
	}
	if _, err := scoped.CallTool(context.Background(), ghServerID, "create_issue", nil); !IsToolNotPermitted(err) {
		t.Fatalf("direct github CallTool: err = %v, want not permitted", err)
	}
}

func TestScopedToolExecutor_BareDeclaredMCPToolName(t *testing.T) {
	// Blueprint agents may declare an installed MCP tool by its exact name.
	ghServerID := uuid.New()
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{"create_issue": ghServerID, "delete_repo": ghServerID}}
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(nil, mock), []string{"create_issue"}, map[uuid.UUID]string{ghServerID: "github"})

	serverID, _, err := scoped.FindToolByName(context.Background(), "create_issue")
	if err != nil || serverID != ghServerID {
		t.Fatalf("declared bare MCP name: (%v, %v)", serverID, err)
	}
	if _, _, err := scoped.FindToolByName(context.Background(), "delete_repo"); !IsToolNotPermitted(err) {
		t.Fatalf("delete_repo: err = %v, want not permitted", err)
	}
	if _, _, err := scoped.FindToolByName(context.Background(), "Create_Issue"); !IsToolNotPermitted(err) {
		t.Fatalf("case-changed name: err = %v, want not permitted", err)
	}
}

func TestScopedToolExecutor_CallToolDelegates(t *testing.T) {
	fsServerID := uuid.New()
	mock := &mockMCPExecutor{tools: map[string]uuid.UUID{"read_file": fsServerID}}
	scoped := NewScopedToolExecutor(NewCompositeToolExecutor(nil, mock), []string{"mcp:filesystem/read_file"}, map[uuid.UUID]string{fsServerID: "filesystem"})

	result, err := scoped.CallTool(context.Background(), fsServerID, "read_file", nil)
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result != "result:read_file" {
		t.Errorf("got %q, want result:read_file", result)
	}
}
