package mcp

import (
	"context"
	"os"
	"reflect"
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

func (s *Service) EnsureRuntimeDefaults(ctx context.Context, cfg ServerConfig) (ServerConfig, error) {
	runtimeCfg, err := ApplyRuntimeDefaults(cfg)
	if err != nil {
		return cfg, err
	}
	if reflect.DeepEqual(runtimeCfg.Args, cfg.Args) &&
		reflect.DeepEqual(runtimeCfg.Env, cfg.Env) &&
		runtimeCfg.Command == cfg.Command &&
		runtimeCfg.Transport == cfg.Transport &&
		runtimeCfg.URL == cfg.URL {
		return runtimeCfg, nil
	}
	installed, err := s.Install(ctx, runtimeCfg)
	if err != nil {
		return runtimeCfg, err
	}
	return *installed, nil
}
