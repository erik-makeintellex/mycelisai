package mcp

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// maxToolErrorBytes caps MCP error text before it reaches mission events,
// the Exchange and HTTP responses.
const maxToolErrorBytes = 2048

const redactedToolText = "[REDACTED]"

// secretKeyName matches a credential-bearing key: snake/kebab names ending in
// key/token/secret (X-Api-Key, OPENAI_API_KEY), camelCase ones (apiKey,
// accessToken, clientSecret) and the bare D7 names.
const secretKeyName = `(?:[A-Za-z0-9_-]*[_-](?i:key|token|secret)|[A-Za-z0-9]*[a-z0-9](?:Key|Token|Secret|KEY|TOKEN|SECRET)|(?i:api_?key|password|passwd|secret|token|credentials?))`

var (
	// Authorization headers keep their scheme: "Authorization: Basic [REDACTED]".
	authorizationHeaderPattern = regexp.MustCompile(`(?i)\b((?:proxy-)?authorization\\?"?\s*[=:]\s*\\?"?)(?:(bearer|basic|token|digest|negotiate)\s+)?[^\s,;"'\\]+`)
	// key=value, key: value and JSON "key": "value" (also inside escaped JSON).
	secretAssignmentPattern = regexp.MustCompile(`\b(` + secretKeyName + `)(\\?"?\s*[=:]\s*)("[^"]*"|'[^']*'|\\"[^"\\]*\\"|[^\s,;&]+)`)
	bearerPattern           = regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/=-]+`)
	basicCredentialPattern  = regexp.MustCompile(`(?i)\b(basic)\s+([A-Za-z0-9+/]{6,}={0,2})`)
	urlUserinfoPattern      = regexp.MustCompile(`([A-Za-z][A-Za-z0-9+.-]*://)[^/\s@]+@`)
	// Well-known token shapes, redacted wherever they appear.
	prefixedTokenPattern = regexp.MustCompile(`\b(github_pat_|gh[pousr]_|sk-(?:proj-|live-|test-)?|xox[abprs]-)([A-Za-z0-9_-]+)`)
	awsKeyIDPattern      = regexp.MustCompile(`\b(?:AKIA|ASIA)[A-Z0-9]{16}\b`)
	jwtPattern           = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}`)
)

// minPrefixedTokenBody is the shortest token body redacted after a known
// prefix; shorter tails ("sk-learn") are ordinary words.
var minPrefixedTokenBody = map[string]int{"sk-": 12, "sk-proj-": 8, "sk-live-": 8, "sk-test-": 8}

// RedactToolText removes credentials from text without capping it: bearer
// and basic credentials, Authorization and *-key headers, key=value and JSON
// "key": "value" secrets (snake, kebab and camelCase keys), URL userinfo, and
// well-known token shapes (ghp_/gho_/github_pat_, sk-, xox?-, AKIA, JWTs).
// It is best-effort and pattern-based, not a secret scanner: a credential in
// an unrecognized shape passes through. Core applies it to retained MCP
// arguments and results and to MCP error text.
func RedactToolText(text string) string {
	text = authorizationHeaderPattern.ReplaceAllStringFunc(text, func(m string) string {
		sub := authorizationHeaderPattern.FindStringSubmatch(m)
		if sub[2] != "" {
			return sub[1] + sub[2] + " " + redactedToolText
		}
		return sub[1] + redactedToolText
	})
	text = secretAssignmentPattern.ReplaceAllStringFunc(text, func(m string) string {
		sub := secretAssignmentPattern.FindStringSubmatch(m)
		return sub[1] + sub[2] + quoteLike(sub[3], redactedToolText)
	})
	text = bearerPattern.ReplaceAllString(text, "Bearer "+redactedToolText)
	text = basicCredentialPattern.ReplaceAllStringFunc(text, func(m string) string {
		sub := basicCredentialPattern.FindStringSubmatch(m)
		if !looksLikeBasicCredential(sub[2]) {
			return m
		}
		return sub[1] + " " + redactedToolText
	})
	text = urlUserinfoPattern.ReplaceAllString(text, "${1}"+redactedToolText+"@")
	text = prefixedTokenPattern.ReplaceAllStringFunc(text, func(m string) string {
		sub := prefixedTokenPattern.FindStringSubmatch(m)
		min, ok := minPrefixedTokenBody[sub[1]]
		if !ok {
			min = 8
		}
		if len(sub[2]) < min {
			return m
		}
		return sub[1] + redactedToolText
	})
	text = awsKeyIDPattern.ReplaceAllString(text, redactedToolText)
	return jwtPattern.ReplaceAllString(text, redactedToolText)
}

// quoteLike wraps replacement in the quoting style of original.
func quoteLike(original, replacement string) string {
	for _, q := range []string{`\"`, `"`, `'`} {
		if len(original) >= 2*len(q) && strings.HasPrefix(original, q) && strings.HasSuffix(original, q) {
			return q + replacement + q
		}
	}
	return replacement
}

// looksLikeBasicCredential separates base64 credentials from prose such as
// "Basic Authentication": it needs a digit, +, /, = or an inner capital.
func looksLikeBasicCredential(value string) bool {
	for i, r := range value {
		if unicode.IsDigit(r) || r == '+' || r == '/' || r == '=' || (i > 0 && unicode.IsUpper(r)) {
			return true
		}
	}
	return false
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
