package swarm

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mycelis/core/internal/codecontext"
)

func TestInternalToolRegistryCodeContextRegistered(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	for _, name := range []string{"code_context.query", "code_context.impact", "code_context.explain"} {
		if !r.Has(name) {
			t.Fatalf("expected %s internal tool to be registered", name)
		}
	}
}

func TestInternalToolRegistryCodeContextQueryCallsService(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "thing.go"), []byte("package main\n\nfunc Thing() {}\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	r := NewInternalToolRegistry(InternalToolDeps{
		CodeContext: codecontext.NewService(codecontext.Config{SourceRoots: []string{root}}),
	})
	out, err := r.handleCodeContextQuery(context.Background(), map[string]any{"query": "Thing"})
	if err != nil {
		t.Fatalf("handleCodeContextQuery: %v", err)
	}
	var resp codecontext.Response
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Status != "ok" || resp.Count == 0 {
		t.Fatalf("resp = %+v", resp)
	}
}
