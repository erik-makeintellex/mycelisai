package swarm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// InternalServerID is the sentinel UUID used for internal (non-MCP) tools.
var InternalServerID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

// CompositeToolExecutor unifies InternalToolRegistry and MCPToolExecutor behind
// the same MCPToolExecutor interface. Internal tools are resolved first; if not
// found, the call falls through to the MCP adapter.
type CompositeToolExecutor struct {
	internal *InternalToolRegistry
	mcp      MCPToolExecutor // existing MCP adapter, may be nil
}

// NewCompositeToolExecutor creates a composite that tries internal tools first.
func NewCompositeToolExecutor(internal *InternalToolRegistry, mcp MCPToolExecutor) *CompositeToolExecutor {
	return &CompositeToolExecutor{
		internal: internal,
		mcp:      mcp,
	}
}

// FindToolByName resolves a tool by name. Internal tools return InternalServerID.
func (c *CompositeToolExecutor) FindToolByName(ctx context.Context, name string) (uuid.UUID, string, error) {
	// 1. Check internal tools first
	if c.internal != nil && c.internal.Has(name) {
		return InternalServerID, name, nil
	}

	// 2. Fall through to MCP
	if c.mcp != nil {
		return c.mcp.FindToolByName(ctx, name)
	}

	return uuid.Nil, "", fmt.Errorf("tool %q not found (no internal or MCP match)", name)
}

// CallTool invokes a tool. Routes to internal registry if serverID is the
// sentinel, otherwise to the MCP adapter.
func (c *CompositeToolExecutor) CallTool(ctx context.Context, serverID uuid.UUID, toolName string, args map[string]any) (string, error) {
	// Route to internal if sentinel
	if serverID == InternalServerID {
		if c.internal == nil {
			return "", fmt.Errorf("internal tool registry not available")
		}
		tool := c.internal.Get(toolName)
		if tool == nil {
			return "", fmt.Errorf("internal tool %q not found", toolName)
		}
		return tool.Handler(ctx, args)
	}

	// Route to MCP
	if c.mcp != nil {
		return c.mcp.CallTool(ctx, serverID, toolName, args)
	}

	return "", fmt.Errorf("MCP tool executor not available for server %s", serverID)
}
