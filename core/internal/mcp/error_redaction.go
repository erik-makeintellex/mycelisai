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

// RedactToolText removes obvious credentials (bearer tokens, URL userinfo,
// key=value secrets) from text without capping it. It is a best-effort
// filter, not a secret scanner. Core applies it to retained MCP arguments.
func RedactToolText(text string) string {
	for _, rule := range toolErrorRedactions {
		text = rule.pattern.ReplaceAllString(text, rule.replacement)
	}
	return text
}

// redactToolErrorText is RedactToolText for MCP server error text, capped.
func redactToolErrorText(text string) string {
	text = RedactToolText(text)
	if len(text) <= maxToolErrorBytes {
		return text
	}
	cut := maxToolErrorBytes
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return strings.TrimRight(text[:cut], " \n") + " ... [truncated]"
}
