package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	mcplib "github.com/mark3labs/mcp-go/mcp"
)

// ── formatCallToolResult (unexported but testable from same package) ──

func TestFormatCallToolResult_TextContent(t *testing.T) {
	result := &mcplib.CallToolResult{
		Content: []mcplib.Content{
			mcplib.NewTextContent("hello world"),
		},
	}
	got := formatCallToolResult(result)
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestFormatCallToolResult_MultipleContent(t *testing.T) {
	result := &mcplib.CallToolResult{
		Content: []mcplib.Content{
			mcplib.NewTextContent("line 1"),
			mcplib.NewTextContent("line 2"),
		},
	}
	got := formatCallToolResult(result)
	if got != "line 1\nline 2" {
		t.Errorf("got %q, want %q", got, "line 1\nline 2")
	}
}

func TestFormatCallToolResult_NoContent(t *testing.T) {
	result := &mcplib.CallToolResult{Content: []mcplib.Content{}}
	got := formatCallToolResult(result)
	if got != "(no text output)" {
		t.Errorf("got %q, want %q", got, "(no text output)")
	}
}

func TestFormatCallToolResult_NonTextContent(t *testing.T) {
	result := &mcplib.CallToolResult{
		Content: []mcplib.Content{
			mcplib.NewImageContent("aGVsbG8=", "image/png"),
		},
	}
	got := formatCallToolResult(result)
	if got != "(no text output)" {
		t.Errorf("got %q, want %q", got, "(no text output)")
	}
}

func TestFormatCallToolResult_NilResult(t *testing.T) {
	got := formatCallToolResult(nil)
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

// ── ToolExecutorAdapter ──────────────────────────────────────

// mockServiceFinder mocks the Service's FindToolByName for adapter tests.
type mockServiceFinder struct {
	tool *ToolDef
	srv  *ServerConfig
	err  error
}

// We can't mock Service directly (concrete struct), so we test the adapter
// through a real Service with sqlmock.

func TestExecutorAdapter_FindToolByName(t *testing.T) {
	svc, mock := newTestService(t)
	pool := NewClientPool(svc)
	adapter := NewToolExecutorAdapter(svc, pool)

	serverID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	toolID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	findToolColumns := []string{
		"id", "server_id", "server_name", "name", "description", "input_schema",
		"srv_id", "srv_name", "transport", "command", "args", "env", "url", "headers", "status", "error_message", "created_at", "updated_at",
	}
	mock.ExpectQuery("SELECT .+ FROM mcp_tools .+ JOIN mcp_servers").
		WithArgs("read_file").
		WillReturnRows(sqlmock.NewRows(findToolColumns).
			AddRow(toolID, serverID, "filesystem", "read_file", "Read", []byte(`{}`),
				serverID, "filesystem", "stdio", "npx", `[]`, `{}`, "", `{}`, "connected", nil, time.Now(), time.Now()))

	gotID, gotName, err := adapter.FindToolByName(context.Background(), "read_file")
	if err != nil {
		t.Fatalf("FindToolByName: %v", err)
	}
	if gotID != serverID {
		t.Errorf("serverID = %v, want %v", gotID, serverID)
	}
	if gotName != "read_file" {
		t.Errorf("toolName = %q", gotName)
	}
}

func TestExecutorAdapter_FindToolByName_NotFound(t *testing.T) {
	svc, mock := newTestService(t)
	pool := NewClientPool(svc)
	adapter := NewToolExecutorAdapter(svc, pool)

	findToolColumns := []string{
		"id", "server_id", "server_name", "name", "description", "input_schema",
		"srv_id", "srv_name", "transport", "command", "args", "env", "url", "headers", "status", "error_message", "created_at", "updated_at",
	}
	mock.ExpectQuery("SELECT .+ FROM mcp_tools .+ JOIN mcp_servers").
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows(findToolColumns))

	_, _, err := adapter.FindToolByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestExecutorAdapter_CallTool_Error(t *testing.T) {
	svc, _ := newTestService(t)
	pool := NewClientPool(svc)
	adapter := NewToolExecutorAdapter(svc, pool)

	// Call a tool on a server that doesn't exist in the pool → error
	serverID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	_, err := adapter.CallTool(context.Background(), serverID, "read_file", nil)
	if err == nil {
		t.Fatal("expected error for unknown server")
	}
}

