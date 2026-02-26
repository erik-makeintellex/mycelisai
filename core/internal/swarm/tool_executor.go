package swarm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/mcp"
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
func NewCompositeToolExecutor(internal *InternalToolRegistry, mcpExec MCPToolExecutor) *CompositeToolExecutor {
	return &CompositeToolExecutor{
		internal: internal,
		mcp:      mcpExec,
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

// ---------------------------------------------------------------------------
// ScopedToolExecutor — per-agent MCP tool filtering
// ---------------------------------------------------------------------------

// ScopedToolExecutor wraps a CompositeToolExecutor with per-agent MCP tool filtering.
// Internal tools pass through unchanged (filtered by buildToolsBlock in agent.go).
// MCP tools are checked against an allow-list derived from the agent's manifest.
//
// Backward compatible: if the agent manifest has zero mcp: references,
// allowAll is true and all MCP tools remain accessible (pre-binding behavior).
type ScopedToolExecutor struct {
	inner       *CompositeToolExecutor
	allowedMCP  []mcp.ToolRef             // parsed from agent manifest
	serverNames map[uuid.UUID]string      // serverID → server name (for ToolRef matching)
	allowAll    bool                       // true when no mcp: refs → backward compat
}

// NewScopedToolExecutor creates a scoped executor from the agent's MCP tool refs.
// serverNames maps server UUIDs to their names for allow-list matching.
// If mcpRefs is empty, allowAll is true (backward compatible — all MCP tools permitted).
func NewScopedToolExecutor(inner *CompositeToolExecutor, mcpRefs []mcp.ToolRef, serverNames map[uuid.UUID]string) *ScopedToolExecutor {
	return &ScopedToolExecutor{
		inner:       inner,
		allowedMCP:  mcpRefs,
		serverNames: serverNames,
		allowAll:    len(mcpRefs) == 0,
	}
}

// FindToolByName delegates to the inner composite executor, then checks MCP tools
// against the agent's allow-list. Internal tools always pass through.
func (s *ScopedToolExecutor) FindToolByName(ctx context.Context, name string) (uuid.UUID, string, error) {
	serverID, toolName, err := s.inner.FindToolByName(ctx, name)
	if err != nil {
		return uuid.Nil, "", err
	}

	// Internal tools always pass (filtered by buildToolsBlock, not here)
	if serverID == InternalServerID {
		return serverID, toolName, nil
	}

	// MCP tool: check allow-list
	if s.allowAll {
		return serverID, toolName, nil
	}

	serverName := s.serverNames[serverID]
	if serverName == "" {
		serverName = serverID.String() // fallback to UUID string
	}

	for _, ref := range s.allowedMCP {
		if ref.MatchesTool(serverName, toolName) {
			return serverID, toolName, nil
		}
	}

	return uuid.Nil, "", fmt.Errorf("tool %q (server %q) not authorized for this agent", toolName, serverName)
}

// CallTool delegates to the inner composite executor.
// The serverID was already validated by FindToolByName.
func (s *ScopedToolExecutor) CallTool(ctx context.Context, serverID uuid.UUID, toolName string, args map[string]any) (string, error) {
	return s.inner.CallTool(ctx, serverID, toolName, args)
}
