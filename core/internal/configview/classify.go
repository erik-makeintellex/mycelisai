package configview

import (
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// keySegments splits a key into lowercase segments: camelCase, snake_case,
// kebab-case, dotted, and spaced keys all normalize the same way.
func keySegments(key string) []string {
	var b strings.Builder
	runes := []rune(key)
	for i, r := range runes {
		if unicode.IsUpper(r) && i > 0 && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1])) {
			b.WriteRune('_')
		}
		switch r {
		case '-', '.', ' ', '/':
			b.WriteRune('_')
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	parts := strings.Split(b.String(), "_")
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func lastSegment(key string) string {
	segs := keySegments(key)
	if len(segs) == 0 {
		return ""
	}
	return segs[len(segs)-1]
}

// IsReferenceKey reports keys such as api_key_env or secret_ref that name
// where a secret lives. A bare `env` or `ref` key is a container, not a
// reference, and is walked normally.
func IsReferenceKey(key string) bool {
	segs := keySegments(key)
	if len(segs) < 2 {
		return false
	}
	if _, ok := referenceSegments[segs[len(segs)-1]]; ok {
		return true
	}
	// *_env_var
	return len(segs) >= 3 && segs[len(segs)-2] == "env" && segs[len(segs)-1] == "var"
}

// IsSensitiveKey reports whether a key holds secret material. Matching is
// case-insensitive and separator-insensitive, at any depth:
//   - any segment in password, passwd, secret(s), credential(s),
//     authorization, bearer, apikey, privatekey
//   - a last segment of `key` or `token` (api_key, OPENAI_API_KEY,
//     GITHUB_TOKEN); `tokens` (max_tokens) and token_budget_* stay visible
//   - a compacted key ending in apikey, privatekey, clientsecret, password,
//     passwd, secret, credential(s), authorization, bearer, or token
func IsSensitiveKey(key string) bool {
	if IsReferenceKey(key) {
		return false
	}
	segs := keySegments(key)
	if len(segs) == 0 {
		return false
	}
	for _, s := range segs {
		if _, ok := sensitiveSegments[s]; ok {
			return true
		}
	}
	last := segs[len(segs)-1]
	if last == "key" || last == "token" {
		return true
	}
	compact := strings.Join(segs, "")
	for _, suffix := range sensitiveSuffixes {
		if strings.HasSuffix(compact, suffix) {
			return true
		}
	}
	return false
}

// secretPrefixes are well-known credential formats. A field must also be at
// least minPrefixedSecretLen long so short words never match.
var secretPrefixes = []string{
	"sk-", "sk_live_", "sk_test_", "rk_live_", "rk_test_", "ghp_", "gho_", "ghu_",
	"ghs_", "ghr_", "github_pat_", "glpat-", "xoxb-", "xoxp-", "xoxa-", "xoxs-",
	"xapp-", "hf_", "AKIA", "ASIA", "AIza", "ya29.", "npm_", "pypi-",
}

const minPrefixedSecretLen = 16

var jwtPattern = regexp.MustCompile(`^eyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}$`)

// assignmentPattern matches env-style NAME=value and --flag=value fields.
var assignmentPattern = regexp.MustCompile(`^(-{0,2}[A-Za-z_][A-Za-z0-9_.-]*)=(.+)$`)

// LooksSecret reports whether a whole value has a known credential shape.
func LooksSecret(s string) bool {
	s = strings.TrimSpace(s)
	if isBearerValue(s) || strings.Contains(s, "-----BEGIN") {
		return true
	}
	for _, f := range strings.Fields(s) {
		if _, ok := scrubField(f); ok {
			return true
		}
	}
	return false
}

// scrubField returns a replacement for one whitespace-free field when it is
// a credential, an env-style assignment of a sensitive name, or a URL with
// an embedded password.
func scrubField(f string) (string, bool) {
	trimmed := strings.Trim(f, `"'`+"`,;()[]{}")
	if len(trimmed) >= minPrefixedSecretLen {
		for _, p := range secretPrefixes {
			if strings.HasPrefix(trimmed, p) {
				return strings.Replace(f, trimmed, Marker, 1), true
			}
		}
		if jwtPattern.MatchString(trimmed) {
			return strings.Replace(f, trimmed, Marker, 1), true
		}
	}
	if m := assignmentPattern.FindStringSubmatch(trimmed); m != nil {
		name := strings.TrimLeft(m[1], "-")
		if IsSensitiveKey(name) {
			return strings.Replace(f, trimmed, m[1]+"="+Marker, 1), true
		}
	}
	if strings.Contains(trimmed, "://") {
		return scrubURL(f, trimmed)
	}
	return "", false
}

// scrubURL hides a userinfo password and sensitive query parameters
// (?token=, ?api_key=) in a URL field.
func scrubURL(f, trimmed string) (string, bool) {
	u, err := url.Parse(trimmed)
	if err != nil {
		return "", false
	}
	out, changed := f, false
	if u.User != nil {
		if pw, set := u.User.Password(); set && pw != "" {
			out = strings.Replace(out, ":"+pw+"@", ":"+Marker+"@", 1)
			changed = true
		}
	}
	for _, pair := range strings.Split(u.RawQuery, "&") {
		name, value, found := strings.Cut(pair, "=")
		if found && value != "" && IsSensitiveKey(name) {
			out = strings.Replace(out, pair, name+"="+Marker, 1)
			changed = true
		}
	}
	return out, changed
}

var headerTokenPattern = regexp.MustCompile(`^[A-Za-z0-9._~+/=-]{8,}$`)

// isBearerValue matches a whole Authorization header value: "Bearer <token>"
// or "Basic <credentials>".
func isBearerValue(s string) bool {
	fields := strings.Fields(s)
	if len(fields) != 2 {
		return false
	}
	switch strings.ToLower(fields[0]) {
	case "bearer", "basic":
		return looksLikeHeaderToken(fields[1])
	}
	return false
}

// looksLikeHeaderToken requires token characters plus a digit or base64
// punctuation (or a long run) so words such as "instructions" never match.
func looksLikeHeaderToken(s string) bool {
	s = strings.Trim(s, `"',;`)
	if !headerTokenPattern.MatchString(s) {
		return false
	}
	return len(s) >= 20 || strings.ContainsAny(s, "0123456789+/=")
}

// isSensitiveFlag matches a bare CLI flag such as --api-key or -password
// whose value follows as the next list element.
func isSensitiveFlag(s string) bool {
	if !strings.HasPrefix(s, "-") || strings.Contains(s, "=") {
		return false
	}
	return IsSensitiveKey(strings.TrimLeft(s, "-"))
}
