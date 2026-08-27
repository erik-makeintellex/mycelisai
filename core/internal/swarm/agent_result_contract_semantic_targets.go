package swarm

import (
	"fmt"
	"strings"
)

var semanticCriterionTargets = map[string][]string{
	"visible game instructions identify controls, objective, and restart": {
		"[data-mycelis-game-instructions]",
	},
	"playable controls move the player and change the visible game surface": {"#game"},
	"attack changes enemy, hazard, or score state": {
		`[data-mycelis-validation-action="attack"]`, "#score",
	},
	"hazard contact changes health state": {
		`[data-mycelis-validation-action="hazard"]`, "#health",
	},
	"key pickup changes key and score state": {
		`[data-mycelis-validation-action="key"]`, "#keyState", "#score",
	},
	"team play-tests the documented winning route through the locked door objective to win state": {
		`[data-mycelis-validation-action="win"]`, "#goalState",
	},
	"fail state can transition through restart to the initial objective": {
		`[data-mycelis-validation-action="fail"]`, `[data-mycelis-validation-action="restart"]`, "#goalState",
	},
	"audio control changes its visible state": {
		`[data-mycelis-validation-action="audio"]`,
	},
	"requested revision changes its declared visible revision state": {
		`[data-mycelis-validation-action="revision"]`, "[data-mycelis-revision-state]",
	},
}

func semanticAcceptanceEntrypointIssues(criteria []string, content string) []string {
	issues := []string{}
	for _, criterion := range criteria {
		for _, target := range semanticCriterionTargets[strings.ToLower(strings.TrimSpace(criterion))] {
			if semanticValidationTargetPresent(content, target) {
				continue
			}
			issues = append(issues, fmt.Sprintf("entrypoint readback is missing semantic validation target %s", target))
		}
	}
	return uniqueResultContractStrings(issues)
}

func semanticValidationTargetPresent(content, selector string) bool {
	if strings.HasPrefix(selector, "#") {
		id := strings.TrimPrefix(selector, "#")
		return strings.Contains(content, `id="`+id+`"`) || strings.Contains(content, `id='`+id+`'`)
	}
	if strings.HasPrefix(selector, "[") && strings.HasSuffix(selector, "]") && strings.Contains(selector, "=") {
		attribute := strings.TrimSuffix(strings.TrimPrefix(selector, "["), "]")
		if strings.Contains(content, attribute) {
			return true
		}
		return strings.Contains(content, strings.ReplaceAll(attribute, `"`, `'`))
	}
	return outputValidationTargetPresent(content, selector)
}
