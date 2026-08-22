package server

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/swarm"
)

func TestClaimQAFixtureResourcesUsesPurgeFence(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	scopeID := "11111111-1111-1111-1111-111111111111"
	now := time.Now().UTC()
	mock.ExpectExec("SELECT pg_advisory_lock").WithArgs(scopeID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text").WithArgs(scopeID, qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows(qaFixtureScopeColumns).AddRow(
			scopeID, qaFixtureTenantID, "playwright", "journey-42", "open",
			now.Add(time.Hour), now, now,
		))
	mock.ExpectExec("INSERT INTO qa_fixture_resources").
		WithArgs(sqlmock.AnyArg(), scopeID, "run", "22222222-2222-2222-2222-222222222222").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec("SELECT pg_advisory_unlock").WithArgs(scopeID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.claimQAFixtureResources(t.Context(), scopeID, []qaFixtureResource{{
		Kind: "run",
		Ref:  "22222222-2222-2222-2222-222222222222",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestClaimConfirmedCreatedTeamLockedClaimsActualResult(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	scopeID := "11111111-1111-1111-1111-111111111111"
	now := time.Now().UTC()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text").WithArgs(scopeID, qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows(qaFixtureScopeColumns).AddRow(
			scopeID, qaFixtureTenantID, "playwright", "journey-42", "open",
			now.Add(time.Hour), now, now,
		))
	mock.ExpectExec("INSERT INTO qa_fixture_resources").
		WithArgs(sqlmock.AnyArg(), scopeID, "team", "requested-team").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id::text").WithArgs(scopeID, qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows(qaFixtureScopeColumns).AddRow(
			scopeID, qaFixtureTenantID, "playwright", "journey-42", "open",
			now.Add(time.Hour), now, now,
		))
	mock.ExpectExec("INSERT INTO qa_fixture_resources").
		WithArgs(sqlmock.AnyArg(), scopeID, "workspace_path", "groups/requested-team").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := s.claimConfirmedCreatedTeamLocked(
		t.Context(), scopeID,
		map[string]any{"team_id": "requested-team"},
		`{"status":"created","team_id":"requested-team","workspace_folder":"groups/requested-team"}`,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfirmedActionCreatedTeamResultsNormalizesTeamIdentity(t *testing.T) {
	created := confirmedActionCreatedTeamResults([]plannedToolExecutionResult{{
		Name:      "create_team",
		Arguments: map[string]any{"team_id": "Delivery_Team"},
		Output:    `{"status":"created","team_id":"delivery-team","workspace_folder":"groups/delivery-team"}`,
	}})
	if !created["delivery-team"] || len(created) != 1 {
		t.Fatalf("expected one normalized created team, got %#v", created)
	}
}

func TestClaimConfirmedCreatedTeamRejectsReturnedIdentityMismatch(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	actualFolder := filepath.Join(workspace, "groups", "existing-team")
	if err := os.MkdirAll(actualFolder, 0o755); err != nil {
		t.Fatal(err)
	}

	err := s.claimConfirmedCreatedTeam(
		t.Context(), "11111111-1111-1111-1111-111111111111",
		map[string]any{"team_id": "requested-team"},
		`{"status":"created","team_id":"existing-team","workspace_folder":"groups/existing-team"}`,
	)
	if !errors.Is(err, errQAFixtureTeamIdentityMismatch) {
		t.Fatalf("expected team identity mismatch, got %v", err)
	}
	if _, statErr := os.Stat(actualFolder); statErr != nil {
		t.Fatalf("mismatched team workspace must not be removed: %v", statErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestClaimConfirmedCreatedTeamLockedIgnoresExistingTeam(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	err := s.claimConfirmedCreatedTeamLocked(
		t.Context(), "11111111-1111-1111-1111-111111111111",
		map[string]any{"team_id": "existing-team"},
		`{"status":"already_exists","team_id":"existing-team"}`,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureQAFixtureTeamCreationAvailableRejectsDurableTeam(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	scopeID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectQuery("SELECT EXISTS").WithArgs(scopeID, "existing-team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WithArgs(qaFixtureTenantID, "existing-team", "22222222-2222-2222-2222-222222222222").
		WillReturnRows(sqlmock.NewRows([]string{"registry", "work", "group", "runtime"}).AddRow(true, false, false, false))

	err := s.ensureQAFixtureTeamCreationAvailable(t.Context(), scopeID, "22222222-2222-2222-2222-222222222222", map[string]any{"team_id": "existing-team"})
	if !errors.Is(err, errQAFixtureTeamPreexisting) {
		t.Fatalf("expected pre-existing team rejection, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureQAFixtureTeamCreationAvailableRejectsPersistedRuntimeManifest(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	scopeID := "11111111-1111-1111-1111-111111111111"
	runID := "22222222-2222-2222-2222-222222222222"
	mock.ExpectQuery("SELECT EXISTS").WithArgs(scopeID, "persisted-team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WithArgs(qaFixtureTenantID, "persisted-team", runID).
		WillReturnRows(sqlmock.NewRows([]string{"registry", "work", "group", "runtime"}).AddRow(false, false, false, true))

	err := s.ensureQAFixtureTeamCreationAvailable(t.Context(), scopeID, runID, map[string]any{"team_id": "persisted-team"})
	if !errors.Is(err, errQAFixtureTeamPreexisting) {
		t.Fatalf("expected persisted runtime team rejection, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureQAFixtureTeamCreationAvailableRejectsExistingWorkspace(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	scopeID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectQuery("SELECT EXISTS").WithArgs(scopeID, "existing-team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WithArgs(qaFixtureTenantID, "existing-team", "22222222-2222-2222-2222-222222222222").
		WillReturnRows(sqlmock.NewRows([]string{"registry", "work", "group", "runtime"}).AddRow(false, false, false, false))
	if err := os.MkdirAll(filepath.Join(workspace, "groups", "existing-team"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := s.ensureQAFixtureTeamCreationAvailable(t.Context(), scopeID, "22222222-2222-2222-2222-222222222222", map[string]any{"team_id": "existing-team"})
	if !errors.Is(err, errQAFixtureTeamPreexisting) {
		t.Fatalf("expected existing workspace rejection, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureQAFixtureTeamCreationAvailableAllowsNewTeam(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	scopeID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectQuery("SELECT EXISTS").WithArgs(scopeID, "new-team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WithArgs(qaFixtureTenantID, "new-team", "22222222-2222-2222-2222-222222222222").
		WillReturnRows(sqlmock.NewRows([]string{"registry", "work", "group", "runtime"}).AddRow(false, false, false, false))

	if err := s.ensureQAFixtureTeamCreationAvailable(t.Context(), scopeID, "22222222-2222-2222-2222-222222222222", map[string]any{"team_id": "new-team"}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureQAFixtureTeamCreationAvailableAllowsOwnedRetry(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	scopeID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectQuery("SELECT EXISTS").WithArgs(scopeID, "owned-team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	if err := s.ensureQAFixtureTeamCreationAvailable(t.Context(), scopeID, "22222222-2222-2222-2222-222222222222", map[string]any{"team_id": "owned-team"}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureQAFixtureTeamCreationAvailableRejectsActiveRuntimeTeam(t *testing.T) {
	wireNATS := withNATS(t)
	withDatabase, mock := withDB(t)
	s := newTestServer(wireNATS, withDatabase, func(s *AdminServer) {
		s.Soma = swarm.NewSoma(s.NC, nil, nil, nil, nil, nil, nil)
		t.Cleanup(s.Soma.Shutdown)
	})
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	if err := s.Soma.SpawnTeam(&swarm.TeamManifest{ID: "active-team", Name: "Active Team"}); err != nil {
		t.Fatal(err)
	}
	scopeID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectQuery("SELECT EXISTS").WithArgs(scopeID, "active-team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	err := s.ensureQAFixtureTeamCreationAvailable(t.Context(), scopeID, "22222222-2222-2222-2222-222222222222", map[string]any{"team_id": "active-team"})
	if !errors.Is(err, errQAFixtureTeamPreexisting) {
		t.Fatalf("expected active team rejection, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestClaimQAFixtureResourcesReusesHeldFence(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	scopeID := "11111111-1111-1111-1111-111111111111"
	now := time.Now().UTC()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text").WithArgs(scopeID, qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows(qaFixtureScopeColumns).AddRow(
			scopeID, qaFixtureTenantID, "playwright", "journey-42", "open",
			now.Add(time.Hour), now, now,
		))
	mock.ExpectExec("INSERT INTO qa_fixture_resources").
		WithArgs(sqlmock.AnyArg(), scopeID, "run", "22222222-2222-2222-2222-222222222222").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx := withQAFixtureFenceHeld(t.Context(), scopeID)
	err := s.claimQAFixtureResources(ctx, scopeID, []qaFixtureResource{{
		Kind: "run",
		Ref:  "22222222-2222-2222-2222-222222222222",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestClaimConfirmedCreatedTeamCleansWorkspaceWhenClaimFails(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	teamFolder := filepath.Join(workspace, "groups", "new-team")
	if err := os.MkdirAll(teamFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	scopeID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text").WithArgs(scopeID, qaFixtureTenantID).
		WillReturnError(errQAFixtureScopeClosed)
	mock.ExpectRollback()

	err := s.claimConfirmedCreatedTeam(
		t.Context(), scopeID,
		map[string]any{"team_id": "new-team"},
		`{"status":"created","team_id":"new-team","workspace_folder":"groups/new-team"}`,
	)
	if !errors.Is(err, errQAFixtureScopeClosed) {
		t.Fatalf("expected fixture claim failure, got %v", err)
	}
	if _, statErr := os.Stat(teamFolder); !os.IsNotExist(statErr) {
		t.Fatalf("expected unclaimed team workspace removed, stat error %v", statErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
