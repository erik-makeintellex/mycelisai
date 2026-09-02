package swarm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeTeamManifestMigrationProtectsExactState(t *testing.T) {
	upPath := filepath.FromSlash("../../migrations/001_current_schema.sql")
	up, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	text := string(up)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS runtime_team_manifests",
		"manifest_digest",
		"CHECK (jsonb_typeof(manifest) = 'object')",
		"CHECK (manifest->>'id' = team_id)",
		"PRIMARY KEY (tenant_id, team_id)",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("migration missing %q", required)
		}
	}

	if strings.Contains(text, "DROP TABLE IF EXISTS runtime_team_manifests") {
		t.Fatal("current schema must not include a destructive runtime-team downgrade")
	}
}
