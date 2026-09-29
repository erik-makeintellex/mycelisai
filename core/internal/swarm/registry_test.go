package swarm

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

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
	})

	org := reg.RuntimeOrganization()
	if org == nil {
		t.Fatal("expected runtime organization")
	}
	if org.ID != "bundle-org" {
		t.Fatalf("expected bundle-org, got %s", org.ID)
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

// CONS-C3: with the teams-directory loader retired, a registry with no
// runtime organization (or one with no teams) loads nothing.
func TestRegistry_EmptyOrganizationLoadsNothing(t *testing.T) {
	for name, reg := range map[string]*Registry{
		"nil organization":   NewRegistryFromRuntimeOrganization(nil),
		"empty organization": NewRegistryFromRuntimeOrganization(&RuntimeOrganization{}),
	} {
		manifests, err := reg.LoadManifests()
		if err != nil || manifests != nil {
			t.Fatalf("%s: LoadManifests() = %v, %v; want nil, nil", name, manifests, err)
		}
	}
}
