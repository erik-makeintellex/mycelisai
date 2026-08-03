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
	outputValidationNamedFunction           = regexp.MustCompile(`(?m)\bfunction\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*\([^)]*\)\s*\{`)
	outputValidationScriptContent           = regexp.MustCompile(`(?is)<script\b[^>]*>(.*?)</script>`)
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
	writeContent := ""
	readContent := ""
	for _, item := range evidence {
		if !evidenceContainsPath([]string{item.Path}, entrypoint) || strings.TrimSpace(item.Content) == "" {
			continue
		}
		switch item.ToolName {
		case "write_file":
			writeContent = item.Content
		case "read_file", "read_text_file":
			readContent = item.Content
		}
	}
	if writeContent != "" {
		return writeContent
	}
	return readContent
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

func outputValidationAnimationLoopIssues(plan *protocol.OutputValidationPlan, content string) []string {
	if plan == nil || !plan.Required || plan.Kind != protocol.OutputValidationInteractiveBrowser {
		return nil
	}
	scanContent := javascriptCodeOnly(outputValidationJavaScript(content))
	issues := make([]string, 0, 1)
	for _, match := range outputValidationNamedFunction.FindAllStringSubmatchIndex(scanContent, -1) {
		name := scanContent[match[2]:match[3]]
		bodyEnd, ok := javascriptBlockEnd(scanContent, match[1]-1)
		if !ok {
			continue
		}
		body := scanContent[match[1]:bodyEnd]
		selfSchedule := regexp.MustCompile(`\brequestAnimationFrame\s*\(\s*` + regexp.QuoteMeta(name) + `\s*\)`)
		if !selfSchedule.MatchString(body) {
			continue
		}
		outside := scanContent[:match[0]] + strings.Repeat(" ", bodyEnd-match[0]+1) + scanContent[bodyEnd+1:]
		bootstrap := regexp.MustCompile(
			`(?:\b` + regexp.QuoteMeta(name) + `\s*\(|\b(?:requestAnimationFrame|setTimeout|setInterval)\s*\(\s*` + regexp.QuoteMeta(name) + `\b|\baddEventListener\s*\([^;]*,\s*` + regexp.QuoteMeta(name) + `\s*\))`,
		)
		if !bootstrap.MatchString(outside) {
			issues = append(issues, fmt.Sprintf("animation loop %s is defined but never started", name))
		}
	}
	return uniqueResultContractStrings(issues)
}

func outputValidationJavaScript(content string) string {
	matches := outputValidationScriptContent.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return content
	}
	var source strings.Builder
	for _, match := range matches {
		source.WriteString(match[1])
		source.WriteByte('\n')
	}
	return source.String()
}

func javascriptBlockEnd(content string, open int) (int, bool) {
	depth, quote, escaped := 0, byte(0), false
	lineComment, blockComment := false, false
	for index := open; index < len(content); index++ {
		current := content[index]
		next := byte(0)
		if index+1 < len(content) {
			next = content[index+1]
		}
		if lineComment {
			if current == '\n' {
				lineComment = false
			}
			continue
		}
		if blockComment {
			if current == '*' && next == '/' {
				blockComment = false
				index++
			}
			continue
		}
		if quote != 0 {
			if escaped {
				escaped = false
			} else if current == '\\' {
				escaped = true
			} else if current == quote {
				quote = 0
			}
			continue
		}
		if current == '/' && next == '/' {
			lineComment = true
			index++
			continue
		}
		if current == '/' && next == '*' {
			blockComment = true
			index++
			continue
		}
		if current == '\'' || current == '"' || current == '`' {
			quote = current
			continue
		}
		switch current {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return index, true
			}
		}
	}
	return 0, false
}

func javascriptCodeOnly(content string) string {
	result := []byte(content)
	quote, escaped := byte(0), false
	lineComment, blockComment := false, false
	for index := 0; index < len(result); index++ {
		current := result[index]
		next := byte(0)
		if index+1 < len(result) {
			next = result[index+1]
		}
		if lineComment {
			result[index] = ' '
			if current == '\n' {
				lineComment = false
				result[index] = '\n'
			}
			continue
		}
		if blockComment {
			result[index] = ' '
			if current == '*' && next == '/' {
				blockComment = false
				result[index+1] = ' '
				index++
			}
			continue
		}
		if quote != 0 {
			result[index] = ' '
			if escaped {
				escaped = false
			} else if current == '\\' {
				escaped = true
			} else if current == quote {
				quote = 0
			}
			continue
		}
		if current == '/' && next == '/' {
			lineComment = true
			result[index], result[index+1] = ' ', ' '
			index++
			continue
		}
		if current == '/' && next == '*' {
			blockComment = true
			result[index], result[index+1] = ' ', ' '
			index++
			continue
		}
		if current == '\'' || current == '"' || current == '`' {
			quote = current
			result[index] = ' '
		}
	}
	return string(result)
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
		" The entrypoint must implement the approved %s and include observation target %s exactly; that action must visibly change the observed surface. The action must mutate state that the DOM or render loop actually consumes, and the rendered before/after state must differ; do not only assign an otherwise-unused intermediate value.",
		action,
		plan.Probe.Observe.Target,
	)
}

func outputValidationCorrectionInstruction(plan *protocol.OutputValidationPlan, issues []string) string {
	joined := strings.Join(issues, " ")
	if strings.Contains(joined, "defined but never started") {
		return " Start the retained animation or render loop explicitly after defining it (for example by invoking the loop once), then read the entrypoint back."
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
