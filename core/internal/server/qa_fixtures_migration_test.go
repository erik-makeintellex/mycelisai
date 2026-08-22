package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQAFixtureConfigDocumentMigrationMatchesRuntimeKind(t *testing.T) {
	upPath := filepath.FromSlash("../../migrations/062_qa_fixture_config_documents.up.sql")
	up, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read migration %s: %v", upPath, err)
	}
	text := string(up)
	if !strings.Contains(text, "chk_qa_fixture_resource_kind") ||
		!strings.Contains(text, "'config_document'") {
		t.Fatalf("migration does not permit the runtime config_document fixture kind: %s", text)
	}

	downPath := filepath.FromSlash("../../migrations/062_qa_fixture_config_documents.down.sql")
	down, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("read migration %s: %v", downPath, err)
	}
	if !strings.Contains(string(down), "RAISE EXCEPTION") ||
		strings.Contains(string(down), "DELETE FROM qa_fixture_resources") {
		t.Fatal("down migration must refuse to discard live config_document fixture claims")
	}
}
