package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/mycelis/core/internal/mcp"
)

// MCPA (contract D1, D7-D9): server delete and toolset writes need
// mcp_config:write and a fail-closed audit; delete resolves the server before
// it disconnects anything.

const mcpaToolSetID = "cccccccc-cccc-cccc-cccc-cccccccccccc"

func withIdentityForMcpa(r *http.Request, identity *RequestIdentity) context.Context {
	return context.WithValue(r.Context(), ctxKeyIdentity, identity)
}

type mcpaRoute struct {
	name, method, pattern, path, body string
	handler                           func(s *AdminServer) http.HandlerFunc
}

// mcpaWriteRoutes is every MCP configuration write route.
func mcpaWriteRoutes() []mcpaRoute {
	return []mcpaRoute{
		{"install", "POST", "POST /api/v1/mcp/library/install", "/api/v1/mcp/library/install", `{"name":"fetch"}`, func(s *AdminServer) http.HandlerFunc { return s.handleMCPLibraryInstall }},
		{"apply", "POST", "POST /api/v1/mcp/library/apply", "/api/v1/mcp/library/apply", `{"name":"fetch"}`, func(s *AdminServer) http.HandlerFunc { return s.handleMCPLibraryApply }},
		{"delete", "DELETE", "DELETE /api/v1/mcp/servers/{id}", "/api/v1/mcp/servers/" + mcpaServerID, "", func(s *AdminServer) http.HandlerFunc { return s.handleMCPDelete }},
		{"toolset_create", "POST", "POST /api/v1/mcp/toolsets", "/api/v1/mcp/toolsets", `{"name":"workspace","tool_refs":["mcp:github/*"]}`, func(s *AdminServer) http.HandlerFunc { return s.handleCreateToolSet }},
		{"toolset_update", "PUT", "PUT /api/v1/mcp/toolsets/{id}", "/api/v1/mcp/toolsets/" + mcpaToolSetID, `{"name":"workspace","tool_refs":["mcp:github/*"]}`, func(s *AdminServer) http.HandlerFunc { return s.handleUpdateToolSet }},
		{"toolset_delete", "DELETE", "DELETE /api/v1/mcp/toolsets/{id}", "/api/v1/mcp/toolsets/" + mcpaToolSetID, "", func(s *AdminServer) http.HandlerFunc { return s.handleDeleteToolSet }},
	}
}

func (route mcpaRoute) do(t *testing.T, s *AdminServer, identity *RequestIdentity) *httptest.ResponseRecorder {
	t.Helper()
	mux := setupMux(t, route.pattern, route.handler(s))
	if identity == nil {
		return doRequest(t, mux, route.method, route.path, route.body)
	}
	return doAuthenticatedRequestAs(t, mux, route.method, route.path, route.body, identity)
}

// mcpaConnectFixture connects a live in-test MCP server under mcpaServerID so
// a Disconnect is observable as the pool's "stopped" status write.
func mcpaConnectFixture(t *testing.T, s *AdminServer, mock sqlmock.Sqlmock) {
	t.Helper()
	fixture := mcpserver.NewMCPServer("mcpa-fixture", "0")
	fixture.AddTool(mcplib.NewTool("noop"), func(context.Context, mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		return mcplib.NewToolResultText("ok"), nil
	})
	endpoint := newLocalHTTPTestServer(t, mcpserver.NewStreamableHTTPServer(fixture))
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM mcp_tools").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO mcp_tools").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec("UPDATE mcp_servers").WillReturnResult(sqlmock.NewResult(0, 1))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := s.MCPPool.Connect(ctx, mcp.ServerConfig{ID: uuid.MustParse(mcpaServerID), Name: "fetch", Transport: "sse", URL: endpoint.URL}); err != nil {
		t.Fatalf("connect fixture: %v", err)
	}
	t.Cleanup(s.MCPPool.ShutdownAll)
	mcpaAssertMet(t, mock)
}

func mcpaExpectDisconnect(mock sqlmock.Sqlmock) {
	mock.ExpectExec("UPDATE mcp_servers").WithArgs("stopped", sqlmock.AnyArg(), uuid.MustParse(mcpaServerID)).WillReturnResult(sqlmock.NewResult(0, 1))
}

// mcpaAssertStillConnected proves the handler did not disconnect the fixture.
func mcpaAssertStillConnected(t *testing.T, s *AdminServer, mock sqlmock.Sqlmock) {
	t.Helper()
	mcpaExpectDisconnect(mock)
	if err := s.MCPPool.Disconnect(uuid.MustParse(mcpaServerID)); err != nil {
		t.Fatalf("fixture was disconnected by the handler: %v", err)
	}
	mcpaAssertMet(t, mock)
}

