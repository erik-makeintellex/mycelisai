package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type ActionCLIConfig struct {
	APIBaseURL     string            `yaml:"api_base_url"`
	APIKey         string            `yaml:"api_key"`
	TimeoutSeconds int               `yaml:"timeout_seconds"`
	Headers        map[string]string `yaml:"headers"`
}

func defaultActionCLIConfig() ActionCLIConfig {
	return ActionCLIConfig{
		APIBaseURL:     "http://localhost:8081",
		TimeoutSeconds: 20,
		Headers:        map[string]string{},
	}
}

func discoverActionConfigPaths() []string {
	return discoverActionConfigPathsWith(os.Getenv, os.Getwd, mustUserHomeDir())
}

func discoverActionConfigPathsWith(getenv func(string) string, getwd func() (string, error), userHome string) []string {
	paths := []string{
		"/etc/mycelis/config.yaml",
		"/etc/mycelis/config.yml",
	}

	if cwd, err := getwd(); err == nil && cwd != "" {
		paths = append(paths,
			filepath.Join(cwd, "mycelis.yaml"),
			filepath.Join(cwd, "mycelis.yml"),
			filepath.Join(cwd, "config", "mycelis.yaml"),
			filepath.Join(cwd, "config", "mycelis.yml"),
		)
	}

	if xdgHome := getenv("XDG_CONFIG_HOME"); xdgHome != "" {
		paths = append(paths,
			filepath.Join(xdgHome, "mycelis", "config.yaml"),
			filepath.Join(xdgHome, "mycelis", "config.yml"),
		)
	}

	if userHome != "" {
		paths = append(paths,
			filepath.Join(userHome, ".config", "mycelis", "config.yaml"),
			filepath.Join(userHome, ".config", "mycelis", "config.yml"),
			filepath.Join(userHome, ".mycelis", "config.yaml"),
			filepath.Join(userHome, ".mycelis", "config.yml"),
		)
	}

	if override := getenv("MYCELIS_CONFIG"); override != "" {
		paths = append(paths, override)
	}

	return paths
}

func mustUserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func mergeActionCLIConfig(dst *ActionCLIConfig, src ActionCLIConfig) {
	if src.APIBaseURL != "" {
		dst.APIBaseURL = src.APIBaseURL
	}
	if src.APIKey != "" {
		dst.APIKey = src.APIKey
	}
	if src.TimeoutSeconds > 0 {
		dst.TimeoutSeconds = src.TimeoutSeconds
	}
	if len(src.Headers) > 0 {
		if dst.Headers == nil {
			dst.Headers = map[string]string{}
		}
		for k, v := range src.Headers {
			dst.Headers[k] = v
		}
	}
}

func loadActionCLIConfigFromPaths(paths []string, base ActionCLIConfig) (ActionCLIConfig, []string, error) {
	cfg := base
	loaded := []string{}
	for _, p := range paths {
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var fileCfg ActionCLIConfig
		if err := yaml.Unmarshal(data, &fileCfg); err != nil {
			return cfg, loaded, fmt.Errorf("parse action config %s: %w", p, err)
		}
		mergeActionCLIConfig(&cfg, fileCfg)
		loaded = append(loaded, p)
	}
	return cfg, loaded, nil
}

func loadActionCLIConfig() (ActionCLIConfig, []string, error) {
	return loadActionCLIConfigFromPaths(discoverActionConfigPaths(), defaultActionCLIConfig())
}

func loadActionRuntimeConfig() (ActionCLIConfig, error) {
	cfg, _, err := loadActionCLIConfig()
	if err != nil {
		return cfg, err
	}

	if envBase := os.Getenv("MYCELIS_API_URL"); envBase != "" {
		cfg.APIBaseURL = envBase
	}
	if envKey := os.Getenv("MYCELIS_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}
	if envTimeout := os.Getenv("MYCELIS_API_TIMEOUT_SEC"); envTimeout != "" {
		if n, parseErr := strconv.Atoi(envTimeout); parseErr == nil && n > 0 {
			cfg.TimeoutSeconds = n
		}
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 20
	}
	return cfg, nil
}
