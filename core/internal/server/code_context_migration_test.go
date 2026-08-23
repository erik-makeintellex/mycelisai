package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodeContextMigrationDefinesNativeIndexTables(t *testing.T) {
	upPath := filepath.FromSlash("../../migrations/064_code_context.up.sql")
	text, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read migration %s: %v", upPath, err)
	}
	for _, token := range []string{
		"code_context_sources",
		"code_context_snapshots",
		"code_context_files",
		"code_context_symbols",
		"code_context_edges",
		"code-context-fixture-v1",
	} {
		if !strings.Contains(string(text), token) {
			t.Fatalf("migration missing %s", token)
		}
	}
}
