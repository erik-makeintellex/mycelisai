package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/swarm"
)

func TestLoadStartupBundleRegistry(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	teamsDir := filepath.Join(root, "teams")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("mkdir teams: %v", err)
	}

	teamYAML := `id: bridge-team
name: Bridge Team
type: action
members:
  - id: bridge-agent
    role: coder
inputs:
  - swarm.team.bridge-team.internal.command
deliveries:
  - swarm.team.bridge-team.signal.status
`
	if err := os.WriteFile(filepath.Join(teamsDir, "bridge-team.yaml"), []byte(teamYAML), 0o644); err != nil {
		t.Fatalf("write team manifest: %v", err)
	}

	bundleYAML := `id: v8-migration-standing-team-bridge
name: V8 Migration Standing-Team Bridge
team_manifest_refs:
  - ../teams/bridge-team.yaml
`
	if err := os.WriteFile(filepath.Join(templatesDir, "v8-migration-standing-team-bridge.yaml"), []byte(bundleYAML), 0o644); err != nil {
		t.Fatalf("write template bundle: %v", err)
	}

	selected, registry, err := loadStartupBundleRegistry(templatesDir, teamsDir)
	if err != nil {
		t.Fatalf("loadStartupBundleRegistry() failed: %v", err)
	}
	if selected.Source != bootstrap.StartupSourceBundle {
		t.Fatalf("expected %s source, got %q", bootstrap.StartupSourceBundle, selected.Source)
	}
	if selected.Bundle == nil || selected.Bundle.ID != "v8-migration-standing-team-bridge" {
		t.Fatalf("selected bundle = %+v", selected.Bundle)
	}
	if selected.Organization == nil || selected.Organization.MigrationFallback {
		t.Fatalf("expected primary runtime organization, got %+v", selected.Organization)
	}
	if selected.Organization.SourceKind != "standing_team_migration_input" {
		t.Fatalf("unexpected source kind: %s", selected.Organization.SourceKind)
	}
	if registry.RuntimeOrganization() == nil || registry.RuntimeOrganization().ID != selected.Organization.ID {
		t.Fatalf("expected registry runtime organization to match selection")
	}
	manifests, err := registry.LoadManifests()
	if err != nil {
		t.Fatalf("registry.LoadManifests() failed: %v", err)
	}
	if len(manifests) != 1 || manifests[0].ID != "bridge-team" {
		t.Fatalf("unexpected startup manifests: %+v", manifests)
	}
}

func TestLoadStartupBundleRegistryFallsBackWithoutBundle(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	teamsDir := filepath.Join(root, "teams")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("mkdir teams: %v", err)
	}

	teamYAML := `id: bridge-team
name: Bridge Team
type: action
members:
  - id: bridge-agent
    role: coder
inputs:
  - swarm.team.bridge-team.internal.command
deliveries:
  - swarm.team.bridge-team.signal.status
`
	if err := os.WriteFile(filepath.Join(teamsDir, "bridge-team.yaml"), []byte(teamYAML), 0o644); err != nil {
		t.Fatalf("write team manifest: %v", err)
	}

	selected, registry, err := loadStartupBundleRegistry(templatesDir, teamsDir)
	if err != nil {
		t.Fatalf("loadStartupBundleRegistry() failed: %v", err)
	}
	if selected.Source != bootstrap.StartupSourceMigrationFallbackTeams {
		t.Fatalf("expected fallback source, got %q", selected.Source)
	}
	if selected.Organization == nil || !selected.Organization.MigrationFallback {
		t.Fatalf("expected migration fallback runtime organization, got %+v", selected.Organization)
	}
	manifests, err := registry.LoadManifests()
	if err != nil {
		t.Fatalf("registry.LoadManifests() failed: %v", err)
	}
	if len(manifests) != 1 || manifests[0].ID != "bridge-team" {
		t.Fatalf("unexpected fallback manifests: %+v", manifests)
	}
}

