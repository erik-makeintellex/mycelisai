package mcp

import (
	"os"
	"strings"
)

func ApplyRuntimeDefaults(cfg ServerConfig) (ServerConfig, error) {
	if cfg.Name != "filesystem" {
		return cfg, nil
	}

	workspaceRoot := ResolveFilesystemWorkspaceRoot()
	cfg.Args = append([]string(nil), cfg.Args...)
	if len(cfg.Args) >= 3 {
		cfg.Args[len(cfg.Args)-1] = workspaceRoot
	}

	if err := os.MkdirAll(workspaceRoot, 0o755); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func ResolveFilesystemWorkspaceRoot() string {
	workspace := strings.TrimSpace(os.Getenv("MYCELIS_WORKSPACE"))
	if workspace == "" {
		return "./workspace"
	}
	return workspace
}
