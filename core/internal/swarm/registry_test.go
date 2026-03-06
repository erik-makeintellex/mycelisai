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

func TestRegistry_LoadStandingPrimeTeamManifests(t *testing.T) {
	reg := NewRegistry(filepath.Join("..", "..", "config", "teams"))

	manifests, err := reg.LoadManifests()
	if err != nil {
		t.Fatalf("LoadManifests() failed: %v", err)
	}

	expected := map[string]struct {
		input      string
		delivery   string
		memberID   string
		memberRole string
	}{
		"prime-architect": {
			input:      "swarm.team.prime-architect.internal.command",
			delivery:   "swarm.team.prime-architect.signal.status",
			memberID:   "prime-architect-agent",
			memberRole: "architect",
		},
		"prime-development": {
			input:      "swarm.team.prime-development.internal.command",
			delivery:   "swarm.team.prime-development.signal.status",
			memberID:   "prime-development-agent",
			memberRole: "coder",
		},
		"agui-design-architect": {
			input:      "swarm.team.agui-design-architect.internal.command",
			delivery:   "swarm.team.agui-design-architect.signal.status",
			memberID:   "agui-design-architect-agent",
			memberRole: "design_architect",
		},
	}

	found := map[string]*TeamManifest{}
	for _, manifest := range manifests {
		if _, ok := expected[manifest.ID]; ok {
			found[manifest.ID] = manifest
		}
	}

	if len(found) != len(expected) {
		t.Fatalf("expected %d standing prime manifests, found %d", len(expected), len(found))
	}

	for id, want := range expected {
		manifest := found[id]
		if manifest == nil {
			t.Fatalf("missing manifest %s", id)
		}
		if len(manifest.Members) != 1 {
			t.Fatalf("manifest %s expected 1 member, got %d", id, len(manifest.Members))
		}
		if manifest.Members[0].ID != want.memberID {
			t.Fatalf("manifest %s expected member ID %s, got %s", id, want.memberID, manifest.Members[0].ID)
		}
		if manifest.Members[0].Role != want.memberRole {
			t.Fatalf("manifest %s expected member role %s, got %s", id, want.memberRole, manifest.Members[0].Role)
		}
		if len(manifest.Inputs) != 1 || manifest.Inputs[0] != want.input {
			t.Fatalf("manifest %s expected input %s, got %v", id, want.input, manifest.Inputs)
		}
		if len(manifest.Deliveries) != 1 || manifest.Deliveries[0] != want.delivery {
			t.Fatalf("manifest %s expected delivery %s, got %v", id, want.delivery, manifest.Deliveries)
		}
	}
}
