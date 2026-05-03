package mcp

import (
	"os"
	"testing"
)

func TestResolveFilesystemWorkspaceRoot(t *testing.T) {
	t.Setenv("MYCELIS_WORKSPACE", "/data/workspace")

	root := ResolveFilesystemWorkspaceRoot()

	if root != "/data/workspace" {
		t.Fatalf("root = %q, want /data/workspace", root)
	}
}

func TestResolveFilesystemWorkspaceRoot_Default(t *testing.T) {
	_ = os.Unsetenv("MYCELIS_WORKSPACE")

	root := ResolveFilesystemWorkspaceRoot()

	if root != "./workspace" {
		t.Fatalf("root = %q, want ./workspace", root)
	}
}

func TestApplyRuntimeDefaults_RewritesFilesystemWorkspaceArg(t *testing.T) {
	workspaceRoot := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspaceRoot)

	originalArgs := []string{"-y", "@modelcontextprotocol/server-filesystem", "./workspace"}
	cfg := ServerConfig{
		Name:      "filesystem",
		Transport: "stdio",
		Command:   "npx",
		Args:      originalArgs,
	}

	got, err := ApplyRuntimeDefaults(cfg)
	if err != nil {
		t.Fatalf("ApplyRuntimeDefaults returned error: %v", err)
	}
	if got.Args[2] != workspaceRoot {
		t.Fatalf("filesystem root arg = %q, want %q", got.Args[2], workspaceRoot)
	}
	if originalArgs[2] != "./workspace" {
		t.Fatalf("original args mutated to %q, want ./workspace", originalArgs[2])
	}
}

func TestApplyRuntimeDefaults_LeavesNonFilesystemConfigUnchanged(t *testing.T) {
	cfg := ServerConfig{Name: "fetch", Args: []string{"-y", "@modelcontextprotocol/server-fetch"}}

	got, err := ApplyRuntimeDefaults(cfg)
	if err != nil {
		t.Fatalf("ApplyRuntimeDefaults returned error: %v", err)
	}
	if got.Args[1] != "@modelcontextprotocol/server-fetch" {
		t.Fatalf("args = %v, want unchanged fetch args", got.Args)
	}
}
