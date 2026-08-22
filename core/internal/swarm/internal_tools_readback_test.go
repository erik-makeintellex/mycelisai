package swarm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFileKeepsModelOutputBoundedButReturnsCompleteRuntimeEvidence(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	content := strings.Repeat("entrypoint-content-", 2500)
	path := filepath.Join(workspace, "groups", "team", "generated", "app", "index.html")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	registry := &InternalToolRegistry{}

	bounded, err := registry.handleReadFile(context.Background(), map[string]any{"path": path})
	if err != nil {
		t.Fatal(err)
	}
	if len(bounded) >= len(content) || !strings.Contains(bounded, "truncated at 32KB") {
		t.Fatalf("ordinary model read was not bounded: %d bytes", len(bounded))
	}

	runtimeContext := WithToolInvocationContext(context.Background(), ToolInvocationContext{RuntimeOwned: true})
	complete, err := registry.handleReadFile(runtimeContext, map[string]any{"path": path})
	if err != nil {
		t.Fatal(err)
	}
	if complete != content {
		t.Fatalf("runtime proof read returned %d bytes, want %d", len(complete), len(content))
	}
}
