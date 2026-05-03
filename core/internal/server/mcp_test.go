package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestHandleMCPList_NilSubsystem(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPList), "GET", "/api/v1/mcp/servers", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

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

func TestHandleMCPToolsList_NilSubsystem(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPToolsList), "GET", "/api/v1/mcp/tools", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

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
