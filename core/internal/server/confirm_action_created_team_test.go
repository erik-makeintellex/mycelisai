package server

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestEnsureQAFixtureTeamCreationAvailableAllowsCurrentRunStagedOwnership(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	scopeID := "11111111-1111-1111-1111-111111111111"
	runID := "22222222-2222-2222-2222-222222222222"
	mock.ExpectQuery("SELECT EXISTS").WithArgs(scopeID, "new-team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("LEFT JOIN outcome_projects").WithArgs(qaFixtureTenantID, "new-team", runID).
		WillReturnRows(sqlmock.NewRows([]string{"registry", "work", "group"}).AddRow(false, false, false))

	if err := s.ensureQAFixtureTeamCreationAvailable(t.Context(), scopeID, runID, map[string]any{"team_id": "new-team"}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureQAFixtureTeamCreationAvailableRejectsPriorRunVisibility(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	scopeID := "11111111-1111-1111-1111-111111111111"
	runID := "22222222-2222-2222-2222-222222222222"
	mock.ExpectQuery("SELECT EXISTS").WithArgs(scopeID, "existing-team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WithArgs(qaFixtureTenantID, "existing-team", runID).
		WillReturnRows(sqlmock.NewRows([]string{"registry", "work", "group"}).AddRow(false, true, false))

	err := s.ensureQAFixtureTeamCreationAvailable(t.Context(), scopeID, runID, map[string]any{"team_id": "existing-team"})
	if !errors.Is(err, errQAFixtureTeamPreexisting) {
		t.Fatalf("expected prior-run team visibility rejection, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
