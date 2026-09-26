package swarm

import (
	"context"
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/mcp"
)

// maxReportedToolNameRunes bounds model-supplied tool names echoed into
// denial messages, events, and logs.
const maxReportedToolNameRunes = 128

// boundedToolName trims a tool name and truncates it to
// maxReportedToolNameRunes runes, appending an ellipsis when cut.
func boundedToolName(name string) string {
	name = strings.TrimSpace(name)
	runes := []rune(name)
	if len(runes) <= maxReportedToolNameRunes {
		return name
	}
	return string(runes[:maxReportedToolNameRunes]) + "…"
}

type callerToolScopeKey struct{}

// withCallerToolScope records the calling agent's declared-tool scope on the
// context handed to a tool handler (set only by ScopedToolExecutor.CallTool).
func withCallerToolScope(ctx context.Context, scope *agentToolScope) context.Context {
	if scope == nil {
		return ctx
	}
	return context.WithValue(ctx, callerToolScopeKey{}, scope)
}

func callerToolScopeFrom(ctx context.Context) (*agentToolScope, bool) {
	scope, ok := ctx.Value(callerToolScopeKey{}).(*agentToolScope)
	return scope, ok && scope != nil
}

// permitsChildEntry reports whether a child agent's declared entry stays
// within this (parent) scope: a bare name the parent may use, an mcp: ref the
// parent's refs cover (a child wildcard needs a parent wildcard for that
// server), or a toolset: entry the parent declares by the same name.
func (s *agentToolScope) permitsChildEntry(ctx context.Context, entry string) bool {
	entry = strings.TrimSpace(entry)
	switch {
	case entry == "":
		return true
	case mcp.IsToolSetRef(entry):
		want := "toolset:" + strings.TrimSpace(mcp.ToolSetName(entry))
		for _, declared := range s.toolsets {
			if declared == want {
				return true
			}
		}
		return false
	case mcp.IsMCPRef(entry):
		ref := mcp.ParseToolRef(entry)
		if ref == nil || strings.TrimSpace(ref.ServerName) == "" || strings.TrimSpace(ref.ToolName) == "" {
			return false
		}
		server, tool := strings.TrimSpace(ref.ServerName), strings.TrimSpace(ref.ToolName)
		if tool != "*" {
			return s.allowsMCP(ctx, server, tool)
		}
		_, expanded := s.expanded(ctx)
		for _, parent := range append(append([]mcp.ToolRef(nil), s.mcpRefs...), expanded...) {
			if parent.ServerName == server && parent.ToolName == "*" {
				return true
			}
		}
		return false
	default:
		return s.allowsName(ctx, entry)
	}
}

// childToolsWithinCaller rejects a runtime-created team whose members declare
// tools outside the creating agent's declared scope. Calls without an agent
// scope (Core-owned restore, operator-approved plans) are not narrowed here.
func childToolsWithinCaller(ctx context.Context, manifest *TeamManifest) error {
	scope, ok := callerToolScopeFrom(ctx)
	if !ok || manifest == nil {
		return nil
	}
	var outside []string
	seen := map[string]bool{}
	for _, member := range manifest.Members {
		for _, tool := range member.Tools {
			name := boundedToolName(tool)
			if seen[name] || scope.permitsChildEntry(ctx, tool) {
				continue
			}
			seen[name] = true
			outside = append(outside, name)
		}
	}
	if len(outside) == 0 {
		return nil
	}
	return fmt.Errorf("create_team refused: team members may only hold tools this agent declares; not permitted: %s", strings.Join(outside, ", "))
}
