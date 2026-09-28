package server

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/mycelis/core/internal/mcp"
)

// MCPA live-install helpers: library entries point at an in-test MCP server
// over loopback streamable HTTP, so a successful install really connects and
// discovers a tool. Nothing spawns a process.

// mcpaFixtureURL starts an in-test MCP server offering one tool ("noop").
func mcpaFixtureURL(t *testing.T) string {
	t.Helper()
	fixture := mcpserver.NewMCPServer("mcpa-fixture", "0")
	fixture.AddTool(mcplib.NewTool("noop"), func(context.Context, mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		return mcplib.NewToolResultText("ok"), nil
	})
	return newLocalHTTPTestServer(t, mcpserver.NewStreamableHTTPServer(fixture)).URL
}

// mcpaLiveLibrary holds fetch, filesystem, tokened and a credentialed
// external github entry, all served by the fixture at url.
func mcpaLiveLibrary(url string) func(*AdminServer) {
	return func(s *AdminServer) {
		entry := func(name string) mcp.LibraryEntry { return mcp.LibraryEntry{Name: name, Transport: "sse", URL: url} }
		tokened := entry("tokened")
		tokened.Env = map[string]string{"FETCH_TOKEN": ""}
		tokened.EnvironmentVariables = []mcp.LibraryEnvVar{{Name: "FETCH_REGION"}}
		github := entry("github")
		github.DeploymentBoundary = "external_saas"
		github.EnvironmentVariables = []mcp.LibraryEnvVar{{Name: "GITHUB_PERSONAL_ACCESS_TOKEN", Required: true, Secret: true}}
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{{
			Name: "Default", Servers: []mcp.LibraryEntry{entry("fetch"), entry("filesystem"), tokened, github},
		}}}
	}
}

// mcpaExpectLiveInstall scripts, in order: the replace lookup, the audit, the
// Install row (rowEnv/rowHeaders are what the DB returns), the pool's tool
// cache and connected status, then ListTools. It returns the audit capture.
func mcpaExpectLiveInstall(t *testing.T, s *AdminServer, mock sqlmock.Sqlmock, name string, envArg driver.Value, url, rowEnv, rowHeaders string, exists bool) *mcpsCapture {
	t.Helper()
	t.Cleanup(s.MCPPool.ShutdownAll)
	mcpaExpectLookup(mock, name, exists)
	audit := mcpaExpectAudit(mock, false)
	id := uuid.NewString()
	now := time.Now()
	mock.ExpectQuery("INSERT INTO mcp_servers").
		WithArgs(name, "sse", "", sqlmock.AnyArg(), envArg, url, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow(id, name, "sse", "", `[]`, rowEnv, url, rowHeaders, "installed", nil, now, now))
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM mcp_tools").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO mcp_tools").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec("UPDATE mcp_servers").WithArgs("connected", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").
		WillReturnRows(sqlmock.NewRows(mcpToolColumns()).AddRow(uuid.NewString(), id, "noop", "", []byte(`{}`)))
	return audit
}

// MCPS-QA item 2 (ported from TestZzqaMCPAH_OtherFieldsCannotReshapeInstall):
// a Connect failure is an honest 502 mcp_connect_failed with a connect-stage
// failure record, never status "installed"; extra body fields reshape nothing.
func TestMcpaInstall_ConnectFailureIsHonestBlocker(t *testing.T) {
	body := `{"name":"fetch","command":"sh","args":["-c","id"],"url":"http://evil","transport":"stdio",` +
		`"headers":{"X":"y"},"environment_variables":[{"name":"NODE_OPTIONS","default_value":"--require /x"}],"env":null}`
	for _, route := range mcpahWriteRoutes() {
		opt, mock := mcpaSharedDB(t)
		s := newTestServer(opt, mcpaLibrary())
		failed := mcpaExpectConnectFailedInstall(mock, "fetch", mcpahEnvKeys{})
		rr := doAuthenticatedRequest(t, route.handler(s), "POST", route.path, body)
		env := mcpsAssertBlocker(t, rr, http.StatusBadGateway, codeMCPConnectFailed)
		if strings.Contains(rr.Body.String(), `"installed"`) || env.Data["server_name"] != "fetch" || env.Data["server_id"] != mcpaServerID {
			t.Fatalf("%s: body = %s", route.name, rr.Body.String())
		}
		mcpaAssertMet(t, mock)
		if record := failed.object(t); record["action"] != "mcp_server_install_failed" || record["failed_stage"] != "connect" {
			t.Fatalf("%s: failure record = %s", route.name, failed.value())
		}
	}
}

// MCPS-QA item 1 (ported from TestZzqaMCPS_TransportErrorEcho): a transport
// or JSON-RPC error from the MCP server is redacted, capped and valid JSON.
func TestMcpaToolCall_TransportErrorIsRedactedValidJSON(t *testing.T) {
	longTail := strings.Repeat("x", 3*mcpErrorTextCap)
	fixture := mcpserver.NewMCPServer("mcpa-failing", "0")
	fixture.AddTool(mcplib.NewTool("write_file"), func(context.Context, mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		return nil, errors.New(`upstream "rejected" token=mcpa-err-secret-1 for https://u:mcpa-err-pw-2@db ` + longTail)
	})
	url := newLocalHTTPTestServer(t, mcpserver.NewStreamableHTTPServer(fixture)).URL
	opt, mock := mcpaSharedDB(t)
	s := newTestServer(opt)
	id := uuid.MustParse(mcpaServerID)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM mcp_tools").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO mcp_tools").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec("UPDATE mcp_servers").WillReturnResult(sqlmock.NewResult(0, 1))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := s.MCPPool.Connect(ctx, mcp.ServerConfig{ID: id, Name: "filesystem", Transport: "sse", URL: url}); err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(s.MCPPool.ShutdownAll)
	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM mcp_servers").WithArgs(id).WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
		AddRow(mcpaServerID, "filesystem", "sse", "", `[]`, `{}`, url, `{}`, "connected", nil, now, now))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").WithArgs(id).WillReturnRows(sqlmock.NewRows(mcpToolColumns()).
		AddRow(uuid.NewString(), mcpaServerID, "write_file", "", []byte(`{}`)))
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(0, 1))

	mux := setupMux(t, mcpsCallRoute, s.handleMCPToolCall)
	rr := doAuthenticatedRequestAs(t, mux, "POST", "/api/v1/mcp/servers/"+mcpaServerID+"/tools/write_file/call", `{"path":"a"}`, mcpsWebAdmin())
	assertStatus(t, rr, http.StatusBadGateway)
	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("502 body is not valid JSON: %v (%q)", err, rr.Body.String())
	}
	for _, secret := range []string{"mcpa-err-secret-1", "mcpa-err-pw-2"} {
		if strings.Contains(rr.Body.String(), secret) {
			t.Fatalf("502 body leaks %q: %s", secret, rr.Body.String())
		}
	}
	if resp.OK || !strings.Contains(resp.Error, "rejected") || len(resp.Error) > mcpErrorTextCap+64 {
		t.Fatalf("error = %d bytes %q, want redacted text capped near %d", len(resp.Error), resp.Error[:min(len(resp.Error), 120)], mcpErrorTextCap)
	}
	mcpaAssertMet(t, mock)
}
