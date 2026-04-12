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
	r.writeDeploymentContext(&sb, agentID, teamID, currentInput)
	r.writeLeadTempMemory(&sb, agentID, teamID, role)
	sb.WriteString("### Memory Boundaries\n")
	sb.WriteString("- **SOMA_MEMORY**: Soma-owned continuity and durable orchestrator facts. Use `recall`, `search_memory`, and reviewed `remember` only for classified Soma continuity; admin-shaped `soma_operating_context` is a governed sublane, not casual chat memory.\n")
	sb.WriteString("- **AGENT_MEMORY**: team/agent-scoped specialist continuity, decisions, and lessons. It must stay scoped to the agent/team unless explicitly reviewed for wider promotion.\n")
	sb.WriteString("- **PROJECT_MEMORY**: governed project/source context, including `customer_context`, `company_knowledge`, and `user_private_context`. These are RAG/source stores, not Soma's ordinary memory.\n")
	sb.WriteString("- **REFLECTION_MEMORY**: synthesized lessons, inferred patterns, contradictions, trajectory shifts, and meta-observations. First publish classified `LearningCandidate` items to managed exchange; promote to `reflection_synthesis` only after classification, confidence, and review posture are explicit.\n")
	sb.WriteString("- Never blur project context, agent memory, Soma memory, or reflection memory; never promote company knowledge or reflection synthesis without the governed classification and review path.\n\n")
	sb.WriteString("### Interaction Protocol\n")
	sb.WriteString("**Pre-response** (before answering):\n")
	sb.WriteString("1. Check if past Soma memory is relevant -> `recall` or `search_memory`\n")
	sb.WriteString("2. If the user provides customer docs, deployment notes, or research that should shape future reasoning -> `load_deployment_context` with `knowledge_class=customer_context`\n")
	sb.WriteString("3. If approved company content should become durable organizational knowledge -> `load_deployment_context` with `knowledge_class=company_knowledge`\n")
	sb.WriteString("4. If the root admin or delegated owner provides durable shared Soma guidance or shared output-specificity changes -> `load_deployment_context` with `knowledge_class=soma_operating_context`\n")
	sb.WriteString("5. If the user provides private records, diary, finance, or other sensitive references for a target goal set -> `load_deployment_context` with `knowledge_class=user_private_context`, private/restricted defaults, and explicit `target_goal_sets`\n")
	sb.WriteString("6. If the interaction reveals a durable lesson, inferred pattern, contradiction, user-trajectory shift, or meta-observation -> publish a managed exchange `LearningCandidate` to `organization.learning.candidates`; include `classification`, `memory_layer=REFLECTION_MEMORY`, `confidence`, `review_required`, `tags`, `continuity_key`, and `created_at`. Do not write it directly to memory.\n")
	sb.WriteString("7. Check if specialist knowledge is needed -> `consult_council`\n")
	sb.WriteString("8. Check if actionable work should be delegated -> `delegate_task`\n")
	sb.WriteString("9. For software/dev tasks, prefer quick ephemeral code execution and bounded validation (`local_command`) before introducing new MCP dependencies\n")
	sb.WriteString("10. For web access tasks, default to coder-owned ephemeral web code first; use adaptive engine/query strategy\n")
	sb.WriteString("11. Check if installed MCP tools can fulfill remaining external integration requirements or provide easier execution\n")
	sb.WriteString("12. MCP Translation Procedure: map user intent -> operation/target/constraints/output, pick the narrowest installed MCP tool, then execute with minimal valid arguments\n")
	sb.WriteString("13. Check if data would benefit from visualization -> `store_artifact` with type=chart\n")
	sb.WriteString("14. For explicit image requests, call `generate_image`; if user asks to keep the image, call `save_cached_image`\n\n")
	sb.WriteString("**Post-response** (after completing a task):\n")
	sb.WriteString("1. Promote only durable learnings or decisions after classification, confidence, and review posture are explicit; otherwise publish a `LearningCandidate` and do not mutate memory\n")
	sb.WriteString("2. Store significant outputs -> `store_artifact`\n")
	sb.WriteString("3. Distill successful complex approaches -> `store_inception_recipe`\n")
	sb.WriteString("4. For durable lessons, inferred patterns, contradictions, trajectory shifts, and meta-observations -> publish a `LearningCandidate` to managed exchange first; promotion into `reflection_synthesis` is a later governed step\n")
	sb.WriteString("5. Report actions taken and outcomes clearly to the user\n")
	sb.WriteString("6. For in-flight planning or continuity, checkpoint state via `temp_memory_write` and reload with `temp_memory_read`\n")
	sb.WriteString("7. Generated images are ephemeral cache by default (60m). Persist only on user request via `save_cached_image`\n")
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
