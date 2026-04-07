package server

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/mcp"
)

func loadStandardMCPLibrary(t *testing.T) *mcp.Library {
	t.Helper()
	lib, err := mcp.LoadLibrary(filepath.Join("..", "..", "config", "mcp-library.yaml"))
	if err != nil {
		t.Fatalf("LoadLibrary: %v", err)
	}
	return lib
}

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

func withExchangeDB(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.Exchange = exchange.NewService(db, nil, nil)
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

func exchangeItemColumns() []string {
	return []string{"id", "channel_id", "channel_name", "schema_id", "payload", "created_by", "addressed_to", "thread_id", "visibility", "sensitivity_class", "source_role", "source_team", "target_role", "target_team", "allowed_consumers", "capability_id", "trust_class", "review_required", "metadata", "summary", "created_at"}
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

func TestHandleMCPActivity_NilExchange(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.handleMCPActivity), "GET", "/api/v1/mcp/activity", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleMCPActivity_ReturnsPersistedMCPUsage(t *testing.T) {
	opt, mock := withExchangeDB(t)
	s := newTestServer(opt)
	now := time.Now()
	channelID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	itemID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	payload := []byte(`{"summary":"Read workspace brief successfully.","state":"completed","server_id":"srv-001","server_name":"filesystem","tool_name":"read_file","result_preview":"Read workspace brief successfully.","run_id":"run-1","source_team":"alpha","agent_id":"soma-admin","created_at":"2026-04-06T12:00:00Z"}`)
	metadata := []byte(`{"source_kind":"mcp","mcp":{"server_id":"srv-001","server_name":"filesystem","tool_name":"read_file","state":"completed","run_id":"run-1","source_team":"alpha","agent_id":"soma-admin"}}`)

	mock.ExpectQuery("SELECT i.id, i.channel_id, c.name, i.schema_id, i.payload, i.created_by").
		WithArgs("browser.research.results", nil, 10).
		WillReturnRows(sqlmock.NewRows(exchangeItemColumns()).
			AddRow(itemID.String(), channelID.String(), "browser.research.results", "ToolResult", payload, "mcp:filesystem", "", nil, "advanced", "team_scoped", "mcp", "alpha", "soma", "", []byte(`[]`), "browser_research", "bounded_external", true, metadata, "Read workspace brief successfully.", now))
	mock.ExpectQuery("SELECT i.id, i.channel_id, c.name, i.schema_id, i.payload, i.created_by").
		WithArgs("media.image.output", nil, 10).
		WillReturnRows(sqlmock.NewRows(exchangeItemColumns()))
	mock.ExpectQuery("SELECT i.id, i.channel_id, c.name, i.schema_id, i.payload, i.created_by").
		WithArgs("api.data.output", nil, 10).
		WillReturnRows(sqlmock.NewRows(exchangeItemColumns()))

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPActivity), "GET", "/api/v1/mcp/activity?limit=10", "")
	assertStatus(t, rr, http.StatusOK)

	var resp struct {
		OK   bool                     `json:"ok"`
		Data []map[string]interface{} `json:"data"`
	}
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatal("expected ok=true")
	}
	if len(resp.Data) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(resp.Data))
	}
	if resp.Data[0]["server_id"] != "srv-001" {
		t.Fatalf("server_id = %v, want srv-001", resp.Data[0]["server_id"])
	}
	if resp.Data[0]["tool_name"] != "read_file" {
		t.Fatalf("tool_name = %v, want read_file", resp.Data[0]["tool_name"])
	}
	if resp.Data[0]["state"] != "completed" {
		t.Fatalf("state = %v, want completed", resp.Data[0]["state"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
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

func TestHandleMCPLibraryInspect_LocalOwnedConfigIsAutoAllowed(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{
			Categories: []mcp.LibraryCategory{
				{
					Name: "Default",
					Servers: []mcp.LibraryEntry{
						{Name: "filesystem", Transport: "stdio", Command: "npx", Tags: []string{"local"}},
					},
				},
			},
		}
	})

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInspect), "POST", "/api/v1/mcp/library/inspect", `{"name":"filesystem"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["decision"] != "allow" {
		t.Fatalf("decision = %v, want allow", resp["decision"])
	}
	governance, ok := resp["governance"].(map[string]any)
	if !ok {
		t.Fatalf("expected governance object, got %T", resp["governance"])
	}
	if governance["approval_required"] != false {
		t.Fatalf("approval_required = %v, want false", governance["approval_required"])
	}
	if governance["approval_reason"] != "user_owned_mcp_config" {
		t.Fatalf("approval_reason = %v, want user_owned_mcp_config", governance["approval_reason"])
	}
}

func TestHandleMCPLibraryInspect_RemoteConfigRequiresApproval(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{
			Categories: []mcp.LibraryCategory{
				{
					Name: "Default",
					Servers: []mcp.LibraryEntry{
						{Name: "remote-knowledge", Transport: "sse", URL: "https://mcp.example.com/sse", Tags: []string{"remote"}},
					},
				},
			},
		}
	})

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInspect), "POST", "/api/v1/mcp/library/inspect", `{"name":"remote-knowledge"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["decision"] != "require_approval" {
		t.Fatalf("decision = %v, want require_approval", resp["decision"])
	}
}

func TestHandleMCPLibraryInspect_StandardLibraryFilesystemIsAutoAllowed(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.MCPLibrary = loadStandardMCPLibrary(t)
	})

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInspect), "POST", "/api/v1/mcp/library/inspect", `{"name":"filesystem"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["decision"] != "allow" {
		t.Fatalf("decision = %v, want allow", resp["decision"])
	}

	governance, ok := resp["governance"].(map[string]any)
	if !ok {
		t.Fatalf("expected governance object, got %T", resp["governance"])
	}
	if governance["approval_reason"] != "user_owned_mcp_config" {
		t.Fatalf("approval_reason = %v, want user_owned_mcp_config", governance["approval_reason"])
	}
}

func TestHandleMCPLibraryInspect_StandardLibraryGitHubStaysAutoAllowedForOwnedConfig(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.MCPLibrary = loadStandardMCPLibrary(t)
	})

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInspect), "POST", "/api/v1/mcp/library/inspect", `{"name":"github"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["decision"] != "allow" {
		t.Fatalf("decision = %v, want allow", resp["decision"])
	}
	if resp["network_locality"] != "local" {
		t.Fatalf("network_locality = %v, want local", resp["network_locality"])
	}
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

func TestHandleMCPLibraryInstall_RemoteConfigReturnsApprovalBoundary(t *testing.T) {
	s := newTestServer(withMCPStubs(), func(s *AdminServer) {
		s.MCPLibrary = &mcp.Library{
			Categories: []mcp.LibraryCategory{
				{
					Name: "Default",
					Servers: []mcp.LibraryEntry{
						{Name: "remote-knowledge", Transport: "sse", URL: "https://mcp.example.com/sse", Tags: []string{"remote"}},
					},
				},
			},
		}
	})

	rr := doAuthenticatedRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"remote-knowledge"}`)
	assertStatus(t, rr, http.StatusAccepted)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["requires_approval"] != true {
		t.Fatalf("requires_approval = %v, want true", resp["requires_approval"])
	}
}

func TestHandleMCPInstall_ForbiddenForStandardLibraryEntryToo(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.MCPLibrary = loadStandardMCPLibrary(t)
	})
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	rr := doRequest(t, mux, "POST", "/api/v1/mcp/install", `{"name":"filesystem","transport":"stdio","command":"npx"}`)
	assertStatus(t, rr, http.StatusForbidden)
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
	mock.ExpectExec("UPDATE mcp_servers").
		WithArgs("error", sqlmock.AnyArg(), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT .+ FROM mcp_tools").
		WithArgs("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa").
		WillReturnRows(sqlmock.NewRows(mcpToolColumns()))

	rr := doRequest(t, http.HandlerFunc(s.handleMCPLibraryInstall), "POST", "/api/v1/mcp/library/install", `{"name":"fetch"}`)
	assertStatus(t, rr, http.StatusOK)
}
