package bootstrap

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/swarm"
)

// BOOT-B (owner decision 2026-09-29): first boot starts only admin-core and
// council-core; prime-architect, prime-development and agui-design-architect
// ship as an optional bundle that is off by default; the V8 migration bridge
// bundle is retired with no alias.

var bootbShippedTemplateDirs = []string{
	filepath.Join("..", "..", "config", "templates"),
	filepath.Join("..", "..", "..", "charts", "mycelis-core", "config", "templates"),
}

var bootbOptionalTeamIDs = []string{"agui-design-architect", "prime-architect", "prime-development"}

func bootbTeamIDs(selection *StartupSelection) []string {
	ids := make([]string, 0, len(selection.Manifests))
	for _, manifest := range selection.Manifests {
		ids = append(ids, manifest.ID)
	}
	return ids
}

func bootbWriteBundle(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func bootbFixtureBundle(id, extends string, teamIDs ...string) string {
	var b strings.Builder
	b.WriteString("id: " + id + "\nname: " + id + "\n")
	if extends != "" {
		b.WriteString("extends: " + extends + "\n")
	}
	b.WriteString("teams:\n")
	for _, team := range teamIDs {
		b.WriteString("  - id: " + team + "\n    name: " + team + "\n    type: action\n")
		b.WriteString("    members:\n      - id: " + team + "-agent\n        role: coder\n")
		b.WriteString("    inputs:\n      - swarm.team." + team + ".internal.command\n")
		b.WriteString("    deliveries:\n      - swarm.team." + team + ".signal.status\n")
	}
	return b.String()
}

func TestBootBShippedDefaultBootsOnlyRuntimeCore(t *testing.T) {
	for _, dir := range bootbShippedTemplateDirs {
		selection, err := ResolveStartupSelection(dir, "")
		if err != nil {
			t.Fatalf("%s: default startup selection failed: %v", dir, err)
		}
		if selection.Bundle.ID != DefaultStartupBundleID || DefaultStartupBundleID != "mycelis-runtime-core" {
			t.Fatalf("%s: default bundle = %q, want mycelis-runtime-core", dir, selection.Bundle.ID)
		}
		if got := bootbTeamIDs(selection); !slices.Equal(got, []string{"admin-core", "council-core"}) {
			t.Fatalf("%s: default boot teams = %v, want exactly [admin-core council-core]", dir, got)
		}
	}
	// Dropped from boot, still reserved: nothing may register these IDs.
	for _, id := range []string{"genesis-core", "telemetry-core"} {
		if !swarm.IsReservedTeamID(id) {
			t.Fatalf("%s must stay a reserved Core team ID after leaving boot", id)
		}
	}
}

func TestBootBOptionalDevSwarmLoadsOnlyWhenSelected(t *testing.T) {
	for _, dir := range bootbShippedTemplateDirs {
		selection, err := ResolveStartupSelection(dir, OptionalDevSwarmBundleID)
		if err != nil {
			t.Fatalf("%s: optional bundle selection failed: %v", dir, err)
		}
		want := append([]string{"admin-core", "council-core"}, bootbOptionalTeamIDs...)
		if got := bootbTeamIDs(selection); !slices.Equal(got, want) {
			t.Fatalf("%s: optional boot teams = %v, want %v", dir, got, want)
		}
		if selection.Organization.ID != "mycelis-dev-swarm-optional" {
			t.Fatalf("%s: optional organization id = %q", dir, selection.Organization.ID)
		}
		defaultSelection, err := ResolveStartupSelection(dir, "")
		if err != nil {
			t.Fatalf("%s: default selection failed: %v", dir, err)
		}
		for _, id := range bootbOptionalTeamIDs {
			if slices.Contains(bootbTeamIDs(defaultSelection), id) {
				t.Fatalf("%s: optional team %s loaded without being selected", dir, id)
			}
		}
	}
}

func bootbRequireRetiredError(t *testing.T, err error, context string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: retired bridge bundle was accepted", context)
	}
	for _, want := range []string{"v8-migration-standing-team-bridge", "retired", "mycelis-runtime-core", "mycelis-dev-swarm-optional"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("%s: error %q does not mention %q", context, err, want)
		}
	}
}

