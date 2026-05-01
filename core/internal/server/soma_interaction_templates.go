package server

import "strings"

type somaInteractionTheme struct {
	ID            string
	Label         string
	Phrases       []string
	Concepts      []string
	MutationTools []string
	Protected     bool
}

type somaInteractionTemplate struct {
	ID                 string
	Label              string
	ThemeIDs           []string
	ActionSummary      string
	ProtectionReason   string
	ConfirmationPrompt string
	TargetMCPHints     []string
	MutationTools      []string
	Protected          bool
}

type somaInteractionMatch struct {
	Template           somaInteractionTemplate
	Themes             []somaInteractionTheme
	Concepts           []string
	TargetMCPHints     []string
	MutationTools      []string
	Protected          bool
	ProtectionReason   string
	ConfirmationPrompt string
}

func defaultSomaInteractionThemes() []somaInteractionTheme {
	return []somaInteractionTheme{
		{
			ID:            "team_manifestation",
			Label:         "Create or reshape a team",
			Phrases:       []string{"create team", "build team", "launch team", "instantiate team", "manifest team", "put together", "orchestrate team", "specialists", "members", "lanes"},
			Concepts:      []string{"team", "specialists", "outputs"},
			MutationTools: []string{"generate_blueprint", "delegate"},
			Protected:     true,
		},
		{
			ID:       "mcp_tool_posture",
			Label:    "Review connected tools",
			Phrases:  []string{"mcp", "tool", "tools", "web search", "github", "fetch", "browser", "host data", "shared-sources"},
			Concepts: []string{"mcp", "tool_visibility", "connected_tools"},
		},
		{
			ID:            "mcp_enablement",
			Label:         "Enable or bind connected tools",
			Phrases:       []string{"enable mcp", "install mcp", "connect mcp", "associate mcp", "configure mcp", "assign tools", "enable tools", "bind tools", "walk me through enabling"},
			Concepts:      []string{"mcp", "tool_binding", "connected_tools"},
			MutationTools: []string{"delegate"},
			Protected:     true,
		},
		{
			ID:        "private_service",
			Label:     "Use a private or production service",
			Phrases:   []string{"private service", "internal service", "production service", "private api", "customer system", "client system", "api key", "token", "credential"},
			Concepts:  []string{"private_service", "credential_boundary", "audit"},
			Protected: true,
		},
		{
			ID:        "private_data",
			Label:     "Use private or governed data",
			Phrases:   []string{"private data", "customer data", "user private", "deployment context", "company knowledge", "sensitive", "confidential", "workspace/shared-sources"},
			Concepts:  []string{"private_data", "visibility_scope", "lineage"},
			Protected: true,
		},
		{
			ID:        "recurring_behavior",
			Label:     "Make behavior repeatable",
			Phrases:   []string{"always", "every time", "from now on", "standing behavior", "repetitively", "recurring", "template", "reuse"},
			Concepts:  []string{"recurring_policy", "conversation_template"},
			Protected: true,
		},
		{
			ID:       "referential_review",
			Label:    "Review prior context before action",
			Phrases:  []string{"review my request", "match prior context", "related prior commands", "tool metadata", "infer the action", "ask me to confirm"},
			Concepts: []string{"referential_review", "confirmation"},
		},
	}
}

