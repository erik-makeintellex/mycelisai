package swarm

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	outputValidationNamedFunction = regexp.MustCompile(`(?m)\bfunction\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*\([^)]*\)\s*\{`)
	outputValidationScriptContent = regexp.MustCompile(`(?is)<script\b[^>]*>(.*?)</script>`)
)

func resultContractDormantAnimationLoopIssues(content string) []string {
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
