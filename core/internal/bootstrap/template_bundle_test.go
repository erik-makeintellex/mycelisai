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

	bundleYAML := `id: v8-migration-standing-team-bridge
name: V8 Migration Standing-Team Bridge
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
	if err := os.WriteFile(filepath.Join(templatesDir, "v8-migration-standing-team-bridge.yaml"), []byte(bundleYAML), 0o644); err != nil {
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
	if bundles[0].ID != "v8-migration-standing-team-bridge" {
		t.Fatalf("expected v8-migration-standing-team-bridge, got %s", bundles[0].ID)
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

func TestTemplateLoader_LoadStandingMigrationBridgeBundle(t *testing.T) {
	loader := NewTemplateLoader(filepath.Join("..", "..", "config", "templates"))
	bundles, err := loader.LoadBundles()
	if err != nil {
		t.Fatalf("LoadBundles() failed: %v", err)
	}

	if len(bundles) != 1 {
		t.Fatalf("expected 1 standing template bundle, got %d", len(bundles))
	}
	if bundles[0].ID != "v8-migration-standing-team-bridge" {
		t.Fatalf("expected v8-migration-standing-team-bridge bundle, got %s", bundles[0].ID)
	}
	if len(bundles[0].Teams) == 0 {
		t.Fatal("expected migration bridge bundle to embed teams")
	}
}

func TestTemplateBundle_InstantiateRuntimeOrganization(t *testing.T) {
	loader := NewTemplateLoader(filepath.Join("..", "..", "config", "templates"))
	bundles, err := loader.LoadBundles()
	if err != nil {
		t.Fatalf("LoadBundles() failed: %v", err)
	}

	org, err := bundles[0].InstantiateRuntimeOrganization()
	if err != nil {
		t.Fatalf("InstantiateRuntimeOrganization() failed: %v", err)
	}
	if org.ID != "v8-migration-standing-team-bridge" {
		t.Fatalf("expected organization id v8-migration-standing-team-bridge, got %s", org.ID)
	}
	if org.SourceKind != "standing_team_migration_input" {
		t.Fatalf("expected standing_team_migration_input, got %s", org.SourceKind)
	}
	if len(org.Teams) == 0 {
		t.Fatal("expected instantiated runtime organization to include teams")
	}
	if org.Teams[0].ID != "admin-core" {
		t.Fatalf("expected first embedded team admin-core, got %s", org.Teams[0].ID)
	}
	if len(org.Teams[0].Members) == 0 || org.Teams[0].Members[0].ID != "admin" {
		t.Fatalf("expected admin-core members to be embedded, got %+v", org.Teams[0].Members)
	}
	if org.ProviderPolicy.Provider != "local-ollama-dev" {
		t.Fatalf("expected migration bridge provider default local-ollama-dev, got %q", org.ProviderPolicy.Provider)
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
		{ID: "v8-migration-standing-team-bridge"},
	}

	selected, err := SelectStartupBundle(bundles, "")
	if err != nil {
		t.Fatalf("SelectStartupBundle() failed: %v", err)
	}
	if selected.ID != "v8-migration-standing-team-bridge" {
		t.Fatalf("expected bridge bundle, got %s", selected.ID)
	}
}
