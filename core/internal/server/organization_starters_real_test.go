package server

import (
	"testing"

	"github.com/mycelis/core/internal/governance"
)

// TestRealStarterCatalogAllFourStartersAvailable loads the real starters
// directory (core/config/templates/starters/) and the real policy.yaml
// (RC-004B) and asserts all four starters validate and resolve their
// approval policy, so the organization-starters catalog can actually offer
// them.
func TestRealStarterCatalogAllFourStartersAvailable(t *testing.T) {
	engine, err := governance.NewEngine("../../config/policy.yaml")
	if err != nil {
		t.Fatalf("load real policy.yaml: %v", err)
	}

	s := newTestServer(
		withTemplateBundlesPath("../../config/templates"),
		withGuard(engine.Config),
	)

	catalog, err := s.loadOrganizationStarterCatalog()
	if err != nil {
		t.Fatalf("loadOrganizationStarterCatalog() error = %v", err)
	}

	wantIDs := map[string]bool{
		"starter-client-delivery-studio":      false,
		"starter-product-delivery-team":       false,
		"starter-operations-desk":             false,
		"starter-governed-enterprise-program": false,
	}

	for _, item := range catalog {
		if _, tracked := wantIDs[item.ID]; !tracked {
			continue
		}
		if item.Source != starterSourceStarter {
			t.Errorf("starter %s: source = %q, want %q", item.ID, item.Source, starterSourceStarter)
		}
		if !item.Available {
			t.Errorf("starter %s unavailable: %s", item.ID, item.UnavailableReason)
			continue
		}
		if item.Starter == nil {
			t.Errorf("starter %s missing starter labels", item.ID)
		}
		if item.DeliveryDefaults == nil {
			t.Errorf("starter %s missing delivery_defaults", item.ID)
			continue
		}
		if len(item.Departments) == 0 {
			t.Errorf("starter %s has no departments", item.ID)
		}
		for _, dept := range item.Departments {
			if dept.SpecialistCount > item.DeliveryDefaults.MaxTeamMembers {
				t.Errorf("starter %s department %s has %d members, above max_team_members %d",
					item.ID, dept.ID, dept.SpecialistCount, item.DeliveryDefaults.MaxTeamMembers)
			}
		}
		wantIDs[item.ID] = true

		// The starter's own resolution path must also succeed, since this
		// is what organization creation uses.
		entry, err := s.resolveStarterCatalogEntry(item.ID)
		if err != nil {
			t.Errorf("resolveStarterCatalogEntry(%s) error = %v", item.ID, err)
			continue
		}
		if entry.Starter == nil || entry.Starter.SomaIntro == "" {
			t.Errorf("starter %s missing soma_intro", item.ID)
		}
	}

	for id, found := range wantIDs {
		if !found {
			t.Errorf("expected starter %s in real catalog, not found or unavailable", id)
		}
	}
}
