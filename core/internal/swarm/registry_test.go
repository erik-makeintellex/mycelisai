package swarm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_LoadManifests(t *testing.T) {
	// 1. Setup Temp Dir
	tmpDir, err := os.MkdirTemp("", "swarm_teams")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 2. Write Dummy Manifest
	yamlContent := `
id: test-team
name: Test Team
type: action
members:
  - id: agent-a
inputs:
  - swarm.global.input.test
deliveries:
  - swarm.test.out
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write yaml: %v", err)
	}

	// 3. Load
	reg := NewRegistry(tmpDir)
	manifests, err := reg.LoadManifests()
	if err != nil {
		t.Fatalf("LoadManifests() failed: %v", err)
	}

	// 4. Verify
	if len(manifests) != 1 {
		t.Errorf("Expected 1 manifest, got %d", len(manifests))
	}
	m := manifests[0]
	if m.ID != "test-team" {
		t.Errorf("Expected ID test-team, got %s", m.ID)
	}
	if len(m.Members) != 1 || m.Members[0].ID != "agent-a" {
		t.Errorf("Member parsing failed: %v", m.Members)
	}
}