func defaultSomaInteractionTemplates() []somaInteractionTemplate {
	return []somaInteractionTemplate{
		{
			ID:                 "team-with-target-tools",
			Label:              "Compact team with target tools",
			ThemeIDs:           []string{"team_manifestation", "mcp_tool_posture"},
			ActionSummary:      "confirm compact team manifestation with specialist roles, retained outputs, and target MCP/tool bindings",
			ProtectionReason:   "creating teams and assigning tools changes governed runtime structure",
			ConfirmationPrompt: "Confirm whether Soma should create the team plan once, then route it through proposal approval.",
			TargetMCPHints:     []string{"fetch", "filesystem", "github", "brave-search", "web_search"},
			MutationTools:      []string{"generate_blueprint", "delegate"},
			Protected:          true,
		},
		{
			ID:                 "protected-service-action",
			Label:              "Private service or credentialed action",
			ThemeIDs:           []string{"private_service"},
			ActionSummary:      "confirm protected private-service action and required MCP/credential boundary",
			ProtectionReason:   "private service access may use credentials, external systems, or production data",
			ConfirmationPrompt: "Confirm the target service, allowed action, and whether this is one-time or reusable.",
			TargetMCPHints:     []string{"github", "fetch", "brave-search"},
			MutationTools:      []string{"delegate"},
			Protected:          true,
		},
		{
			ID:                 "private-data-review",
			Label:              "Private data review",
			ThemeIDs:           []string{"private_data"},
			ActionSummary:      "confirm governed private-data review with visibility and source boundaries",
			ProtectionReason:   "private or customer data must keep source, visibility, and audit boundaries explicit",
			ConfirmationPrompt: "Confirm the data source, visibility scope, and whether outputs can be retained.",
			TargetMCPHints:     []string{"filesystem", "web_search"},
			MutationTools:      []string{"delegate"},
			Protected:          true,
		},
		{
			ID:                 "recurring-interaction-template",
			Label:              "Reusable protected interaction",
			ThemeIDs:           []string{"recurring_behavior"},
			ActionSummary:      "confirm reusable conversation template behavior and approval posture",
			ProtectionReason:   "recurring interaction behavior changes how future Soma requests are interpreted",
			ConfirmationPrompt: "Confirm whether to store this as a reusable conversation template and which protections apply.",
			MutationTools:      []string{"delegate"},
			Protected:          true,
		},
		{
			ID:            "referential-action-review",
			Label:         "Referential review before action",
			ThemeIDs:      []string{"referential_review"},
			ActionSummary: "review current request, related history, and tool metadata before asking for confirmation",
		},
	}
}

func matchSomaInteractionTemplate(text string) somaInteractionMatch {
	lower := normalizeIntentText(text)
	if lower == "" {
		return somaInteractionMatch{}
	}
	themes := matchedSomaInteractionThemes(lower)
	template := bestSomaInteractionTemplate(themes)
	match := somaInteractionMatch{Template: template, Themes: themes}
	for _, theme := range themes {
		match.Concepts = append(match.Concepts, theme.Concepts...)
		match.MutationTools = append(match.MutationTools, theme.MutationTools...)
		if theme.Protected {
			match.Protected = true
		}
	}
	if template.ID != "" {
		match.MutationTools = append(match.MutationTools, template.MutationTools...)
		match.TargetMCPHints = append(match.TargetMCPHints, template.TargetMCPHints...)
		match.ProtectionReason = template.ProtectionReason
		match.ConfirmationPrompt = template.ConfirmationPrompt
		if template.Protected {
			match.Protected = true
		}
	}
	match.Concepts = uniqueOrderedTools(match.Concepts)
	match.MutationTools = uniqueOrderedTools(match.MutationTools)
	match.TargetMCPHints = uniqueOrderedTools(match.TargetMCPHints)
	return match
}

func matchedSomaInteractionThemes(lower string) []somaInteractionTheme {
	var matched []somaInteractionTheme
	for _, theme := range defaultSomaInteractionThemes() {
		for _, phrase := range theme.Phrases {
			if strings.Contains(lower, normalizeIntentText(phrase)) {
				matched = append(matched, theme)
				break
			}
		}
	}
	return matched
}

func bestSomaInteractionTemplate(themes []somaInteractionTheme) somaInteractionTemplate {
	if len(themes) == 0 {
		return somaInteractionTemplate{}
	}
	themeIDs := map[string]bool{}
	for _, theme := range themes {
		themeIDs[theme.ID] = true
	}
	var best somaInteractionTemplate
	bestScore := 0
	for _, tpl := range defaultSomaInteractionTemplates() {
		score := 0
		for _, id := range tpl.ThemeIDs {
			if themeIDs[id] {
				score++
			}
		}
		if len(tpl.ThemeIDs) > 1 && score != len(tpl.ThemeIDs) {
			continue
		}
		if score > bestScore {
			best = tpl
			bestScore = score
		}
	}
	if bestScore == 0 {
		return somaInteractionTemplate{}
	}
	return best
}
