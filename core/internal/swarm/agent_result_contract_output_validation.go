package swarm

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

var (
	outputValidationAttributeSelector       = regexp.MustCompile(`^\[([a-zA-Z_:][a-zA-Z0-9_:.-]*)\]$`)
	resultContractInteractiveHandlerPattern = regexp.MustCompile(`(?i)(?:addEventListener\s*\(\s*["'](?:click|pointerdown|touchstart|keydown|keyup)|on(?:click|pointerdown|touchstart|keydown)\s*=)`)
	resultContractVisibleControlPattern     = regexp.MustCompile(`(?i)\b(?:click|tap|press|use|move|drag|select|arrow|space|wasd|control)\b`)
	resultContractScriptOrStylePattern      = regexp.MustCompile(`(?is)<(?:script|style)\b[^>]*>.*?</(?:script|style)>`)
	resultContractHTMLTagPattern            = regexp.MustCompile(`(?s)<[^>]+>`)
)

func resultContractRequiresPrimaryInteraction(requirement *teamResultRequirement) bool {
	if requirement == nil {
		return false
	}
	values := append(append([]string{}, requirement.ExpectedOutputs...), requirement.AcceptanceCriteria...)
	for _, value := range values {
		lower := strings.ToLower(value)
		if strings.Contains(lower, "playable") || strings.Contains(lower, "browser game") ||
			strings.Contains(lower, "controls respond") || strings.Contains(lower, "primary user workflow") ||
			strings.Contains(lower, "primary control") {
			return true
		}
	}
	return false
}

func latestEntrypointEvidenceContent(evidence []successfulToolEvidence, entrypoint string) string {
	content := ""
	for _, item := range evidence {
		if (item.ToolName == "write_file" || item.ToolName == "read_file" || item.ToolName == "read_text_file") &&
			evidenceContainsPath([]string{item.Path}, entrypoint) && strings.TrimSpace(item.Content) != "" {
			content = item.Content
		}
	}
	return content
}

func resultContractExposesPrimaryControl(content string) bool {
	visibleText := resultContractScriptOrStylePattern.ReplaceAllString(content, " ")
	visibleText = html.UnescapeString(resultContractHTMLTagPattern.ReplaceAllString(visibleText, " "))
	return resultContractVisibleControlPattern.MatchString(visibleText)
}

func outputValidationRequirement(raw any) *protocol.OutputValidationPlan {
	plan, err := protocol.DecodeOutputValidationPlan(raw)
	if err != nil || !plan.Required {
		return nil
	}
	return plan
}

func outputValidationTargetIssues(plan *protocol.OutputValidationPlan, content string) []string {
	if plan == nil || plan.Probe == nil {
		return nil
	}
	issues := make([]string, 0, 2)
	for _, target := range []string{plan.Probe.Action.Target, plan.Probe.Observe.Target} {
		if target == "" || outputValidationTargetPresent(content, target) {
			continue
		}
		issues = append(issues, fmt.Sprintf("entrypoint readback is missing approved validation target %s", target))
	}
	return uniqueResultContractStrings(issues)
}

func outputValidationTargetPresent(content, selector string) bool {
	match := outputValidationAttributeSelector.FindStringSubmatch(strings.TrimSpace(selector))
	if len(match) != 2 {
		return true // Runtime validation owns selectors that are not the standard marker form.
	}
	return strings.Contains(strings.ToLower(content), strings.ToLower(match[1]))
}

func outputValidationExecutionInstruction(plan *protocol.OutputValidationPlan) string {
	if plan == nil || plan.Probe == nil {
		return ""
	}
	action := fmt.Sprintf("action target %s", plan.Probe.Action.Target)
	if strings.TrimSpace(plan.Probe.Action.Key) != "" {
		action = fmt.Sprintf("%s action for key %s", plan.Probe.Action.Kind, plan.Probe.Action.Key)
	}
	return fmt.Sprintf(
		" The entrypoint must implement the approved %s and include observation target %s exactly; that action must visibly change the observed surface.",
		action,
		plan.Probe.Observe.Target,
	)
}

func outputValidationCorrectionInstruction(plan *protocol.OutputValidationPlan, issues []string) string {
	if plan == nil || plan.Probe == nil || !strings.Contains(strings.Join(issues, " "), "approved validation target") {
		return ""
	}
	return fmt.Sprintf(
		" Overwrite the entrypoint so it includes %s on the primary control and %s on the surface changed by that control, then read the entrypoint back.",
		plan.Probe.Action.Target,
		plan.Probe.Observe.Target,
	)
}
