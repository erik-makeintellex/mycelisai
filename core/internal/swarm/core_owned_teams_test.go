package swarm

import (
	"context"
	"errors"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

// S7b follow-up: workspace-wide read_file rests on a Core-owned flag that only
// the standing boot registry sets, and no runtime path may register a
// reserved Core team ID, whether or not the real team is loaded.

func TestCreateTeamRefusesReservedIDWhileCoreTeamUnloaded(t *testing.T) {
	readScopeWorkspace(t)
	soma := NewTestSoma(nil) // no Core team loaded
	reg := NewInternalToolRegistry(InternalToolDeps{})
	reg.SetSoma(soma)
	for _, id := range []string{"admin-core", "council-core", "genesis-core", "telemetry-core", " Admin-Core "} {
		out, err := reg.tools["create_team"].Handler(context.Background(), map[string]any{"team_id": id, "name": "Impostor"})
		if err == nil || !errors.Is(err, ErrReservedTeamID) {
			t.Errorf("create_team %q = (%q, %v), want ErrReservedTeamID", id, out, err)
		}
	}
	if teams := soma.ListTeams(); len(teams) != 0 {
		t.Fatalf("reserved create_team registered %d teams", len(teams))
	}
	// With the real team unloaded, an agent claiming its ID stays confined.
	impostor := &ToolInvocationContext{TeamID: "admin-core", AgentID: "admin", AgentRole: "admin"}
	if out, err := readAs(reg, impostor, "groups/team-b/secret.md"); err == nil {
		t.Fatalf("unloaded admin-core read = %q, want confinement", out)
	}
}

func TestSpawnTeamRefusesReservedID(t *testing.T) {
	soma := NewTestSoma(nil)
	for _, id := range CoreOwnedTeamIDs {
		err := soma.SpawnTeamContext(context.Background(), &TeamManifest{ID: id, Members: []protocol.AgentManifest{{ID: "x"}}})
		if !errors.Is(err, ErrReservedTeamID) || !errors.Is(err, ErrRuntimeTeamInvalid) {
			t.Errorf("SpawnTeamContext(%q) = %v, want reserved refusal", id, err)
		}
	}
}

func TestDurableRestoreDropsReservedTeamIDs(t *testing.T) {
	soma := NewTestSoma(nil)
	soma.durableTeamLoader = staticDurableTeamLoader{manifests: []*TeamManifest{{ID: "council-core"}, {ID: "dynamic-team"}}}
	merged := soma.mergeDurableTeamManifests(nil)
	if len(merged) != 1 || merged[0].ID != "dynamic-team" {
		t.Fatalf("merged = %v, want only dynamic-team", merged)
	}
}

func TestReservedIDWithoutCoreOwnedFlagStaysConfined(t *testing.T) {
	readScopeWorkspace(t)
	soma := NewTestSoma(nil)
	soma.teams["admin-core"] = &Team{Manifest: &TeamManifest{ID: "admin-core"}} // registered, not by the boot registry
	reg := NewInternalToolRegistry(InternalToolDeps{})
	reg.SetSoma(soma)
	inv := &ToolInvocationContext{TeamID: "admin-core", AgentID: "admin", AgentRole: "admin"}
	if out, err := readAs(reg, inv, "groups/team-b/secret.md"); err == nil {
		t.Fatalf("non-Core-owned admin-core read = %q, want confinement", out)
	}
	soma.teams["admin-core"].coreOwned = true
	if out, err := readAs(reg, inv, "groups/team-b/secret.md"); err != nil || out != "secret-b" {
		t.Fatalf("Core-owned admin-core read = (%q, %v), want workspace-wide", out, err)
	}
}

func TestStandingBootTeamIsTheOnlyCoreOwnedSource(t *testing.T) {
	soma := newTestSomaForActivation(t)
	soma.startBootTeam(&TeamManifest{ID: "admin-core"}, nil)
	soma.startBootTeam(&TeamManifest{ID: "team-a"}, nil)
	t.Cleanup(func() { soma.StopTeam("admin-core"); soma.StopTeam("team-a") })
	if !soma.IsCoreOwnedTeam("admin-core") {
		t.Fatal("standing admin-core was not marked Core-owned")
	}
	if soma.IsCoreOwnedTeam("team-a") || soma.IsCoreOwnedTeam("council-core") || (*Soma)(nil).IsCoreOwnedTeam("admin-core") {
		t.Fatal("a non-standing or unloaded team was treated as Core-owned")
	}
}
