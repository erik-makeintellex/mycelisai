package swarm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleWriteFileCreatesProjectPackageSupportFiles(t *testing.T) {
	workspaceRoot := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspaceRoot)

	registry := NewInternalToolRegistry(InternalToolDeps{})
	output, err := registry.handleWriteFile(context.Background(), map[string]any{
		"path":               "workspace/generated/game/index.html",
		"content":            "<!doctype html><title>Game</title>",
		"package_kind":       "project_package",
		"package_title":      "Playable Game",
		"package_folder":     "workspace/generated/game",
		"package_entrypoint": "workspace/generated/game/index.html",
		"package_files":      []any{"index.html", "README.md", "PROOF.md"},
		"validation":         "Browser opened, movement and restart verified.",
	})
	if err != nil {
		t.Fatalf("handleWriteFile returned error: %v", err)
	}
	if !strings.Contains(output, "Project package support files written: 2") {
		t.Fatalf("output = %q, want support file count", output)
	}

	for _, rel := range []string{
		"generated/game/index.html",
		"generated/game/README.md",
		"generated/game/PROOF.md",
	} {
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("expected package file %s: %v", rel, err)
		}
	}

	readme, err := os.ReadFile(filepath.Join(workspaceRoot, "generated", "game", "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if !strings.Contains(string(readme), "Browser opened") || !strings.Contains(string(readme), "workspace/generated/game/index.html") {
		t.Fatalf("README content = %q", readme)
	}
}
