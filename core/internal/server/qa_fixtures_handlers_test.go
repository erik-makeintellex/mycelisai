package server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

var qaFixtureScopeColumns = []string{
	"id", "tenant_id", "owner_ref", "execution_ref", "status",
	"expires_at", "created_at", "updated_at",
}

func TestQAFixtureRoutesAreHiddenWhenDisabled(t *testing.T) {
	t.Setenv("MYCELIS_QA_FIXTURE_MANAGEMENT", "")
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/testing/fixture-scopes", s.HandleCreateQAFixtureScope)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/testing/fixture-scopes", `{}`)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected disabled route to return 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateQAFixtureScope(t *testing.T) {
	t.Setenv("MYCELIS_QA_FIXTURE_MANAGEMENT", "true")
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	now := time.Now().UTC()
	mock.ExpectQuery("INSERT INTO qa_fixture_scopes").
		WithArgs(sqlmock.AnyArg(), qaFixtureTenantID, "playwright", "journey-42", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(qaFixtureScopeColumns).AddRow(
			"11111111-1111-1111-1111-111111111111",
			qaFixtureTenantID,
			"playwright",
			"journey-42",
			"open",
			now.Add(time.Hour),
			now,
			now,
		))
	mux := setupMux(t, "POST /api/v1/testing/fixture-scopes", s.HandleCreateQAFixtureScope)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/testing/fixture-scopes", `{
		"owner_ref":"playwright",
		"execution_ref":"journey-42",
		"ttl_seconds":3600
	}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected fixture scope creation, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"owner_ref":"playwright"`) {
		t.Fatalf("response omitted fixture owner: %s", rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAddQAFixtureResourceRejectsOwnerMismatch(t *testing.T) {
	t.Setenv("MYCELIS_QA_FIXTURE_MANAGEMENT", "true")
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	scopeID := "11111111-1111-1111-1111-111111111111"
	now := time.Now().UTC()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text").
		WithArgs(scopeID, qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows(qaFixtureScopeColumns).AddRow(
			scopeID, qaFixtureTenantID, "another-owner", "journey-42", "open",
			now.Add(time.Hour), now, now,
		))
	mock.ExpectRollback()
	mux := setupMux(
		t,
		"POST /api/v1/testing/fixture-scopes/{id}/resources",
		s.HandleAddQAFixtureResources,
	)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost,
		"/api/v1/testing/fixture-scopes/"+scopeID+"/resources", `{
			"owner_ref":"playwright",
			"execution_ref":"journey-42",
			"resources":[{"kind":"team","ref":"qa-team-123"}]
		}`)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected ownership mismatch to return 403, got %d: %s", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAddQAFixtureResourceAcceptsExactOrganizationScope(t *testing.T) {
	t.Setenv("MYCELIS_QA_FIXTURE_MANAGEMENT", "true")
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	scopeID := "11111111-1111-1111-1111-111111111111"
	organizationID := "qa-organization"
	now := time.Now().UTC()
	s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{ID: organizationID, Name: "QA Organization"},
		QAFixtureScopeID:    scopeID,
	})
	if boundScope, ok := s.organizationStore().QAFixtureScope(organizationID); !ok || boundScope != scopeID {
		t.Fatalf("organization fixture scope = %q, %v", boundScope, ok)
	}
	if err := s.validateQAFixtureResource(
		t.Context(), nil, qaFixtureScope{ID: scopeID},
		qaFixtureResource{Kind: "organization", Ref: organizationID}, nil, nil, nil,
	); err != nil {
		t.Fatalf("direct organization scope validation failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text").WithArgs(scopeID, qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows(qaFixtureScopeColumns).AddRow(
			scopeID, qaFixtureTenantID, "playwright", "journey-42", "open",
			now.Add(time.Hour), now, now,
		))
	mock.ExpectQuery("SELECT resource_kind, resource_ref").WithArgs(scopeID).
		WillReturnRows(sqlmock.NewRows([]string{"resource_kind", "resource_ref"}))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("organization", organizationID, scopeID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("INSERT INTO qa_fixture_resources").
		WithArgs(sqlmock.AnyArg(), scopeID, "organization", organizationID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT resource_kind, resource_ref").WithArgs(scopeID).
		WillReturnRows(sqlmock.NewRows([]string{"resource_kind", "resource_ref"}).
			AddRow("organization", organizationID))
	mock.ExpectCommit()

	mux := setupMux(t, "POST /api/v1/testing/fixture-scopes/{id}/resources", s.HandleAddQAFixtureResources)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost,
		"/api/v1/testing/fixture-scopes/"+scopeID+"/resources", `{
			"owner_ref":"playwright",
			"execution_ref":"journey-42",
			"resources":[{"kind":"organization","ref":"qa-organization"}]
		}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected exact-scope organization registration, got %d: %s", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPurgeQAFixtureScopeDryRunDoesNotMutate(t *testing.T) {
	t.Setenv("MYCELIS_QA_FIXTURE_MANAGEMENT", "true")
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	scopeID := "11111111-1111-1111-1111-111111111111"
	now := time.Now().UTC()
	mock.ExpectExec("SELECT pg_advisory_lock").WithArgs(scopeID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id::text").
		WithArgs(scopeID, qaFixtureTenantID).
		WillReturnRows(sqlmock.NewRows(qaFixtureScopeColumns).AddRow(
			scopeID, qaFixtureTenantID, "playwright", "journey-42", "open",
			now.Add(time.Hour), now, now,
		))
	mock.ExpectQuery("SELECT resource_kind, resource_ref").
		WithArgs(scopeID).
		WillReturnRows(sqlmock.NewRows([]string{"resource_kind", "resource_ref"}).
			AddRow("team", "qa-team-123").
			AddRow("workspace_path", "groups/qa-team-123"))
	mock.ExpectExec("SELECT pg_advisory_unlock").WithArgs(scopeID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mux := setupMux(
		t,
		"POST /api/v1/testing/fixture-scopes/{id}/purge",
		s.HandlePurgeQAFixtureScope,
	)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost,
		"/api/v1/testing/fixture-scopes/"+scopeID+"/purge", `{
			"owner_ref":"playwright",
			"execution_ref":"journey-42",
			"confirm":false
		}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected dry-run plan, got %d: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, expected := range []string{`"confirmed":false`, `"registered_resources":2`, `"nats_untouched":true`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("dry-run response missing %s: %s", expected, body)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPurgeQAFixtureScopeResumesInterruptedPurge(t *testing.T) {
	t.Setenv("MYCELIS_QA_FIXTURE_MANAGEMENT", "true")
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	scopeID := "11111111-1111-1111-1111-111111111111"
	now := time.Now().UTC()
	scopeRows := func(status string) *sqlmock.Rows {
		return sqlmock.NewRows(qaFixtureScopeColumns).AddRow(
			scopeID, qaFixtureTenantID, "playwright", "journey-42", status,
			now.Add(time.Hour), now, now,
		)
	}
	resources := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{"resource_kind", "resource_ref"})
	}
	mock.ExpectExec("SELECT pg_advisory_lock").WithArgs(scopeID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id::text").WithArgs(scopeID, qaFixtureTenantID).WillReturnRows(scopeRows("purging"))
	mock.ExpectQuery("SELECT resource_kind, resource_ref").WithArgs(scopeID).WillReturnRows(resources())
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text").WithArgs(scopeID, qaFixtureTenantID).WillReturnRows(scopeRows("purging"))
	mock.ExpectQuery("SELECT resource_kind, resource_ref").WithArgs(scopeID).WillReturnRows(resources())
	mock.ExpectExec("UPDATE qa_fixture_scopes SET status='purging'").
		WithArgs(scopeID, qaFixtureTenantID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM qa_fixture_resources").WithArgs(scopeID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("UPDATE qa_fixture_scopes").WithArgs(scopeID, "purged", qaFixtureTenantID).
		WillReturnRows(scopeRows("purged"))
	mock.ExpectCommit()
	mock.ExpectExec("SELECT pg_advisory_unlock").WithArgs(scopeID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := setupMux(t, "POST /api/v1/testing/fixture-scopes/{id}/purge", s.HandlePurgeQAFixtureScope)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost,
		"/api/v1/testing/fixture-scopes/"+scopeID+"/purge", `{
			"owner_ref":"playwright",
			"execution_ref":"journey-42",
			"confirm":true
		}`)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"status":"purged"`) {
		t.Fatalf("expected interrupted purge recovery, got %d: %s", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPurgeQAFixtureScopeConfirmedClosesScopeAndRemovesOrganization(t *testing.T) {
	t.Setenv("MYCELIS_QA_FIXTURE_MANAGEMENT", "true")
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{ID: "qa-organization", Name: "QA Organization"},
	})
	scopeID := "11111111-1111-1111-1111-111111111111"
	now := time.Now().UTC()
	scopeRows := func(status string) *sqlmock.Rows {
		return sqlmock.NewRows(qaFixtureScopeColumns).AddRow(
			scopeID, qaFixtureTenantID, "playwright", "journey-42", status,
			now.Add(time.Hour), now, now,
		)
	}
	resourceRows := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{"resource_kind", "resource_ref"}).
			AddRow("organization", "qa-organization")
	}

	mock.ExpectExec("SELECT pg_advisory_lock").WithArgs(scopeID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id::text").WithArgs(scopeID, qaFixtureTenantID).WillReturnRows(scopeRows("open"))
	mock.ExpectQuery("SELECT resource_kind, resource_ref").WithArgs(scopeID).WillReturnRows(resourceRows())
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text").WithArgs(scopeID, qaFixtureTenantID).WillReturnRows(scopeRows("open"))
	mock.ExpectQuery("SELECT resource_kind, resource_ref").WithArgs(scopeID).WillReturnRows(resourceRows())
	mock.ExpectExec("UPDATE qa_fixture_scopes SET status='purging'").
		WithArgs(scopeID, qaFixtureTenantID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM qa_fixture_resources").WithArgs(scopeID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("UPDATE qa_fixture_scopes").WithArgs(scopeID, "purged", qaFixtureTenantID).
		WillReturnRows(scopeRows("purged"))
	mock.ExpectCommit()
	mock.ExpectExec("SELECT pg_advisory_unlock").WithArgs(scopeID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := setupMux(t, "POST /api/v1/testing/fixture-scopes/{id}/purge", s.HandlePurgeQAFixtureScope)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost,
		"/api/v1/testing/fixture-scopes/"+scopeID+"/purge", `{
			"owner_ref":"playwright",
			"execution_ref":"journey-42",
			"confirm":true
		}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected confirmed purge, got %d: %s", rr.Code, rr.Body.String())
	}
	if _, ok := s.organizationStore().Get("qa-organization"); ok {
		t.Fatal("confirmed purge left the owned organization in memory")
	}
	for _, expected := range []string{`"status":"purged"`, `"nats_untouched":true`, `"qa-organization"`} {
		if !strings.Contains(rr.Body.String(), expected) {
			t.Fatalf("confirmed purge response missing %s: %s", expected, rr.Body.String())
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
