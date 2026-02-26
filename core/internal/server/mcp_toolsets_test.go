package server

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/mcp"
)

// ── Test Helpers ──────────────────────────────────────────────

func withMCPToolSets(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.MCPToolSets = mcp.NewToolSetService(db)
	}, mock
}

func toolSetColumns() []string {
	return []string{"id", "name", "description", "tool_refs", "tenant_id", "created_at", "updated_at"}
}

// ── handleListToolSets ────────────────────────────────────────

func TestHandleListToolSets_HappyPath(t *testing.T) {
	opt, mock := withMCPToolSets(t)
	s := newTestServer(opt)
	now := time.Now()

	rows := sqlmock.NewRows(toolSetColumns()).
		AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "workspace", "File I/O", `["mcp:filesystem/*"]`, "default", now, now).
		AddRow("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "research", "Web tools", `["mcp:fetch/*"]`, "default", now, now)
	mock.ExpectQuery("SELECT .+ FROM mcp_tool_sets").WillReturnRows(rows)

	mux := setupMux(t, "GET /api/v1/mcp/toolsets", s.handleListToolSets)
	rr := doRequest(t, mux, "GET", "/api/v1/mcp/toolsets", "")

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data := resp["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("expected 2 tool sets, got %d", len(data))
	}
}

func TestHandleListToolSets_NilService(t *testing.T) {
	s := newTestServer() // no MCPToolSets wired
	mux := setupMux(t, "GET /api/v1/mcp/toolsets", s.handleListToolSets)
	rr := doRequest(t, mux, "GET", "/api/v1/mcp/toolsets", "")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ── handleCreateToolSet ───────────────────────────────────────

func TestHandleCreateToolSet_HappyPath(t *testing.T) {
	opt, mock := withMCPToolSets(t)
	s := newTestServer(opt)
	now := time.Now()

	mock.ExpectQuery("INSERT INTO mcp_tool_sets").
		WithArgs("development", "Dev tools", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("cccccccc-cccc-cccc-cccc-cccccccccccc", now, now))

	mux := setupMux(t, "POST /api/v1/mcp/toolsets", s.handleCreateToolSet)
	body := `{"name":"development","description":"Dev tools","tool_refs":["mcp:filesystem/*","mcp:github/*"]}`
	rr := doRequest(t, mux, "POST", "/api/v1/mcp/toolsets", body)

	assertStatus(t, rr, http.StatusCreated)
	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true")
	}
}

func TestHandleCreateToolSet_MissingName(t *testing.T) {
	opt, _ := withMCPToolSets(t)
	s := newTestServer(opt)

	mux := setupMux(t, "POST /api/v1/mcp/toolsets", s.handleCreateToolSet)
	rr := doRequest(t, mux, "POST", "/api/v1/mcp/toolsets", `{"description":"no name"}`)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCreateToolSet_BadJSON(t *testing.T) {
	opt, _ := withMCPToolSets(t)
	s := newTestServer(opt)

	mux := setupMux(t, "POST /api/v1/mcp/toolsets", s.handleCreateToolSet)
	rr := doRequest(t, mux, "POST", "/api/v1/mcp/toolsets", `{invalid}`)

	assertStatus(t, rr, http.StatusBadRequest)
}

// ── handleDeleteToolSet ───────────────────────────────────────

func TestHandleDeleteToolSet_HappyPath(t *testing.T) {
	opt, mock := withMCPToolSets(t)
	s := newTestServer(opt)

	mock.ExpectExec("DELETE FROM mcp_tool_sets WHERE id = \\$1").
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := setupMux(t, "DELETE /api/v1/mcp/toolsets/{id}", s.handleDeleteToolSet)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mcp/toolsets/ccc33333-cccc-cccc-cccc-cccccccccccc", "")

	assertStatus(t, rr, http.StatusOK)
}

func TestHandleDeleteToolSet_BadUUID(t *testing.T) {
	opt, _ := withMCPToolSets(t)
	s := newTestServer(opt)

	mux := setupMux(t, "DELETE /api/v1/mcp/toolsets/{id}", s.handleDeleteToolSet)
	rr := doRequest(t, mux, "DELETE", "/api/v1/mcp/toolsets/not-a-uuid", "")

	assertStatus(t, rr, http.StatusBadRequest)
}

// ── handleUpdateToolSet ───────────────────────────────────────

func TestHandleUpdateToolSet_HappyPath(t *testing.T) {
	opt, mock := withMCPToolSets(t)
	s := newTestServer(opt)
	now := time.Now()

	mock.ExpectQuery("UPDATE mcp_tool_sets SET").
		WithArgs("workspace", "Updated", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(toolSetColumns()).
			AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "workspace", "Updated", `["mcp:filesystem/*"]`, "default", now, now))

	mux := setupMux(t, "PUT /api/v1/mcp/toolsets/{id}", s.handleUpdateToolSet)
	body := `{"name":"workspace","description":"Updated","tool_refs":["mcp:filesystem/*"]}`
	rr := doRequest(t, mux, "PUT", "/api/v1/mcp/toolsets/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", body)

	assertStatus(t, rr, http.StatusOK)
}

func TestHandleUpdateToolSet_NotFound(t *testing.T) {
	opt, mock := withMCPToolSets(t)
	s := newTestServer(opt)

	mock.ExpectQuery("UPDATE mcp_tool_sets SET").
		WithArgs("workspace", "", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	mux := setupMux(t, "PUT /api/v1/mcp/toolsets/{id}", s.handleUpdateToolSet)
	rr := doRequest(t, mux, "PUT", "/api/v1/mcp/toolsets/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", `{"name":"workspace"}`)

	assertStatus(t, rr, http.StatusNotFound)
}

func TestHandleUpdateToolSet_BadUUID(t *testing.T) {
	opt, _ := withMCPToolSets(t)
	s := newTestServer(opt)

	mux := setupMux(t, "PUT /api/v1/mcp/toolsets/{id}", s.handleUpdateToolSet)
	rr := doRequest(t, mux, "PUT", "/api/v1/mcp/toolsets/not-a-uuid", `{"name":"workspace"}`)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleUpdateToolSet_MissingName(t *testing.T) {
	opt, _ := withMCPToolSets(t)
	s := newTestServer(opt)

	mux := setupMux(t, "PUT /api/v1/mcp/toolsets/{id}", s.handleUpdateToolSet)
	rr := doRequest(t, mux, "PUT", "/api/v1/mcp/toolsets/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", `{"description":"missing"}`)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCreateToolSet_NilService(t *testing.T) {
	s := newTestServer()

	mux := setupMux(t, "POST /api/v1/mcp/toolsets", s.handleCreateToolSet)
	rr := doRequest(t, mux, "POST", "/api/v1/mcp/toolsets", `{"name":"workspace"}`)

	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleUpdateToolSet_NilService(t *testing.T) {
	s := newTestServer()

	mux := setupMux(t, "PUT /api/v1/mcp/toolsets/{id}", s.handleUpdateToolSet)
	rr := doRequest(t, mux, "PUT", "/api/v1/mcp/toolsets/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", `{"name":"workspace"}`)

	assertStatus(t, rr, http.StatusServiceUnavailable)
}
