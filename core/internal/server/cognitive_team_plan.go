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
	agents := specialistAgentsForTeamRequest(teamID, lower)
	staffingMode := "lead_only_start"
	initialMemberCount := 1
	if len(agents) > 0 {
		staffingMode = "specialist_delivery"
		initialMemberCount = len(agents)
	}

	args := map[string]any{
		"team_id":                     teamID,
		"name":                        name,
		"role":                        role,
		"goal":                        trimmed,
		"staffing_mode":               staffingMode,
		"initial_member_count":        initialMemberCount,
		"recommended_member_limit":    maxTeamMemberLimit(3, initialMemberCount),
		"expansion_policy":            "operator_adds_members_or_team_lead_requests_temp_specialist_with_reason",
		"temporary_addition_guidance": "Add specialists only after the lead names the missing capability, owned task, proof expected, and removal point.",
	}
	if len(agents) > 0 {
		args["agents"] = agents
		args["required_capabilities"] = []string{"team_orchestration", "generate_image", "save_cached_image"}
	}

	return protocol.PlannedToolCall{Name: "create_team", Arguments: args}, true
}

func specialistAgentsForTeamRequest(teamID, lower string) []map[string]any {
	if !requestContainsAny(lower, []string{"comic", "artist", "character", "dialogue", "lines", "specialist", "specialists", "members"}) {
		return nil
	}
	base := slugID(teamID)
	if base == "" {
		base = "runtime-team"
	}
	if strings.Contains(lower, "comic") || requestContainsAny(lower, []string{"artist", "character", "dialogue", "lines"}) {
		return []map[string]any{
			specialistAgent(base, "lead", "creative lead", "Coordinate the comic page work, keep scope bounded, and assemble the final output/proof for Soma."),
			specialistAgent(base, "story", "story lead", "Define the page beat, panel-by-panel story intent, and emotional arc."),
			specialistAgent(base, "characters", "character designer", "Define visual identity, silhouettes, props, and consistency notes for the characters."),
			specialistAgent(base, "dialogue", "dialogue writer", "Write concise speech-balloon guidance and captions for the page."),
			specialistAgent(base, "layout", "panel layout artist", "Own panel composition, camera flow, gutters, and generation prompt visual structure."),
			specialistAgent(base, "proof", "proof editor", "Check output readiness, local/private media boundary, retained artifact path, and recovery notes."),
		}
	}
	return []map[string]any{
		specialistAgent(base, "lead", "team lead", "Coordinate the requested work and return retained outputs through Soma."),
	}
}

func specialistAgent(teamID, suffix, role, prompt string) map[string]any {
	return map[string]any{
		"id":            teamID + "-" + suffix,
		"role":          role,
		"system_prompt": prompt,
		"tools":         []string{"generate_image", "save_cached_image", "store_artifact"},
	}
}

func maxTeamMemberLimit(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func toolsForPlannedCalls(planned []protocol.PlannedToolCall, fallback []string) []string {
	if len(planned) == 0 {
		return uniqueOrderedTools(fallback)
	}
	tools := make([]string, 0, len(planned))
	for _, call := range planned {
		if tool := strings.TrimSpace(firstNonEmptyString(call.ToolRef, call.Name)); tool != "" {
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
	for _, marker := range []string{".", " and get ", " then ", " role ", " with ", " no web", " return "} {
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
