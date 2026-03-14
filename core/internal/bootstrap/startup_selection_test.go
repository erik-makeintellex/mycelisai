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
	teamsDir := filepath.Join(root, "teams")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("mkdir teams: %v", err)
	}

	writeStartupSelectionTeam(t, teamsDir)

	bundleYAML := `id: v8-migration-standing-team-bridge
name: V8 Migration Standing-Team Bridge
team_manifest_refs:
  - ../teams/bridge-team.yaml
`
	if err := os.WriteFile(filepath.Join(templatesDir, "v8-migration-standing-team-bridge.yaml"), []byte(bundleYAML), 0o644); err != nil {
		t.Fatalf("write template bundle: %v", err)
	}

	selection, err := ResolveStartupSelection(templatesDir, teamsDir, "")
	if err != nil {
		t.Fatalf("ResolveStartupSelection() failed: %v", err)
	}
	if selection.Source != "bundle" {
		t.Fatalf("expected bundle source, got %q", selection.Source)
	}
	if selection.Bundle == nil || selection.Bundle.ID != "v8-migration-standing-team-bridge" {
		t.Fatalf("unexpected selected bundle: %+v", selection.Bundle)
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
	if selection.Source != "fallback_teams" {
		t.Fatalf("expected fallback_teams source, got %q", selection.Source)
	}
	if selection.Bundle != nil {
		t.Fatalf("expected no selected bundle, got %+v", selection.Bundle)
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
team_manifest_refs:
  - ../teams/missing.yaml
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
