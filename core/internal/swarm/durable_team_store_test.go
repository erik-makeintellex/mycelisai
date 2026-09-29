package swarm

import (
	"context"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

func TestPostgresDurableTeamLoaderRestoresExactManifest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	want := completeDurableManifest()
	raw, digest := encodedManifest(t, want)
	mock.ExpectQuery("SELECT team_id, schema_version, manifest_digest, manifest").
		WithArgs(durableRuntimeTenant).
		WillReturnRows(sqlmock.NewRows([]string{"team_id", "schema_version", "manifest_digest", "manifest"}).
			AddRow(want.ID, "v1", digest, raw))
	mock.ExpectQuery("WITH restorable AS").
		WithArgs(legacyRuntimeTeamArgs()...).
		WillReturnRows(sqlmock.NewRows([]string{
			"team_id", "name", "purpose", "allowed_capabilities", "capability_requirements", "coordinator_profile",
		}))

	got, err := NewPostgresDurableTeamLoader(db).LoadRuntimeTeams(context.Background())
	if err != nil {
		t.Fatalf("LoadRuntimeTeams: %v", err)
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0], want) {
		t.Fatalf("restored manifest mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresDurableTeamLoaderQuarantinesCorruptRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	good := completeDurableManifest()
	raw, digest := encodedManifest(t, good)
	other := manifestRevision("other-team", "1.0.0", "sha256:profile-a")
	otherRaw, otherDigest := encodedManifest(t, other)
	mock.ExpectQuery("SELECT team_id, schema_version, manifest_digest, manifest").
		WithArgs(durableRuntimeTenant).
		WillReturnRows(sqlmock.NewRows([]string{"team_id", "schema_version", "manifest_digest", "manifest"}).
			AddRow(good.ID, "v1", digest, raw).
			AddRow("tampered-team", "v1", "sha256:tampered", raw).
			AddRow(other.ID, "v1", otherDigest, otherRaw))
	mock.ExpectQuery("WITH restorable AS").
		WithArgs(legacyRuntimeTeamArgs()...).
		WillReturnRows(sqlmock.NewRows([]string{
			"team_id", "name", "purpose", "allowed_capabilities", "capability_requirements", "coordinator_profile",
		}).AddRow("legacy-bad", "Legacy", "Broken tools", []byte(`{not-json`), []byte(`[]`), ""))

	got, degraded, err := NewPostgresDurableTeamLoader(db).LoadRuntimeTeamsWithReport(context.Background())
	if err != nil {
		t.Fatalf("LoadRuntimeTeamsWithReport: %v", err)
	}
	if len(got) != 2 || got[0].ID != good.ID || got[1].ID != other.ID {
		t.Fatalf("restored = %#v, want the two valid rows", got)
	}
	if len(degraded) != 2 || degraded[0].TeamID != "tampered-team" || !strings.Contains(degraded[0].Reason, "digest validation") ||
		degraded[1].TeamID != "legacy-bad" {
		t.Fatalf("degradations = %#v", degraded)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSomaRecordsLoaderRestorationDegradations(t *testing.T) {
	_, nc := startTestNATS(t)
	soma := NewSoma(nc, &governance.Guard{}, NewRegistryFromRuntimeOrganization(&RuntimeOrganization{}), nil, nil, nil, nil)
	soma.SetDurableTeamLoader(reportingLoader{
		manifests: []*TeamManifest{manifestRevision("kept-team", "1.0.0", "sha256:profile-a")},
		degraded:  []DurableTeamRestoreDegradation{{TeamID: "bad-team", Reason: "unsupported schema"}},
	})
	if err := soma.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(soma.Shutdown)
	if ids := soma.ListTeams(); len(ids) != 1 || ids[0].ID != "kept-team" {
		t.Fatalf("restored = %#v", ids)
	}
	if got := soma.RestorationDegradations(); len(got) != 1 || got[0].TeamID != "bad-team" {
		t.Fatalf("degradations = %#v", got)
	}
}

type reportingLoader struct {
	manifests []*TeamManifest
	degraded  []DurableTeamRestoreDegradation
}

func (l reportingLoader) LoadRuntimeTeams(context.Context) ([]*TeamManifest, error) {
	return l.manifests, nil
}

func (l reportingLoader) LoadRuntimeTeamsWithReport(context.Context) ([]*TeamManifest, []DurableTeamRestoreDegradation, error) {
	return l.manifests, l.degraded, nil
}

func TestPostgresDurableTeamStoreIsIdempotentAndRejectsReplacement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	store := NewPostgresDurableTeamLoader(db)
	manifest := completeDurableManifest()
	_, digest := encodedManifest(t, manifest)

	mock.ExpectExec("INSERT INTO runtime_team_manifests").
		WithArgs(durableRuntimeTenant, manifest.ID, digest, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.SaveRuntimeTeam(context.Background(), manifest); err != nil {
		t.Fatalf("first save: %v", err)
	}

	mock.ExpectExec("INSERT INTO runtime_team_manifests").
		WithArgs(durableRuntimeTenant, manifest.ID, digest, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT manifest_digest FROM runtime_team_manifests").
		WithArgs(durableRuntimeTenant, manifest.ID).
		WillReturnRows(sqlmock.NewRows([]string{"manifest_digest"}).AddRow(digest))
	if err := store.SaveRuntimeTeam(context.Background(), manifest); err != nil {
		t.Fatalf("idempotent save: %v", err)
	}

	changed := *manifest
	changed.Description = "A later activation must not replace the approved manifest."
	_, changedDigest := encodedManifest(t, &changed)
	mock.ExpectExec("INSERT INTO runtime_team_manifests").
		WithArgs(durableRuntimeTenant, manifest.ID, changedDigest, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT manifest_digest FROM runtime_team_manifests").
		WithArgs(durableRuntimeTenant, manifest.ID).
		WillReturnRows(sqlmock.NewRows([]string{"manifest_digest"}).AddRow(digest))
	if err := store.SaveRuntimeTeam(context.Background(), &changed); err == nil {
		t.Fatal("expected conflicting approved manifest to fail")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSomaSpawnAndStopOwnDurableManifest(t *testing.T) {
	_, nc := startTestNATS(t)
	store := &memoryDurableTeamStore{}
	soma := NewSoma(nc, &governance.Guard{}, NewRegistryFromRuntimeOrganization(&RuntimeOrganization{}), nil, nil, nil, nil)
	soma.SetDurableTeamStore(store)
	manifest := completeDurableManifest()
	if err := soma.SpawnTeam(manifest); err != nil {
		t.Fatalf("SpawnTeam: %v", err)
	}
	if len(store.saved) != 1 || !reflect.DeepEqual(store.saved[0], manifest) {
		t.Fatalf("saved = %#v", store.saved)
	}
	found, err := soma.StopTeamDurably(manifest.ID)
	if err != nil || !found {
		t.Fatalf("StopTeamDurably = (%v, %v)", found, err)
	}
	if !reflect.DeepEqual(store.deleted, []string{manifest.ID}) {
		t.Fatalf("deleted = %#v", store.deleted)
	}
}

func TestDeactivateMissionDeletesDurableManifests(t *testing.T) {
	_, nc := startTestNATS(t)
	store := &memoryDurableTeamStore{}
	soma := NewSoma(nc, &governance.Guard{}, NewRegistryFromRuntimeOrganization(&RuntimeOrganization{}), nil, nil, nil, nil)
	soma.SetDurableTeamStore(store)
	for _, teamID := range []string{"mission-7.builder", "mission-7.reviewer", "other.builder"} {
		if err := soma.SpawnTeam(manifestRevision(teamID, "1.0.0", "sha256:profile-a")); err != nil {
			t.Fatalf("SpawnTeam(%s): %v", teamID, err)
		}
	}

	if got := soma.DeactivateMission("mission-7"); got != 2 {
		t.Fatalf("DeactivateMission stopped %d teams, want 2", got)
	}
	if got := len(store.saved); got != 1 || store.saved[0].ID != "other.builder" {
		t.Fatalf("persisted teams after deactivation = %#v", store.saved)
	}
}

func TestDurableProfilesRemainProspectiveAcrossActivationAndRollback(t *testing.T) {
	_, nc := startTestNATS(t)
	store := &memoryDurableTeamStore{}
	first := NewSoma(nc, &governance.Guard{}, NewRegistryFromRuntimeOrganization(&RuntimeOrganization{}), nil, nil, nil, nil)
	first.SetDurableTeamStore(store)

	teamA := manifestRevision("team-before-b", "1.0.0", "sha256:profile-a")
	teamB := manifestRevision("team-after-b", "2.0.0", "sha256:profile-b")
	teamAfterRollback := manifestRevision("team-after-rollback", "1.0.0", "sha256:profile-a")
	for _, manifest := range []*TeamManifest{teamA, teamB, teamAfterRollback} {
		if err := first.SpawnTeam(manifest); err != nil {
			t.Fatalf("SpawnTeam(%s): %v", manifest.ID, err)
		}
	}
	first.Shutdown()

	restarted := NewSoma(nc, &governance.Guard{}, NewRegistryFromRuntimeOrganization(&RuntimeOrganization{}), nil, nil, nil, nil)
	restarted.SetDurableTeamStore(store)
	if err := restarted.Start(); err != nil {
		t.Fatalf("restart Soma: %v", err)
	}
	t.Cleanup(restarted.Shutdown)

	digests := map[string]string{}
	for _, manifest := range restarted.ListTeams() {
		digests[manifest.ID] = manifest.Members[0].Profile.Digest
	}
	want := map[string]string{
		"team-before-b":       "sha256:profile-a",
		"team-after-b":        "sha256:profile-b",
		"team-after-rollback": "sha256:profile-a",
	}
	if !reflect.DeepEqual(digests, want) {
		t.Fatalf("restored profile lineage = %#v, want %#v", digests, want)
	}
}

type memoryDurableTeamStore struct {
	saved   []*TeamManifest
	deleted []string
}

func (s *memoryDurableTeamStore) LoadRuntimeTeams(context.Context) ([]*TeamManifest, error) {
	return append([]*TeamManifest(nil), s.saved...), nil
}

func (s *memoryDurableTeamStore) SaveRuntimeTeam(_ context.Context, manifest *TeamManifest) error {
	raw, _ := json.Marshal(manifest)
	var clone TeamManifest
	_ = json.Unmarshal(raw, &clone)
	s.saved = append(s.saved, &clone)
	return nil
}

func (s *memoryDurableTeamStore) DeleteRuntimeTeam(_ context.Context, teamID string) error {
	s.deleted = append(s.deleted, teamID)
	kept := s.saved[:0]
	for _, manifest := range s.saved {
		if manifest.ID != teamID {
			kept = append(kept, manifest)
		}
	}
	s.saved = kept
	return nil
}

func encodedManifest(t *testing.T, manifest *TeamManifest) ([]byte, string) {
	t.Helper()
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return raw, fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
}

func legacyRuntimeTeamArgs() []driver.Value {
	return []driver.Value{
		durableRuntimeTenant,
		string(protocol.TeamWorkStateNew), string(protocol.TeamWorkStateBriefed),
		string(protocol.TeamWorkStateQueued), string(protocol.TeamWorkStateRunning),
		string(protocol.TeamWorkStateNeedsOperator), string(protocol.TeamWorkStateReviewing),
		string(protocol.TeamWorkStateDegraded), string(protocol.TeamWorkStatePaused),
	}
}

func completeDurableManifest() *TeamManifest {
	return &TeamManifest{
		ID: "delivery-team", Name: "Delivery Team", Type: TeamTypeAction,
		Description: "Own the approved delivery.", Provider: "local-primary",
		AskRouting: map[string]string{"implementation": "implementer"},
		Members: []protocol.AgentManifest{{
			ID: "delivery-lead", ProfileRef: "delivery.builder",
			Profile: &protocol.WorkerProfileSnapshot{
				ID: "delivery.builder", Version: "2.0.0", Digest: "sha256:profile-a", RecordID: "record-a",
				TenantID: "default", Scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "workspace-7"},
			},
			Role: "builder", SystemPrompt: "Build and verify the approved output.", Model: "model-a", Provider: "provider-a",
			Inputs: []string{"work_brief"}, Outputs: []string{"deliverable"}, Tools: []string{"write_file", "store_artifact"},
			Context: []protocol.AgentContextBinding{{Kind: "approved_mount", Ref: "source-a", Access: "read"}},
			Usage:   protocol.AgentUsagePolicy{Selection: "soma_or_manual", Scope: "workspace"}, MaxIterations: 8,
			Verification: &protocol.Verification{Strategy: protocol.VerifyEmpirical, Rubric: []string{"Output opens"}, ValidationCommand: "go test ./..."},
		}},
		Inputs:     []string{"swarm.team.delivery-team.internal.command"},
		Deliveries: []string{"swarm.team.delivery-team.signal.result"},
		Schedule:   &protocol.ScheduleConfig{Type: "interval", Interval: "1h"},
	}
}

func manifestRevision(teamID, version, digest string) *TeamManifest {
	manifest := completeDurableManifest()
	manifest.ID = teamID
	manifest.Name = teamID
	manifest.Members[0].ID = teamID + "-lead"
	manifest.Members[0].Profile.Version = version
	manifest.Members[0].Profile.Digest = digest
	manifest.Inputs = []string{"swarm.team." + teamID + ".internal.command"}
	manifest.Deliveries = []string{"swarm.team." + teamID + ".signal.result"}
	return manifest
}
