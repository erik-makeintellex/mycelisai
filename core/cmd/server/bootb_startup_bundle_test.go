package main

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// BOOT-B: the real startup path (loadStartupBundleRegistry reads
// MYCELIS_BOOTSTRAP_TEMPLATE_ID) boots only admin-core and council-core from
// the shipped templates, and a deployment pinned to the retired bridge
// bundle fails closed with a message naming the replacements.
func TestBootBStartupRegistryBootsRuntimeCoreByDefault(t *testing.T) {
	shipped := filepath.Join("..", "..", "config", "templates")

	t.Setenv("MYCELIS_BOOTSTRAP_TEMPLATE_ID", "")
	selection, registry, err := loadStartupBundleRegistry(shipped)
	if err != nil {
		t.Fatalf("default startup failed: %v", err)
	}
	manifests, err := registry.LoadManifests()
	if err != nil {
		t.Fatalf("LoadManifests: %v", err)
	}
	ids := make([]string, 0, len(manifests))
	for _, manifest := range manifests {
		ids = append(ids, manifest.ID)
	}
	if selection.Bundle.ID != "mycelis-runtime-core" || !slices.Equal(ids, []string{"admin-core", "council-core"}) {
		t.Fatalf("default boot = %s %v, want mycelis-runtime-core [admin-core council-core]", selection.Bundle.ID, ids)
	}

	t.Setenv("MYCELIS_BOOTSTRAP_TEMPLATE_ID", "v8-migration-standing-team-bridge")
	if _, _, err := loadStartupBundleRegistry(shipped); err == nil {
		t.Fatal("startup accepted the retired bridge bundle")
	} else if !strings.Contains(err.Error(), "mycelis-runtime-core") || !strings.Contains(err.Error(), "mycelis-dev-swarm-optional") {
		t.Fatalf("retired bundle error %q must name mycelis-runtime-core and mycelis-dev-swarm-optional", err)
	}
}
