package server

import (
	"net/http"
	"testing"

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
