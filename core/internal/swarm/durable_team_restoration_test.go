package swarm

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

func TestPostgresDurableTeamLoaderBuildsRuntimeManifest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery("WITH restorable AS").
		WithArgs(
			string(protocol.TeamWorkStateNew),
			string(protocol.TeamWorkStateBriefed),
			string(protocol.TeamWorkStateQueued),
			string(protocol.TeamWorkStateRunning),
			string(protocol.TeamWorkStateNeedsOperator),
			string(protocol.TeamWorkStateReviewing),
			string(protocol.TeamWorkStateDegraded),
			string(protocol.TeamWorkStatePaused),
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"team_id", "name", "purpose", "allowed_capabilities",
			"capability_requirements", "coordinator_profile",
		}).AddRow(
			"delivery-team", "Delivery Team", "Finish the retained package.",
			[]byte(`["write_file","store_artifact"]`),
			[]byte(`["store_artifact","team_orchestration"]`), "delivery lead",
		))

	manifests, err := NewPostgresDurableTeamLoader(db).LoadRuntimeTeams(context.Background())
	if err != nil {
		t.Fatalf("LoadRuntimeTeams: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("manifest count = %d, want 1", len(manifests))
	}
	manifest := manifests[0]
	if manifest.ID != "delivery-team" || manifest.Name != "Delivery Team" {
		t.Fatalf("restored identity = %#v", manifest)
	}
	if manifest.Description != "Finish the retained package." {
		t.Fatalf("description = %q", manifest.Description)
	}
	if len(manifest.Members) != 1 || manifest.Members[0].Role != "delivery lead" {
		t.Fatalf("restored members = %#v", manifest.Members)
	}
	if got := manifest.Members[0].Tools; len(got) != 3 || got[0] != "write_file" || got[2] != "team_orchestration" {
		t.Fatalf("restored tools = %#v", got)
	}
	if len(manifest.Inputs) != 1 || manifest.Inputs[0] != "swarm.team.delivery-team.internal.command" {
		t.Fatalf("restored inputs = %#v", manifest.Inputs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
}

func TestSomaStartRestoresDurableTeamsWithStandingPrecedence(t *testing.T) {
	_, nc := startTestNATS(t)
	standing := &TeamManifest{
		ID:   "standing-team",
		Name: "Standing Authority",
		Type: TeamTypeAction,
	}
	restoredDuplicate := &TeamManifest{ID: "standing-team", Name: "Stale Durable Copy", Type: TeamTypeAction}
	restoredDynamic := buildRuntimeTeamManifest(map[string]any{
		"team_id": "dynamic-team",
		"name":    "Dynamic Team",
	})
	soma := NewSoma(nc, &governance.Guard{}, NewRegistryFromManifests([]*TeamManifest{standing}), nil, nil, nil, nil)
	soma.SetDurableTeamLoader(staticDurableTeamLoader{manifests: []*TeamManifest{restoredDuplicate, restoredDynamic}})
	if err := soma.Start(); err != nil {
		t.Fatalf("Soma.Start: %v", err)
	}
	t.Cleanup(soma.Shutdown)

	teams := soma.ListTeams()
	if len(teams) != 2 {
		t.Fatalf("active teams = %#v, want standing and restored dynamic", teams)
	}
	byID := map[string]*TeamManifest{}
	for _, manifest := range teams {
		byID[manifest.ID] = manifest
	}
	if byID["standing-team"].Name != "Standing Authority" {
		t.Fatalf("standing manifest lost precedence: %#v", byID["standing-team"])
	}
	if byID["dynamic-team"] == nil || len(byID["dynamic-team"].Members) != 1 {
		t.Fatalf("dynamic team was not restored: %#v", byID["dynamic-team"])
	}
}

type staticDurableTeamLoader struct {
	manifests []*TeamManifest
	err       error
}

func (l staticDurableTeamLoader) LoadRuntimeTeams(context.Context) ([]*TeamManifest, error) {
	return l.manifests, l.err
}
