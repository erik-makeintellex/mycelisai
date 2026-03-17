package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeStartupSelectionTeam(t *testing.T, teamsDir string) {
	t.Helper()
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
}

func TestResolveStartupSelectionUsesBundleWhenPresent(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	teamsDir := filepath.Join(root, "migration-only-fallback-teams")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

	bundleYAML := `id: v8-migration-standing-team-bridge
name: V8 Migration Standing-Team Bridge
teams:
  - id: bridge-team
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
	if err := os.WriteFile(filepath.Join(templatesDir, "v8-migration-standing-team-bridge.yaml"), []byte(bundleYAML), 0o644); err != nil {
		t.Fatalf("write template bundle: %v", err)
	}

	selection, err := ResolveStartupSelection(templatesDir, teamsDir, "")
	if err != nil {
		t.Fatalf("ResolveStartupSelection() failed: %v", err)
	}
	if selection.Source != StartupSourceBundle {
		t.Fatalf("expected bundle source, got %q", selection.Source)
	}
	if selection.Bundle == nil || selection.Bundle.ID != "v8-migration-standing-team-bridge" {
		t.Fatalf("unexpected selected bundle: %+v", selection.Bundle)
	}
	if selection.Organization == nil {
		t.Fatal("expected instantiated runtime organization")
	}
	if selection.Organization.MigrationFallback {
		t.Fatal("expected bundle path to be primary runtime organization, not fallback")
	}
	if selection.Organization.SourceKind != "standing_team_migration_input" {
		t.Fatalf("unexpected organization source kind: %s", selection.Organization.SourceKind)
	}
	if len(selection.Organization.Teams) != 1 || selection.Organization.Teams[0].ID != "bridge-team" {
		t.Fatalf("unexpected organization teams: %+v", selection.Organization.Teams)
	}
	if len(selection.Manifests) != 1 || selection.Manifests[0].ID != "bridge-team" {
		t.Fatalf("unexpected startup manifests: %+v", selection.Manifests)
	}
}

func TestResolveStartupSelectionFallsBackWhenBundleMissing(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	teamsDir := filepath.Join(root, "teams")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("mkdir teams: %v", err)
	}

	writeStartupSelectionTeam(t, teamsDir)

	selection, err := ResolveStartupSelection(templatesDir, teamsDir, "")
	if err != nil {
		t.Fatalf("ResolveStartupSelection() failed: %v", err)
	}
	if selection.Source != StartupSourceMigrationFallbackTeams {
		t.Fatalf("expected %s source, got %q", StartupSourceMigrationFallbackTeams, selection.Source)
	}
	if selection.Bundle != nil {
		t.Fatalf("expected no selected bundle, got %+v", selection.Bundle)
	}
	if selection.Organization == nil {
		t.Fatal("expected fallback runtime organization")
	}
	if !selection.Organization.MigrationFallback {
		t.Fatal("expected fallback runtime organization to be marked migration-only")
	}
	if selection.Organization.SourceKind != "standing_team_migration_fallback" {
		t.Fatalf("unexpected fallback source kind: %s", selection.Organization.SourceKind)
	}
	if len(selection.Manifests) != 1 || selection.Manifests[0].ID != "bridge-team" {
		t.Fatalf("unexpected fallback manifests: %+v", selection.Manifests)
	}
}

func TestResolveStartupSelectionErrorsWhenBundleInvalid(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	teamsDir := filepath.Join(root, "teams")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("mkdir teams: %v", err)
	}

	writeStartupSelectionTeam(t, teamsDir)

	invalidBundle := `id: broken-bridge
name: Broken Bridge
teams:
  - id: broken-team
    name: Broken Team
    members: []
    inputs:
      - swarm.team.broken.internal.command
    deliveries:
      - swarm.team.broken.signal.status
`
	if err := os.WriteFile(filepath.Join(templatesDir, "broken-bridge.yaml"), []byte(invalidBundle), 0o644); err != nil {
		t.Fatalf("write invalid template bundle: %v", err)
	}

	if _, err := ResolveStartupSelection(templatesDir, teamsDir, ""); err == nil {
		t.Fatal("expected invalid bundle error")
	} else if !strings.Contains(err.Error(), "load bootstrap template bundles") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveStartupSelectionDoesNotSilentlyFallbackWhenTemplateRequested(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	teamsDir := filepath.Join(root, "teams")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("mkdir teams: %v", err)
	}

	writeStartupSelectionTeam(t, teamsDir)

	if _, err := ResolveStartupSelection(templatesDir, teamsDir, "expected-bundle"); err == nil {
		t.Fatal("expected requested bundle error")
	} else if !strings.Contains(err.Error(), "migration-only compatibility") {
		t.Fatalf("unexpected error: %v", err)
	}
}
