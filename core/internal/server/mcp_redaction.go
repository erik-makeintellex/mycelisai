package server

import (
	"encoding/json"
	"strings"
	"unicode"

	"github.com/mycelis/core/internal/mcp"
)

const redactedMCPSecretValue = "[redacted]"

// redactedMCPArgumentValue replaces a sensitive retained MCP argument (MCPS D7).
const redactedMCPArgumentValue = "[REDACTED]"

var sensitiveMCPArgumentKeys = map[string]bool{
	"api_key": true, "apikey": true, "authorization": true, "credential": true, "credentials": true,
	"password": true, "passwd": true, "access_token": true, "refresh_token": true, "secret": true, "token": true,
}

// isSensitiveMCPArgumentKey matches the D7 key list in any key style: case,
// camelCase (apiKey, APIKey, clientSecret), kebab-case (x-api-key) and
// snake_case are read as snake_case, plus any key ending _key/_token/_secret.
func isSensitiveMCPArgumentKey(key string) bool {
	k := snakeCaseMCPKey(strings.TrimSpace(key))
	if sensitiveMCPArgumentKeys[k] {
		return true
	}
	return strings.HasSuffix(k, "_key") || strings.HasSuffix(k, "_token") || strings.HasSuffix(k, "_secret")
}

// snakeCaseMCPKey lower-cases key, maps "-" and " " to "_", and splits
// camelCase words: accessToken -> access_token, APIKey -> api_key.
func snakeCaseMCPKey(key string) string {
	runes := []rune(key)
	var b strings.Builder
	for i, r := range runes {
		switch {
		case r == '-' || r == ' ':
			b.WriteByte('_')
			continue
		case unicode.IsUpper(r) && i > 0:
			prev := runes[i-1]
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextLower) {
				b.WriteByte('_')
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// redactMCPToolArguments returns a deep copy of args that is safe to retain:
// sensitive keys become [REDACTED] and every string passes the credential
// patterns of mcp.RedactToolText. The tool itself still receives args.
func redactMCPToolArguments(args map[string]any) map[string]any {
	if args == nil {
		return nil
	}
	out := make(map[string]any, len(args))
	for key, value := range args {
		if isSensitiveMCPArgumentKey(key) {
			out[key] = redactedMCPArgumentValue
			continue
		}
		out[key] = redactMCPArgumentValue(value)
	}
	return out
}

// redactMCPToolResult returns the copy of an MCP tool result that the
// Exchange retains: the result as generic JSON with sensitive keys replaced
// and every string passed through mcp.RedactToolText. The caller's HTTP
// response still carries the raw result. A result that cannot be encoded is
// not retained verbatim: only a marker is kept.
func redactMCPToolResult(result any) any {
	raw, err := json.Marshal(result)
	if err != nil {
		return map[string]any{"redaction": "result could not be encoded for retention"}
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return map[string]any{"redaction": "result could not be encoded for retention"}
	}
	return redactMCPArgumentValue(generic)
}

func redactMCPArgumentValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return redactMCPToolArguments(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = redactMCPArgumentValue(item)
		}
		return out
	case string:
		return mcp.RedactToolText(typed)
	default:
		return value
	}
}

func redactMCPServerConfig(cfg mcp.ServerConfig) mcp.ServerConfig {
	cfg.Env = redactStringMap(cfg.Env)
	cfg.Headers = redactStringMap(cfg.Headers)
	return cfg
}

func redactStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return values
	}
	redacted := make(map[string]string, len(values))
	for key, value := range values {
		if value == "" {
			redacted[key] = ""
			continue
		}
		redacted[key] = redactedMCPSecretValue
	}
	return redacted
}