func TestLoadStartupBundleRegistryRequiresExplicitSelectionForMultipleBundles(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	teamsDir := filepath.Join(root, "teams")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("mkdir teams: %v", err)
	}

	teamYAML := `id: bridge-team
name: Bridge Team
type: action
members:
  - id: bridge-agent
    role: coder
inputs:
  - swarm.team.bridge-team.internal.command
deliveries:
  - swarm.team.bridge-team.signal.status
`
	if err := os.WriteFile(filepath.Join(teamsDir, "bridge-team.yaml"), []byte(teamYAML), 0o644); err != nil {
		t.Fatalf("write team manifest: %v", err)
	}

	firstBundle := `id: bridge-a
name: Bridge A
team_manifest_refs:
  - ../teams/bridge-team.yaml
`
	secondBundle := `id: bridge-b
name: Bridge B
team_manifest_refs:
  - ../teams/bridge-team.yaml
`
	if err := os.WriteFile(filepath.Join(templatesDir, "bridge-a.yaml"), []byte(firstBundle), 0o644); err != nil {
		t.Fatalf("write first bundle: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templatesDir, "bridge-b.yaml"), []byte(secondBundle), 0o644); err != nil {
		t.Fatalf("write second bundle: %v", err)
	}

	t.Setenv("MYCELIS_BOOTSTRAP_TEMPLATE_ID", "")
	if _, _, err := loadStartupBundleRegistry(templatesDir, teamsDir); err == nil {
		t.Fatal("expected bundle selection error")
	} else if !strings.Contains(err.Error(), "set MYCELIS_BOOTSTRAP_TEMPLATE_ID") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadStartupBundleRegistryDoesNotFallbackWhenTemplateRequested(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	teamsDir := filepath.Join(root, "teams")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("mkdir teams: %v", err)
	}

	teamYAML := `id: bridge-team
name: Bridge Team
type: action
members:
  - id: bridge-agent
    role: coder
inputs:
  - swarm.team.bridge-team.internal.command
deliveries:
  - swarm.team.bridge-team.signal.status
`
	if err := os.WriteFile(filepath.Join(teamsDir, "bridge-team.yaml"), []byte(teamYAML), 0o644); err != nil {
		t.Fatalf("write team manifest: %v", err)
	}

	t.Setenv("MYCELIS_BOOTSTRAP_TEMPLATE_ID", "missing-bundle")
	if _, _, err := loadStartupBundleRegistry(templatesDir, teamsDir); err == nil {
		t.Fatal("expected requested bundle error")
	} else if !strings.Contains(err.Error(), "migration-only compatibility") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveStartupProviderRoutingUsesRuntimeOrganizationPolicy(t *testing.T) {
	selection := &bootstrap.StartupSelection{
		Organization: &swarm.RuntimeOrganization{
			ID: "bundle-org",
			ProviderPolicy: swarm.ProviderPolicy{
				Provider: "org-provider",
				Teams: map[string]swarm.ProviderScope{
					"prime-development": {Provider: "team-provider"},
				},
				Agents: map[string]swarm.ProviderScope{
					"prime-development-agent": {Provider: "agent-provider"},
				},
			},
		},
		Source: bootstrap.StartupSourceBundle,
	}

	routing := resolveStartupProviderRouting(
		selection,
		`{"prime-development":"legacy-team-provider"}`,
		`{"prime-development-agent":"legacy-agent-provider"}`,
	)
	if routing.Source != "runtime_organization" {
		t.Fatalf("routing source = %q", routing.Source)
	}
	if routing.Policy.Provider != "org-provider" {
		t.Fatalf("provider = %q", routing.Policy.Provider)
	}
	if routing.Policy.Teams["prime-development"].Provider != "team-provider" {
		t.Fatalf("team provider = %q", routing.Policy.Teams["prime-development"].Provider)
	}
	if routing.Policy.Agents["prime-development-agent"].Provider != "agent-provider" {
		t.Fatalf("agent provider = %q", routing.Policy.Agents["prime-development-agent"].Provider)
	}
	if !routing.IgnoredLegacyEnvMaps {
		t.Fatal("expected legacy env maps to be ignored on bundle path")
	}
}

func TestResolveStartupProviderRoutingKeepsMigrationFallbackIsolated(t *testing.T) {
	selection := &bootstrap.StartupSelection{
		Organization: &swarm.RuntimeOrganization{
			ID:                "migration-fallback-standing-teams",
			MigrationFallback: true,
		},
		Source: bootstrap.StartupSourceMigrationFallbackTeams,
	}

	routing := resolveStartupProviderRouting(
		selection,
		`{"council-core":"legacy-team-provider"}`,
		`{"council-architect":"legacy-agent-provider"}`,
	)
	if !routing.Policy.IsEmpty() {
		t.Fatalf("expected no provider policy on isolated fallback path, got %+v", routing.Policy)
	}
	if routing.Source != "" {
		t.Fatalf("expected no provider routing source, got %q", routing.Source)
	}
	if !routing.IgnoredLegacyEnvMaps {
		t.Fatal("expected retired env-map compatibility inputs to be ignored")
	}
}
