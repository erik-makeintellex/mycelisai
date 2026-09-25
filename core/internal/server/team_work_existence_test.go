package server

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/swarm"
)

const teamWorkBody = `{"execution_shape":"deliverable","state":"queued","objective":"Ship the approved output"}`

func postTeamWork(t *testing.T, s *AdminServer, teamID string) int {
	t.Helper()
	mux := setupMux(t, "POST /api/v1/teams/{id}/work", s.HandleCreateTeamWork)
	return doRequest(t, mux, http.MethodPost, "/api/v1/teams/"+teamID+"/work", teamWorkBody).Code
}

func expectTeamWorkInsert(mock sqlmock.Sqlmock) {
	now := time.Now().UTC()
	mock.ExpectQuery("INSERT INTO team_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))
}

func TestHandleCreateTeamWork_RejectsUnknownTeam(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt, func(s *AdminServer) { s.Soma = swarm.NewTestSoma(nil) })
	mock.ExpectQuery("SELECT EXISTS \\(SELECT 1 FROM runtime_team_manifests").
		WithArgs("ghost-team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	if got := postTeamWork(t, s, "ghost-team"); got != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleCreateTeamWork_AcceptsRunningTeamWithoutDurableLookup(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt, func(s *AdminServer) {
		s.Soma = swarm.NewTestSoma([]*swarm.TeamManifest{{ID: "live-team", Name: "Live Team"}})
	})
	expectTeamWorkInsert(mock)

	if got := postTeamWork(t, s, "live-team"); got != http.StatusCreated {
		t.Fatalf("status = %d, want 201", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleCreateTeamWork_AcceptsDurablyKnownTeam(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	mock.ExpectQuery("SELECT EXISTS \\(SELECT 1 FROM runtime_team_manifests").
		WithArgs("durable-team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	expectTeamWorkInsert(mock)

	if got := postTeamWork(t, s, "durable-team"); got != http.StatusCreated {
		t.Fatalf("status = %d, want 201", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleCreateTeamWork_FailsClosedWhenTeamLookupUnavailable(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	mock.ExpectQuery("SELECT EXISTS \\(SELECT 1 FROM runtime_team_manifests").
		WithArgs("any-team").
		WillReturnError(errors.New("connection reset"))

	if got := postTeamWork(t, s, "any-team"); got != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", got)
	}
	if got := postTeamWork(t, newTestServer(), "any-team"); got != http.StatusServiceUnavailable {
		t.Fatalf("no-DB status = %d, want 503", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
