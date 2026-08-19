package swarm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeTeamManifestMigrationProtectsExactState(t *testing.T) {
	upPath := filepath.FromSlash("../../migrations/063_runtime_team_manifests.up.sql")
	up, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	text := string(up)
	for _, required := range []string{
		"runtime_team_manifests", "manifest_digest", "manifest->>'id' = team_id", "PRIMARY KEY (tenant_id, team_id)",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("migration missing %q", required)
		}
	}

	down, err := os.ReadFile(filepath.FromSlash("../../migrations/063_runtime_team_manifests.down.sql"))
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	if !strings.Contains(string(down), "RAISE EXCEPTION") {
		t.Fatal("down migration must refuse to discard persisted team manifests")
	}
}
