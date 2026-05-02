package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveWorkspaceRoot() string {
	workspace := strings.TrimSpace(os.Getenv("MYCELIS_WORKSPACE"))
	if workspace == "" {
		return "./workspace"
	}
	return workspace
}

func ensureStorageLayout(workspaceRoot, dataDir string) error {
	dirs := []string{
		workspaceRoot,
		filepath.Join(workspaceRoot, "saved-media"),
		dataDir,
	}
	for _, dir := range dirs {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create storage path %s: %w", dir, err)
		}
	}
	return nil
}
