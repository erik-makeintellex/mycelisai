package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveWorkspaceRootUsesEnv(t *testing.T) {
	t.Setenv("MYCELIS_WORKSPACE", "/data/workspace")

	if got := resolveWorkspaceRoot(); got != "/data/workspace" {
		t.Fatalf("resolveWorkspaceRoot() = %q, want /data/workspace", got)
	}
}

func TestResolveWorkspaceRootDefault(t *testing.T) {
	_ = os.Unsetenv("MYCELIS_WORKSPACE")

	if got := resolveWorkspaceRoot(); got != "./workspace" {
		t.Fatalf("resolveWorkspaceRoot() = %q, want ./workspace", got)
	}
}

func TestEnsureStorageLayoutCreatesWorkspaceAndArtifacts(t *testing.T) {
	root := t.TempDir()
	workspaceRoot := filepath.Join(root, "workspace")
	dataDir := filepath.Join(root, "artifacts")

	if err := ensureStorageLayout(workspaceRoot, dataDir); err != nil {
		t.Fatalf("ensureStorageLayout: %v", err)
	}

	for _, path := range []string{
		workspaceRoot,
		filepath.Join(workspaceRoot, "saved-media"),
		dataDir,
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", path)
		}
	}
}
