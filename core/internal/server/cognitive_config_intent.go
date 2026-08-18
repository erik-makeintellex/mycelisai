package server

import (
	"strings"

	"github.com/mycelis/core/internal/somacommands"
)

func outcomeTemplateMutationTools(lower string) ([]string, bool) {
	if !requestContainsAny(lower, []string{"outcome template", "outcome-template"}) {
		return nil, false
	}
	if outcomeTemplateWorkApplication(lower) {
		return nil, false
	}

	var tools []string
	if matchesConfiguredSomaCommandQuote(lower, "activate_config_document") ||
		requestContainsAny(lower, []string{"roll back", "rollback"}) ||
		requestContainsAny(lower, []string{"make active", "set active", "from now on"}) ||
		hasAnyExactWord(lower, "activate") {
		tools = append(tools, "activate_config_document")
	}
	if matchesConfiguredSomaCommandQuote(lower, "store_config_document") ||
		hasAnyExactWord(lower, "save", "store", "persist") {
		tools = append(tools, "store_config_document")
	}
	return uniqueOrderedTools(tools), true
}

func outcomeTemplateWorkApplication(lower string) bool {
	if !requestContainsAny(lower, []string{"outcome template", "outcome-template"}) {
		return false
	}
	if requestContainsAny(lower, []string{
		"save this outcome template", "save the outcome template",
		"store this outcome template", "store the outcome template",
		"preview this outcome template", "validate this outcome template",
	}) {
		return false
	}
	if !hasAnyExactWord(lower, "use", "using", "apply") {
		return false
	}
	if requestContainsAny(lower, []string{"for this work", "for this task", "shape this work"}) {
		return true
	}
	return hasAnyExactWord(
		lower, "build", "create", "write", "generate", "produce",
		"deliver", "implement", "update", "run", "execute",
	)
}

func inferReadOnlyConfigToolsFromText(text string) []string {
	lower := strings.ToLower(strings.TrimSpace(text))
	if !requestContainsAny(lower, []string{"outcome template", "outcome-template"}) {
		return nil
	}
	if matchesConfiguredSomaCommandQuote(lower, "preview_config_document") ||
		requestContainsAny(lower, []string{"preview", "dry run", "dry-run"}) ||
		hasAnyExactWord(lower, "draft", "validate", "review", "create", "compose", "design") {
		return []string{"preview_config_document"}
	}
	return nil
}

func matchesConfiguredSomaCommandQuote(text, handler string) bool {
	registry, err := somacommands.LoadDefault()
	if err != nil {
		return false
	}
	command, ok := registry.ByHandler()[handler]
	if !ok {
		return false
	}
	quote := normalizeIntentText(command.UserQuote)
	return quote != "" && strings.Contains(normalizeIntentText(text), quote)
}

func hasAnyExactWord(text string, words ...string) bool {
	for _, word := range words {
		if hasExactWord(text, word) {
			return true
		}
	}
	return false
}
