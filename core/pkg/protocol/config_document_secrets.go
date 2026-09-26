package protocol

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// configDocumentSecretMode selects one of two rule sets inside the single
// ConfigDocument secret classifier.
//
//   - configDocumentSecretsStored is the admission rule set every existing
//     revision was stored under. Read, compile, activation, and rollback
//     revalidate with it, so tightening detection never strands a stored or
//     active revision.
//   - configDocumentSecretsStrict applies to new writes (store, dry-run) and to
//     export redaction. It is always a superset of the stored rules, so any
//     document admitted today still revalidates when it is read.
type configDocumentSecretMode uint8

const (
	configDocumentSecretsStored configDocumentSecretMode = iota
	configDocumentSecretsStrict
)

var (
	configDocumentStoredSecretPrefixes = []string{"sk-", "rk_live_", "pk_live_", "ghp_", "github_pat_", "xoxb-", "xoxp-", "xoxa-", "xoxr-"}

	// Known credential prefixes need at least 16 token characters after the
	// prefix, so prose such as "sk-learn" or "xoxb-docs" is not a credential.
	configDocumentPrefixedTokenPattern = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9_-])(?:sk-|sk_live_|sk_test_|rk_live_|rk_test_|pk_live_|ghp_|gho_|ghu_|ghs_|ghr_|github_pat_|xox[bpar]-|glpat-)[A-Za-z0-9_-]{16,}`)
	configDocumentHFTokenPattern       = regexp.MustCompile(`(?:^|[^A-Za-z0-9_-])hf_[A-Za-z0-9]{30,}`)
	configDocumentGoogleKeyPattern     = regexp.MustCompile(`(?:^|[^A-Za-z0-9_-])AIza[0-9A-Za-z_-]{30,}`)
	configDocumentAWSKeyPattern        = regexp.MustCompile(`(?:^|[^A-Za-z0-9])(?:AKIA|ASIA)[A-Z0-9]{16}(?:$|[^A-Za-z0-9])`)
	configDocumentJWTPattern           = regexp.MustCompile(`eyJ[A-Za-z0-9_-]{8,}\.eyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}`)
	configDocumentPrivateKeyPattern    = regexp.MustCompile(`-----BEGIN [A-Z0-9 ]*PRIVATE KEY[A-Z ]*-----`)
	configDocumentBearerPattern        = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9])bearer\s+[A-Za-z0-9._~+/=-]{16,}`)
	configDocumentURLPasswordPattern   = regexp.MustCompile(`[A-Za-z][A-Za-z0-9+.-]*://[^\s/@:]*:([^\s/@]+)@`)
	configDocumentAssignmentPattern    = regexp.MustCompile(`(?:^|[^A-Za-z0-9_.])([A-Za-z_][A-Za-z0-9_.-]*)\s*=\s*("[^"]*"|'[^']*'|[^\s;,&"']+)`)
	configDocumentPlainScalarPattern   = regexp.MustCompile(`^(?i:true|false|yes|no|on|off|null|none|[0-9]+(?:\.[0-9]+)?)$`)
)

func isConfigDocumentSecretRef(raw string) bool {
	if raw == "" || raw != strings.TrimSpace(raw) {
		return false
	}
	if configDocumentEnvRefPattern.MatchString(raw) {
		return true
	}
	if strings.HasPrefix(raw, "env:") {
		return configDocumentEnvRefPattern.MatchString(strings.TrimPrefix(raw, "env:"))
	}
	for _, prefix := range []string{"secret:", "vault:", "sm://"} {
		if strings.HasPrefix(raw, prefix) {
			path := strings.TrimPrefix(raw, prefix)
			if prefix != "sm://" {
				path = strings.TrimPrefix(path, "//")
			}
			return configDocumentPathRefPattern.MatchString(path)
		}
	}
	return false
}

func validateConfigDocumentSpecSecrets(value any, path string, issues *[]ConfigDocumentValidationIssue, mode configDocumentSecretMode) {
	rawSecret := func(field, message string) {
		*issues = append(*issues, ConfigDocumentValidationIssue{Code: "spec.raw_secret", Field: field, Message: message})
	}
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child := typed[key]
			childPath := path + "." + key
			if mode == configDocumentSecretsStrict && looksLikeRawConfigDocumentSecret(key, mode) {
				// Never echo a secret-looking key back through the issue field.
				rawSecret(path+".[redacted-key]", "object keys must not contain raw secret-looking values")
				continue
			}
			if configDocumentSecretRefField(key, mode) {
				validateConfigDocumentSpecSecretRef(child, childPath, issues)
				continue
			}
			if configDocumentSensitiveField(key) {
				ref, ok := child.(string)
				if !ok || !isConfigDocumentSecretRef(ref) {
					rawSecret(childPath, "secret-bearing fields must contain a managed secret reference, not a raw credential")
				}
				continue
			}
			if text, ok := child.(string); ok && mode == configDocumentSecretsStrict && configDocumentStrictSensitiveField(key) {
				if text != "" && !isConfigDocumentSecretRef(text) {
					rawSecret(childPath, "secret-bearing fields must contain a managed secret reference, not a raw credential")
				}
				continue
			}
			if raw, ok := child.(string); ok && looksLikeRawConfigDocumentSecret(raw, mode) {
				rawSecret(childPath, "raw secret-looking values are not allowed; use a managed secret reference")
				continue
			}
			validateConfigDocumentSpecSecrets(child, childPath, issues, mode)
		}
	case []any:
		for index, child := range typed {
			validateConfigDocumentSpecSecrets(child, fmt.Sprintf("%s[%d]", path, index), issues, mode)
		}
	case string:
		// Map values are checked by the parent; this catches array elements.
		if mode == configDocumentSecretsStrict && looksLikeRawConfigDocumentSecret(typed, mode) {
			rawSecret(path, "raw secret-looking values are not allowed; use a managed secret reference")
		}
	}
}

