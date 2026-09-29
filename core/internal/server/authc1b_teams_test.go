package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/swarm"
)

// AUTH-C1b: DELETE /api/v1/teams/{id} needs root admin + groups:write, audits
// first, and refuses Core-owned teams for everyone. An OPTIONS preflight never
// reaches a handler.

func authc1bDeleteServer(t *testing.T, ids ...string) (*AdminServer, sqlmock.Sqlmock, *http.ServeMux) {
	t.Helper()
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	var manifests []*swarm.TeamManifest
	for _, id := range ids {
		manifests = append(manifests, &swarm.TeamManifest{ID: id, Name: id})
	}
	s.Soma = swarm.NewTestSoma(manifests)
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)
	return s, mock, mux
}

func authc1bDelete(t *testing.T, mux http.Handler, id string, who *RequestIdentity) *httptest.ResponseRecorder {
	t.Helper()
	return doAuthenticatedRequestAs(t, mux, http.MethodDelete, "/api/v1/teams/"+id, "", who)
}

func TestAuthC1bDeleteTeamRefusedWithoutAuthority(t *testing.T) {
	for name, c := range map[string]struct {
		who    *RequestIdentity
		status int
		scope  string
	}{
		"anonymous":            {nil, http.StatusUnauthorized, ""},
		"signed-in non-admin":  {standardUserIdentity(), http.StatusForbidden, ""},
		"non-admin with *":     {authc1UserWithScopes("*"), http.StatusForbidden, ""},
		"admin groups:read":    {adminWithScopes("groups:read", scopeGovernanceWrite), http.StatusForbidden, scopeRuntimeTeamSpawn},
		"admin role mis-cased": {&RequestIdentity{UserID: "x", Username: "x", Role: "Admin", Scopes: []string{"*"}}, http.StatusForbidden, ""},
	} {
		t.Run(name, func(t *testing.T) {
			s, mock, mux := authc1bDeleteServer(t, "user-team")
			rr := authc1bDelete(t, mux, "user-team", c.who)
			assertStatus(t, rr, c.status)
			if c.status == http.StatusForbidden {
				data := authc1BlockerData(t, rr.Body.String())
				if data["code"] != codeAdminRequired || (c.scope != "" && data["required_scope"] != c.scope) {
					t.Fatalf("blocker data = %v", data)
				}
			}
			if !s.Soma.HasTeam("user-team") {
				t.Fatal("a refused DELETE stopped the team")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("refusal touched the audit store: %v", err)
			}
		})
	}
}

func TestAuthC1bDeleteTeamAdminWithScopeAudited(t *testing.T) {
	s, mock, mux := authc1bDeleteServer(t, "user-team", "kept-team")
	expectAudits(mock, 2) // requested, then the result
	rr := authc1bDelete(t, mux, "user-team", adminWithScopes(scopeRuntimeTeamSpawn))
	assertStatus(t, rr, http.StatusOK)
	if s.Soma.HasTeam("user-team") || !s.Soma.HasTeam("kept-team") {
		t.Fatalf("teams after delete = %v", authc1TeamIDs(s))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("delete was not audited before and after: %v", err)
	}
}

func TestAuthC1bDeleteTeamAuditFailureDeletesNothing(t *testing.T) {
	s, mock, mux := authc1bDeleteServer(t, "user-team")
	mock.ExpectExec("INSERT INTO log_entries").WillReturnError(errors.New("audit down"))
	rr := authc1bDelete(t, mux, "user-team", adminWithScopes("*"))
	assertStatus(t, rr, http.StatusServiceUnavailable)
	if data := authc1BlockerData(t, rr.Body.String()); data["code"] != codeServiceUnavailable {
		t.Fatalf("code = %v", data["code"])
	}
	if !s.Soma.HasTeam("user-team") {
		t.Fatal("an unaudited DELETE stopped the team")
	}
}

func TestAuthC1bDeleteCoreOwnedTeamRefusedForEveryone(t *testing.T) {
	for _, id := range append(append([]string{}, swarm.CoreOwnedTeamIDs...), "Admin-Core") {
		for name, who := range map[string]*RequestIdentity{"admin *": adminWithScopes("*"), "admin groups:write": adminWithScopes(scopeRuntimeTeamSpawn)} {
			t.Run(id+"/"+name, func(t *testing.T) {
				s, mock, mux := authc1bDeleteServer(t, id)
				expectAudits(mock, 1) // the refused attempt is recorded
				rr := authc1bDelete(t, mux, id, who)
				assertStatus(t, rr, http.StatusForbidden)
				if data := authc1BlockerData(t, rr.Body.String()); data["code"] != codeCoreTeamProtected {
					t.Fatalf("code = %v, want %s", data["code"], codeCoreTeamProtected)
				}
				if !s.Soma.HasTeam(id) {
					t.Fatalf("Core-owned team %q was stopped", id)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("refused Core-team delete was not audited: %v", err)
				}
			})
		}
	}
}

func TestAuthC1bDeleteTeamNotFoundAndNoSoma(t *testing.T) {
	_, mock, mux := authc1bDeleteServer(t, "kept-team")
	expectAudits(mock, 2)
	assertStatus(t, authc1bDelete(t, mux, "missing-team", adminWithScopes("*")), http.StatusNotFound)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	s := newTestServer()
	noSoma := http.NewServeMux()
	s.RegisterRoutes(noSoma)
	rr := authc1bDelete(t, noSoma, "kept-team", adminWithScopes("*"))
	assertStatus(t, rr, http.StatusServiceUnavailable)
	if data := authc1BlockerData(t, rr.Body.String()); data["code"] != codeTeamServiceOffline {
		t.Fatalf("code = %v", data["code"])
	}
}

func TestAuthC1bPreflightNeverReachesHandlers(t *testing.T) {
	s, _, mux := authc1bDeleteServer(t, "c1b-secret-team")
	h := AuthMiddleware("authc1b-api-key-0123456789abcdef-0123456789", mux)
	for _, p := range []string{"/api/v1/teams", "/api/swarm/teams", "/api/v1/teams/c1b-secret-team", "/api/v1/user/settings", "/api/v1/trust/threshold"} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodOptions, p, nil))
		if rr.Code != http.StatusOK || strings.TrimSpace(rr.Body.String()) != "" {
			t.Errorf("OPTIONS %s -> %d %q, want an empty 200", p, rr.Code, rr.Body.String())
		}
	}
	if !s.Soma.HasTeam("c1b-secret-team") {
		t.Fatal("a preflight changed runtime state")
	}
	reached := false
	sentinel := AuthMiddleware("authc1b-api-key-0123456789abcdef-0123456789", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true }))
	sentinel.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodOptions, "/api/v1/anything", nil))
	if reached {
		t.Fatal("an unauthenticated OPTIONS reached the handler")
	}
}
