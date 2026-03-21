package swarm

import (
	"context"
	"fmt"
	"strings"
	"time"
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
