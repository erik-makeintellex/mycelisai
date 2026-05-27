package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/searchcap"
)

const somaAttunementContextHeader = "[SOMA UNDERSTANDING ATTUNEMENT]"

func buildSomaAttunementContext(latestRequest string, status searchcap.Status) string {
	trimmed := strings.TrimSpace(latestRequest)
	if trimmed == "" {
		return ""
	}

	lines := []string{
		somaAttunementContextHeader,
		"Treat the latest operator message as intent expression, not as a bare command or database query.",
		"Before answering or proposing execution, infer the desired outcome, audience, output form, constraints, and uncertainty.",
		"Use current workspace/deployment context before asking the operator to repeat information.",
	}

	lines = append(lines, fmt.Sprintf("Likely output: %s.", inferAttunementOutputForm(trimmed)))
	lines = append(lines, fmt.Sprintf("Knowledge posture: %s.", attunementKnowledgePosture(trimmed, status)))
	lines = append(lines, "If external or current knowledge is needed and the configured source is unavailable, surface a blocker or recovery path instead of inventing facts.")
	lines = append(lines, "Ask at most one clarifying question only when the missing information would materially change a governed action or the quality of the deliverable; otherwise proceed with explicit assumptions.")
	lines = append(lines, "Do not expose raw topology by default. Translate capability, team, proof, and recovery details into the operator-facing Soma workflow.")
	return strings.Join(lines, "\n")
}

func prependSomaAttunementContext(messages []chatRequestMessage, latestRequest string, status searchcap.Status) []chatRequestMessage {
	context := buildSomaAttunementContext(latestRequest, status)
	if strings.TrimSpace(context) == "" {
		return messages
	}
	out := make([]chatRequestMessage, 0, len(messages)+1)
	out = append(out, chatRequestMessage{Role: "system", Content: context})
	out = append(out, messages...)
	return out
}

func inferAttunementOutputForm(text string) string {
	lower := strings.ToLower(strings.Join(strings.Fields(text), " "))
	switch {
	case requestContainsAny(lower, []string{"report", "research", "investigate", "findings", "subject"}):
		return "a sourced report or investigation package with assumptions, sources, outputs, and proof"
	case requestContainsAny(lower, []string{"comic", "image", "illustration", "picture", "visual", "artwork", "media"}):
		return "a retained visual/media deliverable with prompt, artifact, output path, and proof"
	case requestContainsAny(lower, []string{"game", "playable", "browser game", "html", "canvas"}):
		return "a retained playable project package with files, validation, openable entrypoint, and proof"
	case requestContainsAny(lower, []string{"team", "specialist", "members", "orchestrate", "delegate"}):
		return "a bounded Soma-directed team/work item with owner, expected output, status, and proof"
	case requestContainsAny(lower, []string{"configure", "setting", "settings", "auth", "provider", "mcp", "tool"}):
		return "a governed configuration or capability change with risk, approval, recovery, and proof"
	case requestContainsAny(lower, []string{"file", "document", "markdown", "readme", "docs", "documentation"}):
		return "a retained file/document output with clear location and proof"
	default:
		return "a concise answer or governed work proposal matched to the operator's intended outcome"
	}
}

func attunementKnowledgePosture(text string, status searchcap.Status) string {
	lower := strings.ToLower(strings.Join(strings.Fields(text), " "))
	needsFreshness := requestContainsAny(lower, []string{
		"latest", "current", "today", "recent", "news", "up to date", "research", "look up", "lookup", "search", "browse", "online",
	})
	source := attunementSearchSourceLabel(status)
	if needsFreshness {
		return fmt.Sprintf("consult configured source context when safe, disclose the source boundary as %s, and treat external results as leads requiring interpretation", source)
	}
	if status.Enabled && status.SupportsLocalSources {
		return "prefer local Mycelis context and retained deployment knowledge; use external search only when the operator asks for current or outside-world facts"
	}
	return "use available conversation/workspace context; name assumptions when source lookup is not configured or not needed"
}

func attunementSearchSourceLabel(status searchcap.Status) string {
	provider := strings.TrimSpace(status.Provider)
	switch {
	case !status.Enabled || provider == "" || provider == searchcap.ProviderDisabled:
		return "search unavailable"
	case status.SupportsLocalSources && provider == searchcap.ProviderLocalSources:
		return "Local Mycelis context"
	case status.SupportsPublicWeb:
		return provider
	default:
		return provider
	}
}
