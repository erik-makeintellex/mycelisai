package server

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/mcp"
)

// MCPA-H: curated MCP install/apply spawn a process inside Core, so they need
// root admin with mcp_config:write before anything else runs.

type mcpahRoute struct {
	name    string
	path    string
	handler func(s *AdminServer) http.HandlerFunc
}

func mcpahWriteRoutes() []mcpahRoute {
	return []mcpahRoute{
		{"install", "/api/v1/mcp/library/install", func(s *AdminServer) http.HandlerFunc { return s.handleMCPLibraryInstall }},
		{"apply", "/api/v1/mcp/library/apply", func(s *AdminServer) http.HandlerFunc { return s.handleMCPLibraryApply }},
	}
}

// mcpahLibrary holds stub entries whose transport never spawns a process.
func mcpahLibrary() func(*AdminServer) {
	return func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{{
			Name: "Default",
			Servers: []mcp.LibraryEntry{
				{Name: "fetch", Transport: "unsupported"},
				{Name: "filesystem", Transport: "unsupported"},
				{Name: "tokened", Transport: "unsupported", Env: map[string]string{"FETCH_TOKEN": ""},
					EnvironmentVariables: []mcp.LibraryEnvVar{{Name: "FETCH_REGION"}}},
			},
		}}}
	}
}

func mcpahAssertNoDBCalls(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected MCP service calls: %v", err)
	}
}

func TestMcpahLibraryWrite_StandardUserForbiddenBeforeAnySideEffect(t *testing.T) {
	bodies := []string{
		`{"name":"fetch"}`,
		`{"name":"filesystem","env":{"NODE_OPTIONS":"--require /tmp/mcpah-pwn.js"}}`,
		`{"name":"fetch","governance_context":{"actor_role":"owner","owner_user_id":"root"}}`,
		`not json`,
	}
	for _, route := range mcpahWriteRoutes() {
		for _, body := range bodies {
			opt, mock := withMCPDB(t)
			s := newTestServer(opt, mcpahLibrary())
			rr := doAuthenticatedRequestAs(t, route.handler(s), "POST", route.path, body, standardUserIdentity())
			assertStatus(t, rr, http.StatusForbidden)
			env := decodeBlocker(t, rr)
			if env.Data["code"] != codeAdminRequired {
				t.Fatalf("%s %s: code = %v, want admin_required", route.name, body, env.Data["code"])
			}
			assertUserSafe(t, "mcpah "+route.name, userVisible(env)...)
			mcpahAssertNoDBCalls(t, mock)
		}
	}
}

func TestMcpahLibraryWrite_AuthorityRunsBeforeSubsystemChecks(t *testing.T) {
	for _, route := range mcpahWriteRoutes() {
		s := newTestServer() // MCP, pool and library all nil
		rr := doAuthenticatedRequestAs(t, route.handler(s), "POST", route.path, `{"name":"fetch"}`, standardUserIdentity())
		assertStatus(t, rr, http.StatusForbidden)
		rr = doRequest(t, route.handler(s), "POST", route.path, `{"name":"fetch"}`)
		assertStatus(t, rr, http.StatusUnauthorized)
	}
}

func TestMcpahLibraryWrite_NoIdentityUnauthorized(t *testing.T) {
	for _, route := range mcpahWriteRoutes() {
		opt, mock := withMCPDB(t)
		s := newTestServer(opt, mcpahLibrary())
		rr := doRequest(t, route.handler(s), "POST", route.path, `{"name":"fetch"}`)
		assertStatus(t, rr, http.StatusUnauthorized)
		mcpahAssertNoDBCalls(t, mock)
	}
}

func TestMcpahLibraryWrite_AdminWithoutScopeNamesRequiredScope(t *testing.T) {
	for _, route := range mcpahWriteRoutes() {
		opt, mock := withMCPDB(t)
		s := newTestServer(opt, mcpahLibrary())
		for _, scope := range []string{"outputs:read", "mcp:write", "mcp:*"} {
			rr := doAuthenticatedRequestAs(t, route.handler(s), "POST", route.path, `{"name":"fetch"}`, adminWithScopes(scope))
			assertStatus(t, rr, http.StatusForbidden)
			env := decodeBlocker(t, rr)
			if env.Data["code"] != codeAdminRequired || env.Data["required_scope"] != scopeMCPConfigWrite {
				t.Fatalf("%s with %s: data = %v, want admin_required + required_scope %s", route.name, scope, env.Data, scopeMCPConfigWrite)
			}
		}
		mcpahAssertNoDBCalls(t, mock)
	}
}

func TestMcpahLibraryWrite_ScopedAndRootAdminsInstallWithoutEnv(t *testing.T) {
	admins := map[string]*RequestIdentity{
		"root":   localAdminIdentityForTest(),
		"scoped": adminWithScopes(scopeMCPConfigWrite),
	}
	url := mcpaFixtureURL(t)
	for label, identity := range admins {
		for _, route := range mcpahWriteRoutes() {
			for _, name := range []string{"fetch", "filesystem"} {
				opt, mock := mcpaSharedDB(t)
				s := newTestServer(opt, mcpaLiveLibrary(url))
				mcpaExpectLiveInstall(t, s, mock, name, sqlmock.AnyArg(), url, `{}`, `{}`, false)
				rr := doAuthenticatedRequestAs(t, route.handler(s), "POST", route.path, `{"name":"`+name+`"}`, identity)
				if rr.Code != http.StatusOK {
					t.Fatalf("%s %s %s: status = %d body %s", label, route.name, name, rr.Code, rr.Body.String())
				}
				mcpahAssertNoDBCalls(t, mock)
			}
		}
	}
}
