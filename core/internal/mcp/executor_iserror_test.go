package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/client"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// attachInProcessServer registers a live in-process MCP server in the pool so
// the adapter is exercised through a real mcp-go client round trip.
func attachInProcessServer(t *testing.T, pool *ClientPool, srv *server.MCPServer) uuid.UUID {
	t.Helper()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("in-process client: %v", err)
	}
	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	init := mcplib.InitializeRequest{}
	init.Params.ProtocolVersion = mcplib.LATEST_PROTOCOL_VERSION
	init.Params.ClientInfo = mcplib.Implementation{Name: "test", Version: "0"}
	if _, err := c.Initialize(ctx, init); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	id := uuid.New()
	pool.mu.Lock()
	pool.clients[id] = &ManagedClient{ServerID: id, Client: c, Connected: true}
	pool.mu.Unlock()
	return id
}

func newFailingAndPassingServer() *server.MCPServer {
	srv := server.NewMCPServer("fixture", "0")
	srv.AddTool(mcplib.NewTool("write_file"), func(context.Context, mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		return mcplib.NewToolResultError("access denied: path outside allowed directories"), nil
	})
	srv.AddTool(mcplib.NewTool("read_file"), func(context.Context, mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		return mcplib.NewToolResultText("file body"), nil
	})
	return srv
}

func TestExecutorAdapter_CallTool_IsErrorResultIsFailure(t *testing.T) {
	svc, _ := newTestService(t)
	pool := NewClientPool(svc)
	id := attachInProcessServer(t, pool, newFailingAndPassingServer())

	out, err := NewToolExecutorAdapter(svc, pool).CallTool(context.Background(), id, "write_file", map[string]any{"path": "/etc/x"})
	if err == nil {
		t.Fatalf("isError result returned success output %q", out)
	}
	var toolErr *ToolResultError
	if !errors.As(err, &toolErr) || !strings.Contains(toolErr.Message, "access denied") {
		t.Fatalf("err = %v, want ToolResultError carrying the server text", err)
	}
	if out != "" {
		t.Fatalf("failed call leaked output %q", out)
	}
}

func TestExecutorAdapter_CallTool_SuccessUnchanged(t *testing.T) {
	svc, _ := newTestService(t)
	pool := NewClientPool(svc)
	id := attachInProcessServer(t, pool, newFailingAndPassingServer())

	out, err := NewToolExecutorAdapter(svc, pool).CallTool(context.Background(), id, "read_file", nil)
	if err != nil || out != "file body" {
		t.Fatalf("CallTool = %q, %v; want file body", out, err)
	}
}

func TestCallToolResultError_NoDetailStillFails(t *testing.T) {
	err := CallToolResultError("probe", &mcplib.CallToolResult{IsError: true})
	if err == nil || !strings.Contains(err.Error(), "no error detail") {
		t.Fatalf("err = %v, want explicit failure without detail", err)
	}
	if CallToolResultError("probe", &mcplib.CallToolResult{}) != nil || CallToolResultError("probe", nil) != nil {
		t.Fatal("non-error results must not be failures")
	}
}
