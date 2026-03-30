package main

import (
	"fmt"
	"os"
	"strings"
)

type databaseRuntimeConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func resolveDatabaseConfig() databaseRuntimeConfig {
	return databaseRuntimeConfig{
		Host:     envOrDefault("DB_HOST", "mycelis-core-postgresql"),
		Port:     envOrDefault("DB_PORT", "5432"),
		User:     envOrDefault("DB_USER", "mycelis"),
		Password: envOrDefault("DB_PASSWORD", "password"),
		Name:     envOrDefault("DB_NAME", "cortex"),
	}
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
