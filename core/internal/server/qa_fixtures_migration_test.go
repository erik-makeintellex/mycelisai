package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQAFixtureConfigDocumentMigrationMatchesRuntimeKind(t *testing.T) {
	upPath := filepath.FromSlash("../../migrations/001_current_schema.sql")
	up, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read migration %s: %v", upPath, err)
	}
	text := string(up)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS qa_fixture_resources",
		"chk_qa_fixture_resource_kind",
		"'config_document'",
		"CREATE UNIQUE INDEX IF NOT EXISTS uq_qa_fixture_resource_claim",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("current schema missing QA fixture ownership fragment %q", required)
		}
	}
	if strings.Contains(text, "DROP TABLE IF EXISTS qa_fixture_resources") {
		t.Fatal("current schema must not include a destructive QA fixture downgrade")
	}
}
