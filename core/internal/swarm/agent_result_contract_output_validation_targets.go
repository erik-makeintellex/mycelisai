package swarm

import (
	"regexp"
	"strings"
)

func outputValidationCanvasTargetRendered(content, target string) bool {
	script := outputValidationJavaScript(content)
	if !outputValidationDrawingEffectPattern.MatchString(script) {
		return false
	}
	for _, tag := range outputValidationTargetTags(content, target) {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(tag)), "<canvas") {
			return true
		}
	}
	return false
}

func outputValidationTargetTags(content, target string) []string {
	match := outputValidationAttributeSelector.FindStringSubmatch(strings.TrimSpace(target))
	if len(match) != 2 {
		return nil
	}
	attribute := regexp.QuoteMeta(match[1])
	pattern := regexp.MustCompile(`(?is)<[^>]*\s` + attribute + `(?:\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+))?[^>]*>`)
	return pattern.FindAllString(content, -1)
}

func outputValidationTargetHasDirectMutation(script, candidate, propertyPattern string) bool {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return false
	}
	if selectorMutatesProperty(script, candidate, propertyPattern) {
		return true
	}
	if strings.HasPrefix(candidate, "#") && selectorMutatesProperty(script, strings.TrimPrefix(candidate, "#"), propertyPattern) {
		return true
	}
	return targetVariableMutates(script, candidate, propertyPattern)
}

func outputValidationTargetHasMethodMutation(script, candidate string) bool {
	methodPattern := `(?:classList\.(?:add|remove|toggle|replace)\s*\(|style\.[A-Za-z][A-Za-z0-9-]*\s*=|dataset\.[A-Za-z_$][A-Za-z0-9_$]*\s*=|setAttribute\s*\()`
	return outputValidationTargetHasDirectMutation(script, candidate, methodPattern)
}

func selectorMutatesProperty(script, selector, propertyPattern string) bool {
	quoted := regexp.QuoteMeta(selector)
	return regexp.MustCompile(`(?is)(?:querySelector|getElementById)\(\s*["']` + quoted + `["']\s*\)\s*\.\s*` + propertyPattern + `\s*=`).MatchString(script)
}

func targetVariableMutates(script, selector, propertyPattern string) bool {
	quoted := regexp.QuoteMeta(selector)
	variablePattern := regexp.MustCompile(`(?is)(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*(?:document\.)?(?:querySelector|getElementById)\(\s*["']` + quoted + `["']\s*\)`)
	for _, match := range variablePattern.FindAllStringSubmatch(script, -1) {
		if len(match) >= 2 && regexp.MustCompile(`(?m)\b`+regexp.QuoteMeta(match[1])+`\s*\.\s*`+propertyPattern+`\s*=?`).MatchString(script) {
			return true
		}
	}
	return false
}
