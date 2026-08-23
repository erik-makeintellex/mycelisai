package server

import (
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

var (
	localHTMLAssetRefPattern      = regexp.MustCompile(`(?i)(?:src|href)\s*=\s*["']([^"']+)["']`)
	interactiveHandlerPattern     = regexp.MustCompile(`(?i)(?:addEventListener\s*\(\s*["'](?:click|pointerdown|touchstart|keydown|keyup)|on(?:click|pointerdown|touchstart|keydown)\s*=)`)
	interactiveEffectPattern      = regexp.MustCompile(`(?is)(?:\.(?:textContent|innerText|innerHTML|value|checked|disabled|hidden|className)\s*=|\.classList\.(?:add|remove|toggle|replace)\s*\(|\.style\.[A-Za-z][A-Za-z0-9-]*\s*=|\.dataset\.[A-Za-z_$][A-Za-z0-9_$]*\s*=|\.setAttribute\s*\(|\.(?:appendChild|append|prepend|remove|replaceChildren|insertAdjacentHTML)\s*\(|\b(?:requestAnimationFrame|setTimeout|setInterval)\s*\(|\b(?:fillRect|clearRect|strokeRect|drawImage|fillText|strokeText|putImageData|arc|lineTo|moveTo|bezierCurveTo|quadraticCurveTo)\s*\(|\.(?:play|pause)\s*\(|\b(?:localStorage|sessionStorage)\.setItem\s*\()`)
	namedFunctionPattern          = regexp.MustCompile(`(?m)\bfunction\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*\([^)]*\)\s*\{`)
	actionableButtonPattern       = regexp.MustCompile(`(?is)<button\b[^>]*>(.*?)</button>`)
	visibleControlLanguagePattern = regexp.MustCompile(`(?i)\b(?:click|tap|press|use|move|drag|select|arrow|space|wasd|control|start|restart|run|play|submit|save|reset|open|add|next)\b`)
	scriptOrStylePattern          = regexp.MustCompile(`(?is)<(?:script|style)\b[^>]*>.*?</(?:script|style)>`)
	htmlTagPattern                = regexp.MustCompile(`(?s)<[^>]+>`)
)

func deliverableResultMissingOutputs(item protocol.TeamWorkItem, payloadKind protocol.SignalPayloadKind, outputRefs []protocol.TeamOutputRef) bool {
	return deliverableResultOutputIssue(item, payloadKind, outputRefs) != ""
}

func deliverableResultOutputIssue(item protocol.TeamWorkItem, payloadKind protocol.SignalPayloadKind, outputRefs []protocol.TeamOutputRef) string {
	if payloadKind != protocol.PayloadKindResult ||
		(item.ExecutionShape != protocol.TeamExecutionShapeDeliverable && item.ExecutionShape != protocol.TeamExecutionShapeDelegatedWork) ||
		len(item.ExpectedOutputs) == 0 {
		return ""
	}
	if len(outputRefs) == 0 {
		return "missing_retained_output"
	}
	if !teamWorkExpectsProjectPackage(item) {
		return ""
	}
	teamRoot := "groups/" + strings.Trim(strings.TrimSpace(item.TeamID), "/") + "/"
	for _, ref := range outputRefs {
		storageRef := strings.Trim(strings.ReplaceAll(strings.TrimSpace(ref.StorageRef), "\\", "/"), "/")
		if strings.EqualFold(strings.TrimSpace(ref.Kind), "project_package") &&
			strings.TrimSpace(ref.Entrypoint) != "" &&
			strings.HasPrefix(storageRef+"/", teamRoot) {
			return projectPackageFileIssue(item, ref)
		}
	}
	return "invalid_deliverable_shape"
}

func projectPackageFileIssue(item protocol.TeamWorkItem, ref protocol.TeamOutputRef) string {
	storageRef := strings.Trim(strings.ReplaceAll(strings.TrimSpace(ref.StorageRef), "\\", "/"), "/")
	entrypoint := strings.Trim(strings.ReplaceAll(strings.TrimSpace(ref.Entrypoint), "\\", "/"), "/")
	entryPath := entrypoint
	if !strings.HasPrefix(entrypoint+"/", storageRef+"/") {
		entryPath = storageRef + "/" + entrypoint
	}
	target, _, err := resolveWorkspaceFilePath(entryPath)
	if err != nil {
		return "invalid_deliverable_shape"
	}
	info, err := os.Stat(target)
	if err != nil || info.IsDir() {
		return "incomplete_deliverable_files"
	}
	if !strings.EqualFold(filepath.Ext(target), ".html") {
		return ""
	}
	content, err := os.ReadFile(target)
	if err != nil || len(content) > maxWorkspaceViewBytes {
		return "incomplete_deliverable_files"
	}
	packageFolder, _, err := resolveWorkspacePath(storageRef, false)
	if err != nil {
		return "invalid_deliverable_shape"
	}
	combinedExecutableContent := string(content)
	for _, match := range localHTMLAssetRefPattern.FindAllStringSubmatch(string(content), -1) {
		assetRef := strings.TrimSpace(match[1])
		lower := strings.ToLower(assetRef)
		if assetRef == "" || strings.HasPrefix(assetRef, "#") || strings.HasPrefix(lower, "data:") ||
			strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "//") {
			continue
		}
		if cut := strings.IndexAny(assetRef, "?#"); cut >= 0 {
			assetRef = assetRef[:cut]
		}
		assetTarget := filepath.Clean(filepath.Join(filepath.Dir(target), filepath.FromSlash(assetRef)))
		rel, relErr := filepath.Rel(packageFolder, assetTarget)
		if relErr != nil || pathEscapesWorkspace(rel) {
			return "incomplete_deliverable_files"
		}
		assetInfo, statErr := os.Stat(assetTarget)
		if statErr != nil || assetInfo.IsDir() {
			return "incomplete_deliverable_files"
		}
		if strings.EqualFold(filepath.Ext(assetTarget), ".js") {
			assetContent, readErr := os.ReadFile(assetTarget)
			if readErr != nil || len(assetContent) > maxWorkspaceViewBytes {
				return "incomplete_deliverable_files"
			}
			combinedExecutableContent += "\n" + string(assetContent)
		}
	}
	if teamWorkRequiresPrimaryInteraction(item) &&
		(!htmlExposesInspectablePrimaryInteraction(string(content), combinedExecutableContent)) {
		return "unverified_primary_interaction"
	}
	return ""
}

