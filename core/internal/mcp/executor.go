package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	mcplib "github.com/mark3labs/mcp-go/mcp"
)

// ToolExecutorAdapter bridges the mcp.Service (name lookup) and ClientPool (execution)
// to satisfy the swarm.MCPToolExecutor interface via Go structural typing.
type ToolExecutorAdapter struct {
	Service *Service
	Pool    *ClientPool
}

// NewToolExecutorAdapter creates an adapter wiring the MCP registry and pool.
func NewToolExecutorAdapter(svc *Service, pool *ClientPool) *ToolExecutorAdapter {
	return &ToolExecutorAdapter{Service: svc, Pool: pool}
}

// FindToolByName looks up an MCP tool by name and returns the server ID that hosts it.
func (a *ToolExecutorAdapter) FindToolByName(ctx context.Context, name string) (uuid.UUID, string, error) {
	tool, _, err := a.Service.FindToolByName(ctx, name)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("find tool %q: %w", name, err)
	}
	return tool.ServerID, tool.Name, nil
}

// CallTool invokes a tool on the specified MCP server and returns the text result.
func (a *ToolExecutorAdapter) CallTool(ctx context.Context, serverID uuid.UUID, toolName string, args map[string]any) (string, error) {
	result, err := a.Pool.CallTool(ctx, serverID, toolName, args)
	if err != nil {
		return "", err
	}
	if err := CallToolResultError(toolName, result); err != nil {
		return "", err
	}
	return formatCallToolResult(result), nil
}

// ToolResultError is an MCP tool result the server itself marked isError.
// The call reached the server, but the tool failed; callers must treat it as
// a tool failure, never as output.
type ToolResultError struct {
	Tool    string
	Message string
}

func (e *ToolResultError) Error() string {
	return fmt.Sprintf("MCP tool %s reported an error: %s", e.Tool, e.Message)
}

// CallToolResultError returns a *ToolResultError when the MCP server flagged
// the result with isError, carrying the redacted, capped text as the reason.
func CallToolResultError(toolName string, result *mcplib.CallToolResult) error {
	if result == nil || !result.IsError {
		return nil
	}
	message := strings.TrimSpace(formatCallToolResult(result))
	if message == "" || message == noTextOutput {
		message = "no error detail returned"
	}
	return &ToolResultError{Tool: toolName, Message: redactToolErrorText(message)}
}

const noTextOutput = "(no text output)"

// formatCallToolResult extracts text content from an MCP CallToolResult.
func formatCallToolResult(result *mcplib.CallToolResult) string {
	if result == nil {
		return ""
	}
	var parts []string
	for _, c := range result.Content {
		if tc, ok := c.(mcplib.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	if len(parts) == 0 {
		return noTextOutput
	}
	return strings.Join(parts, "\n")
}
