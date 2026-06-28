package server

import "strings"

func contentMarketingHandoffMarkdown(request, sourceTeamID, marketingTeamID, evidencePath string) string {
	var b strings.Builder
	b.WriteString("# Marketing Handoff\n\n")
	b.WriteString("Soma retained this handoff so the marketing lane can react to verified output changes without inventing claims.\n\n")
	b.WriteString("## Operator request\n\n")
	b.WriteString(request)
	b.WriteString("\n\n")
	b.WriteString("## Team routing\n\n")
	b.WriteString("- Source delivery team: ")
	b.WriteString(sourceTeamID)
	b.WriteString("\n")
	b.WriteString("- Marketing team: ")
	b.WriteString(marketingTeamID)
	b.WriteString("\n")
	b.WriteString("- Evidence examples: ")
	b.WriteString(evidencePath)
	b.WriteString("\n\n")
	writeMarkdownList(&b, "Marketing responsibilities", []string{
		"Use retained examples as the evidence base.",
		"Generate messaging, positioning, and launch copy only from confirmed source facts.",
		"Ask Soma for clarification or more proof if a claim is not supported.",
	})
	writeMarkdownList(&b, "Required marketing output", []string{
		"Short positioning statement",
		"Three feature bullets grounded in evidence examples",
		"One release/social blurb",
		"Assumptions or missing-proof notes",
	})
	return b.String()
}

func sourceChangeEvidenceMarkdown(request, sourceTeamID, marketingTeamID string) string {
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(sourceEvidenceTitle(request))
	b.WriteString("\n\n")
	b.WriteString("Soma retained this as the evidence package the marketing team should use after the source team improves the deliverable.\n\n")
	b.WriteString("## Operator request\n\n")
	b.WriteString(request)
	b.WriteString("\n\n")
	b.WriteString("## Teams\n\n")
	b.WriteString("- Source delivery team: ")
	b.WriteString(sourceTeamID)
	b.WriteString("\n")
	b.WriteString("- Marketing team to notify: ")
	b.WriteString(marketingTeamID)
	b.WriteString("\n\n")
	writeMarkdownList(&b, "Examples the source team must provide", sourceEvidenceExamplePrompts(request))
	writeMarkdownList(&b, "Marketing-safe claim boundary", []string{
		"Use only behaviors shown in these examples as confirmed claims.",
		"Label unverified ideas as proposed positioning.",
		"Send the marketing handoff back through Soma when the source team updates this evidence.",
	})
	return b.String()
}

func sourceEvidenceTitle(request string) string {
	lower := strings.ToLower(request)
	switch {
	case requestContainsAny(lower, []string{"game", "gameplay", "playable"}):
		return "Gameplay Change Examples"
	case strings.Contains(lower, "media"):
		return "Media Change Examples"
	case strings.Contains(lower, "document") || strings.Contains(lower, "report"):
		return "Content Change Examples"
	case strings.Contains(lower, "app") || strings.Contains(lower, "application"):
		return "Usage Change Examples"
	default:
		return "Source Change Examples"
	}
}

func sourceEvidencePathForMarketingRequest(request, sourceTeamID string) string {
	lower := strings.ToLower(request)
	filename := "SOURCE_CHANGE_EXAMPLES.md"
	switch {
	case requestContainsAny(lower, []string{"game", "gameplay", "playable"}):
		filename = "GAMEPLAY_CHANGE_EXAMPLES.md"
	case strings.Contains(lower, "media"):
		filename = "MEDIA_CHANGE_EXAMPLES.md"
	case strings.Contains(lower, "document") || strings.Contains(lower, "report"):
		filename = "CONTENT_CHANGE_EXAMPLES.md"
	case strings.Contains(lower, "app") || strings.Contains(lower, "application"):
		filename = "USAGE_CHANGE_EXAMPLES.md"
	}
	return groupWorkspaceFolderForTeamID(sourceTeamID) + "/proof/" + filename
}

func sourceImprovementExitCriteria(request string) []string {
	lower := strings.ToLower(request)
	if requestContainsAny(lower, []string{"game", "gameplay", "playable"}) {
		return []string{"game change is visible in play", "gameplay examples show changed behavior", "marketing team has a retained handoff input"}
	}
	return []string{"deliverable change is visible or reviewable", "evidence examples show changed behavior or content", "marketing team has a retained handoff input"}
}

func sourceEvidenceRequirements(request string) []string {
	lower := strings.ToLower(request)
	if requestContainsAny(lower, []string{"game", "gameplay", "playable"}) {
		return []string{"direct launch path", "gameplay examples", "proof notes"}
	}
	return []string{"direct open path", "source change examples", "proof notes"}
}

func sourceExpectedOutputs(request string) []string {
	lower := strings.ToLower(request)
	if requestContainsAny(lower, []string{"game", "gameplay", "playable"}) {
		return []string{"improved game output", "gameplay examples showing the changes"}
	}
	return []string{"improved source output", "evidence examples showing the changes"}
}

func sourceEvidenceExamplePrompts(request string) []string {
	lower := strings.ToLower(request)
	switch {
	case requestContainsAny(lower, []string{"game", "gameplay", "playable"}):
		return []string{
			"Start-state screenshot or note showing the changed objective or level entry.",
			"Mid-play screenshot or note showing the changed mechanic, enemy, pickup, audio cue, or route.",
			"Win/fail/restart screenshot or note showing the changed outcome behavior.",
			"Direct launch path and controls used for the proof run.",
		}
	case strings.Contains(lower, "media"):
		return []string{
			"Preview or saved artifact path showing the changed media direction.",
			"Notes naming the changed visual/audio/content element.",
			"Review notes separating confirmed asset state from proposed campaign claims.",
		}
	case strings.Contains(lower, "document") || strings.Contains(lower, "report"):
		return []string{
			"Before/after notes or excerpts showing the changed document section.",
			"Readback path for the updated source document.",
			"Claims or assumptions the marketing team may safely reference.",
		}
	default:
		return []string{
			"Open/view path showing the changed deliverable.",
			"Notes naming the changed behavior, data, content, or user-facing state.",
			"Evidence the marketing team can cite without inventing claims.",
		}
	}
}

func groupIDsFromText(text string) []string {
	matches := groupScopedPathPattern.FindAllStringSubmatch(text, -1)
	var ids []string
	for _, match := range matches {
		if len(match) == 2 {
			ids = append(ids, strings.TrimSpace(match[1]))
		}
	}
	return uniqueOrderedTools(ids)
}

func extractLabeledTokenAfterMarker(text, marker, label string) string {
	lower := strings.ToLower(text)
	idx := strings.Index(lower, marker)
	if idx < 0 {
		return ""
	}
	return extractLabeledToken(text[idx:], label)
}
