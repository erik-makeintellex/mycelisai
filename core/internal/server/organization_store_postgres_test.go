package server

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

const testOrgID = "3f2a1c9e-6b0d-4f7e-9a51-2c8d7e4b1a00"

func newSQLMockOrganizationStore(t *testing.T) (*OrganizationStore, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return NewOrganizationStore(db), mock
}

// captureArg records the argument it matches, so a second store instance can
// be fed exactly the bytes the first one wrote.
type captureArg struct{ value []byte }

func (c *captureArg) Match(v driver.Value) bool {
	switch typed := v.(type) {
	case []byte:
		c.value = append([]byte(nil), typed...)
	case string:
		c.value = []byte(typed)
	default:
		return false
	}
	return true
}

func testOrganizationHome() OrganizationHomePayload {
	return OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{ID: testOrgID, Name: "Atlas", Purpose: "Persist me", TemplateID: "engineering-starter"},
		QAFixtureScopeID:    "scope-a",
	}
}

func TestOrganizationStorePostgres_RestartReadsBackIdenticalHomeAndScope(t *testing.T) {
	storeA, mockA := newSQLMockOrganizationStore(t)
	document := &captureArg{}
	mockA.ExpectExec("INSERT INTO organizations").
		WithArgs(testOrgID, organizationTenantID, "Atlas", "Persist me", "engineering-starter", "scope-a", document).
		WillReturnResult(sqlmock.NewResult(0, 1))
	created, err := storeA.Save(context.Background(), testOrganizationHome())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(document.value), "scope-a") {
		t.Fatalf("fixture scope leaked into document: %s", document.value)
	}

	storeB, mockB := newSQLMockOrganizationStore(t) // brand-new instance
	mockB.ExpectQuery("SELECT document, qa_fixture_scope_id FROM organizations").
		WithArgs(testOrgID, organizationTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"document", "qa_fixture_scope_id"}).AddRow(document.value, "scope-a"))
	mockB.ExpectQuery("SELECT document, qa_fixture_scope_id FROM organizations").
		WithArgs(testOrgID, organizationTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"document", "qa_fixture_scope_id"}).AddRow(document.value, "scope-a"))
	loaded, err := storeB.Get(context.Background(), testOrgID)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, _ := json.Marshal(created)
	gotJSON, _ := json.Marshal(loaded)
	if string(wantJSON) != string(gotJSON) || loaded.QAFixtureScopeID != "scope-a" {
		t.Fatalf("restart mismatch:\nwant %s\ngot  %s scope=%q", wantJSON, gotJSON, loaded.QAFixtureScopeID)
	}
	if scope, ok, err := storeB.QAFixtureScope(context.Background(), testOrgID); err != nil || !ok || scope != "scope-a" {
		t.Fatalf("QAFixtureScope after restart = %q %v %v", scope, ok, err)
	}
	for _, mock := range []sqlmock.Sqlmock{mockA, mockB} {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOrganizationStorePostgres_DuplicateInsertIsConflict(t *testing.T) {
	store, mock := newSQLMockOrganizationStore(t)
	mock.ExpectExec("INSERT INTO organizations").WillReturnError(&pgconn.PgError{Code: "23505"})
	if _, err := store.Save(context.Background(), testOrganizationHome()); !errors.Is(err, ErrOrganizationConflict) {
		t.Fatalf("duplicate insert err = %v, want conflict", err)
	}
}

func TestOrganizationStorePostgres_GetNotFoundAndOutage(t *testing.T) {
	store, mock := newSQLMockOrganizationStore(t)
	mock.ExpectQuery("SELECT document").WillReturnError(sql.ErrNoRows)
	if _, err := store.Get(context.Background(), testOrgID); !errors.Is(err, ErrOrganizationNotFound) {
		t.Fatalf("missing row err = %v, want not found", err)
	}
	if _, err := store.Get(context.Background(), "not-a-uuid"); !errors.Is(err, ErrOrganizationNotFound) {
		t.Fatalf("non-uuid id err = %v, want not found without a query", err)
	}
	mock.ExpectQuery("SELECT document").WillReturnError(errors.New("connection refused"))
	_, err := store.Get(context.Background(), testOrgID)
	if !errors.Is(err, ErrOrganizationStoreUnavailable) || errors.Is(err, ErrOrganizationNotFound) {
		t.Fatalf("outage err = %v, want storage unavailable", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestOrganizationStorePostgres_UpdateLocksRowAndPreservesIdentity(t *testing.T) {
	store, mock := newSQLMockOrganizationStore(t)
	current, _ := json.Marshal(testOrganizationHome())
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT document, qa_fixture_scope_id FROM organizations\s+WHERE id=\$1::uuid AND tenant_id=\$2\s+FOR UPDATE`).
		WithArgs(testOrgID, organizationTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"document", "qa_fixture_scope_id"}).AddRow(current, "scope-a"))
	mock.ExpectExec("UPDATE organizations").
		WithArgs(testOrgID, organizationTenantID, "Atlas", "Persist me", "engineering-starter", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	updated, err := store.Update(context.Background(), testOrgID, func(home OrganizationHomePayload) OrganizationHomePayload {
		home.ID = "forged-id"
		home.QAFixtureScopeID = "forged-scope"
		home.AIEngineProfileID = "high_reasoning"
		return home
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != testOrgID || updated.QAFixtureScopeID != "scope-a" || updated.AIEngineProfileID != "high_reasoning" {
		t.Fatalf("update did not preserve identity: %+v", updated)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestOrganizationStorePostgres_UpdateWriteFailureRollsBack(t *testing.T) {
	store, mock := newSQLMockOrganizationStore(t)
	current, _ := json.Marshal(testOrganizationHome())
	mock.ExpectBegin()
	mock.ExpectQuery("FOR UPDATE").
		WillReturnRows(sqlmock.NewRows([]string{"document", "qa_fixture_scope_id"}).AddRow(current, "scope-a"))
	mock.ExpectExec("UPDATE organizations").WillReturnError(errors.New("disk full"))
	mock.ExpectRollback()
	_, err := store.Update(context.Background(), testOrgID, func(home OrganizationHomePayload) OrganizationHomePayload { return home })
	if !errors.Is(err, ErrOrganizationStoreUnavailable) {
		t.Fatalf("write failure err = %v, want storage unavailable", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestOrganizationStorePostgres_ListIsTenantScopedAndOrdered(t *testing.T) {
	store, mock := newSQLMockOrganizationStore(t)
	docA, _ := json.Marshal(OrganizationHomePayload{OrganizationSummary: OrganizationSummary{ID: "a", Name: "Atlas"}})
	docB, _ := json.Marshal(OrganizationHomePayload{OrganizationSummary: OrganizationSummary{ID: "b", Name: "Beacon"}})
	mock.ExpectQuery(`WHERE tenant_id=\$1\s+ORDER BY name COLLATE "C", id`).WithArgs(organizationTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "document", "qa_fixture_scope_id"}).
			AddRow(testOrgID, docA, "").AddRow("6f1d0a7e-0000-4000-8000-000000000001", docB, "scope-b"))
	summaries, err := store.List(context.Background())
	if err != nil || len(summaries) != 2 || summaries[0].Name != "Atlas" || summaries[0].ID != testOrgID {
		t.Fatalf("list = %+v err=%v", summaries, err)
	}
}

func TestOrganizationHandlers_StorageFailureIs503Not404(t *testing.T) {
	store, mock := newSQLMockOrganizationStore(t)
	mock.ExpectQuery("SELECT document").WillReturnError(errors.New("connection reset"))
	s := newTestServer(func(s *AdminServer) { s.Organizations = store })
	mux := setupMux(t, "GET /api/v1/organizations/{id}/home", s.handleGetOrganizationHome)
	rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/organizations/"+testOrgID+"/home", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
	if !strings.Contains(rr.Body.String(), "Organization storage unavailable") {
		t.Fatalf("missing recovery hint: %s", rr.Body.String())
	}
}

func TestOrganizationHandlers_NoDatabaseIs503WithoutMemoryFallback(t *testing.T) {
	s := &AdminServer{TemplateBundlesPath: writeStarterBundle(t)} // production shape: no DB, no injected store
	list := doAuthenticatedRequest(t, http.HandlerFunc(s.handleListOrganizations), http.MethodGet, "/api/v1/organizations", "")
	assertStatus(t, list, http.StatusServiceUnavailable)
	create := doAuthenticatedRequest(t, http.HandlerFunc(s.handleCreateOrganization), http.MethodPost, "/api/v1/organizations",
		`{"name":"Atlas","purpose":"Persist me","start_mode":"empty"}`)
	assertStatus(t, create, http.StatusServiceUnavailable)
	if s.Organizations != nil {
		t.Fatal("organization store was cached without a database")
	}
}
