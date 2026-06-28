package swarm

import (
	"context"
	"strings"
	"testing"
)

func TestInternalDocsToolsAreRegisteredAndCitable(t *testing.T) {
	registry := NewInternalToolRegistry(InternalToolDeps{})

	for _, name := range []string{"list_docs", "read_doc", "search_docs"} {
		if !registry.Has(name) {
			t.Fatalf("expected internal tool %s to be registered", name)
		}
	}

	listed, err := registry.Get("list_docs").Handler(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("list_docs: %v", err)
	}
	if !strings.Contains(listed, "soma-chat") || !strings.Contains(listed, "mycelis-canonical-prd") {
		t.Fatalf("expected curated docs in list_docs output: %s", listed)
	}

	read, err := registry.Get("read_doc").Handler(context.Background(), map[string]any{"slug": "memory-guide"})
	if err != nil {
		t.Fatalf("read_doc: %v", err)
	}
	if !strings.Contains(read, "docs/user/memory.md") || !strings.Contains(read, "Memory") {
		t.Fatalf("expected citable memory doc content: %s", read)
	}

	search, err := registry.Get("search_docs").Handler(context.Background(), map[string]any{"query": "deployment context", "limit": float64(2)})
	if err != nil {
		t.Fatalf("search_docs: %v", err)
	}
	if !strings.Contains(search, `"results"`) || !strings.Contains(search, `"path"`) {
		t.Fatalf("expected citable docs search results: %s", search)
	}
}

func TestInternalDocsToolsValidateRequiredInputs(t *testing.T) {
	registry := NewInternalToolRegistry(InternalToolDeps{})

	if _, err := registry.Get("read_doc").Handler(context.Background(), map[string]any{}); err == nil {
		t.Fatalf("expected read_doc to require slug")
	}
	if _, err := registry.Get("search_docs").Handler(context.Background(), map[string]any{}); err == nil {
		t.Fatalf("expected search_docs to require query")
	}
}
