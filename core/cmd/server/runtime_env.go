package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/mycelis/core/internal/server"
)

type databaseRuntimeConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type localAuthRuntimeConfig struct {
	PrimaryAPIKey      string
	PrimaryUsername    string
	PrimaryUserID      string
	BreakGlassAPIKey   string
	BreakGlassUsername string
	BreakGlassUserID   string
	DeploymentContract server.DeploymentContract
}

func resolveDatabaseConfig() databaseRuntimeConfig {
	cfg := databaseRuntimeConfig{
		Host:     envOrDefault("DB_HOST", "mycelis-core-postgresql"),
		Port:     envOrDefault("DB_PORT", "5432"),
		User:     envOrDefault("DB_USER", "mycelis"),
		Password: envOrDefault("DB_PASSWORD", "password"),
		Name:     envOrDefault("DB_NAME", "cortex"),
	}
	if rawAddr := strings.TrimSpace(os.Getenv("MYCELIS_SERVICE_DB_ADDRESS")); rawAddr != "" {
		host, port := parseServiceHostPort(rawAddr)
		if host != "" {
			cfg.Host = host
		}
		if port != "" && strings.TrimSpace(os.Getenv("DB_PORT")) == "" {
			cfg.Port = port
		}
	}
	return cfg
}

func parseServiceHostPort(raw string) (string, string) {
	trimmed := strings.TrimSpace(raw)
	if strings.Contains(trimmed, "://") {
		if parsed, err := url.Parse(trimmed); err == nil && parsed.Host != "" {
			return parsed.Hostname(), parsed.Port()
		}
	}
	if host, port, err := net.SplitHostPort(trimmed); err == nil {
		return strings.TrimSpace(host), strings.TrimSpace(port)
	}
	return strings.Trim(trimmed, "[]"), ""
}

func resolveNATSURL(defaultURL string) string {
	if addr := strings.TrimSpace(os.Getenv("MYCELIS_SERVICE_NATS_ADDRESS")); addr != "" {
		return addr
	}
	if addr := strings.TrimSpace(os.Getenv("NATS_URL")); addr != "" {
		return addr
	}
	if strings.TrimSpace(defaultURL) != "" {
		return strings.TrimSpace(defaultURL)
	}
	return "nats://localhost:4222"
}

func (cfg databaseRuntimeConfig) connectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func defaultMCPBootstrapEnabled() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("MYCELIS_DISABLE_DEFAULT_MCP_BOOTSTRAP")))
	switch value {
	case "1", "true", "yes", "on":
		return false
	default:
		return true
	}
}

func resolveLocalAuthRuntimeConfig() localAuthRuntimeConfig {
	return localAuthRuntimeConfig{
		PrimaryAPIKey:      envOrDefault("MYCELIS_API_KEY", ""),
		PrimaryUsername:    envOrDefault("MYCELIS_LOCAL_ADMIN_USERNAME", "admin"),
		PrimaryUserID:      envOrDefault("MYCELIS_LOCAL_ADMIN_USER_ID", "00000000-0000-0000-0000-000000000000"),
		BreakGlassAPIKey:   envOrDefault("MYCELIS_BREAK_GLASS_API_KEY", ""),
		BreakGlassUsername: envOrDefault("MYCELIS_BREAK_GLASS_USERNAME", "recovery-admin"),
		BreakGlassUserID:   envOrDefault("MYCELIS_BREAK_GLASS_USER_ID", "00000000-0000-0000-0000-000000000001"),
		DeploymentContract: server.ResolveDeploymentContract(),
	}
}

func (cfg localAuthRuntimeConfig) breakGlassWarnings() []string {
	warnings := make([]string, 0, 3)
	if cfg.DeploymentContract.RequiresBreakGlassRecovery() && strings.TrimSpace(cfg.BreakGlassAPIKey) == "" {
		warnings = append(warnings, "deployment auth contract requires MYCELIS_BREAK_GLASS_API_KEY for enterprise-like recovery posture")
	}
	if strings.TrimSpace(cfg.BreakGlassAPIKey) != "" {
		keyConfigured := true
		usernameConfigured := strings.TrimSpace(os.Getenv("MYCELIS_BREAK_GLASS_USERNAME")) != ""
		userIDConfigured := strings.TrimSpace(os.Getenv("MYCELIS_BREAK_GLASS_USER_ID")) != ""
		if !(keyConfigured && usernameConfigured && userIDConfigured) {
			warnings = append(warnings, "partial break-glass config detected; set MYCELIS_BREAK_GLASS_API_KEY, MYCELIS_BREAK_GLASS_USERNAME, and MYCELIS_BREAK_GLASS_USER_ID together")
		}
		if cfg.BreakGlassAPIKey == cfg.PrimaryAPIKey && cfg.PrimaryAPIKey != "" {
			warnings = append(warnings, "break-glass API key matches MYCELIS_API_KEY; use a distinct recovery credential")
		}
	}
	if strings.TrimSpace(os.Getenv("MYCELIS_BREAK_GLASS_USERNAME")) != "" || strings.TrimSpace(os.Getenv("MYCELIS_BREAK_GLASS_USER_ID")) != "" {
		if strings.TrimSpace(cfg.BreakGlassAPIKey) == "" {
			warnings = append(warnings, "break-glass metadata is set without MYCELIS_BREAK_GLASS_API_KEY")
		}
	}
	if cfg.BreakGlassAPIKey != "" && cfg.BreakGlassUsername == cfg.PrimaryUsername && cfg.BreakGlassUserID == cfg.PrimaryUserID {
		warnings = append(warnings, "break-glass principal duplicates the primary local admin identity")
	}
	return warnings
}

func init() {
	for _, warning := range resolveLocalAuthRuntimeConfig().breakGlassWarnings() {
		log.Printf("WARN: auth posture: %s", warning)
	}
}
