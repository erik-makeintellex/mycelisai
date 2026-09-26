package swarm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mycelis/core/internal/mcp"
)

// runtimeOwnedBaseTools are the only tools Core may run for an agent without
// the agent declaring them, and only as runtime-owned mechanical steps
// (ToolInvocationContext.RuntimeOwned): the council preflight consultation
// and the entrypoint readback after a declared write. Models never get them
// implicitly.
var runtimeOwnedBaseTools = map[string]struct{}{
	"consult_council": {},
	"read_file":       {},
}

// ToolSetResolver expands "toolset:<name>" entries into their tool refs.
// mcp.ToolSetService satisfies it.
type ToolSetResolver interface {
	ResolveRefs(ctx context.Context, tools []string) ([]string, error)
}

// ToolNotPermittedError is the normalized denial returned to the model. It
// names only the requested tool, never servers, IDs, or policy internals.
type ToolNotPermittedError struct{ Tool string }

func (e *ToolNotPermittedError) Error() string {
	return fmt.Sprintf("tool %q is not permitted for this agent", boundedToolName(e.Tool))
}

// IsToolNotPermitted reports whether err is a declared-tool denial.
func IsToolNotPermitted(err error) bool {
	var denied *ToolNotPermittedError
	return errors.As(err, &denied)
}

// agentToolScope is the immutable declared-tool policy for one agent. Entries
// are trimmed and matched exactly and case-sensitively. "mcp:<server>/<tool>"
// and "mcp:<server>/*" grant MCP tools; "toolset:<name>" entries expand
// through the resolver on first use and grant nothing if they cannot resolve.
// An empty declared list grants nothing.
type agentToolScope struct {
	names    map[string]struct{}
	mcpRefs  []mcp.ToolRef
	toolsets []string
	resolver ToolSetResolver

	mu         sync.Mutex
	resolved   bool
	setNames   map[string]struct{}
	setMCPRefs []mcp.ToolRef
}

func newAgentToolScope(declared []string, resolver ToolSetResolver) *agentToolScope {
	scope := &agentToolScope{names: map[string]struct{}{}, resolver: resolver}
	for _, entry := range declared {
		entry = strings.TrimSpace(entry)
		switch {
		case entry == "":
		case mcp.IsToolSetRef(entry):
			if name := strings.TrimSpace(mcp.ToolSetName(entry)); name != "" {
				scope.toolsets = append(scope.toolsets, "toolset:"+name)
			}
		default:
			addScopeEntry(entry, scope.names, &scope.mcpRefs)
		}
	}
	scope.resolved = len(scope.toolsets) == 0
	return scope
}

func addScopeEntry(entry string, names map[string]struct{}, refs *[]mcp.ToolRef) {
	if !mcp.IsMCPRef(entry) {
		names[entry] = struct{}{}
		return
	}
	ref := mcp.ParseToolRef(entry)
	if ref == nil || strings.TrimSpace(ref.ServerName) == "" || strings.TrimSpace(ref.ToolName) == "" {
		return
	}
	*refs = append(*refs, mcp.ToolRef{ServerName: strings.TrimSpace(ref.ServerName), ToolName: strings.TrimSpace(ref.ToolName)})
}

// expanded returns the toolset-derived grants, resolving them once. A failed
// resolution grants nothing now and is retried on the next call.
func (s *agentToolScope) expanded(ctx context.Context) (map[string]struct{}, []mcp.ToolRef) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.resolved {
		return s.setNames, s.setMCPRefs
	}
	if s.resolver == nil {
		return nil, nil
	}
	resolveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	refs, err := s.resolver.ResolveRefs(resolveCtx, s.toolsets)
	if err != nil {
		return nil, nil
	}
	names := map[string]struct{}{}
	var mcpRefs []mcp.ToolRef
	for _, entry := range refs {
		entry = strings.TrimSpace(entry)
		if entry == "" || mcp.IsToolSetRef(entry) {
			continue
		}
		addScopeEntry(entry, names, &mcpRefs)
	}
	s.setNames, s.setMCPRefs, s.resolved = names, mcpRefs, true
	return names, mcpRefs
}

// allowsName reports whether an exact (bare) tool name is declared directly
// or through a resolved toolset.
func (s *agentToolScope) allowsName(ctx context.Context, name string) bool {
	if _, ok := s.names[name]; ok {
		return true
	}
	if len(s.toolsets) == 0 {
		return false
	}
	names, _ := s.expanded(ctx)
	_, ok := names[name]
	return ok
}

// allowsMCP reports whether a declared mcp: ref matches server + tool.
func (s *agentToolScope) allowsMCP(ctx context.Context, serverName, toolName string) bool {
	for _, ref := range s.mcpRefs {
		if ref.MatchesTool(serverName, toolName) {
			return true
		}
	}
	if len(s.toolsets) == 0 {
		return false
	}
	_, refs := s.expanded(ctx)
	for _, ref := range refs {
		if ref.MatchesTool(serverName, toolName) {
			return true
		}
	}
	return false
}

// mayGrantMCP reports whether any mcp: grant could exist, so MCP lookups are
// skipped for agents that cannot use MCP at all.
func (s *agentToolScope) mayGrantMCP(ctx context.Context) bool {
	if len(s.mcpRefs) > 0 {
		return true
	}
	if len(s.toolsets) == 0 {
		return false
	}
	_, refs := s.expanded(ctx)
	return len(refs) > 0
}

func runtimeOwnedBaseAllowed(ctx context.Context, name string) bool {
	if _, ok := runtimeOwnedBaseTools[name]; !ok {
		return false
	}
	inv, ok := ToolInvocationContextFromContext(ctx)
	return ok && inv.RuntimeOwned
}

// toolSetResolverFor finds the toolset registry behind an MCP executor.
func toolSetResolverFor(exec MCPToolExecutor) ToolSetResolver {
	switch typed := exec.(type) {
	case *mcp.ToolExecutorAdapter:
		if typed != nil && typed.Service != nil && typed.Service.ToolSets != nil {
			return typed.Service.ToolSets
		}
	case ToolSetResolver:
		return typed
	}
	return nil
}