func mcpaExpectServerGet(mock sqlmock.Sqlmock, found bool) {
	rows := sqlmock.NewRows(mcpServerColumns())
	if found {
		now := time.Now()
		rows.AddRow(mcpaServerID, "fetch", "sse", "", `[]`, `{}`, "http://fixture", `{}`, "connected", nil, now, now)
	}
	mock.ExpectQuery("SELECT .+ FROM mcp_servers").WithArgs(uuid.MustParse(mcpaServerID)).WillReturnRows(rows)
}

func mcpaExpectToolSetGet(mock sqlmock.Sqlmock, found bool) {
	rows := sqlmock.NewRows(toolSetColumns())
	if found {
		now := time.Now()
		rows.AddRow(mcpaToolSetID, "workspace", "", `["mcp:filesystem/*"]`, "all", "", "default", now, now)
	}
	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets").WithArgs(uuid.MustParse(mcpaToolSetID)).WillReturnRows(rows)
}

func TestMcpaWrites_NonAdminsRefusedBeforeAnySideEffect(t *testing.T) {
	callers := map[string]*RequestIdentity{"standard": standardUserIdentity(), "scopeless_admin": adminWithScopes("outputs:read", "mcp:write"), "none": nil}
	for label, identity := range callers {
		for _, route := range mcpaWriteRoutes() {
			opt, mock := mcpaSharedDB(t)
			s := newTestServer(opt, mcpaLibrary())
			rr := route.do(t, s, identity)
			if identity == nil {
				assertStatus(t, rr, http.StatusUnauthorized)
			} else {
				env := mcpsAssertBlocker(t, rr, http.StatusForbidden, codeAdminRequired)
				if label == "scopeless_admin" && env.Data["required_scope"] != scopeMCPConfigWrite {
					t.Fatalf("%s: required_scope = %v", route.name, env.Data["required_scope"])
				}
				if label == "standard" {
					assertUserSafe(t, "mcpa "+route.name, userVisible(env)...)
				}
			}
			mcpaAssertMet(t, mock)
		}
	}
}

func TestMcpaDelete_ResolvesAuditsThenDisconnectsAndDeletes(t *testing.T) {
	opt, mock := mcpaSharedDB(t)
	s := newTestServer(opt)
	mcpaConnectFixture(t, s, mock)
	mcpaExpectServerGet(mock, true)
	audit := mcpaExpectAudit(mock, false)
	mcpaExpectDisconnect(mock)
	mock.ExpectExec("DELETE FROM mcp_servers").WithArgs(uuid.MustParse(mcpaServerID)).WillReturnResult(sqlmock.NewResult(0, 1))
	rr := mcpaWriteRoutes()[2].do(t, s, localAdminIdentityForTest())
	assertStatus(t, rr, http.StatusOK)
	mcpaAssertMet(t, mock)
	if record := audit.object(t); record["action"] != "mcp_server_deleted" || record["server_name"] != "fetch" || record["server_id"] != mcpaServerID {
		t.Fatalf("audit = %v", record)
	}
}

func TestMcpaDelete_UnknownIDIs404WithoutDisconnect(t *testing.T) {
	opt, mock := mcpaSharedDB(t)
	s := newTestServer(opt)
	mcpaConnectFixture(t, s, mock)
	mcpaExpectServerGet(mock, false)
	rr := mcpaWriteRoutes()[2].do(t, s, localAdminIdentityForTest())
	mcpsAssertBlocker(t, rr, http.StatusNotFound, "mcp_server_not_found")
	mcpaAssertStillConnected(t, s, mock)
}

func TestMcpaDelete_ErrorIsFixedAndRecorded(t *testing.T) {
	opt, mock := mcpaSharedDB(t)
	s := newTestServer(opt)
	mcpaExpectServerGet(mock, true)
	mcpaExpectAudit(mock, false)
	mock.ExpectExec("DELETE FROM mcp_servers").WillReturnError(errors.New("pq: password=" + mcpaSecretValue))
	failed := mcpaExpectAudit(mock, false)
	rr := mcpaWriteRoutes()[2].do(t, s, localAdminIdentityForTest())
	assertStatus(t, rr, http.StatusInternalServerError)
	if strings.Contains(rr.Body.String(), mcpaSecretValue) || strings.Contains(rr.Body.String(), "pq:") {
		t.Fatalf("raw delete error reached the response: %s", rr.Body.String())
	}
	mcpaAssertMet(t, mock)
	if failed.object(t)["action"] != "mcp_server_delete_failed" {
		t.Fatalf("failure record = %s", failed.value())
	}
}

func TestMcpaWrites_AuditUnavailableIs503WithZeroSideEffects(t *testing.T) {
	for _, failInsert := range []bool{false, true} {
		for _, route := range mcpaWriteRoutes() {
			opt, mock := mcpaSharedDB(t)
			s := newTestServer(opt, mcpaLibrary())
			switch route.name {
			case "install", "apply":
				mcpaExpectLookup(mock, "fetch", false)
			case "delete":
				mcpaConnectFixture(t, s, mock)
				mcpaExpectServerGet(mock, true)
			case "toolset_update", "toolset_delete":
				mcpaExpectToolSetGet(mock, true)
			}
			if failInsert {
				mcpaExpectAudit(mock, true)
			} else {
				s.DB = nil // no audit store at all
			}
			rr := route.do(t, s, localAdminIdentityForTest())
			mcpsAssertBlocker(t, rr, http.StatusServiceUnavailable, codeServiceUnavailable)
			mcpaAssertMet(t, mock)
			if route.name == "delete" {
				mcpaAssertStillConnected(t, s, mock)
			}
		}
	}
}

