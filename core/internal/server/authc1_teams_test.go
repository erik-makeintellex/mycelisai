package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// AUTH-C1 A1: a raw manifest spawn (POST /api/swarm/teams, POST /api/v1/teams)
// needs root admin + groups:write, audits first, and a denial spawns nothing.
// The full-manifest list (GET /api/swarm/teams) needs root admin + groups:read.

const authc1TeamBody = `{"id":"authc1-team","name":"AuthC1 Team","type":"action"}`

var authc1SpawnPaths = []string{"/api/swarm/teams", "/api/v1/teams"}

// authc1TeamServer wires the real route table, an audit DB and a live Soma on
// an embedded NATS, so a spawn that is not refused really starts a team.
func authc1TeamServer(t *testing.T) (*AdminServer, sqlmock.Sqlmock, *http.ServeMux) {
	t.Helper()
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt, withNATS(t))
	s.Soma = swarm.NewSoma(s.NC, nil, nil, nil, nil, nil, nil)
	t.Cleanup(s.Soma.Shutdown)
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)
	return s, mock, mux
}

func authc1TeamIDs(s *AdminServer) []string {
	var ids []string
	for _, m := range s.Soma.ListTeams() {
		ids = append(ids, m.ID)
	}
	return ids
}

func authc1BlockerData(t *testing.T, body string) map[string]any {
	t.Helper()
	var env struct {
		OK   bool           `json:"ok"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &env); err != nil || env.OK {
		t.Fatalf("expected a normalized blocker envelope, got %s", body)
	}
	return env.Data
}

func TestAuthC1TeamSpawnRefusedWithoutAuthority(t *testing.T) {
	for _, path := range authc1SpawnPaths {
		for name, c := range map[string]struct {
			who    *RequestIdentity
			status int
			scope  string
		}{
			"anonymous":            {nil, http.StatusUnauthorized, ""},
			"signed-in non-admin":  {standardUserIdentity(), http.StatusForbidden, ""},
			"non-admin with scope": {authc1UserWithScopes("groups:write", "*"), http.StatusForbidden, ""},
			"admin groups:read":    {adminWithScopes("groups:read", scopeGovernanceWrite), http.StatusForbidden, "groups:write"},
		} {
			t.Run(path+"/"+name, func(t *testing.T) {
				s, mock, mux := authc1TeamServer(t)
				rr := doAuthenticatedRequestAs(t, mux, http.MethodPost, path, authc1TeamBody, c.who)
				assertStatus(t, rr, c.status)
				if c.status == http.StatusForbidden {
					data := authc1BlockerData(t, rr.Body.String())
					if data["code"] != codeAdminRequired {
						t.Fatalf("code = %v, want %s", data["code"], codeAdminRequired)
					}
					if c.scope != "" && data["required_scope"] != c.scope {
						t.Fatalf("required_scope = %v, want %s", data["required_scope"], c.scope)
					}
				}
				if ids := authc1TeamIDs(s); len(ids) != 0 {
					t.Fatalf("refused spawn started teams %v", ids)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("refusal touched the audit store: %v", err)
				}
			})
		}
	}
}

func TestAuthC1TeamSpawnAdminWithScopeSucceeds(t *testing.T) {
	for _, path := range authc1SpawnPaths {
		t.Run(path, func(t *testing.T) {
			s, mock, mux := authc1TeamServer(t)
			expectAudits(mock, 2) // requested, then the result
			rr := doAuthenticatedRequestAs(t, mux, http.MethodPost, path, authc1TeamBody, adminWithScopes("groups:write"))
			assertStatus(t, rr, http.StatusCreated)
			if ids := authc1TeamIDs(s); len(ids) != 1 || ids[0] != "authc1-team" {
				t.Fatalf("teams = %v, want authc1-team", ids)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("spawn was not audited before and after: %v", err)
			}
		})
	}
}

func TestAuthC1TeamSpawnAuditFailureSpawnsNothing(t *testing.T) {
	for _, path := range authc1SpawnPaths {
		t.Run(path, func(t *testing.T) {
			s, mock, mux := authc1TeamServer(t)
			mock.ExpectExec("INSERT INTO log_entries").WillReturnError(errors.New("audit down"))
			rr := doAuthenticatedRequestAs(t, mux, http.MethodPost, path, authc1TeamBody, adminWithScopes("groups:write"))
			assertStatus(t, rr, http.StatusServiceUnavailable)
			if data := authc1BlockerData(t, rr.Body.String()); data["code"] != codeServiceUnavailable {
				t.Fatalf("code = %v, want %s", data["code"], codeServiceUnavailable)
			}
			if ids := authc1TeamIDs(s); len(ids) != 0 {
				t.Fatalf("unaudited spawn started teams %v", ids)
			}
		})
	}
}

func TestAuthC1SwarmTeamListFollowsGroupsRead(t *testing.T) {
	s, _, mux := authc1TeamServer(t)
	s.Soma = swarm.NewTestSoma([]*swarm.TeamManifest{{ID: "listed-team", Name: "Listed", Members: []protocol.AgentManifest{{ID: "a1", SystemPrompt: "secret prompt"}}}})
	for name, c := range map[string]struct {
		who    *RequestIdentity
		status int
	}{
		"anonymous":            {nil, http.StatusUnauthorized},
		"signed-in non-admin":  {standardUserIdentity(), http.StatusForbidden},
		"admin groups:write":   {adminWithScopes("groups:write"), http.StatusForbidden},
		"admin groups:read":    {adminWithScopes("groups:read"), http.StatusOK},
		"admin wildcard scope": {adminWithScopes("*"), http.StatusOK},
	} {
		t.Run(name, func(t *testing.T) {
			rr := doAuthenticatedRequestAs(t, mux, http.MethodGet, "/api/swarm/teams", "", c.who)
			assertStatus(t, rr, c.status)
			leaked := strings.Contains(rr.Body.String(), "secret prompt")
			if leaked != (c.status == http.StatusOK) {
				t.Fatalf("manifest visibility wrong for %s: %s", name, rr.Body.String())
			}
		})
	}
}

func TestAuthC1TeamSpawnWithoutSomaIsBlocker(t *testing.T) {
	s := newTestServer()
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleTeams), http.MethodPost, "/api/v1/teams", authc1TeamBody, adminWithScopes("groups:write"))
	assertStatus(t, rr, http.StatusServiceUnavailable)
	if data := authc1BlockerData(t, rr.Body.String()); data["code"] != codeTeamServiceOffline {
		t.Fatalf("code = %v, want %s", data["code"], codeTeamServiceOffline)
	}
}

func authc1UserWithScopes(scopes ...string) *RequestIdentity {
	id := standardUserIdentity()
	id.Scopes = scopes
	return id
}
