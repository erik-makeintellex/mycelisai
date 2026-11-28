package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/memory"
)

func TestAuthMiddleware_BlocksViewerOnMutatingRoutesAndAllowsReads(t *testing.T) {
	t.Setenv("MYCELIS_OPERATOR_API_KEY", "op-key-123")
	t.Setenv("MYCELIS_VIEWER_API_KEY", "view-key-456")

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := AuthMiddleware("owner-key-000", okHandler)

	mutatingCases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/memory/promote"},
		{http.MethodPost, "/api/v1/memory/resolve-conflict"},
		{http.MethodPost, "/api/v1/runs/run-1/interject"},
		{http.MethodPost, "/api/v1/groups"},
		{http.MethodPut, "/api/v1/governance/policy"},
		{http.MethodDelete, "/api/v1/missions/m-1"},
	}

	for _, tc := range mutatingCases {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
		req.Header.Set("Authorization", "Bearer view-key-456")
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("viewer %s %s: expected 403, got %d (%s)", tc.method, tc.path, rr.Code, rr.Body.String())
		}

		reqOp := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
		reqOp.Header.Set("Authorization", "Bearer op-key-123")
		rrOp := httptest.NewRecorder()
		wrapped.ServeHTTP(rrOp, reqOp)
		if rrOp.Code != http.StatusOK {
			t.Fatalf("operator %s %s: expected 200, got %d (%s)", tc.method, tc.path, rrOp.Code, rrOp.Body.String())
		}
	}

	// Viewer read-only routes remain allowed
	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/user/me"},
		{http.MethodGet, "/api/v1/memory/search?q=test"},
		{http.MethodPost, "/api/v1/search"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
		req.Header.Set("Authorization", "Bearer view-key-456")
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("viewer read %s %s: expected 200, got %d", tc.method, tc.path, rr.Code)
		}
	}
}

func TestHandleMemoryGovernance_Returns500OnDBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	s := newTestServer()
	s.Mem = memory.NewServiceWithDB(db)

	mock.ExpectQuery("UPDATE context_vectors").WillReturnError(errors.New("db write failure"))
	rrPromote := doRequest(t, http.HandlerFunc(s.HandleMemoryPromote), http.MethodPost, "/api/v1/memory/promote",
		`{"id":"vec-err","target_layer":"PROJECT_MEMORY"}`)
	assertStatus(t, rrPromote, http.StatusInternalServerError)

	mock.ExpectQuery("UPDATE context_vectors").WillReturnError(errors.New("db write failure"))
	rrResolve := doRequest(t, http.HandlerFunc(s.HandleMemoryResolveConflict), http.MethodPost, "/api/v1/memory/resolve-conflict",
		`{"id":"vec-err","resolution":"keep"}`)
	assertStatus(t, rrResolve, http.StatusInternalServerError)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
