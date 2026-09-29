package server

import (
	"errors"
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/overseer"
)

// AUTH-C1 A6: the V7 provisioning draft route is gone from the route table.
func TestAuthC1ProvisionDraftRouteRemoved(t *testing.T) {
	s := newTestServer()
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)
	for _, method := range []string{http.MethodPost, http.MethodGet} {
		rr := doAuthenticatedRequestAs(t, mux, method, "/api/v1/provision/draft", `{"intent":"build a team"}`, adminWithScopes("*"))
		assertStatus(t, rr, http.StatusNotFound)
	}
}

// AUTH-C1 A4: PUT /api/v1/trust/threshold is root admin + governance:write,
// audited first; a refusal leaves the Overseer threshold unchanged.
func authc1TrustServer(t *testing.T) (*AdminServer, *overseer.Engine, *http.ServeMux, func() error) {
	t.Helper()
	dbOpt, mock := withDB(t)
	ov := overseer.NewEngine(nil)
	s := newTestServer(dbOpt, func(s *AdminServer) { s.Overseer = ov })
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)
	return s, ov, mux, mock.ExpectationsWereMet
}

func TestAuthC1TrustThresholdPutNeedsGovernanceWrite(t *testing.T) {
	for name, c := range map[string]struct {
		who    *RequestIdentity
		status int
	}{
		"anonymous":             {nil, http.StatusUnauthorized},
		"signed-in non-admin":   {standardUserIdentity(), http.StatusForbidden},
		"non-admin wildcard":    {authc1UserWithScopes("*"), http.StatusForbidden},
		"admin governance:read": {adminWithScopes(scopeGovernanceRead), http.StatusForbidden},
	} {
		t.Run(name, func(t *testing.T) {
			_, ov, mux, expect := authc1TrustServer(t)
			before := ov.GetAutoExecuteThreshold()
			rr := doAuthenticatedRequestAs(t, mux, http.MethodPut, "/api/v1/trust/threshold", `{"threshold":0.1}`, c.who)
			assertStatus(t, rr, c.status)
			if c.status == http.StatusForbidden {
				if data := authc1BlockerData(t, rr.Body.String()); data["code"] != codeAdminRequired {
					t.Fatalf("code = %v", data["code"])
				}
			}
			if ov.GetAutoExecuteThreshold() != before {
				t.Fatalf("refused PUT changed the threshold to %v", ov.GetAutoExecuteThreshold())
			}
			if err := expect(); err != nil {
				t.Fatalf("refusal touched the audit store: %v", err)
			}
		})
	}
}

func TestAuthC1TrustThresholdAdminAuditsAndUpdates(t *testing.T) {
	dbOpt, mock := withDB(t)
	ov := overseer.NewEngine(nil)
	s := newTestServer(dbOpt, func(s *AdminServer) { s.Overseer = ov })
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)
	expectAudits(mock, 2)
	rr := doAuthenticatedRequestAs(t, mux, http.MethodPut, "/api/v1/trust/threshold", `{"threshold":0.4}`, adminWithScopes(scopeGovernanceWrite))
	assertStatus(t, rr, http.StatusOK)
	if ov.GetAutoExecuteThreshold() != 0.4 {
		t.Fatalf("threshold = %v, want 0.4", ov.GetAutoExecuteThreshold())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("threshold change not audited before and after: %v", err)
	}
	// Reads stay available to any signed-in user.
	get := doAuthenticatedRequestAs(t, mux, http.MethodGet, "/api/v1/trust/threshold", "", standardUserIdentity())
	assertStatus(t, get, http.StatusOK)
}

func TestAuthC1TrustThresholdAuditFailureChangesNothing(t *testing.T) {
	dbOpt, mock := withDB(t)
	mock.ExpectExec("INSERT INTO log_entries").WillReturnError(errors.New("audit down"))
	for name, opt := range map[string]func(*AdminServer){"insert-error": dbOpt, "no-audit-db": func(*AdminServer) {}} {
		ov := overseer.NewEngine(nil)
		s := newTestServer(opt, func(s *AdminServer) { s.Overseer = ov })
		before := ov.GetAutoExecuteThreshold()
		rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleTrustThreshold), http.MethodPut, "/api/v1/trust/threshold", `{"threshold":0.2}`, adminWithScopes(scopeGovernanceWrite))
		assertStatus(t, rr, http.StatusServiceUnavailable)
		if data := authc1BlockerData(t, rr.Body.String()); data["code"] != codeServiceUnavailable {
			t.Fatalf("%s: code = %v", name, data["code"])
		}
		if ov.GetAutoExecuteThreshold() != before {
			t.Fatalf("%s: threshold changed without audit", name)
		}
	}
}
