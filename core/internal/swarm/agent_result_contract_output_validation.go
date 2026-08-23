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
	outputValidationDrawingEffectPattern    = regexp.MustCompile(`(?is)\b(?:fillRect|clearRect|strokeRect|drawImage|fillText|strokeText|putImageData|arc|lineTo|moveTo|bezierCurveTo|quadraticCurveTo)\s*\(`)
	resultContractInteractiveHandlerPattern = regexp.MustCompile(`(?i)(?:addEventListener\s*\(\s*["'](?:click|pointerdown|touchstart|keydown|keyup)|on(?:click|pointerdown|touchstart|keydown)\s*=)`)
	resultContractInteractiveEffectPattern  = regexp.MustCompile(`(?is)(?:\.(?:textContent|innerText|innerHTML|value|checked|disabled|hidden|className)\s*=|\.classList\.(?:add|remove|toggle|replace)\s*\(|\.style\.[A-Za-z][A-Za-z0-9-]*\s*=|\.dataset\.[A-Za-z_$][A-Za-z0-9_$]*\s*=|\.setAttribute\s*\(|\.(?:appendChild|append|prepend|remove|replaceChildren|insertAdjacentHTML)\s*\(|\b(?:requestAnimationFrame|setTimeout|setInterval)\s*\(|\b(?:fillRect|clearRect|strokeRect|drawImage|fillText|strokeText|putImageData|arc|lineTo|moveTo|bezierCurveTo|quadraticCurveTo)\s*\(|\.(?:play|pause)\s*\(|\b(?:localStorage|sessionStorage)\.setItem\s*\()`)
	resultContractVisibleControlPattern     = regexp.MustCompile(`(?i)\b(?:click|tap|press|use|move|drag|select|arrow|space|wasd|control|start|restart|run|play|submit|save|reset|open|add|next)\b`)
	resultContractScriptOrStylePattern      = regexp.MustCompile(`(?is)<(?:script|style)\b[^>]*>.*?</(?:script|style)>`)
	resultContractHTMLTagPattern            = regexp.MustCompile(`(?s)<[^>]+>`)
	textChangeMutationProperty              = `(?:textContent|innerText|innerHTML|value)`
)

func resultContractRequiresPrimaryInteraction(requirement *teamResultRequirement) bool {
	if requirement == nil {
		return false
	}
	if requirement.OutputValidation != nil &&
		requirement.OutputValidation.Required &&
		requirement.OutputValidation.Kind == protocol.OutputValidationInteractiveBrowser {
		return true
	}
	values := append(append([]string{}, requirement.ExpectedOutputs...), requirement.AcceptanceCriteria...)
	for _, value := range values {
		lower := strings.ToLower(value)
		if strings.Contains(lower, "playable") || strings.Contains(lower, "browser game") ||
			strings.Contains(lower, "interactive browser") || strings.Contains(lower, "interactive app") ||
			strings.Contains(lower, "web app") || strings.Contains(lower, "web application") ||
			strings.Contains(lower, "controls respond") || strings.Contains(lower, "primary user workflow") ||
			strings.Contains(lower, "primary control") || strings.Contains(lower, "primary interaction") {
			return true
		}
	}
	return false
}

func latestEntrypointEvidenceContent(evidence []successfulToolEvidence, entrypoint string) string {
	writeContent := ""
	currentReadContent := ""
	for _, item := range evidence {
		if !evidenceContainsPath([]string{item.Path}, entrypoint) || strings.TrimSpace(item.Content) == "" {
			continue
		}
		switch item.ToolName {
		case "write_file":
			writeContent = item.Content
			currentReadContent = ""
		case "read_file", "read_text_file":
			if item.Content == writeContent {
				currentReadContent = item.Content
			} else {
				currentReadContent = ""
			}
		}
	}
	if currentReadContent != "" {
		return currentReadContent
	}
	return writeContent
}

