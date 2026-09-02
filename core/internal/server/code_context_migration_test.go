package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodeContextMigrationDefinesNativeIndexTables(t *testing.T) {
	upPath := filepath.FromSlash("../../migrations/001_current_schema.sql")
	text, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read migration %s: %v", upPath, err)
	}
	for _, token := range []string{
		"CREATE TABLE IF NOT EXISTS code_context_sources",
		"CREATE TABLE IF NOT EXISTS code_context_snapshots",
		"CREATE TABLE IF NOT EXISTS code_context_files",
		"CREATE TABLE IF NOT EXISTS code_context_symbols",
		"CREATE TABLE IF NOT EXISTS code_context_edges",
		"code-context-fixture-v1",
		"CHECK (provenance IN ('extracted', 'inferred'))",
	} {
		if !strings.Contains(string(text), token) {
			t.Fatalf("migration missing %s", token)
		}
	}
}