func TestMcpaReads_StayOpenWithoutAuditStore(t *testing.T) {
	opt, mock := mcpaSharedDB(t)
	s := newTestServer(opt, mcpaLibrary())
	s.DB = nil
	mock.ExpectQuery("SELECT .+ FROM mcp_servers").WillReturnRows(sqlmock.NewRows(mcpServerColumns()))
	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets").WillReturnRows(sqlmock.NewRows(toolSetColumns()))
	std := standardUserIdentity()
	reads := []struct {
		method, pattern, path, body string
		handler                     http.HandlerFunc
	}{
		{"GET", "GET /api/v1/mcp/servers", "/api/v1/mcp/servers", "", s.handleMCPList},
		{"GET", "GET /api/v1/mcp/toolsets", "/api/v1/mcp/toolsets", "", s.handleListToolSets},
		{"GET", "GET /api/v1/mcp/library", "/api/v1/mcp/library", "", s.handleMCPLibrary},
		{"POST", "POST /api/v1/mcp/library/inspect", "/api/v1/mcp/library/inspect", `{"name":"github"}`, s.handleMCPLibraryInspect},
	}
	for _, read := range reads {
		rr := doAuthenticatedRequestAs(t, setupMux(t, read.pattern, read.handler), read.method, read.path, read.body, std)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: status = %d body %s", read.path, rr.Code, rr.Body.String())
		}
	}
	mcpaAssertMet(t, mock)
}

func TestMcpaToolSets_WritesAuditRefsBeforeAndAfter(t *testing.T) {
	now := time.Now()
	cases := []struct {
		route          mcpaRoute
		action         string
		before, after  []string
		expectMutation func(sqlmock.Sqlmock)
	}{
		{mcpaWriteRoutes()[3], "mcp_toolset_created", []string{}, []string{"mcp:github/*"}, func(m sqlmock.Sqlmock) {
			m.ExpectQuery("INSERT INTO mcp_tool_sets").WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(mcpaToolSetID, now, now))
		}},
		{mcpaWriteRoutes()[4], "mcp_toolset_updated", []string{"mcp:filesystem/*"}, []string{"mcp:github/*"}, func(m sqlmock.Sqlmock) {
			m.ExpectQuery("UPDATE mcp_tool_sets SET").WillReturnRows(sqlmock.NewRows(toolSetColumns()).
				AddRow(mcpaToolSetID, "workspace", "", `["mcp:github/*"]`, "all", "", "default", now, now))
		}},
		{mcpaWriteRoutes()[5], "mcp_toolset_deleted", []string{"mcp:filesystem/*"}, []string{}, func(m sqlmock.Sqlmock) {
			m.ExpectExec("DELETE FROM mcp_tool_sets").WillReturnResult(sqlmock.NewResult(0, 1))
		}},
	}
	for _, tc := range cases {
		opt, mock := mcpaSharedDB(t)
		s := newTestServer(opt)
		if tc.action != "mcp_toolset_created" {
			mcpaExpectToolSetGet(mock, true)
		}
		audit := mcpaExpectAudit(mock, false)
		tc.expectMutation(mock)
		rr := tc.route.do(t, s, mcpsWebAdmin())
		if rr.Code != http.StatusOK && rr.Code != http.StatusCreated {
			t.Fatalf("%s: status = %d body %s", tc.action, rr.Code, rr.Body.String())
		}
		mcpaAssertMet(t, mock)
		record := audit.object(t)
		if record["action"] != tc.action || record["toolset_name"] != "workspace" {
			t.Fatalf("%s: audit = %v", tc.action, record)
		}
		if got := mcpaStrings(record["tool_refs_before"]); !slices.Equal(got, tc.before) {
			t.Fatalf("%s: tool_refs_before = %v, want %v", tc.action, got, tc.before)
		}
		if got := mcpaStrings(record["tool_refs_after"]); !slices.Equal(got, tc.after) {
			t.Fatalf("%s: tool_refs_after = %v, want %v", tc.action, got, tc.after)
		}
	}
}

func TestMcpaToolSets_UnknownIDIs404WithoutAudit(t *testing.T) {
	for _, route := range mcpaWriteRoutes()[4:] {
		opt, mock := mcpaSharedDB(t)
		s := newTestServer(opt)
		mcpaExpectToolSetGet(mock, false)
		rr := route.do(t, s, localAdminIdentityForTest())
		assertStatus(t, rr, http.StatusNotFound)
		mcpaAssertMet(t, mock)
	}
}
