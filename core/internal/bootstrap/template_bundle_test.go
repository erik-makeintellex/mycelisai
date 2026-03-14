package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateLoader_LoadBundles(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	teamsDir := filepath.Join(root, "teams")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("mkdir teams: %v", err)
	}

	teamYAML := `id: legacy-team
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
	if err := os.WriteFile(filepath.Join(teamsDir, "legacy-team.yaml"), []byte(teamYAML), 0o644); err != nil {
		t.Fatalf("write team manifest: %v", err)
	}

	bundleYAML := `id: v8-migration-standing-team-bridge
name: V8 Migration Standing-Team Bridge
source_kind: standing_team_migration_input
team_manifest_refs:
  - ../teams/legacy-team.yaml
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
	manifests, err := bundles[0].LoadTeamManifests()
	if err != nil {
		t.Fatalf("LoadTeamManifests() failed: %v", err)
	}
	if len(manifests) != 1 || manifests[0].ID != "legacy-team" {
		t.Fatalf("expected 1 legacy-team manifest, got %+v", manifests)
	}
}

func TestTemplateLoader_LoadBundlesRejectsMissingManifestRef(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

	bundleYAML := `id: broken-bridge
name: Broken Bridge
team_manifest_refs:
  - ../teams/missing.yaml
`
	if err := os.WriteFile(filepath.Join(templatesDir, "broken-bridge.yaml"), []byte(bundleYAML), 0o644); err != nil {
		t.Fatalf("write template bundle: %v", err)
	}

	loader := NewTemplateLoader(templatesDir)
	if _, err := loader.LoadBundles(); err == nil {
		t.Fatal("expected LoadBundles() to fail for missing manifest ref")
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
	if len(bundles[0].TeamManifestRefs) == 0 {
		t.Fatal("expected migration bridge bundle to reference standing manifests")
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
	if org.MigrationFallback {
		t.Fatal("expected bundle-instantiated organization to be primary path, not migration fallback")
	}
	if len(org.Teams) == 0 {
		t.Fatal("expected instantiated runtime organization to include teams")
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
