package main

import (
	"fmt"
	"log"
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
	SSLMode  string
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
	return databaseRuntimeConfig{
		Host:     envOrDefault("DB_HOST", "mycelis-core-postgresql"),
		Port:     envOrDefault("DB_PORT", "5432"),
		User:     envOrDefault("DB_USER", "mycelis"),
		Password: envOrDefault("DB_PASSWORD", "password"),
		Name:     envOrDefault("DB_NAME", "cortex"),
		SSLMode:  envOrDefault("DB_SSLMODE", "disable"),
	}
}

func (cfg databaseRuntimeConfig) connectionString() string {
	postgresURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   "/" + strings.TrimLeft(cfg.Name, "/"),
	}
	query := postgresURL.Query()
	if strings.TrimSpace(cfg.SSLMode) != "" {
		query.Set("sslmode", strings.TrimSpace(cfg.SSLMode))
	}
	postgresURL.RawQuery = query.Encode()
	return postgresURL.String()
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
