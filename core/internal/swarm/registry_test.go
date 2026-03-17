package swarm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
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
		input       string
		deliveries  []string
		memberIDs   []string
		memberRoles []string
	}{
		"prime-architect": {
			input:       "swarm.team.prime-architect.internal.command",
			deliveries:  []string{"swarm.team.prime-architect.signal.status"},
			memberIDs:   []string{"prime-architect-agent"},
			memberRoles: []string{"architect"},
		},
		"prime-development": {
			input:       "swarm.team.prime-development.internal.command",
			deliveries:  []string{"swarm.team.prime-development.signal.status"},
			memberIDs:   []string{"prime-development-agent"},
			memberRoles: []string{"coder"},
		},
		"agui-design-architect": {
			input:       "swarm.team.agui-design-architect.internal.command",
			deliveries:  []string{"swarm.team.agui-design-architect.signal.status"},
			memberIDs:   []string{"agui-design-architect-agent"},
			memberRoles: []string{"design_architect"},
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
		if len(manifest.Members) != len(want.memberIDs) {
			t.Fatalf("manifest %s expected %d members, got %d", id, len(want.memberIDs), len(manifest.Members))
		}
		for idx, memberID := range want.memberIDs {
			if manifest.Members[idx].ID != memberID {
				t.Fatalf("manifest %s expected member ID %s at index %d, got %s", id, memberID, idx, manifest.Members[idx].ID)
			}
		}
		for idx, memberRole := range want.memberRoles {
			if manifest.Members[idx].Role != memberRole {
				t.Fatalf("manifest %s expected member role %s at index %d, got %s", id, memberRole, idx, manifest.Members[idx].Role)
			}
		}
		if len(manifest.Inputs) != 1 || manifest.Inputs[0] != want.input {
			t.Fatalf("manifest %s expected input %s, got %v", id, want.input, manifest.Inputs)
		}
		if len(manifest.Deliveries) != len(want.deliveries) {
			t.Fatalf("manifest %s expected %d deliveries, got %d", id, len(want.deliveries), len(manifest.Deliveries))
		}
		for idx, delivery := range want.deliveries {
			if manifest.Deliveries[idx] != delivery {
				t.Fatalf("manifest %s expected delivery %s at index %d, got %s", id, delivery, idx, manifest.Deliveries[idx])
			}
		}
	}
}

func TestRegistry_RuntimeOrganizationPrimaryPath(t *testing.T) {
	reg := NewRegistryFromRuntimeOrganization(&RuntimeOrganization{
		ID:             "bundle-org",
		Name:           "Bundle Org",
		SourceKind:     "template_bundle",
		KernelMode:     "bundle-native",
		CouncilMode:    "bundle-native",
		ProviderPolicy: ProviderPolicy{Metadata: map[string]string{"posture": "bundle-native"}},
		Teams: []*TeamManifest{{
			ID:          "bundle-team",
			Name:        "Bundle Team",
			Type:        TeamTypeAction,
			Description: "Team loaded from embedded bundle content.",
			Members:     []protocol.AgentManifest{{ID: "bundle-agent", Role: "coder"}},
			Inputs:      []string{"swarm.team.bundle-team.internal.command"},
			Deliveries:  []string{"swarm.team.bundle-team.signal.status"},
		}},
		MigrationFallback: false,
	})

	org := reg.RuntimeOrganization()
	if org == nil {
		t.Fatal("expected runtime organization")
	}
	if org.ID != "bundle-org" {
		t.Fatalf("expected bundle-org, got %s", org.ID)
	}
	if org.MigrationFallback {
		t.Fatal("expected primary runtime organization path, not migration fallback")
	}

	manifests, err := reg.LoadManifests()
	if err != nil {
		t.Fatalf("LoadManifests() failed: %v", err)
	}
	if len(manifests) != 1 || manifests[0].ID != "bundle-team" {
		t.Fatalf("unexpected manifests: %+v", manifests)
	}
	if len(manifests[0].Members) != 1 || manifests[0].Members[0].ID != "bundle-agent" {
		t.Fatalf("expected embedded bundle member, got %+v", manifests[0].Members)
	}
}
