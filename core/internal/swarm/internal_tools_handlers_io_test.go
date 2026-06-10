package swarm

import (
	"context"
	"encoding/json"
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
		"package_files":      []any{"index.html", "README.md", "PROOF.md", "project-package.json"},
		"package_usage":      "Use arrow keys or WASD to move. Press R to restart.",
		"validation":         "Browser opened, movement and restart verified.",
	})
	if err != nil {
		t.Fatalf("handleWriteFile returned error: %v", err)
	}
	if !strings.Contains(output, "Project package support files written: 3") {
		t.Fatalf("output = %q, want support file count", output)
	}

	for _, rel := range []string{
		"generated/game/index.html",
		"generated/game/README.md",
		"generated/game/PROOF.md",
		"generated/game/project-package.json",
	} {
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("expected package file %s: %v", rel, err)
		}
	}

	readme, err := os.ReadFile(filepath.Join(workspaceRoot, "generated", "game", "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	readmeText := string(readme)
	if !strings.Contains(readmeText, "Browser opened") ||
		!strings.Contains(readmeText, "workspace/generated/game/index.html") ||
		!strings.Contains(readmeText, "## Included files") ||
		!strings.Contains(readmeText, "project-package.json") ||
		!strings.Contains(readmeText, "## Recovery") {
		t.Fatalf("README content = %q", readme)
	}

	manifestBytes, err := os.ReadFile(filepath.Join(workspaceRoot, "generated", "game", "project-package.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("manifest is not JSON: %v", err)
	}
	if manifest["title"] != "Playable Game" || manifest["kind"] != "project_package" || manifest["entrypoint"] != "workspace/generated/game/index.html" {
		t.Fatalf("manifest = %#v, want title, kind, and entrypoint", manifest)
	}
	if open, ok := manifest["open"].(map[string]any); !ok || !strings.Contains(open["resources_url"].(string), "/resources?tab=workspace&path=workspace/generated/game") {
		t.Fatalf("manifest open hints = %#v", manifest["open"])
	}
	if recovery, ok := manifest["recovery"].(map[string]any); !ok || !strings.Contains(recovery["hint"].(string), "Resources") {
		t.Fatalf("manifest recovery hints = %#v", manifest["recovery"])
	}
}
