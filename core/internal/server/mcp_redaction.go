package server

import (
	"strings"

	"github.com/mycelis/core/internal/mcp"
)

const redactedMCPSecretValue = "[redacted]"

// redactedMCPArgumentValue replaces a sensitive retained MCP argument (MCPS D7).
const redactedMCPArgumentValue = "[REDACTED]"

var sensitiveMCPArgumentKeys = map[string]bool{
	"api_key": true, "authorization": true, "credential": true, "password": true,
	"access_token": true, "refresh_token": true, "secret": true, "token": true,
}

// isSensitiveMCPArgumentKey matches the D7 key list case-insensitively, with
// "-" read as "_" (x-api-key), plus any key ending _key, _token or _secret.
func isSensitiveMCPArgumentKey(key string) bool {
	k := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(key)), "-", "_")
	if sensitiveMCPArgumentKeys[k] {
		return true
	}
	return strings.HasSuffix(k, "_key") || strings.HasSuffix(k, "_token") || strings.HasSuffix(k, "_secret")
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
