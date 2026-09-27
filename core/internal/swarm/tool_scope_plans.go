package swarm

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/mcp"
)

// AgentDeclaredTools returns a copy of the declared tool list of agentID in
// the live team teamID. The team bound keeps a same-named member of another
// team from lending its tools. ok is false when the team or agent is unknown.
func (s *Soma) AgentDeclaredTools(teamID, agentID string) ([]string, bool) {
	if s == nil {
		return nil, false
	}
	teamID, agentID = strings.TrimSpace(teamID), strings.TrimSpace(agentID)
	s.mu.RLock()
	defer s.mu.RUnlock()
	team := s.teams[teamID]
	if team == nil || team.Manifest == nil {
		return nil, false
	}
	for _, member := range team.Manifest.Members {
		if member.ID == agentID {
			return append([]string{}, member.Tools...), true
		}
	}
	return nil, false
}

// MCPServerNames returns a copy of the MCP server-ID-to-name map agents use
// for mcp: grants, so approved plans match refs the same way.
func (s *Soma) MCPServerNames() map[uuid.UUID]string {
	if s == nil {
		return nil
	}
	names := make(map[uuid.UUID]string, len(s.mcpServerNames))
	for id, name := range s.mcpServerNames {
		names[id] = name
	}
	return names
}

// PermitsPlannedCall reports whether an approved-plan call stays inside the
// agent's declared scope: a tool_ref must be a concrete mcp:<server>/<tool>
// the scope grants (a bare declaration never covers a ref); otherwise the
// bare name must be declared. Runtime-owned base tools are never implied.
func (s *ScopedToolExecutor) PermitsPlannedCall(ctx context.Context, name, toolRef string) bool {
	if toolRef = strings.TrimSpace(toolRef); toolRef != "" {
		ref := mcp.ParseToolRef(toolRef)
		if ref == nil || strings.TrimSpace(ref.ServerName) == "" || strings.TrimSpace(ref.ToolName) == "" || ref.ToolName == "*" {
			return false
		}
		return s.scope.allowsMCP(ctx, strings.TrimSpace(ref.ServerName), strings.TrimSpace(ref.ToolName))
	}
	name = strings.TrimSpace(name)
	return name != "" && s.scope.allowsName(ctx, name)
}
