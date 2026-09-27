package swarm

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func projectPackageWriteArgs(path, content string) map[string]any {
	return map[string]any{
		"path":               path,
		"content":            content,
		"package_kind":       "project_package",
		"package_title":      "Playable Game",
		"package_folder":     "workspace/generated/game",
		"package_entrypoint": "workspace/generated/game/index.html",
		"package_files":      []any{"index.html", "README.md", "PROOF.md", "project-package.json"},
		"package_usage":      "Use arrow keys or WASD to move. Press R to restart.",
		"validation":         "Browser opened, movement and restart verified.",
	}
}

// write_file writes only the manifest on its own. README.md and PROOF.md are
// never synthesized, and the model's validation claim is not recorded as proof.
func TestHandleWriteFileWritesOnlyManifestForProjectPackage(t *testing.T) {
	workspaceRoot := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspaceRoot)

	registry := NewInternalToolRegistry(InternalToolDeps{})
	output, err := registry.handleWriteFile(context.Background(), projectPackageWriteArgs("workspace/generated/game/index.html", "<!doctype html><html><body>Game</body></html>"))
	if err != nil {
		t.Fatalf("handleWriteFile returned error: %v", err)
	}
	message, artifacts, ok := extractToolOutputArtifacts(output)
	if !ok || !strings.Contains(message, "Package manifest written") {
		t.Fatalf("structured write output = %q", output)
	}
	if len(artifacts) != 1 || artifacts[0].Type != "project_package" || artifacts[0].Entrypoint != "workspace/generated/game/index.html" {
		t.Fatalf("write artifacts = %#v", artifacts)
	}
	if got := strings.Join(artifacts[0].Files, ","); got != "index.html,project-package.json" {
		t.Fatalf("artifact files = %s, want only the files this call wrote", got)
	}
	if artifacts[0].Validation != "" {
		t.Fatalf("artifact validation = %q, want no unverified claim", artifacts[0].Validation)
	}
	for _, rel := range []string{"README.md", "PROOF.md", "validation-notes.md"} {
		if _, err := os.Stat(filepath.Join(workspaceRoot, "generated", "game", rel)); !os.IsNotExist(err) {
			t.Fatalf("%s was synthesized (stat err %v)", rel, err)
		}
	}
	entrypointBytes, err := os.ReadFile(filepath.Join(workspaceRoot, "generated", "game", "index.html"))
	if err != nil || !strings.Contains(string(entrypointBytes), "<title>Playable Game</title>") {
		t.Fatalf("entrypoint = %q, %v; want normalized title", entrypointBytes, err)
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
	if _, claimed := manifest["validation"]; claimed {
		t.Fatalf("manifest carries a validation claim: %#v", manifest)
	}
	if open, ok := manifest["open"].(map[string]any); !ok || !strings.Contains(open["resources_url"].(string), "/resources?tab=workspace&path=workspace/generated/game") {
		t.Fatalf("manifest open hints = %#v", manifest["open"])
	}
}

func TestHandleWriteFileKeepsModelAuthoredSupportFiles(t *testing.T) {
	workspaceRoot := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspaceRoot)
	registry := NewInternalToolRegistry(InternalToolDeps{})

	if _, err := registry.handleWriteFile(context.Background(), projectPackageWriteArgs("workspace/generated/game/README.md", "# Model README")); err != nil {
		t.Fatalf("README write: %v", err)
	}
	modelManifest := `{"title":"Model manifest"}`
	if _, err := registry.handleWriteFile(context.Background(), projectPackageWriteArgs("workspace/generated/game/project-package.json", modelManifest)); err != nil {
		t.Fatalf("manifest write: %v", err)
	}
	if _, err := registry.handleWriteFile(context.Background(), projectPackageWriteArgs("workspace/generated/game/index.html", "<!doctype html><title>x</title>")); err != nil {
		t.Fatalf("entrypoint write: %v", err)
	}
	readme, _ := os.ReadFile(filepath.Join(workspaceRoot, "generated", "game", "README.md"))
	if string(readme) != "# Model README" {
		t.Fatalf("README = %q, want the model's bytes", readme)
	}
	manifest, _ := os.ReadFile(filepath.Join(workspaceRoot, "generated", "game", "project-package.json"))
	if string(manifest) != modelManifest {
		t.Fatalf("manifest = %q, want the model-authored manifest kept", manifest)
	}
}
