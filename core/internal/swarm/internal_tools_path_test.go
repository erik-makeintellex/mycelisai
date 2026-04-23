package swarm

import (
	"path/filepath"
	"testing"
)

func TestNormalizeWorkspaceRelativePathStripsWorkspaceAlias(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"workspace/logs/test.py":   filepath.Join("logs", "test.py"),
		"./workspace/logs/test.py": filepath.Join("logs", "test.py"),
		"workspace":                ".",
		"logs/test.py":             filepath.Join("logs", "test.py"),
	}

	for input, want := range cases {
		if got := normalizeWorkspaceRelativePath(input); got != want {
			t.Fatalf("normalizeWorkspaceRelativePath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestValidateToolPathTreatsWorkspacePrefixAsAlias(t *testing.T) {
	tempDir := t.TempDir()
	workspaceRoot := filepath.Join(tempDir, "workspace-root")
	t.Setenv("MYCELIS_WORKSPACE", workspaceRoot)

	got, err := validateToolPath("workspace/logs/hello.py")
	if err != nil {
		t.Fatalf("validateToolPath returned error: %v", err)
	}

	want := filepath.Join(filepath.Clean(workspaceRoot), "logs", "hello.py")
	if got != want {
		t.Fatalf("validateToolPath returned %q, want %q", got, want)
	}
}
