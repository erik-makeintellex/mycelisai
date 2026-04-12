package swarm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mycelis/core/internal/memory"
)

func (r *InternalToolRegistry) writeRecalledMemory(sb *strings.Builder, agentID, currentInput string) {
	if r.brain == nil || r.mem == nil || currentInput == "" {
		return
	}

	// Embed the current input (truncated) for semantic search.
	query := currentInput
	if len(query) > 200 {
		query = query[:200]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	vec, err := r.brain.Embed(ctx, query, "")
	if err != nil {
		return // silent - embedding not available
	}

	summaries, err := r.mem.RecallConversations(ctx, vec, agentID, 3)
	if err != nil || len(summaries) == 0 {
		return
	}

	sb.WriteString("### Previous Context (from past conversations)\n")
	for _, s := range summaries {
		age := time.Since(s.CreatedAt)
		var ageStr string
		switch {
		case age < time.Hour:
			ageStr = fmt.Sprintf("%d min ago", int(age.Minutes()))
		case age < 24*time.Hour:
			ageStr = fmt.Sprintf("%d hours ago", int(age.Hours()))
		default:
			ageStr = fmt.Sprintf("%d days ago", int(age.Hours()/24))
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s", ageStr, s.Summary))
		if len(s.KeyTopics) > 0 {
			sb.WriteString(fmt.Sprintf(" (topics: %s)", strings.Join(s.KeyTopics, ", ")))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

func (r *InternalToolRegistry) writeDeploymentContext(sb *strings.Builder, agentID, teamID, currentInput string) {
	if r.brain == nil || r.mem == nil || strings.TrimSpace(currentInput) == "" {
		return
	}

	query := currentInput
	if len(query) > 240 {
		query = query[:240]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	vec, err := r.brain.Embed(ctx, query, "")
	if err != nil {
		return
	}

	// Restrict retrieval to the governed deployment-context vector classes so
	// ordinary Soma memory and durable company/customer knowledge never blur.
	results, err := r.mem.SemanticSearchWithOptions(ctx, vec, memory.SemanticSearchOptions{
		Limit:               3,
		TenantID:            "default",
		TeamID:              strings.TrimSpace(teamID),
		AgentID:             strings.TrimSpace(agentID),
		Types:               []string{"customer_context", "company_knowledge", "soma_operating_context", "user_private_context"},
		AllowGlobal:         true,
		AllowLegacyUnscoped: false,
	})
	if err != nil || len(results) == 0 {
		return
	}

	customerResults := make([]memory.VectorResult, 0, len(results))
	companyResults := make([]memory.VectorResult, 0, len(results))
	somaResults := make([]memory.VectorResult, 0, len(results))
	privateResults := make([]memory.VectorResult, 0, len(results))
	for _, result := range results {
		switch strings.TrimSpace(stringMeta(result.Metadata, "knowledge_class")) {
		case "company_knowledge":
			companyResults = append(companyResults, result)
		case "soma_operating_context":
			somaResults = append(somaResults, result)
		case "user_private_context":
			privateResults = append(privateResults, result)
		default:
			customerResults = append(customerResults, result)
		}
	}

	if len(customerResults) > 0 {
		sb.WriteString("### Customer Context Store (operator-provided documents)\n")
		for _, result := range customerResults {
			writeKnowledgeResult(sb, result, "operator provided")
		}
		sb.WriteString("\n")
	}
	if len(companyResults) > 0 {
		sb.WriteString("### Company Knowledge Store (approved Soma/company content)\n")
		for _, result := range companyResults {
			writeKnowledgeResult(sb, result, "approved company knowledge")
		}
		sb.WriteString("\n")
	}
	if len(somaResults) > 0 {
		sb.WriteString("### Admin-Shaped Soma Context (organization-owned Soma operating guidance)\n")
		for _, result := range somaResults {
			writeKnowledgeResult(sb, result, "admin-shaped Soma context")
		}
		sb.WriteString("\n")
	}
	if len(privateResults) > 0 {
		sb.WriteString("### User-Private Context Store (private records and goal-scoped references)\n")
		for _, result := range privateResults {
			writeKnowledgeResult(sb, result, "private user context")
		}
		sb.WriteString("\n")
	}
}

func writeKnowledgeResult(sb *strings.Builder, result memory.VectorResult, fallbackSourceLabel string) {
	title := "Knowledge entry"
	sourceLabel := fallbackSourceLabel
	if result.Metadata != nil {
		if value, ok := result.Metadata["artifact_title"].(string); ok && strings.TrimSpace(value) != "" {
			title = strings.TrimSpace(value)
		}
		if value, ok := result.Metadata["source_label"].(string); ok && strings.TrimSpace(value) != "" {
			sourceLabel = strings.TrimSpace(value)
		}
	}
	preview := strings.TrimSpace(result.Content)
	if len(preview) > 220 {
		preview = preview[:220] + "..."
	}
	sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", title, sourceLabel, preview))
}

func stringMeta(meta map[string]any, key string) string {
	if meta == nil {
		return ""
	}
	if value, ok := meta[key].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return ""
}
