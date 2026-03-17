package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveStartupSelectionUsesBundleWhenPresent(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
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

	selection, err := ResolveStartupSelection(templatesDir, "")
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

func TestResolveStartupSelectionFailsWhenBundleMissing(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

	if _, err := ResolveStartupSelection(templatesDir, ""); err == nil {
		t.Fatal("expected missing bundle error")
	} else if !strings.Contains(err.Error(), "requires a valid bootstrap template bundle") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveStartupSelectionErrorsWhenBundleInvalid(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

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

	if _, err := ResolveStartupSelection(templatesDir, ""); err == nil {
		t.Fatal("expected invalid bundle error")
	} else if !strings.Contains(err.Error(), "load bootstrap template bundles") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveStartupSelectionDoesNotSilentlyFallbackWhenTemplateRequested(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

	if _, err := ResolveStartupSelection(templatesDir, "expected-bundle"); err == nil {
		t.Fatal("expected requested bundle error")
	} else if !strings.Contains(err.Error(), "requires a valid bootstrap template bundle") {
		t.Fatalf("unexpected error: %v", err)
	}
}
