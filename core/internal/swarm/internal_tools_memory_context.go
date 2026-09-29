package swarm

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
)

// writeRecalledMemory injects past conversation summaries the turn's user may
// read (MEM-LANES): their own, or org-wide ones; none when access failed.
func (r *InternalToolRegistry) writeRecalledMemory(sb *strings.Builder, access RecallAccess, agentID, currentInput string) {
	if r.brain == nil || r.mem == nil || currentInput == "" {
		return
	}
	reader, err := access.laneReader()
	if err != nil {
		return
	}

	// Embed the operator's request (bounded) for semantic search.
	query := recallQuery(currentInput, 200)
	if query == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Conversation summaries are only vector-indexed; skip without a working
	// embedding engine instead of paying a failing round trip every turn.
	if !r.brain.EmbeddingAvailable(ctx) {
		return
	}
	vec, err := r.brain.Embed(ctx, query, "")
	if err != nil {
		return
	}

	summaries, err := r.mem.RecallConversations(ctx, vec, agentID, 3, reader)
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

// governedClasses are the deployment-context vector classes. Leads receive
// all of them; workers receive only customer and company knowledge.
var (
	governedClasses = []string{"customer_context", "company_knowledge", "soma_operating_context", "user_private_context", "reflection_synthesis"}
	workerClasses   = []string{"customer_context", "company_knowledge"}
)

// ContextSource is one governed source injected into a prompt, with the exact
// excerpt the model saw (used to decide whether the reply used it).
type ContextSource struct {
	protocol.ContextSourceRef
	Excerpt string
}

// governedRecallOptions scopes ambient recall. Goal-scoped rows are never
// ambient; restricted rows reach a worker only from its own team.
func governedRecallOptions(agentID, teamID string, lead bool) memory.SemanticSearchOptions {
	opts := memory.SemanticSearchOptions{Limit: 5, TenantID: "default", TeamID: strings.TrimSpace(teamID),
		AgentID: strings.TrimSpace(agentID), Types: governedClasses, AllowGlobal: true}
	if !lead {
		opts.Types = workerClasses
		opts.ExcludeSensitivity = []string{"restricted"}
	}
	return opts
}

// writeDeploymentContext injects governed sources the turn's requesting user
// may read (SRU). A failed identity or membership lookup injects none and
// says so; it never falls back to the unscoped set.
func (r *InternalToolRegistry) writeDeploymentContext(sb *strings.Builder, access RecallAccess, agentID, teamID, currentInput string, lead bool) []ContextSource {
	if r.mem == nil || strings.TrimSpace(currentInput) == "" {
		return nil
	}
	query := recallQuery(currentInput, 240)
	if query == "" {
		return nil
	}
	reader, err := access.governedReader()
	if err != nil {
		sb.WriteString("### Saved Memory\n" + recallUnavailableNote + " Do not claim saved memory you did not receive.\n\n")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := governedRecallOptions(agentID, teamID, lead)
	opts.Reader = reader
	// Semantic when an embedding engine works; PostgreSQL keyword ranking otherwise.
	results, _, err := r.mem.RecallGoverned(ctx, r.brain, query, opts)
	if err != nil || len(results) == 0 {
		return nil
	}
	sections := []struct{ class, heading, fallback string }{
		{"customer_context", "### Customer Context Store (operator-provided documents)", "operator provided"},
		{"company_knowledge", "### Company Knowledge Store (approved Soma/company content)", "approved company knowledge"},
		{"soma_operating_context", "### Admin-Shaped Soma Context (organization-owned Soma operating guidance)", "admin-shaped Soma context"},
		{"user_private_context", "### User-Private Context Store (private records and goal-scoped references)", "private user context"},
		{"reflection_synthesis", "### Reflection / Synthesis Memory (lessons, patterns, contradictions, and trajectory shifts)", "reflection synthesis"},
	}
	var sources []ContextSource
	for _, section := range sections {
		wrote := false
		for _, result := range results {
			class := stringMeta(result.Metadata, "knowledge_class")
			if class != section.class && !(section.class == "customer_context" && !knownGovernedClass(class)) {
				continue
			}
			if !wrote {
				sb.WriteString(section.heading + "\n")
				wrote = true
			}
			sources = append(sources, writeKnowledgeResult(sb, result, section.fallback))
		}
		if wrote {
			sb.WriteString("\n")
		}
	}
	sb.WriteString("When your reply uses one of these sources, cite it as [source: <title>]. Do not claim a source you did not use.\n\n")
	return sources
}

// requestMarker ends every header Core's chat handler wraps around the latest
// turn: the governance profile, then the route header (server cognitive.go,
// action_governance_profile.go). It is the shared protocol constant so all
// three wrap sites and this unwrap site cannot drift apart.
const requestMarker = protocol.ChatOriginalRequestMarker

// recallQuery is the text recall ranks on: the operator's own request. A turn
// that starts with a bracketed header is unwrapped past each header's
// "Original request:" marker, so boilerplate never fills the bounded query.
// Only the ranking text changes; the caller's scope clauses are untouched.
// The result is capped at max bytes on a rune boundary.
func recallQuery(input string, max int) string {
	query := strings.TrimSpace(input)
	for strings.HasPrefix(query, "[") {
		idx := strings.Index(query, requestMarker)
		if idx < 0 {
			break
		}
		rest := query[idx+len(requestMarker):]
		if rest != "" && rest[0] != '\n' {
			break
		}
		query = strings.TrimSpace(rest)
	}
	if len(query) > max {
		cut := max
		for cut > 0 && !utf8.RuneStart(query[cut]) {
			cut--
		}
		query = query[:cut]
	}
	return query
}

func knownGovernedClass(class string) bool {
	for _, known := range governedClasses {
		if class == known {
			return true
		}
	}
	return false
}

func writeKnowledgeResult(sb *strings.Builder, result memory.VectorResult, fallbackSourceLabel string) ContextSource {
	title := "Knowledge entry"
	if value := stringMeta(result.Metadata, "artifact_title"); value != "" {
		title = value
	}
	sourceLabel := fallbackSourceLabel
	if value := stringMeta(result.Metadata, "source_label"); value != "" {
		sourceLabel = value
	}
	// One line per source, so prompt line filtering keeps or drops it whole.
	preview := strings.Join(strings.Fields(result.Content), " ")
	if len(preview) > 600 {
		preview = preview[:600] + "..."
	}
	sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", title, sourceLabel, preview))
	return ContextSource{Excerpt: preview, ContextSourceRef: protocol.ContextSourceRef{
		ArtifactID: stringMeta(result.Metadata, "artifact_id"), Title: title,
		KnowledgeClass: stringMeta(result.Metadata, "knowledge_class"), RetrievalMode: result.RetrievalMode,
	}}
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

// goalSetArg accepts an explicit goal_set tool argument as a string or list.
func goalSetArg(v any) []string {
	if single := strings.TrimSpace(stringValue(v)); single != "" {
		return []string{single}
	}
	return stringSlice(v)
}
