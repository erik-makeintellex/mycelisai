package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/mcp"
)

// withMCPStubs wires non-nil MCP + MCPPool so handlers pass the nil guard.
// The stubs have nil DB underneath but validation tests fail before DB calls.
func withMCPStubs() func(*AdminServer) {
	return func(s *AdminServer) {
		s.MCP = mcp.NewService(nil)
		s.MCPPool = mcp.NewClientPool(s.MCP)
	}
}

func withMCPDB(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.MCP = mcp.NewService(db)
		s.MCPPool = mcp.NewClientPool(s.MCP)
	}, mock
}

func mcpServerColumns() []string {
	return []string{"id", "name", "transport", "command", "args", "env", "url", "headers", "status", "error_message", "created_at", "updated_at"}
}

func mcpToolColumns() []string {
	return []string{"id", "server_id", "name", "description", "input_schema"}
}

func mcpToolWithServerColumns() []string {
	return []string{"id", "server_id", "server_name", "name", "description", "input_schema"}
}

// ── POST /api/v1/mcp/install ───────────────────────────────────────

func TestHandleMCPInstall_NilSubsystem(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPInstall), "POST", "/api/v1/mcp/install", `{"name":"test"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPInstall_MissingName(t *testing.T) {
	s := newTestServer(withMCPStubs())
	rr := doRequest(t, http.HandlerFunc(s.handleMCPInstall), "POST", "/api/v1/mcp/install", `{"transport":"stdio","command":"echo"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleMCPInstall_MissingTransport(t *testing.T) {
	s := newTestServer(withMCPStubs())
	rr := doRequest(t, http.HandlerFunc(s.handleMCPInstall), "POST", "/api/v1/mcp/install", `{"name":"test"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleMCPInstall_StdioNoCommand(t *testing.T) {
	s := newTestServer(withMCPStubs())
	rr := doRequest(t, http.HandlerFunc(s.handleMCPInstall), "POST", "/api/v1/mcp/install", `{"name":"test","transport":"stdio"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleMCPInstall_SSENoURL(t *testing.T) {
	s := newTestServer(withMCPStubs())
	rr := doRequest(t, http.HandlerFunc(s.handleMCPInstall), "POST", "/api/v1/mcp/install", `{"name":"test","transport":"sse"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleMCPInstall_InvalidJSON(t *testing.T) {
	s := newTestServer(withMCPStubs())
	rr := doRequest(t, http.HandlerFunc(s.handleMCPInstall), "POST", "/api/v1/mcp/install", "not-json")
	assertStatus(t, rr, http.StatusBadRequest)
}

// ── GET /api/v1/mcp/servers ────────────────────────────────────────

func TestHandleMCPList_NilSubsystem(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPList), "GET", "/api/v1/mcp/servers", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── DELETE /api/v1/mcp/servers/{id} ────────────────────────────────

func TestHandleMCPDelete_NilSubsystem(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "DELETE /api/v1/mcp/servers/{id}", s.handleMCPDelete)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mcp/servers/some-id", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPDelete_InvalidUUID(t *testing.T) {
	s := newTestServer(withMCPStubs())
	mux := setupMux(t, "DELETE /api/v1/mcp/servers/{id}", s.handleMCPDelete)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mcp/servers/not-a-uuid", "")
	assertStatus(t, rr, http.StatusBadRequest)
}

// ── POST /api/v1/mcp/servers/{id}/tools/{tool}/call ────────────────

func TestHandleMCPToolCall_NilSubsystem(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/mcp/servers/{id}/tools/{tool}/call", s.handleMCPToolCall)
	rr := doRequest(t, mux, "POST", "/api/v1/mcp/servers/some-id/tools/some-tool/call", `{"arguments":{}}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPToolCall_InvalidUUID(t *testing.T) {
	s := newTestServer(withMCPStubs())
	mux := setupMux(t, "POST /api/v1/mcp/servers/{id}/tools/{tool}/call", s.handleMCPToolCall)
	rr := doRequest(t, mux, "POST", "/api/v1/mcp/servers/not-a-uuid/tools/test/call", `{"arguments":{}}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

// ── GET /api/v1/mcp/tools ──────────────────────────────────────────

func TestHandleMCPToolsList_NilSubsystem(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPToolsList), "GET", "/api/v1/mcp/tools", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── GET /api/v1/mcp/library ────────────────────────────────────────

func TestHandleMCPLibrary_NilLibrary(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibrary), "GET", "/api/v1/mcp/library", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPLibrary_WithLibrary(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{}}
	})
	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibrary), "GET", "/api/v1/mcp/library", "")
	assertStatus(t, rr, http.StatusOK)
}

// ── POST /api/v1/mcp/library/install ───────────────────────────────

func TestHandleMCPLibraryInstall_NilSubsystem(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"test"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPLibraryInstall_MissingName(t *testing.T) {
	s := newTestServer(withMCPStubs(), func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{}}
	})
	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"env":{}}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleMCPLibraryInstall_NotFoundInLibrary(t *testing.T) {
	s := newTestServer(withMCPStubs(), func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{Categories: []mcp.LibraryCategory{}}
	})
	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"nonexistent"}`)
	assertStatus(t, rr, http.StatusNotFound)
}

// ── Happy-path DB-backed tests ───────────────────────────────────────

func TestHandleMCPList_HappyPath(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt)
	now := time.Now()

	serverID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	serverUUID := uuid.MustParse(serverID)
	mock.ExpectQuery("SELECT .+ FROM mcp_servers").
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow(serverID, "filesystem", "stdio", "npx", `[]`, `{}`, "", `{}`, "connected", nil, now, now))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").
		WithArgs(serverUUID).
		WillReturnRows(sqlmock.NewRows(mcpToolColumns()).
			AddRow("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", serverID, "read_file", "Read file", []byte(`{}`)))

	rr := doRequest(t, http.HandlerFunc(s.handleMCPList), "GET", "/api/v1/mcp/servers", "")
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleMCPList_EmptyDB(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt)

	mock.ExpectQuery("SELECT .+ FROM mcp_servers").
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()))

	rr := doRequest(t, http.HandlerFunc(s.handleMCPList), "GET", "/api/v1/mcp/servers", "")
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleMCPDelete_HappyPath(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt)
	serverUUID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	mock.ExpectExec("DELETE FROM mcp_servers WHERE id = \\$1").
		WithArgs(serverUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := setupMux(t, "DELETE /api/v1/mcp/servers/{id}", s.handleMCPDelete)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mcp/servers/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "")
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleMCPToolsList_HappyPath(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt)

	mock.ExpectQuery("SELECT .+ FROM mcp_tools .+ JOIN mcp_servers").
		WillReturnRows(sqlmock.NewRows(mcpToolWithServerColumns()).
			AddRow("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "filesystem", "read_file", "Read file", []byte(`{}`)))

	rr := doRequest(t, http.HandlerFunc(s.handleMCPToolsList), "GET", "/api/v1/mcp/tools", "")
	assertStatus(t, rr, http.StatusOK)
}

func TestHandleMCPInstall_Forbidden(t *testing.T) {
	s := newTestServer()
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	rr := doRequest(t, mux, "POST", "/api/v1/mcp/install", `{"name":"test"}`)
	assertStatus(t, rr, http.StatusForbidden)
}

func TestHandleMCPLibraryInstall_HappyPath(t *testing.T) {
	opt, mock := withMCPDB(t)
	s := newTestServer(opt, func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{
			Categories: []mcp.LibraryCategory{
				{
					Name: "Default",
					Servers: []mcp.LibraryEntry{
						{Name: "fetch", Transport: "unsupported"},
					},
				},
			},
		}
	})
	now := time.Now()

	mock.ExpectQuery("INSERT INTO mcp_servers").
		WithArgs("fetch", "unsupported", "", sqlmock.AnyArg(), sqlmock.AnyArg(), "", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(mcpServerColumns()).
			AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "fetch", "unsupported", "", `[]`, `{}`, "", `{}`, "installed", nil, now, now))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").
		WithArgs("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa").
		WillReturnRows(sqlmock.NewRows(mcpToolColumns()))

	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"fetch"}`)
	assertStatus(t, rr, http.StatusOK)
}
