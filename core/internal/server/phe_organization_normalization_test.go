package server

import "testing"

// TestPheNormalizeOrganizationHome_CountOnlyDepartmentsAreNeverInvented is the
// regression test for F18: when a caller supplies only DepartmentCount with no
// real Departments records, normalizeOrganizationHome must not fabricate
// department names/IDs (e.g. "Core Delivery Department") from that count. The
// list stays empty and the count is exposed as-is.
func TestPheNormalizeOrganizationHome_CountOnlyDepartmentsAreNeverInvented(t *testing.T) {
	home := normalizeOrganizationHome(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:              "org-count-only",
			Name:            "Count Only Org",
			DepartmentCount: 3,
			SpecialistCount: 6,
		},
	})

	if len(home.Departments) != 0 {
		t.Fatalf("expected no invented department records, got %+v", home.Departments)
	}
	if home.DepartmentCount != 3 {
		t.Fatalf("expected the reported count to be preserved honestly, got %d", home.DepartmentCount)
	}
}

// TestPheNormalizeOrganizationHome_RealDepartmentsStillNormalized guards against a
// regression where fixing F18 also breaks normalization of departments that were
// actually supplied (only the count-only fabrication path is removed).
func TestPheNormalizeOrganizationHome_RealDepartmentsStillNormalized(t *testing.T) {
	home := normalizeOrganizationHome(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:              "org-real-departments",
			Name:            "Real Departments Org",
			DepartmentCount: 1,
			SpecialistCount: 2,
		},
		Departments: []OrganizationDepartmentSummary{
			{SpecialistCount: 2},
		},
	})

	if len(home.Departments) != 1 {
		t.Fatalf("expected the one supplied department to survive normalization, got %+v", home.Departments)
	}
	if home.Departments[0].Name != "Core Delivery Department" {
		t.Fatalf("expected the existing gap-filling default name for a supplied-but-unnamed department, got %q", home.Departments[0].Name)
	}
}
