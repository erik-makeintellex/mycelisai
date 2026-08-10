package server

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestValidateQAFixtureResourceRejectsUnprovenDurableObjects(t *testing.T) {
	s := newTestServer()
	scope := qaFixtureScope{ID: "11111111-1111-1111-1111-111111111111"}
	for _, resource := range []qaFixtureResource{
		{Kind: "group", Ref: "22222222-2222-2222-2222-222222222222"},
		{Kind: "run", Ref: "33333333-3333-3333-3333-333333333333"},
		{Kind: "outcome", Ref: "44444444-4444-4444-4444-444444444444"},
	} {
		err := s.validateQAFixtureResource(
			t.Context(), nil, scope, resource, map[string][]string{}, nil, nil,
		)
		if !errors.Is(err, errQAFixtureResourceUnowned) {
			t.Errorf("%s validation error = %v, want unowned provenance", resource.Kind, err)
		}
	}
}

func TestAddQAFixtureResourceRejectsUncorrelatedTeam(t *testing.T) {
	t.Setenv("MYCELIS_QA_FIXTURE_MANAGEMENT", "true")
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
	mock.ExpectQuery("SELECT resource_kind, resource_ref").WithArgs(scopeID).
		WillReturnRows(sqlmock.NewRows([]string{"resource_kind", "resource_ref"}))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("team", "unrelated-team", scopeID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT COALESCE\\(run_id::text").
		WithArgs("unrelated-team", qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"run_id"}))
	mock.ExpectRollback()

	mux := setupMux(t, "POST /api/v1/testing/fixture-scopes/{id}/resources", s.HandleAddQAFixtureResources)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost,
		"/api/v1/testing/fixture-scopes/"+scopeID+"/resources", `{
			"owner_ref":"playwright",
			"execution_ref":"journey-42",
			"resources":[{"kind":"team","ref":"unrelated-team"}]
		}`)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected uncorrelated team rejection, got %d: %s", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAddQAFixtureResourceRejectsUnprovenRun(t *testing.T) {
	t.Setenv("MYCELIS_QA_FIXTURE_MANAGEMENT", "true")
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	scopeID := "11111111-1111-1111-1111-111111111111"
	runID := "22222222-2222-2222-2222-222222222222"
	now := time.Now().UTC()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text").WithArgs(scopeID, qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows(qaFixtureScopeColumns).AddRow(
			scopeID, qaFixtureTenantID, "playwright", "journey-42", "open",
			now.Add(time.Hour), now, now,
		))
	mock.ExpectQuery("SELECT resource_kind, resource_ref").WithArgs(scopeID).
		WillReturnRows(sqlmock.NewRows([]string{"resource_kind", "resource_ref"}))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("run", runID, scopeID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectRollback()

	mux := setupMux(t, "POST /api/v1/testing/fixture-scopes/{id}/resources", s.HandleAddQAFixtureResources)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost,
		"/api/v1/testing/fixture-scopes/"+scopeID+"/resources", `{
			"owner_ref":"playwright",
			"execution_ref":"journey-42",
			"resources":[{"kind":"run","ref":"22222222-2222-2222-2222-222222222222"}]
		}`)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected exact-provenance rejection, got %d: %s", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateQAFixtureArtifactRejectsTeamOnlyCorrelation(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	artifactID := "55555555-5555-5555-5555-555555555555"
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COALESCE\\(trace_id").WithArgs(artifactID).
		WillReturnRows(sqlmock.NewRows([]string{"trace_id"}).AddRow("unclaimed-run"))
	mock.ExpectRollback()
	tx, err := s.getDB().BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	err = s.validateQAFixtureResource(
		t.Context(), tx, qaFixtureScope{},
		qaFixtureResource{Kind: "artifact", Ref: artifactID},
		map[string][]string{"team": {"fixture-team"}}, nil, nil,
	)
	_ = tx.Rollback()
	if !errors.Is(err, errQAFixtureResourceUnowned) {
		t.Fatalf("artifact validation error = %v, want unowned provenance", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
