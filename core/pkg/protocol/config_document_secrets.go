package protocol

import (
	"fmt"
	"sort"
	"strings"
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

func validateConfigDocumentSpecSecrets(value any, path string, issues *[]ConfigDocumentValidationIssue) {
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
			if configDocumentSecretRefField(key) {
				validateConfigDocumentSpecSecretRef(child, childPath, issues)
				continue
			}
			if configDocumentSensitiveField(key) {
				ref, ok := child.(string)
				if !ok || !isConfigDocumentSecretRef(ref) {
					*issues = append(*issues, ConfigDocumentValidationIssue{
						Code: "spec.raw_secret", Field: childPath,
						Message: "secret-bearing fields must contain a managed secret reference, not a raw credential",
					})
				}
				continue
			}
			if raw, ok := child.(string); ok && looksLikeRawConfigDocumentSecret(raw) {
				*issues = append(*issues, ConfigDocumentValidationIssue{
					Code: "spec.raw_secret", Field: childPath,
					Message: "raw secret-looking values are not allowed; use a managed secret reference",
				})
				continue
			}
			validateConfigDocumentSpecSecrets(child, childPath, issues)
		}
	case []any:
		for index, child := range typed {
			validateConfigDocumentSpecSecrets(child, fmt.Sprintf("%s[%d]", path, index), issues)
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

func configDocumentSecretRefField(key string) bool {
	normalized := normalizeConfigDocumentSecretKey(key)
	for _, suffix := range []string{"_refs", "_ref", "refs", "ref"} {
		if strings.HasSuffix(normalized, suffix) {
			base := strings.TrimSuffix(normalized, suffix)
			base = strings.TrimSuffix(base, "_")
			return configDocumentSensitiveField(base)
		}
	}
	return false
}

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

func normalizeConfigDocumentSecretKey(key string) string {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return strings.ReplaceAll(normalized, "-", "_")
}

func looksLikeRawConfigDocumentSecret(raw string) bool {
	value := strings.TrimSpace(raw)
	if isConfigDocumentSecretRef(value) {
		return false
	}
	lower := strings.ToLower(value)
	for _, prefix := range []string{"sk-", "rk_live_", "pk_live_", "ghp_", "github_pat_", "xoxb-", "xoxp-", "xoxa-", "xoxr-"} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return strings.HasPrefix(value, "AKIA") || strings.HasPrefix(value, "ASIA") ||
		strings.Contains(value, "-----BEGIN PRIVATE KEY-----")
}