func htmlExposesInspectablePrimaryInteraction(content, executableContent string) bool {
	return interactiveHandlerPattern.MatchString(executableContent) &&
		htmlExposesPrimaryControl(content) &&
		interactiveEffectPattern.MatchString(executableContent) &&
		!htmlHasDormantAnimationLoop(executableContent)
}

func htmlExposesPrimaryControl(content string) bool {
	for _, match := range actionableButtonPattern.FindAllStringSubmatch(content, -1) {
		label := strings.ToLower(strings.TrimSpace(html.UnescapeString(htmlTagPattern.ReplaceAllString(match[1], " "))))
		if label != "" && !strings.Contains(label, "restart") && !strings.Contains(label, "reset") {
			return true
		}
	}
	visibleText := scriptOrStylePattern.ReplaceAllString(content, " ")
	visibleText = html.UnescapeString(htmlTagPattern.ReplaceAllString(visibleText, " "))
	return visibleControlLanguagePattern.MatchString(visibleText)
}

func teamWorkRequiresPrimaryInteraction(item protocol.TeamWorkItem) bool {
	if item.WorkIntent != nil && item.WorkIntent.OutputContract != nil {
		plan := protocol.NormalizeOutputValidationPlan(item.WorkIntent.OutputContract.OutputValidation)
		if plan != nil && plan.Required && plan.Kind == protocol.OutputValidationInteractiveBrowser {
			return true
		}
	}
	values := append([]string{item.Objective}, item.ExpectedOutputs...)
	if item.WorkIntent != nil {
		values = append(values, item.WorkIntent.Objective)
		if item.WorkIntent.OutputContract != nil {
			values = append(values, item.WorkIntent.OutputContract.Validation...)
		}
	}
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

func htmlHasDormantAnimationLoop(content string) bool {
	scanContent := javascriptCodeOnly(content)
	for _, match := range namedFunctionPattern.FindAllStringSubmatchIndex(scanContent, -1) {
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
			return true
		}
	}
	return false
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

func teamWorkExpectsProjectPackage(item protocol.TeamWorkItem) bool {
	if item.WorkIntent != nil && item.WorkIntent.OutputContract != nil &&
		strings.EqualFold(strings.TrimSpace(item.WorkIntent.OutputContract.Shape), "app_package") {
		return true
	}
	for _, expected := range item.ExpectedOutputs {
		lower := strings.ToLower(expected)
		if strings.Contains(lower, "project package") || strings.Contains(lower, "application package") ||
			strings.Contains(lower, "browser game") || strings.Contains(lower, "playable app") {
			return true
		}
	}
	return false
}
