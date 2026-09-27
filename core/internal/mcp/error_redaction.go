package mcp

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// maxToolErrorBytes caps MCP error text before it reaches mission events,
// the Exchange and HTTP responses.
const maxToolErrorBytes = 2048

var toolErrorRedactions = []struct {
	pattern     *regexp.Regexp
	replacement string
}{
	{regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/=-]+`), "Bearer [REDACTED]"},
	{regexp.MustCompile(`([A-Za-z][A-Za-z0-9+.-]*://)[^/\s@]+@`), "${1}[REDACTED]@"},
	{regexp.MustCompile(`(?i)\b([A-Za-z0-9_]*(?:_key|_token|_secret)|api_?key|password|passwd|secret|token)(\s*[=:]\s*)("[^"]*"|'[^']*'|[^\s,;&]+)`), "${1}${2}[REDACTED]"},
}

// redactToolErrorText removes obvious credentials from MCP server error text
// and caps its length. It is a best-effort filter, not a secret scanner.
func redactToolErrorText(text string) string {
	for _, rule := range toolErrorRedactions {
		text = rule.pattern.ReplaceAllString(text, rule.replacement)
	}
	if len(text) <= maxToolErrorBytes {
		return text
	}
	cut := maxToolErrorBytes
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return strings.TrimRight(text[:cut], " \n") + " ... [truncated]"
}
