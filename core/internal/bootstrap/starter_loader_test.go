package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var fixturePolicy = map[string]struct{}{"fixture-approve-external": {}}

func readFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(fixtureStarterPath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return string(data)
}

func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func entryByID(t *testing.T, entries []*StarterEntry, id string) *StarterEntry {
	t.Helper()
	for _, e := range entries {
		if e.ID == id {
			return e
		}
	}
	t.Fatalf("entry %q not found in %d entries", id, len(entries))
	return nil
}

func TestLoadStarterEntriesMissingDirIsEmpty(t *testing.T) {
	entries, err := LoadStarterEntries(filepath.Join(t.TempDir(), "starters"), nil, fixturePolicy)
	if err != nil || len(entries) != 0 {
		t.Fatalf("expected no entries and no error, got %d, %v", len(entries), err)
	}
}

func TestLoadStarterEntriesIgnoresTestdataSubdir(t *testing.T) {
	entries, err := LoadStarterEntries(StartersPath(filepath.Join("..", "..", "config", "templates")), nil, fixturePolicy)
	if err != nil {
		t.Fatalf("LoadStarterEntries() failed: %v", err)
	}
	for _, e := range entries {
		if e.ID == "fixture-minimal-starter" {
			t.Fatal("testdata fixture must not be listed as a runtime starter")
		}
	}
}

func TestLoadStarterEntriesAvailableAndUnavailable(t *testing.T) {
	dir := t.TempDir()
	fixture := readFixture(t)
	writeFile(t, dir, "a-good.yaml", fixture)
	writeFile(t, dir, "b-unknown-policy.yaml", strings.NewReplacer(
		"id: fixture-minimal-starter", "id: unknown-policy",
		"approval_policy_ref: fixture-approve-external", "approval_policy_ref: nope").Replace(fixture))
	writeFile(t, dir, "c-no-starter.yaml", strings.NewReplacer(
		"id: fixture-minimal-starter", "id: no-starter",
		"\nstarter:", "\nlabels_moved:").Replace(fixture))
	writeFile(t, dir, "d-broken.yaml", "id: [unterminated")
	writeFile(t, dir, "e-duplicate.yml", fixture)
	writeFile(t, dir, "f-bad-mode.yaml", strings.NewReplacer(
		"id: fixture-minimal-starter", "id: bad-mode",
		"work_mode: propose_only", "work_mode: yolo").Replace(fixture))
	writeFile(t, dir, "notes.txt", "ignored")

	entries, err := LoadStarterEntries(dir, nil, fixturePolicy)
	if err != nil {
		t.Fatalf("LoadStarterEntries() failed: %v", err)
	}
	if len(entries) != 6 {
		t.Fatalf("expected 6 entries, got %d", len(entries))
	}

	good := entries[0]
	if good.ID != "fixture-minimal-starter" || !good.Available || good.Bundle == nil {
		t.Fatalf("expected good starter available, got %+v", good)
	}

	expectUnavailable := map[string]string{
		"unknown-policy": `approval policy "nope" is not a group in policy.yaml`,
		"no-starter":     "starter bundles must declare a starter block",
		"d-broken":       "parse d-broken.yaml",
		"bad-mode":       `unknown work_mode "yolo"`,
	}
	for id, want := range expectUnavailable {
		e := entryByID(t, entries, id)
		if e.Available || !strings.Contains(e.UnavailableReason, want) {
			t.Errorf("%s: want unavailable with %q, got available=%v reason=%q", id, want, e.Available, e.UnavailableReason)
		}
	}
	if entryByID(t, entries, "d-broken").Bundle != nil || entryByID(t, entries, "bad-mode").Bundle != nil {
		t.Error("invalid starters must not expose a bundle")
	}
	dup := entries[4]
	if dup.File != "e-duplicate.yml" || dup.Available || !strings.Contains(dup.UnavailableReason, "duplicate starter id") {
		t.Errorf("expected duplicate starter unavailable, got %+v", dup)
	}
}

func TestLoadStarterEntriesPolicyNotLoaded(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "starter.yaml", readFixture(t))
	entries, err := LoadStarterEntries(dir, nil, nil)
	if err != nil {
		t.Fatalf("LoadStarterEntries() failed: %v", err)
	}
	if len(entries) != 1 || entries[0].Available || !strings.Contains(entries[0].UnavailableReason, "governance policy is not loaded") {
		t.Fatalf("expected unavailable starter, got %+v", entries[0])
	}
}

func TestLoadStarterEntriesReservedIDClash(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "starter.yaml", readFixture(t))
	reserved := map[string]struct{}{"fixture-minimal-starter": {}}
	entries, err := LoadStarterEntries(dir, reserved, fixturePolicy)
	if err != nil {
		t.Fatalf("LoadStarterEntries() failed: %v", err)
	}
	if entries[0].Available || !strings.Contains(entries[0].UnavailableReason, "already used by a top-level template bundle") {
		t.Fatalf("expected reserved id clash, got %+v", entries[0])
	}
}

func TestBootSelectionIgnoresStartersDir(t *testing.T) {
	templatesDir := t.TempDir()
	writeFile(t, templatesDir, "bridge.yaml", `id: only-boot-bundle
name: Only Boot Bundle
teams:
  - id: bridge-team
    name: Bridge Team
    members:
      - id: bridge-agent
        role: coder
    inputs: [swarm.team.bridge-team.internal.command]
    deliveries: [swarm.team.bridge-team.signal.status]
`)
	fixture := readFixture(t)
	writeFile(t, StartersPath(templatesDir), "one.yaml", fixture)
	writeFile(t, StartersPath(templatesDir), "two.yaml", strings.Replace(fixture, "id: fixture-minimal-starter", "id: second-starter", 1))

	selection, err := ResolveStartupSelection(templatesDir, "")
	if err != nil {
		t.Fatalf("host boot without MYCELIS_BOOTSTRAP_TEMPLATE_ID must still select one bundle: %v", err)
	}
	if selection.Bundle.ID != "only-boot-bundle" {
		t.Fatalf("boot selected %q, want only-boot-bundle", selection.Bundle.ID)
	}
	if _, err := ResolveStartupSelection(templatesDir, "fixture-minimal-starter"); err == nil {
		t.Fatal("a starter must never be selectable as the boot bundle")
	}
}

func TestBootSelectionWithRepoTemplatesUnchanged(t *testing.T) {
	selection, err := ResolveStartupSelection(filepath.Join("..", "..", "config", "templates"), "")
	if err != nil {
		t.Fatalf("ResolveStartupSelection() failed: %v", err)
	}
	if selection.Bundle.ID != "v8-migration-standing-team-bridge" {
		t.Fatalf("boot selected %q", selection.Bundle.ID)
	}
}
