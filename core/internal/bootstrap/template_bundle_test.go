package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateLoader_LoadBundles(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

	bundleYAML := `id: fixture-startup-bundle
name: Fixture Startup Bundle
source_kind: standing_team_migration_input
teams:
  - id: legacy-team
    name: Legacy Team
    type: action
    members:
      - id: legacy-agent
        role: coder
    inputs:
      - swarm.team.legacy-team.internal.command
    deliveries:
      - swarm.team.legacy-team.signal.status
`
	if err := os.WriteFile(filepath.Join(templatesDir, "fixture-startup-bundle.yaml"), []byte(bundleYAML), 0o644); err != nil {
		t.Fatalf("write template bundle: %v", err)
	}

	loader := NewTemplateLoader(templatesDir)
	bundles, err := loader.LoadBundles()
	if err != nil {
		t.Fatalf("LoadBundles() failed: %v", err)
	}
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}
	if bundles[0].ID != "fixture-startup-bundle" {
		t.Fatalf("expected fixture-startup-bundle, got %s", bundles[0].ID)
	}
	if bundles[0].TemplateVersion != "v1alpha1" {
		t.Fatalf("expected default template version, got %s", bundles[0].TemplateVersion)
	}
	if len(bundles[0].Teams) != 1 || bundles[0].Teams[0].ID != "legacy-team" {
		t.Fatalf("expected embedded legacy-team manifest, got %+v", bundles[0].Teams)
	}
}

func TestTemplateLoader_LoadBundlesRejectsInvalidEmbeddedTeam(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

	bundleYAML := `id: broken-bridge
name: Broken Bridge
teams:
  - id: broken-team
    name: Broken Team
    members: []
    inputs:
      - swarm.team.broken-team.internal.command
    deliveries:
      - swarm.team.broken-team.signal.status
`
	if err := os.WriteFile(filepath.Join(templatesDir, "broken-bridge.yaml"), []byte(bundleYAML), 0o644); err != nil {
		t.Fatalf("write template bundle: %v", err)
	}

	loader := NewTemplateLoader(templatesDir)
	if _, err := loader.LoadBundles(); err == nil {
		t.Fatal("expected LoadBundles() to fail for invalid embedded team")
	}
}

func TestTemplateLoader_LoadShippedBootBundles(t *testing.T) {
	loader := NewTemplateLoader(filepath.Join("..", "..", "config", "templates"))
	bundles, err := loader.LoadBundles()
	if err != nil {
		t.Fatalf("LoadBundles() failed: %v", err)
	}

	if len(bundles) != 2 || bundles[0].ID != OptionalDevSwarmBundleID || bundles[1].ID != DefaultStartupBundleID {
		ids := make([]string, 0, len(bundles))
		for _, bundle := range bundles {
			ids = append(ids, bundle.ID)
		}
		t.Fatalf("shipped bundles = %v, want [%s %s]", ids, OptionalDevSwarmBundleID, DefaultStartupBundleID)
	}
}

func TestTemplateBundle_InstantiateRuntimeOrganization(t *testing.T) {
	loader := NewTemplateLoader(filepath.Join("..", "..", "config", "templates"))
	bundles, err := loader.LoadBundles()
	if err != nil {
		t.Fatalf("LoadBundles() failed: %v", err)
	}

	for _, bundle := range bundles {
		org, err := bundle.InstantiateRuntimeOrganization()
		if err != nil {
			t.Fatalf("%s: InstantiateRuntimeOrganization() failed: %v", bundle.ID, err)
		}
		if org.ID != bundle.ID {
			t.Fatalf("expected organization id %s, got %s", bundle.ID, org.ID)
		}
		if org.SourceKind != "standing_team_migration_input" {
			t.Fatalf("%s: expected standing_team_migration_input, got %s", bundle.ID, org.SourceKind)
		}
		if len(org.Teams) == 0 || org.Teams[0].ID != "admin-core" {
			t.Fatalf("%s: expected first embedded team admin-core, got %+v", bundle.ID, org.Teams)
		}
		if len(org.Teams[0].Members) == 0 || org.Teams[0].Members[0].ID != "admin" {
			t.Fatalf("%s: expected admin-core members to be embedded, got %+v", bundle.ID, org.Teams[0].Members)
		}
		// Shipped teams must not pin a provider; they route through configured cognitive profiles.
		if org.ProviderPolicy.Provider != "" {
			t.Fatalf("%s: expected provider unpinned, got %q", bundle.ID, org.ProviderPolicy.Provider)
		}
		if len(org.ProviderPolicy.Kernel.RoleProviders) != 0 || len(org.ProviderPolicy.Council.RoleProviders) != 0 {
			t.Fatalf("%s: expected no pinned role providers, got kernel=%v council=%v", bundle.ID, org.ProviderPolicy.Kernel.RoleProviders, org.ProviderPolicy.Council.RoleProviders)
		}
	}
}

func TestTemplateBundleInstantiateRuntimeOrganizationCarriesStructuredProviderPolicy(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

	bundleYAML := `id: structured-policy-bundle
name: Structured Policy Bundle
provider_policy:
  provider: org-provider
  allowed_providers:
    - org-provider
    - council-provider
  council:
    role_providers:
      architect: council-provider
teams:
  - id: council-core
    name: Council
    type: action
    members:
      - id: council-architect
        role: architect
    inputs:
      - swarm.global.broadcast
    deliveries:
      - swarm.team.council-core.signal.status
`
	if err := os.WriteFile(filepath.Join(templatesDir, "structured-policy-bundle.yaml"), []byte(bundleYAML), 0o644); err != nil {
		t.Fatalf("write template bundle: %v", err)
	}

	loader := NewTemplateLoader(templatesDir)
	bundles, err := loader.LoadBundles()
	if err != nil {
		t.Fatalf("LoadBundles() failed: %v", err)
	}

	org, err := bundles[0].InstantiateRuntimeOrganization()
	if err != nil {
		t.Fatalf("InstantiateRuntimeOrganization() failed: %v", err)
	}
	if org.ProviderPolicy.Provider != "org-provider" {
		t.Fatalf("organization provider = %q", org.ProviderPolicy.Provider)
	}
	if org.ProviderPolicy.Council.RoleProviders["architect"] != "council-provider" {
		t.Fatalf("architect role provider = %q", org.ProviderPolicy.Council.RoleProviders["architect"])
	}
}

func TestSelectStartupBundle(t *testing.T) {
	bundles := []*TemplateBundle{
		{ID: "fixture-startup-bundle"},
	}

	selected, err := SelectStartupBundle(bundles, "")
	if err != nil {
		t.Fatalf("SelectStartupBundle() failed: %v", err)
	}
	if selected.ID != "fixture-startup-bundle" {
		t.Fatalf("expected bridge bundle, got %s", selected.ID)
	}
}
