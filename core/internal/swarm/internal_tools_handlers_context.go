package swarm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// BuildContext generates a live system state block for injection into an agent's system prompt.
func (r *InternalToolRegistry) BuildContext(agentID, teamID, role string, teamInputs, teamDeliveries []string, currentInput string) string {
	var sb strings.Builder
	sb.WriteString("\n\n## Runtime Context (Live System State)\n")
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n\n", time.Now().Format(time.RFC3339)))
	r.writeTeamRoster(&sb)
	r.writeAgentTopology(&sb, agentID, teamID, teamInputs, teamDeliveries)
	r.writeCognitiveStatus(&sb)
	r.writeMCPServers(&sb)
	r.writeRecalledMemory(&sb, agentID, currentInput)
	r.writeLeadTempMemory(&sb, agentID, teamID, role)
	sb.WriteString("### Interaction Protocol\n")
	sb.WriteString("**Pre-response** (before answering):\n")
	sb.WriteString("1. Check if past context is relevant -> `recall` or `search_memory`\n")
	sb.WriteString("2. Check if specialist knowledge is needed -> `consult_council`\n")
	sb.WriteString("3. Check if actionable work should be delegated -> `delegate_task`\n")
	sb.WriteString("4. For software/dev tasks, prefer quick ephemeral code execution and bounded validation (`local_command`) before introducing new MCP dependencies\n")
	sb.WriteString("5. For web access tasks, default to coder-owned ephemeral web code first; use adaptive engine/query strategy\n")
	sb.WriteString("6. Check if installed MCP tools can fulfill remaining external integration requirements or provide easier execution\n")
	sb.WriteString("7. MCP Translation Procedure: map user intent -> operation/target/constraints/output, pick the narrowest installed MCP tool, then execute with minimal valid arguments\n")
	sb.WriteString("8. Check if data would benefit from visualization -> `store_artifact` with type=chart\n")
	sb.WriteString("9. For explicit image requests, call `generate_image`; if user asks to keep the image, call `save_cached_image`\n\n")
	sb.WriteString("**Post-response** (after completing a task):\n")
	sb.WriteString("1. Promote only durable learnings or decisions -> `remember`\n")
	sb.WriteString("2. Store significant outputs -> `store_artifact`\n")
	sb.WriteString("3. Distill successful complex approaches -> `store_inception_recipe`\n")
	sb.WriteString("4. Report actions taken and outcomes clearly to the user\n")
	sb.WriteString("5. For in-flight planning or continuity, checkpoint state via `temp_memory_write` and reload with `temp_memory_read`\n")
	sb.WriteString("6. Generated images are ephemeral cache by default (60m). Persist only on user request via `save_cached_image`\n")
	return sb.String()
}

func (r *InternalToolRegistry) isLeadAgent(agentID, teamID, role string) bool {
	id := strings.ToLower(agentID)
	tid := strings.ToLower(teamID)
	rl := strings.ToLower(role)
	return tid == "admin-core" || tid == "council-core" || id == "admin" || strings.HasPrefix(id, "council-") || strings.Contains(rl, "lead") || rl == "architect" || rl == "coder" || rl == "creative" || rl == "sentry"
}

func (r *InternalToolRegistry) writeLeadTempMemory(sb *strings.Builder, agentID, teamID, role string) {
	if !r.isLeadAgent(agentID, teamID, role) {
		return
	}
	sb.WriteString("### Persistent Temp Memory Channels (Restart-Safe)\n")
	sb.WriteString("Use these checkpoints to continue work consistently across provider/service restarts.\n")
	if r.mem == nil {
		sb.WriteString("- Memory backend unavailable; rely on current contract and checkpoint once memory is restored.\n")
		sb.WriteString("Stability rules: preserve user interaction style, output shape, and action sequencing unless user explicitly changes intent.\n\n")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	channels := []string{"interaction.contract", "lead.shared", fmt.Sprintf("lead.%s", agentID)}
	if strings.TrimSpace(teamID) != "" {
		channels = append(channels, fmt.Sprintf("team.%s.planning", teamID))
	}
	for _, ch := range channels {
		entries, err := r.mem.GetTempMemory(ctx, "default", ch, 3)
		if err != nil || len(entries) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("- **%s**\n", ch))
		for _, e := range entries {
			content := strings.TrimSpace(e.Content)
			if len(content) > 220 {
				content = content[:220] + "..."
			}
			sb.WriteString(fmt.Sprintf("  - [%s] %s\n", e.OwnerAgentID, content))
		}
	}
	sb.WriteString("Stability rules: preserve user interaction style, output shape, and action sequencing unless user explicitly changes intent.\n\n")
}
