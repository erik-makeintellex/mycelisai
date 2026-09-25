package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func organizationStoreMux(s *AdminServer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations", s.handleCreateOrganization)
	mux.HandleFunc("GET /api/v1/organizations", s.handleListOrganizations)
	mux.HandleFunc("GET /api/v1/organizations/{id}/home", s.handleGetOrganizationHome)
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/ai-engine", s.handleUpdateOrganizationAIEngine)
	return mux
}

func TestNewOrganizationStoreForDB_SelectsBackend(t *testing.T) {
	if got := NewOrganizationStoreForDB(nil).Backend(); got != "memory" {
		t.Fatalf("nil DB should fall back to memory, got %s", got)
	}
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	if got := NewOrganizationStoreForDB(db).Backend(); got != "postgres" {
		t.Fatalf("DB should select postgres, got %s", got)
	}
}

func TestOrganizationHandlers_InMemoryFallbackRoundTrip(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	s.Organizations = NewOrganizationStoreForDB(nil)
	mux := organizationStoreMux(s)

	createRR := doRequest(t, mux, "POST", "/api/v1/organizations", `{"name":"Fallback Org","purpose":"Run without a database","start_mode":"empty"}`)
	assertStatus(t, createRR, http.StatusCreated)
	var created protocol.APIResponse
	assertJSON(t, createRR, &created)
	id, _ := created.Data.(map[string]any)["id"].(string)
	if id == "" {
		t.Fatalf("expected id, got %+v", created.Data)
	}

	updateRR := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+id+"/ai-engine", `{"profile_id":"high_reasoning"}`)
	assertStatus(t, updateRR, http.StatusOK)

	homeRR := doRequest(t, mux, "GET", "/api/v1/organizations/"+id+"/home", "")
	assertStatus(t, homeRR, http.StatusOK)
	var home protocol.APIResponse
	assertJSON(t, homeRR, &home)
	if home.Data.(map[string]any)["ai_engine_profile_id"] != "high_reasoning" {
		t.Fatalf("update did not persist in fallback store: %+v", home.Data)
	}

	listRR := doRequest(t, mux, "GET", "/api/v1/organizations", "")
	assertStatus(t, listRR, http.StatusOK)
	var list protocol.APIResponse
	assertJSON(t, listRR, &list)
	if items, _ := list.Data.([]any); len(items) != 1 {
		t.Fatalf("expected one listed organization, got %+v", list.Data)
	}

	missingRR := doRequest(t, mux, "GET", "/api/v1/organizations/missing/home", "")
	assertStatus(t, missingRR, http.StatusNotFound)
}

func TestOrganizationHandlers_PostgresStorageErrorsReturn500(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	s.Organizations = NewOrganizationStoreForDB(db)
	mux := organizationStoreMux(s)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO organizations")).WillReturnError(errors.New("relation \"organizations\" does not exist"))
	createRR := doRequest(t, mux, "POST", "/api/v1/organizations", `{"name":"Durable","purpose":"Fail loudly","start_mode":"empty"}`)
	assertStatus(t, createRR, http.StatusServiceUnavailable)

	mock.ExpectQuery(regexp.QuoteMeta(organizationGetSQL)).WillReturnError(errors.New("connection refused"))
	homeRR := doRequest(t, mux, "GET", "/api/v1/organizations/org-1/home", "")
	assertStatus(t, homeRR, http.StatusInternalServerError)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(organizationLockSQL)).WillReturnRows(sqlmock.NewRows([]string{"document"}))
	mock.ExpectRollback()
	updateRR := doRequest(t, mux, "PATCH", "/api/v1/organizations/org-1/ai-engine", `{"profile_id":"high_reasoning"}`)
	assertStatus(t, updateRR, http.StatusNotFound)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRespondOrganizationStoreError_MissingTableNamesMigration(t *testing.T) {
	rr := httptest.NewRecorder()
	respondOrganizationStoreError(rr, errors.New(`pq: relation "organizations" does not exist`))
	if rr.Code != http.StatusServiceUnavailable || !strings.Contains(rr.Body.String(), "compose.migrate") {
		t.Fatalf("missing-table response = %d %s", rr.Code, rr.Body.String())
	}
	rr = httptest.NewRecorder()
	respondOrganizationStoreError(rr, errors.New("connection reset"))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("generic storage error code = %d, want 500", rr.Code)
	}
}