func resultContractExposesPrimaryControl(content string) bool {
	visibleText := resultContractScriptOrStylePattern.ReplaceAllString(content, " ")
	visibleText = html.UnescapeString(resultContractHTMLTagPattern.ReplaceAllString(visibleText, " "))
	return resultContractVisibleControlPattern.MatchString(visibleText)
}

func resultContractExposesInspectablePrimaryInteraction(content string) bool {
	return resultContractInteractiveHandlerPattern.MatchString(content) &&
		resultContractExposesPrimaryControl(content) &&
		resultContractInteractiveEffectPattern.MatchString(content) &&
		len(resultContractDormantAnimationLoopIssues(content)) == 0
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

func outputValidationAnimationLoopIssues(plan *protocol.OutputValidationPlan, content string) []string {
	if plan == nil || !plan.Required || plan.Kind != protocol.OutputValidationInteractiveBrowser {
		return nil
	}
	return resultContractDormantAnimationLoopIssues(content)
}

func outputValidationTextChangeIssues(plan *protocol.OutputValidationPlan, content string) []string {
	if plan == nil || plan.Probe == nil || plan.Probe.Observe.Kind != protocol.OutputValidationObserveTextChange {
		return nil
	}
	target := strings.TrimSpace(plan.Probe.Observe.Target)
	if target == "" || outputValidationTextChangeTargetMutated(content, target) {
		return nil
	}
	return []string{"entrypoint readback does not mutate approved text-change observation target " + target}
}

func outputValidationVisualChangeIssues(plan *protocol.OutputValidationPlan, content string) []string {
	if plan == nil || plan.Probe == nil || plan.Probe.Observe.Kind != protocol.OutputValidationObserveVisualChange {
		return nil
	}
	target := strings.TrimSpace(plan.Probe.Observe.Target)
	if target == "" || outputValidationVisualChangeTargetMutated(content, target) {
		return nil
	}
	return []string{"entrypoint readback does not mutate approved visual-change observation target " + target}
}

func outputValidationTextChangeTargetMutated(content, target string) bool {
	script := outputValidationJavaScript(content)
	targets := []string{target}
	targets = append(targets, outputValidationTargetElementIDs(content, target)...)
	for _, candidate := range targets {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		quoted := regexp.QuoteMeta(candidate)
		if regexp.MustCompile(`(?is)(?:querySelector|getElementById)\(\s*["']` + quoted + `["']\s*\)\s*\.\s*` + textChangeMutationProperty + `\s*=`).MatchString(script) {
			return true
		}
		if strings.HasPrefix(candidate, "#") {
			id := strings.TrimPrefix(candidate, "#")
			if regexp.MustCompile(`(?is)(?:getElementById)\(\s*["']` + regexp.QuoteMeta(id) + `["']\s*\)\s*\.\s*` + textChangeMutationProperty + `\s*=`).MatchString(script) {
				return true
			}
		}
		variablePattern := regexp.MustCompile(`(?is)(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*(?:document\.)?(?:querySelector|getElementById)\(\s*["']` + quoted + `["']\s*\)`)
		for _, match := range variablePattern.FindAllStringSubmatch(script, -1) {
			if len(match) < 2 {
				continue
			}
			if regexp.MustCompile(`(?m)\b` + regexp.QuoteMeta(match[1]) + `\s*\.\s*` + textChangeMutationProperty + `\s*=`).MatchString(script) {
				return true
			}
		}
	}
	return false
}

func outputValidationVisualChangeTargetMutated(content, target string) bool {
	if outputValidationCanvasTargetRendered(content, target) {
		return true
	}
	script := outputValidationJavaScript(content)
	targets := []string{target}
	targets = append(targets, outputValidationTargetElementIDs(content, target)...)
	for _, candidate := range targets {
		if outputValidationTargetHasDirectMutation(script, candidate, `(?:textContent|innerText|innerHTML|className|value|hidden|disabled)`) ||
			outputValidationTargetHasMethodMutation(script, candidate) {
			return true
		}
	}
	return false
}

func outputValidationTargetElementIDs(content, target string) []string {
	match := outputValidationAttributeSelector.FindStringSubmatch(strings.TrimSpace(target))
	if len(match) != 2 {
		return nil
	}
	attribute := regexp.QuoteMeta(match[1])
	targetTagPattern := regexp.MustCompile(`(?is)<[^>]*\s` + attribute + `(?:\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+))?[^>]*>`)
	idPattern := regexp.MustCompile(`(?is)\sid\s*=\s*["']([^"']+)["']`)
	ids := []string{}
	for _, tag := range targetTagPattern.FindAllString(content, -1) {
		match := idPattern.FindStringSubmatch(tag)
		if len(match) == 2 && strings.TrimSpace(match[1]) != "" {
			ids = append(ids, "#"+strings.TrimSpace(match[1]))
		}
	}
	return uniqueResultContractStrings(ids)
}

func outputValidationTargetPresent(content, selector string) bool {
	match := outputValidationAttributeSelector.FindStringSubmatch(strings.TrimSpace(selector))
	if len(match) != 2 {
		return true // Runtime validation owns selectors that are not the standard marker form.
	}
	attribute := regexp.QuoteMeta(match[1])
	pattern := regexp.MustCompile(`(?i)<[^>]*\s` + attribute + `(?:\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+))?(?:\s|/?>)`)
	return pattern.MatchString(content)
}

func outputValidationExecutionInstruction(plan *protocol.OutputValidationPlan) string {
	if plan == nil || plan.Probe == nil {
		return ""
	}
	action := fmt.Sprintf("action target %s", plan.Probe.Action.Target)
	if strings.TrimSpace(plan.Probe.Action.Key) != "" {
		action = fmt.Sprintf("%s action for key %s", plan.Probe.Action.Kind, plan.Probe.Action.Key)
	}
	instruction := fmt.Sprintf(
		" The entrypoint must implement the approved %s and include observation target %s exactly; that action must visibly change the observed surface. Bind handlers through the approved marker or a stable element ID, never a positional selector such as nth-child. Every selector used before addEventListener must resolve to an element in the same entrypoint. Give the approved primary control one unambiguous state-changing effect; broader or secondary handlers must not also match that control or immediately undo its effect. The action must mutate state that the DOM or render loop actually consumes, and the rendered before/after state must differ; do not only assign an otherwise-unused intermediate value.",
		action,
		plan.Probe.Observe.Target,
	)
	if plan.Probe.Observe.Kind == protocol.OutputValidationObserveTextChange {
		instruction += " For text_change validation, the approved action must update the observed element's textContent, innerText, innerHTML, or value to different user-visible text; do not rely only on canvas pixels, CSS class changes, console output, or hidden state."
	}
	return instruction
}

func outputValidationCorrectionInstruction(plan *protocol.OutputValidationPlan, issues []string) string {
	joined := strings.Join(issues, " ")
	if strings.Contains(joined, "defined but never started") {
		return " Start the retained animation or render loop explicitly after defining it (for example by invoking the loop once), then read the entrypoint back."
	}
	if strings.Contains(joined, "visual-change observation target") {
		return " Overwrite the entrypoint so the approved visual observation marker is on the surface that actually changes, such as the canvas, or mutate that marked surface visibly during the approved action; then read the entrypoint back."
	}
	if strings.Contains(joined, "text-change observation target") {
		return " Overwrite the entrypoint so the primary action handler updates the approved observation surface's textContent, innerText, innerHTML, or value, then read the entrypoint back."
	}
	if plan == nil || plan.Probe == nil || !strings.Contains(joined, "approved validation target") {
		return ""
	}
	return fmt.Sprintf(
		" Overwrite the entrypoint so it includes %s on the primary control and %s on the surface changed by that control, then read the entrypoint back.",
		plan.Probe.Action.Target,
		plan.Probe.Observe.Target,
	)
}
