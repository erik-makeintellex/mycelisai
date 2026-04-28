package swarm

import (
	"path/filepath"
	"testing"
)

func TestNormalizeWorkspaceRelativePathStripsWorkspaceAlias(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"workspace/logs/test.py":   filepath.Join("logs", "test.py"),
		"/workspace/logs/test.py":  filepath.Join("logs", "test.py"),
		"./workspace/logs/test.py": filepath.Join("logs", "test.py"),
		"workspace":                ".",
		"/workspace":               ".",
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

	for _, input := range []string{"workspace/logs/hello.py", "/workspace/logs/hello.py"} {
		got, err := validateToolPath(input)
		if err != nil {
			t.Fatalf("validateToolPath(%q) returned error: %v", input, err)
		}

		want := filepath.Join(filepath.Clean(workspaceRoot), "logs", "hello.py")
		if got != want {
			t.Fatalf("validateToolPath(%q) returned %q, want %q", input, got, want)
		}
	}
}