func TestBootBRetiredBridgeBundleFailsClosed(t *testing.T) {
	for _, requested := range []string{"v8-migration-standing-team-bridge", " V8-Migration-Standing-Team-Bridge "} {
		_, err := ResolveStartupSelection(bootbShippedTemplateDirs[0], requested)
		bootbRequireRetiredError(t, err, "shipped dir, requested "+requested)
	}
	// An empty or missing templates dir still names the replacement bundles.
	_, err := ResolveStartupSelection(t.TempDir(), "v8-migration-standing-team-bridge")
	bootbRequireRetiredError(t, err, "empty dir")

	// Adversarial: an old ConfigMap still mounts a file declaring the retired ID.
	dir := t.TempDir()
	bootbWriteBundle(t, dir, "legacy.yaml", bootbFixtureBundle("v8-migration-standing-team-bridge", "", "admin-core"))
	_, err = ResolveStartupSelection(dir, "")
	bootbRequireRetiredError(t, err, "lone mounted legacy bundle")
	_, err = ResolveStartupSelection(dir, "v8-migration-standing-team-bridge")
	bootbRequireRetiredError(t, err, "mounted legacy bundle requested by id")
}

func TestBootBDefaultBundleWinsAmongSeveralWhenUnset(t *testing.T) {
	dir := t.TempDir()
	bootbWriteBundle(t, dir, "a-other.yaml", bootbFixtureBundle("a-other", "", "other-team"))
	bootbWriteBundle(t, dir, "core.yaml", bootbFixtureBundle(DefaultStartupBundleID, "", "core-team"))
	selection, err := ResolveStartupSelection(dir, "")
	if err != nil {
		t.Fatalf("selection failed: %v", err)
	}
	if selection.Bundle.ID != DefaultStartupBundleID || !slices.Equal(bootbTeamIDs(selection), []string{"core-team"}) {
		t.Fatalf("selected %q with teams %v, want the default bundle only", selection.Bundle.ID, bootbTeamIDs(selection))
	}
	explicit, err := ResolveStartupSelection(dir, "a-other")
	if err != nil || explicit.Bundle.ID != "a-other" {
		t.Fatalf("explicit selection = %+v, %v; an explicit id must win over the default", explicit, err)
	}
}

func TestBootBExtendsMergesBaseTeamsFirst(t *testing.T) {
	dir := t.TempDir()
	bootbWriteBundle(t, dir, "base.yaml", bootbFixtureBundle("base", "", "base-a", "base-b"))
	bootbWriteBundle(t, dir, "ext.yaml", bootbFixtureBundle("ext", "base", "ext-a"))
	selection, err := ResolveStartupSelection(dir, "ext")
	if err != nil {
		t.Fatalf("selection failed: %v", err)
	}
	if got := bootbTeamIDs(selection); !slices.Equal(got, []string{"base-a", "base-b", "ext-a"}) {
		t.Fatalf("extended teams = %v", got)
	}
	base, err := ResolveStartupSelection(dir, "base")
	if err != nil || !slices.Equal(bootbTeamIDs(base), []string{"base-a", "base-b"}) {
		t.Fatalf("base selection = %v, %v; extending must not change the base bundle", bootbTeamIDs(base), err)
	}
}

func TestBootBExtendsFailsClosed(t *testing.T) {
	cases := map[string][]string{
		"missing base": {bootbFixtureBundle("ext", "absent", "ext-a")},
		"chained":      {bootbFixtureBundle("root", "", "r"), bootbFixtureBundle("mid", "root", "m"), bootbFixtureBundle("leaf", "mid", "l")},
		"self":         {bootbFixtureBundle("self", "self", "s")},
		"duplicate":    {bootbFixtureBundle("base", "", "shared"), bootbFixtureBundle("ext", "base", "shared")},
	}
	for name, bodies := range cases {
		dir := t.TempDir()
		for idx, body := range bodies {
			bootbWriteBundle(t, dir, string(rune('a'+idx))+".yaml", body)
		}
		if bundles, err := NewTemplateLoader(dir).LoadBundles(); err == nil {
			t.Fatalf("%s: LoadBundles accepted an invalid extends graph: %d bundles", name, len(bundles))
		}
	}
}
