package server

import (
	"strings"
	"unicode"

	"github.com/mycelis/core/pkg/protocol"
)

func inferCreateTeamPlanFromRequest(text string) (protocol.PlannedToolCall, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	if trimmed == "" || !strings.Contains(lower, "team") {
		return protocol.PlannedToolCall{}, false
	}
	if !strings.Contains(lower, "create") && !strings.Contains(lower, "form") && !strings.Contains(lower, "assemble") && !strings.Contains(lower, "need") {
		return protocol.PlannedToolCall{}, false
	}

	teamID := extractLabeledToken(trimmed, "team_id")
	if teamID == "" {
		teamID = extractLabeledToken(trimmed, "team id")
	}
	name := extractNamedTeam(trimmed)
	if name == "" {
		name = extractLabeledToken(trimmed, "team_name")
	}
	if name == "" {
		if teamID != "" {
			name = humanizeSlugID(teamID)
		} else if strings.Contains(lower, "research") {
			name = "AI Research Team"
		} else {
			name = "Soma Requested Team"
		}
	}
	if teamID == "" {
		teamID = slugID(name)
	}
	role := "worker"
	if strings.Contains(lower, "research") {
		role = "researcher"
	}

	return protocol.PlannedToolCall{
		Name: "create_team",
		Arguments: map[string]any{
			"team_id": teamID,
			"name":    name,
			"role":    role,
			"goal":    trimmed,
		},
	}, true
}

func toolsForPlannedCalls(planned []protocol.PlannedToolCall, fallback []string) []string {
	if len(planned) == 0 {
		return uniqueOrderedTools(fallback)
	}
	tools := make([]string, 0, len(planned))
	for _, call := range planned {
		if tool := strings.TrimSpace(call.Name); tool != "" {
			tools = append(tools, tool)
		}
	}
	if len(tools) == 0 {
		return uniqueOrderedTools(fallback)
	}
	return uniqueOrderedTools(tools)
}

func extractLabeledToken(text, label string) string {
	lower := strings.ToLower(text)
	idx := strings.Index(lower, label)
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(text[idx+len(label):])
	rest = strings.TrimLeft(rest, " :=#")
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[0], `"'.,;`)
}

func extractNamedTeam(text string) string {
	lower := strings.ToLower(text)
	idx := strings.Index(lower, " named ")
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(text[idx+len(" named "):])
	if rest == "" {
		return ""
	}
	for _, marker := range []string{".", " role ", " with ", " no web", " return "} {
		if cut := strings.Index(strings.ToLower(rest), marker); cut >= 0 {
			rest = strings.TrimSpace(rest[:cut])
		}
	}
	return strings.Trim(rest, `"'.,;`)
}

func humanizeSlugID(text string) string {
	normalized := strings.NewReplacer("-", " ", "_", " ").Replace(strings.TrimSpace(text))
	fields := strings.Fields(normalized)
	for i, field := range fields {
		runes := []rune(field)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		fields[i] = string(runes)
	}
	return strings.Join(fields, " ")
}

func slugID(text string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
