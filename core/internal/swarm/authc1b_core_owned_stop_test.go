package swarm

import (
	"errors"
	"testing"
)

// AUTH-C1b: StopTeamDurably refuses a Core-owned (boot-registered) team for
// every caller, so no route or purge path can stop or durably delete it.
func TestAuthC1bStopTeamDurablyRefusesCoreOwnedTeam(t *testing.T) {
	soma := NewTestSoma([]*TeamManifest{{ID: "admin-core"}, {ID: "user-team"}})
	soma.teams["admin-core"].coreOwned = true // as the standing boot registry marks it
	found, err := soma.StopTeamDurably("admin-core")
	if !found || !errors.Is(err, ErrCoreOwnedTeam) {
		t.Fatalf("StopTeamDurably(admin-core) = (%v, %v), want (true, ErrCoreOwnedTeam)", found, err)
	}
	if soma.StopTeam(" admin-core ") || !soma.HasTeam("admin-core") {
		t.Fatal("a Core-owned team was stopped")
	}
	if found, err := soma.StopTeamDurably("user-team"); !found || err != nil || soma.HasTeam("user-team") {
		t.Fatalf("StopTeamDurably(user-team) = (%v, %v), want the runtime team stopped", found, err)
	}
}
