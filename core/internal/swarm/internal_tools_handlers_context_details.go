package swarm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (r *InternalToolRegistry) writeTeamRoster(sb *strings.Builder) {
	sb.WriteString("### Active Teams\n")
	if r.somaRef == nil {
		sb.WriteString("- (Soma offline — team roster unavailable)\n\n")
		return
	}
	manifests := r.somaRef.ListTeams()
	if len(manifests) == 0 {
		sb.WriteString("- No active teams\n\n")
		return
	}
	for _, m := range manifests {
		desc := m.Description
		if desc == "" {
			desc = fmt.Sprintf("%d agent(s)", len(m.Members))
		}
		sb.WriteString(fmt.Sprintf("- **%s** (`%s`, %s): %s\n", m.Name, m.ID, m.Type, desc))
		for _, a := range m.Members {
			tools := ""
			if len(a.Tools) > 0 {
				tools = " [" + strings.Join(a.Tools, ", ") + "]"
			}
			sb.WriteString(fmt.Sprintf("  - `%s` (%s)%s\n", a.ID, a.Role, tools))
		}
	}
	sb.WriteString("\n")
}

func (r *InternalToolRegistry) writeAgentTopology(sb *strings.Builder, agentID, teamID string, inputs, deliveries []string) {
	sb.WriteString("### Your Identity & NATS Topology\n")
	sb.WriteString(fmt.Sprintf("- **Agent ID**: `%s`\n", agentID))
	sb.WriteString(fmt.Sprintf("- **Team ID**: `%s`\n", teamID))
	sb.WriteString(fmt.Sprintf("- **Team command bus**: `swarm.team.%s.internal.command`\n", teamID))
	sb.WriteString(fmt.Sprintf("- **Team status bus**: `swarm.team.%s.signal.status`\n", teamID))
	sb.WriteString(fmt.Sprintf("- **Team result bus**: `swarm.team.%s.signal.result`\n", teamID))
	sb.WriteString(fmt.Sprintf("- **Legacy internal worker buses**: `swarm.team.%s.internal.trigger`, `swarm.team.%s.internal.response`\n", teamID, teamID))
	sb.WriteString(fmt.Sprintf("- **Direct address**: `swarm.council.%s.request`\n", agentID))
	if len(inputs) > 0 {
		sb.WriteString(fmt.Sprintf("- **Team inputs**: %s\n", strings.Join(inputs, ", ")))
	}
	if len(deliveries) > 0 {
		sb.WriteString(fmt.Sprintf("- **Team deliveries**: %s\n", strings.Join(deliveries, ", ")))
	}
	sb.WriteString("- **Global broadcast**: `swarm.global.broadcast`\n")
	sb.WriteString("- **Heartbeat**: `swarm.global.heartbeat` (5s, protobuf)\n\n")
}

func (r *InternalToolRegistry) writeCognitiveStatus(sb *strings.Builder) {
	sb.WriteString("### Cognitive Engine\n")
	if r.brain == nil || r.brain.Config == nil {
		sb.WriteString("- (Cognitive engine offline)\n\n")
		return
	}
	cfg := r.brain.Config
	for id, prov := range cfg.Providers {
		if prov.Endpoint == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("- **%s** (%s): model=`%s`, endpoint=`%s`\n", id, prov.Type, prov.ModelID, prov.Endpoint))
	}
	sb.WriteString("- **Profile routing**: ")
	parts := make([]string, 0, len(cfg.Profiles))
	for profile, provID := range cfg.Profiles {
		parts = append(parts, fmt.Sprintf("%s→%s", profile, provID))
	}
	sb.WriteString(strings.Join(parts, ", "))
	sb.WriteString("\n\n")
}

func (r *InternalToolRegistry) writeMCPServers(sb *strings.Builder) {
	sb.WriteString("### Installed MCP Servers & Tools\n")
	if r.db == nil {
		sb.WriteString("- (Database offline — MCP registry unavailable)\n\n")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	srvRows, err := r.db.QueryContext(ctx, `SELECT id, name, transport, status FROM mcp_servers ORDER BY name`)
	if err != nil {
		sb.WriteString(fmt.Sprintf("- (query failed: %v)\n\n", err))
		return
	}
	defer srvRows.Close()
	type srvInfo struct{ id, name, transport, status string }
	var servers []srvInfo
	for srvRows.Next() {
		var s srvInfo
		if err := srvRows.Scan(&s.id, &s.name, &s.transport, &s.status); err == nil {
			servers = append(servers, s)
		}
	}
	if len(servers) == 0 {
		sb.WriteString("- No MCP servers installed. Install via `/settings` -> MCP Tools tab.\n\n")
		return
	}
	type mcpTool struct{ name, desc string }
	toolMap := make(map[string][]mcpTool)
	if toolRows, err := r.db.QueryContext(ctx, `SELECT server_id, name, COALESCE(description, '') FROM mcp_tools ORDER BY name`); err == nil {
		defer toolRows.Close()
		for toolRows.Next() {
			var serverID, toolName, toolDesc string
			if err := toolRows.Scan(&serverID, &toolName, &toolDesc); err == nil {
				toolMap[serverID] = append(toolMap[serverID], mcpTool{name: toolName, desc: toolDesc})
			}
		}
	}
	for _, s := range servers {
		statusLabel := s.status
		if statusLabel == "connected" {
			statusLabel = "online"
		}
		tools := toolMap[s.id]
		if len(tools) == 0 {
			sb.WriteString(fmt.Sprintf("- **%s** (%s, %s): no tools discovered\n", s.name, s.transport, statusLabel))
			continue
		}
		toolNames := make([]string, len(tools))
		for i, t := range tools {
			toolNames[i] = t.name
		}
		sb.WriteString(fmt.Sprintf("- **%s** (%s, %s): tools=[%s]\n", s.name, s.transport, statusLabel, strings.Join(toolNames, ", ")))
	}
	sb.WriteString("\n")
}