func validateConfigDocumentSpecSecretRef(value any, path string, issues *[]ConfigDocumentValidationIssue) {
	valid := false
	switch typed := value.(type) {
	case string:
		valid = isConfigDocumentSecretRef(typed)
	case []any:
		valid = len(typed) > 0
		for _, item := range typed {
			ref, ok := item.(string)
			if !ok || !isConfigDocumentSecretRef(ref) {
				valid = false
				break
			}
		}
	}
	if !valid {
		*issues = append(*issues, ConfigDocumentValidationIssue{
			Code: "spec.invalid_secret_ref", Field: path,
			Message: "secret reference fields must contain managed secret references",
		})
	}
}

func configDocumentSecretRefField(key string, mode configDocumentSecretMode) bool {
	normalized := normalizeConfigDocumentSecretKey(key)
	for _, suffix := range []string{"_refs", "_ref", "refs", "ref"} {
		if strings.HasSuffix(normalized, suffix) {
			base := strings.TrimSuffix(normalized, suffix)
			base = strings.TrimSuffix(base, "_")
			return configDocumentSensitiveField(base) ||
				(mode == configDocumentSecretsStrict && configDocumentStrictSensitiveField(base))
		}
	}
	return false
}

// configDocumentSensitiveField is the stored-rule key list: its value must be a
// managed reference in every mode.
func configDocumentSensitiveField(key string) bool {
	normalized := normalizeConfigDocumentSecretKey(key)
	switch normalized {
	case "secret", "secrets", "password", "passwd", "token", "auth_token", "access_token",
		"api_key", "apikey", "client_secret", "clientsecret", "private_key", "privatekey",
		"access_key", "accesskey", "credential", "credentials":
		return true
	}
	for _, suffix := range []string{
		"_secret", "secret", "_secrets", "secrets", "_password", "password", "_passwd", "passwd",
		"_token", "token", "_api_key", "apikey", "_private_key", "privatekey",
		"_access_key", "accesskey", "_credential", "credential", "_credentials", "credentials",
	} {
		if strings.HasSuffix(normalized, suffix) {
			return true
		}
	}
	return false
}

// configDocumentStrictSensitiveField adds keys whose string values must be
// managed references under strict rules. Object or array values under these
// keys (for example an `auth` block holding `token_ref`) are walked instead.
func configDocumentStrictSensitiveField(key string) bool {
	normalized := normalizeConfigDocumentSecretKey(key)
	switch normalized {
	case "authorization", "proxy_authorization", "auth", "passphrase", "signing_key", "signingkey":
		return true
	}
	return strings.HasSuffix(normalized, "_passphrase") || strings.HasSuffix(normalized, "_signing_key")
}

func normalizeConfigDocumentSecretKey(key string) string {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return strings.ReplaceAll(normalized, "-", "_")
}

// looksLikeRawConfigDocumentSecret is the single raw-credential detector. The
// stored rules are unchanged from original admission; strict rules add
// embedded-token, URL-password, and sensitive NAME=value detection on top.
func looksLikeRawConfigDocumentSecret(raw string, mode configDocumentSecretMode) bool {
	value := strings.TrimSpace(raw)
	if isConfigDocumentSecretRef(value) {
		return false
	}
	lower := strings.ToLower(value)
	for _, prefix := range configDocumentStoredSecretPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	if strings.HasPrefix(value, "AKIA") || strings.HasPrefix(value, "ASIA") ||
		strings.Contains(value, "-----BEGIN PRIVATE KEY-----") {
		return true
	}
	if mode != configDocumentSecretsStrict {
		return false
	}
	for _, pattern := range []*regexp.Regexp{
		configDocumentPrefixedTokenPattern, configDocumentHFTokenPattern, configDocumentGoogleKeyPattern,
		configDocumentAWSKeyPattern, configDocumentJWTPattern, configDocumentPrivateKeyPattern, configDocumentBearerPattern,
	} {
		if pattern.MatchString(value) {
			return true
		}
	}
	for _, match := range configDocumentURLPasswordPattern.FindAllStringSubmatch(value, -1) {
		if !isConfigDocumentSecretPlaceholder(match[1]) {
			return true
		}
	}
	for _, match := range configDocumentAssignmentPattern.FindAllStringSubmatch(value, -1) {
		name, assigned := match[1], strings.Trim(match[2], `"'`)
		if !configDocumentSensitiveField(name) && !configDocumentStrictSensitiveField(name) {
			continue
		}
		if assigned == "" || isConfigDocumentSecretPlaceholder(assigned) || configDocumentPlainScalarPattern.MatchString(assigned) {
			continue
		}
		return true
	}
	return false
}

// isConfigDocumentSecretPlaceholder accepts prefixed refs and interpolation
// markers that stand in for a value without carrying one. A bare NAME is not
// a placeholder here: in "PASSWORD=HUNTER2" the right side is the value.
func isConfigDocumentSecretPlaceholder(value string) bool {
	if value == ConfigDocumentRedactedValue ||
		(isConfigDocumentSecretRef(value) && !configDocumentEnvRefPattern.MatchString(value)) {
		return true
	}
	for _, prefix := range []string{"$", "{{", "<", "%"} {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return strings.Trim(value, "*x.") == ""
}
