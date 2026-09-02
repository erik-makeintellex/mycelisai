package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddress string
	DatabaseURL   string
	CoreToken     string
	MaxRuns       int
	LeaseDuration time.Duration
}

func FromEnv() (Config, error) {
	config := Config{
		ListenAddress: envOr("FRAMEWORK_RUNS_LISTEN_ADDRESS", "127.0.0.1:8091"),
		DatabaseURL:   strings.TrimSpace(os.Getenv("FRAMEWORK_RUNS_DATABASE_URL")),
		CoreToken:     os.Getenv("FRAMEWORK_RUNS_CORE_TOKEN"),
		MaxRuns:       10_000,
		LeaseDuration: 30 * time.Second,
	}
	if config.DatabaseURL == "" {
		return Config{}, errors.New("FRAMEWORK_RUNS_DATABASE_URL is required")
	}
	if config.CoreToken != strings.TrimSpace(config.CoreToken) || len(config.CoreToken) < 32 {
		return Config{}, errors.New("FRAMEWORK_RUNS_CORE_TOKEN must be canonical and at least 32 bytes")
	}
	if raw := strings.TrimSpace(os.Getenv("FRAMEWORK_RUNS_MAX_RUNS")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return Config{}, errors.New("FRAMEWORK_RUNS_MAX_RUNS must be positive")
		}
		config.MaxRuns = value
	}
	if raw := strings.TrimSpace(os.Getenv("FRAMEWORK_RUNS_LEASE_SECONDS")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 3600 {
			return Config{}, errors.New("FRAMEWORK_RUNS_LEASE_SECONDS must be from 1 through 3600")
		}
		config.LeaseDuration = time.Duration(value) * time.Second
	}
	return config, nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
