package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if selected.Bundle == nil || selected.Bundle.ID != "v8-migration-standing-team-bridge" {
		t.Fatalf("selected bundle = %+v", selected.Bundle)
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
	if selected.Source != "fallback_teams" {
		t.Fatalf("expected fallback source, got %q", selected.Source)
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
