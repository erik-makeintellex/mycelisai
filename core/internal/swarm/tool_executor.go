package swarm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// InternalServerID is the sentinel UUID used for internal (non-MCP) tools.
var InternalServerID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

// CompositeToolExecutor unifies InternalToolRegistry and MCPToolExecutor behind
// the same MCPToolExecutor interface. Internal tools are resolved first; if not
// found, the call falls through to the MCP adapter. It is unscoped: agents only
// ever receive it wrapped in a ScopedToolExecutor.
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
	if inv, ok := ToolInvocationContextFromContext(ctx); ok && inv.PlanningOnly && blocksProposalPlanningTool(toolName) {
		return "", fmt.Errorf("tool %q is blocked during proposal planning; confirmation is required before execution", toolName)
	}

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
// ScopedToolExecutor: the single per-agent declared-tool choke point
// ---------------------------------------------------------------------------

// ScopedToolExecutor enforces an agent's declared tool list on every lookup
// and every call. A call is allowed only when the canonical (trimmed, exact,
// case-sensitive) name is declared, matches a declared mcp: ref or resolved
// toolset: entry, or is a runtime-owned base tool (runtimeOwnedBaseTools).
// Everything else returns ToolNotPermittedError without executing.
type ScopedToolExecutor struct {
	inner       MCPToolExecutor
	mcpLookup   MCPToolExecutor // MCP side of a composite, for mcp:-first resolution
	scope       *agentToolScope
	serverNames map[uuid.UUID]string // serverID -> server name (for ToolRef matching)
}

// NewScopedToolExecutor scopes a composite executor to the agent's declared
// tools. serverNames maps MCP server UUIDs to names for mcp: ref matching.
func NewScopedToolExecutor(inner *CompositeToolExecutor, declared []string, serverNames map[uuid.UUID]string) *ScopedToolExecutor {
	scoped := &ScopedToolExecutor{serverNames: serverNames}
	var mcpExec MCPToolExecutor
	if inner != nil {
		scoped.inner = inner
		mcpExec = inner.mcp
		scoped.mcpLookup = inner.mcp
	}
	scoped.scope = newAgentToolScope(declared, toolSetResolverFor(mcpExec))
	return scoped
}

// scopeToolExecutor wraps any executor in the declared-tool scope unless it is
// already scoped, so no agent receives an unscoped executor.
func scopeToolExecutor(exec MCPToolExecutor, declared []string) MCPToolExecutor {
	switch typed := exec.(type) {
	case nil:
		return nil
	case *ScopedToolExecutor:
		return typed
	case *CompositeToolExecutor:
		if typed == nil {
			return nil
		}
		return NewScopedToolExecutor(typed, declared, nil)
	default:
		return &ScopedToolExecutor{inner: exec, scope: newAgentToolScope(declared, toolSetResolverFor(exec))}
	}
}

// PermitsTool reports whether the agent may use the named tool by its bare
// name, without resolving servers. Planning capture uses it.
func (s *ScopedToolExecutor) PermitsTool(ctx context.Context, name string) bool {
	name = strings.TrimSpace(name)
	return name != "" && (s.scope.allowsName(ctx, name) || runtimeOwnedBaseAllowed(ctx, name))
}

// FindToolByName resolves a tool only if the agent may use it. Declared mcp:
// refs win a name collision with an internal tool (existing behavior).
func (s *ScopedToolExecutor) FindToolByName(ctx context.Context, name string) (uuid.UUID, string, error) {
	canonical := strings.TrimSpace(name)
	if canonical == "" {
		return uuid.Nil, "", &ToolNotPermittedError{Tool: canonical}
	}
	if s.mcpLookup != nil && s.scope.mayGrantMCP(ctx) {
		serverID, toolName, err := s.mcpLookup.FindToolByName(ctx, canonical)
		if err == nil && serverID != uuid.Nil && toolName != "" && s.scope.allowsMCP(ctx, s.serverName(serverID), toolName) {
			return serverID, toolName, nil
		}
	}
	permitted := s.PermitsTool(ctx, canonical)
	// A non-composite inner has no separate MCP side, so mcp: grants are
	// checked after its lookup instead.
	mcpAfterLookup := s.mcpLookup == nil && s.scope.mayGrantMCP(ctx)
	if s.inner == nil || (!permitted && !mcpAfterLookup) {
		return uuid.Nil, "", &ToolNotPermittedError{Tool: canonical}
	}
	serverID, toolName, err := s.inner.FindToolByName(ctx, canonical)
	if err != nil {
		if !permitted {
			return uuid.Nil, "", &ToolNotPermittedError{Tool: canonical}
		}
		return uuid.Nil, "", err
	}
	if toolName != "" && s.callPermitted(ctx, serverID, toolName) {
		return serverID, toolName, nil
	}
	return uuid.Nil, "", &ToolNotPermittedError{Tool: canonical}
}

// CallTool re-checks the scope so a caller holding a server ID cannot skip
// FindToolByName, then delegates with the canonical name. Handlers see the
// caller's scope (callerToolScopeFrom), so create_team cannot widen it.
func (s *ScopedToolExecutor) CallTool(ctx context.Context, serverID uuid.UUID, toolName string, args map[string]any) (string, error) {
	canonical := strings.TrimSpace(toolName)
	if canonical == "" || s.inner == nil || !s.callPermitted(ctx, serverID, canonical) {
		return "", &ToolNotPermittedError{Tool: canonical}
	}
	return s.inner.CallTool(withCallerToolScope(ctx, s.scope), serverID, canonical, args)
}

// callPermitted checks a resolved (serverID, tool). Internal tools need a bare
// declaration or a runtime-owned base tool; MCP tools need a matching mcp: ref
// or an exact bare declaration of that MCP tool name.
func (s *ScopedToolExecutor) callPermitted(ctx context.Context, serverID uuid.UUID, toolName string) bool {
	if serverID == InternalServerID {
		return s.PermitsTool(ctx, toolName)
	}
	return s.scope.allowsMCP(ctx, s.serverName(serverID), toolName) || s.scope.allowsName(ctx, toolName)
}

func (s *ScopedToolExecutor) serverName(serverID uuid.UUID) string {
	if name := s.serverNames[serverID]; name != "" {
		return name
	}
	return serverID.String()
}
