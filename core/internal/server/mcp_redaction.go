package server

import "github.com/mycelis/core/internal/mcp"

const redactedMCPSecretValue = "[redacted]"

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
